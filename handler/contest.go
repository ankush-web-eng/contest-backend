package handler

import (
	"net/http"
	"time"

	"github.com/ankush-web-eng/contest-backend/config"
	"github.com/ankush-web-eng/contest-backend/models"
	"github.com/gin-gonic/gin"
)

func RegisterContestRoutes(r *gin.Engine) {
	contestRoutes := r.Group("/contest")
	{
		contestRoutes.POST("/create", createContest)
		contestRoutes.POST("/update/problems", updateContestProblems)
	}
}

type createContestRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	StartTime   string `json:"start_time" binding:"required"`
	EndTime     string `json:"end_time" binding:"required"`

	IsPublic    bool `json:"is_public" binding:"required"`
	MaxDuration int  `json:"max_duration"`

	CreatorID uint `json:"creator_id" binding:"required"`

	Status      string `json:"status" binding:"required"`
	RatingFloor int    `json:"rating_floor"`
	RatingCeil  int    `json:"rating_ceil"`

	IsRated       bool   `json:"is_rated" binding:"required"`
	RatingType    string `json:"rating_type" binding:"required"`
	RatingKFactor int    `json:"rating_k_factor" binding:"required"`
}

func createContest(c *gin.Context) {
	var reqBody createContestRequest

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

type updateContestRequest struct {
	Problems []struct {
		ContestID   uint   `json:"contest_id"`
		Title       string `json:"title"`
		Description string `json:"description"`

		TimeLimit   int    `json:"time_limit"`
		MemoryLimit int    `json:"memory_limit"`
		Difficulty  string `json:"difficulty"`
		Score       int    `json:"score"`
		Rating      int    `json:"rating"`

		SampleInput    string `json:"sample_input"`
		SampleOutput   string `json:"sample_output"`
		TestCasesCount int    `json:"test_cases_count"`

		TestCases []struct {
			ProblemID uint   `json:"problem_id"`
			Input     string `json:"input"`
			Output    string `json:"output"`
			IsHidden  bool   `json:"is_hidden"`
		} `json:"test_cases"`
	} `json:"problems"`
	ContestId uint `json:"contest_id"`
}

func updateContestProblems(c *gin.Context) {
	var reqBody updateContestRequest

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
