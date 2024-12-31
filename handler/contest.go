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
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time" binding:"required"`
	EndTime     time.Time `json:"end_time" binding:"required"`

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

	contest := models.Contest{
		Name:          reqBody.Name,
		Description:   reqBody.Description,
		StartTime:     reqBody.StartTime.UTC(),
		EndTime:       reqBody.EndTime.UTC(),
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
}

func updateContestProblems(c *gin.Context) {

}
