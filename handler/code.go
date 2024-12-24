package handler

import (
	"github.com/ankush-web-eng/contest-backend/types"
	"github.com/gin-gonic/gin"
)

func RegisterCodeRoutes(r *gin.Engine) {
	codeRouter := r.Group("/code")
	{
		codeRouter.POST("/submit", submitCode)
	}
}

func submitCode(c *gin.Context) {
	var reqType types.Request
	if err := c.ShouldBindJSON(&reqType); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": "Code submitted successfully",
		"code":    reqType.Code,
		"userId":  reqType.UserId,
		"lang":    reqType.Language,
	})
}
