package handler

import (
	"github.com/ankush-web-eng/contest-backend/config"
	"github.com/ankush-web-eng/contest-backend/helpers"
	"github.com/ankush-web-eng/contest-backend/models"
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(r *gin.Engine) {
	authRouter := r.Group("/auth")
	{
		authRouter.POST("/signin", signin)
		authRouter.POST("/signup", signup)
		authRouter.GET("/verify", verify)
	}
}

type SignInRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type SignUpRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

func signin(c *gin.Context) {
	var reqBody SignInRequest

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(500, gin.H{"error": "Request Body is invalid!!"})
		return
	}

	var db = config.GetDB()

	var user models.User

	if err := db.Where("email = ?", &reqBody.Email).First(&user).Error; err != nil {
		c.JSON(404, gin.H{"message": "User does not exist, please login!!"})
		return
	}

	if !helpers.CheckPasswordHash(reqBody.Password, user.Password) {
		c.JSON(401, gin.H{"message": "Password is incorrect!!"})
		return
	}

	sessionToken, _ := helpers.GenerateSessionToken()
	user.SessionToken = sessionToken
	db.Save(&user)

	c.SetCookie("session_token", sessionToken, 3600000, "/", "localhost", true, true)

	c.JSON(200, gin.H{"message": "Signin successful"})
}

func signup(c *gin.Context) {
	var reqBody SignUpRequest

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(500, gin.H{"message": "Request Body is invalid!!"})
		return
	}

	var db = config.GetDB()

	hashedPassword, err := helpers.HashPassword(reqBody.Password)
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to hash password"})
		return
	}

	user := models.User{
		FirstName: reqBody.FirstName,
		LastName:  reqBody.LastName,
		Email:     reqBody.Email,
		Password:  hashedPassword,
	}

	if err := db.Create(&user).Error; err != nil {
		c.JSON(500, gin.H{"message": "Failed to create user"})
		return
	}

	c.JSON(200, gin.H{"message": "Signup successful"})
}

func verify(c *gin.Context) {
	sessionToken, err := c.Cookie("session_token")
	if err != nil {
		c.JSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	var db = config.GetDB()
	var user models.User

	if err := db.Where("session_token = ?", sessionToken).First(&user).Error; err != nil {
		c.JSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	c.JSON(200, gin.H{
		"message": "Verification successful",
		"user":    user,
	})
}
