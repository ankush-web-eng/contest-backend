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
		authRouter.POST("/signup", signup)
		authRouter.POST("/verify-email", verifyEmail)
		authRouter.POST("/signin", signin)
		authRouter.GET("/verify", verify)
		authRouter.GET("/signout", signout)
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

func signup(c *gin.Context) {
	var reqBody SignUpRequest

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(500, gin.H{"message": "Request Body is invalid!!"})
		return
	}

	var db = config.GetDB()

	var user models.User
	if err := db.Where("email = ?", reqBody.Email).First(&user).Error; err == nil {
		if user.IsVerified {
			c.JSON(400, gin.H{"message": "User already exists!!"})
			return
		}
	}

	hashedPassword, err := helpers.HashPassword(reqBody.Password)
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to hash password"})
		return
	}

	verifyToken, err := helpers.GenerateVerifyToken()
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to generate verify token"})
		return
	}

	sessionToken, err := helpers.GenerateSessionToken()
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to generate session token"})
		return
	}

	if err := helpers.SendEmail(helpers.EmailDetails{
		From:    "ankushsingh.dev@gmail.com",
		To:      reqBody.Email,
		Subject: "Verify your email",
		Body:    "Your verification token is " + verifyToken,
	}); err != nil {
		c.JSON(500, gin.H{"message": "Failed to send email"})
		return
	}

	newUser := models.User{
		FirstName:    reqBody.FirstName,
		LastName:     reqBody.LastName,
		Email:        reqBody.Email,
		Password:     hashedPassword,
		SessionToken: sessionToken,
		IsVerified:   false,
		VerifyToken:  verifyToken,
	}

	if err := db.Create(&newUser).Error; err != nil {
		c.JSON(500, gin.H{"message": "Failed to create user"})
		return
	}

	c.JSON(200, gin.H{"message": "Signup successful"})
}

func verifyEmail(c *gin.Context) {
	var reqBody struct {
		Email       string `json:"email" binding:"required"`
		VerifyToken string `json:"verify_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(500, gin.H{"message": "Request Body is invalid!!"})
		return
	}

	var db = config.GetDB()
	var user models.User

	if err := db.Where("email = ?", reqBody.Email).First(&user).Error; err != nil {
		c.JSON(404, gin.H{"message": "User does not exist!!"})
		return
	}

	if user.VerifyToken != reqBody.VerifyToken {
		c.JSON(401, gin.H{"message": "Invalid verification token!!"})
		return
	}

	user.IsVerified = true

	db.Save(&user)

	c.JSON(200, gin.H{"message": "Email verified successfully"})
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

	if !user.IsVerified {
		c.JSON(401, gin.H{"message": "User is not verified!! please verify your email"})
		return
	}

	sessionToken, _ := helpers.GenerateSessionToken()
	user.SessionToken = sessionToken
	db.Save(&user)

	c.SetCookie("session_token", sessionToken, 3600000, "/", "localhost", false, true)

	c.JSON(200, gin.H{
		"message": "Signin successful",
		"user":    user,
	})
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

func signout(c *gin.Context) {
	sessionToken, err := c.Cookie("session_token")
	if err != nil {
		c.JSON(401, gin.H{"message": "Unauthorized & No previous session exists!!"})
		return
	}

	var db = config.GetDB()
	var user models.User

	if err := db.Where("session_token = ?", sessionToken).First(&user).Error; err != nil {
		c.JSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	user.SessionToken = ""
	db.Save(&user)

	c.SetCookie("session_token", "", -1, "/", "localhost", false, true)
	c.JSON(200, gin.H{
		"message": "Signout successful",
	})
}
