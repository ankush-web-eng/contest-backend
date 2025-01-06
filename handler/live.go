package handler

import (
	"github.com/ankush-web-eng/contest-backend/config"
	"github.com/ankush-web-eng/contest-backend/models"
	"github.com/gin-gonic/gin"
)

func RegisterLiveRoutes(r *gin.Engine) {
	liveRouter := r.Group("/live")
	{
		liveRouter.GET("/get/:contestId/:problemId", getAllProblems)
	}
}

func getAllProblems(c *gin.Context) {
	contestId := c.Param("contestId")
	problemId := c.Param("problemId")

	sessionToken, err := c.Cookie("session_token")
	if err != nil {
		c.JSON(400, gin.H{"message": "Session Token is invalid, login and try again!!"})
		return
	}

	var db = config.GetDB()
	var user models.User

	if err := db.Where("session_token = ?", sessionToken).First(&user).Error; err != nil {
		c.SetCookie("session_token", "", -1, "/", "localhost", false, true)
		c.JSON(401, gin.H{"message": "Unauthorized access!!"})
		return
	}

	var problem models.Problem
	if err := db.Preload("TestCases").Preload("Submissions").Where("contest_id = ? AND id = ?", contestId, problemId).First(&problem).Error; err != nil {
		c.JSON(404, gin.H{"message": "Problem not found!!"})
		return
	}

	c.JSON(200, gin.H{"problem": problem})
}
