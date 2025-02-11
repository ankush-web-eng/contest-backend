package handler

import (
	"net/http"
	"time"

	"github.com/ankush-web-eng/contest-backend/config"
	"github.com/ankush-web-eng/contest-backend/models"
	"github.com/ankush-web-eng/contest-backend/types"
	"github.com/gin-gonic/gin"
)

func RegisterContestRoutes(r *gin.Engine) {
	contestRoutes := r.Group("/contest")
	{
		contestRoutes.GET("/get", getContest)
		contestRoutes.GET("/get-all", getContests)
		contestRoutes.POST("/create", createContest)
		contestRoutes.POST("/update/problems", updateContestProblems)
		contestRoutes.GET("/get-one/:id", getSingleContest)
	}
}

func getSingleContest(c *gin.Context) {
	contestID := c.Param("id")

	var db = config.GetDB()
	var contest models.Contest

	if err := db.Preload("Problems.TestCases").First(&contest, contestID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Contest not found!!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"contest": contest})
}

func getContest(c *gin.Context) {
	sessionToken, err := c.Cookie("session_token")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Session Token is invalid, login and try again!!"})
		return
	}

	var db = config.GetDB()
	var user models.User

	if err := db.Where("session_token = ?", sessionToken).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized access!!"})
		return
	}

	var contests []models.Contest
	if err := db.Find(&contests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not fetch contests, please try again later!!"})
		return
	}

	var userContests []models.Contest
	var otherContests []models.Contest

	for _, contest := range contests {
		var userContest models.UserContest
		if err := db.Where("user_id = ? AND contest_id = ?", user.ID, contest.ID).First(&userContest).Error; err != nil {
			otherContests = append(otherContests, contest)
		} else {
			userContests = append(userContests, contest)
		}
	}

	c.JSON(http.StatusOK, gin.H{"user_contests": userContests, "other_contests": otherContests})
}

func getContests(c *gin.Context) {
	var db = config.GetDB()

	var contests []models.Contest
	if err := db.Find(&contests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not fetch contests, please try again later!!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_contests": contests})
}

func createContest(c *gin.Context) {
	var reqBody types.CreateContestRequest

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Request type is invalid, please fix the sent data and it's types!!"})
		return
	}

	sessionToken, err := c.Cookie("session_token")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Session Token is invalid, login and try again!!"})
		return
	}

	var db = config.GetDB()
	var user models.User

	if err := db.Where("session_token = ?", sessionToken).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized access!!"})
		return
	}

	if !user.IsAdmin {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized access. You are not admin!!"})
		return
	}

	startTime, err := time.Parse(time.RFC3339, reqBody.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid start time format"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, reqBody.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid end time format"})
		return
	}

	contest := models.Contest{
		Name:          reqBody.Name,
		Description:   reqBody.Description,
		StartTime:     startTime.UTC(),
		EndTime:       endTime.UTC(),
		IsPublic:      reqBody.IsPublic,
		MaxDuration:   reqBody.MaxDuration,
		CreatorID:     user.ID,
		Status:        reqBody.Status,
		RatingFloor:   reqBody.RatingFloor,
		RatingCeil:    reqBody.RatingCeil,
		IsRated:       reqBody.IsRated,
		RatingType:    reqBody.RatingType,
		RatingKFactor: reqBody.RatingKFactor,
	}

	if err := db.Create(&contest).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not create contest, please try again later!!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Contest created successfully!!", "contest": contest})
}

func updateContestProblems(c *gin.Context) {
	var reqBody types.UpdateContestRequest

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Request type is invalid, please fix the sent data and its types!!"})
		return
	}

	sessionToken, err := c.Cookie("session_token")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Session Token is invalid, login and try again!!"})
		return
	}

	var db = config.GetDB()
	var user models.User

	if err := db.Where("session_token = ?", sessionToken).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized!!"})
		return
	}

	if !user.IsAdmin {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Only admins can edit contests!!"})
		return
	}

	var contest models.Contest
	if err := db.Where("id = ?", reqBody.ContestId).First(&contest).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Contest not found!!"})
		return
	}

	for _, problem := range reqBody.Problems {
		var contestProblem models.Problem
		contestProblem.ContestID = problem.ContestID
		contestProblem.Title = problem.Title
		contestProblem.Description = problem.Description
		contestProblem.TimeLimit = problem.TimeLimit
		contestProblem.MemoryLimit = problem.MemoryLimit
		contestProblem.Difficulty = problem.Difficulty
		contestProblem.Score = problem.Score
		contestProblem.Rating = problem.Rating
		contestProblem.SampleInput = problem.SampleInput
		contestProblem.SampleOutput = problem.SampleOutput
		contestProblem.TestCasesCount = problem.TestCasesCount

		if err := db.Create(&contestProblem).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not create contest problem, please try again later!!"})
			return
		}

		for _, testCase := range problem.TestCases {
			var problemTestCase models.TestCase
			problemTestCase.ProblemID = contestProblem.ID
			problemTestCase.Input = testCase.Input
			problemTestCase.Output = testCase.Output
			problemTestCase.IsHidden = testCase.IsHidden

			if err := db.Create(&problemTestCase).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not create test case, please try again later!!"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Contest problems updated successfully!!"})
}
