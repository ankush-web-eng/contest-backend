package handler

import (
	"strings"

	"github.com/ankush-web-eng/contest-backend/config"
	"github.com/ankush-web-eng/contest-backend/models"
	"github.com/ankush-web-eng/contest-backend/types"
	"github.com/ankush-web-eng/contest-backend/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterCodeRoutes(r *gin.Engine) {
	codeRouter := r.Group("/code")
	{
		codeRouter.POST("/submit", submitCode)
	}
}

func submitCode(c *gin.Context) {
	var req types.SubmitCodeRequest

	SessionToken, err := c.Cookie("session_token")
	if err != nil {
		c.JSON(400, gin.H{"message": "Session token not found"})
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}

	db := config.GetDB()
	var user models.User

	if err := db.Where("session_token = ?", SessionToken).First(&user).Error; err != nil {
		c.JSON(404, gin.H{"message": "User not found"})
		return
	}

	languageId, err := utils.GetLanguageId(req.Language)
	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}

	var problem models.Problem
	if err := db.Where("id = ?", req.ProblemID).First(&problem).Error; err != nil {
		c.JSON(404, gin.H{"message": "Problem not found"})
		return
	}

	var testCases []models.TestCase
	if err := db.Where("problem_id = ?", req.ProblemID).Find(&testCases).Error; err != nil {
		c.JSON(500, gin.H{"message": "Error fetching test cases"})
		return
	}

	var cases []models.TestCase
	results := []map[string]interface{}{}

	cases = testCases

	if len(cases) == 0 {
		c.JSON(400, gin.H{"message": "No test cases found for this problem"})
		return
	}
	abortDueToTLE := false

	for _, testCase := range cases {
		if abortDueToTLE {
			break
		}

		result, err := utils.SubmitCodeToJudge(req.Code, languageId, models.TestCase{
			Input:       testCase.Input,
			Output:      testCase.Output,
			TimeLimit:   problem.TimeLimit,
			MemoryLimit: problem.MemoryLimit,
			ID:          testCase.ID,
		})

		if err != nil {
			c.JSON(500, gin.H{"message": "Judge error: " + err.Error()})
			return
		}

		if status, ok := result["status"].(map[string]interface{}); ok {
			if description, ok := status["description"].(string); ok && description == "Time Limit Exceeded" {
				abortDueToTLE = true
			}
		}

		stdout := result["stdout"].(string)
		isCorrect := strings.TrimSpace(stdout) == strings.TrimSpace(testCase.Output)
		status := "Accepted"
		if !isCorrect {
			status = "Wrong Answer"
		}

		results = append(results, map[string]interface{}{
			"status":     status,
			"testCaseId": testCase.ID,
			"stdout":     stdout,
		})
	}

	allPassed := true
	for _, result := range results {
		if result["status"] != "Accepted" {
			allPassed = false
			break
		}
	}

	overallStatus := "Accepted"
	if !allPassed {
		overallStatus = "Failed"
	}

	if problem.TestCasesCount > 0 {
		tx := db.Begin() // db transaction

		submission := models.Submission{
			Status:    overallStatus,
			Language:  req.Language,
			ProblemID: req.ProblemID,
			UserID:    user.ID,
		}
		if err := tx.Create(&submission).Error; err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"message": "Error creating submission"})
			return
		}

		if err := tx.Model(&problem).UpdateColumn("attempt_count", gorm.Expr("attempt_count + ?", 1)).Error; err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"message": "Error updating problem statistics"})
			return
		}

		if allPassed {
			if err := tx.Model(&problem).UpdateColumn("success_count", gorm.Expr("success_count + ?", 1)).Error; err != nil {
				tx.Rollback()
				c.JSON(500, gin.H{"message": "Error updating problem statistics"})
				return
			}
		}

		if err := tx.Commit().Error; err != nil {
			c.JSON(500, gin.H{"message": "Error committing transaction"})
			return
		}
	}

	response := gin.H{
		"status":  overallStatus,
		"results": results,
	}

	c.JSON(200, response)
}
