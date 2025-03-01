package handler

import (
	"github.com/ankush-web-eng/contest-backend/config"
	"github.com/ankush-web-eng/contest-backend/helpers"
	"github.com/ankush-web-eng/contest-backend/models"
	"github.com/ankush-web-eng/contest-backend/types"
	"github.com/ankush-web-eng/contest-backend/utils"
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
		authRouter.POST("/reset/send-email", sendResetMail)
		authRouter.POST("/reset/verify-otp", verifyOTP)
		authRouter.POST("/reset/change-password", changePassword)
	}
}

func signup(c *gin.Context) {
	var reqBody types.SignUpRequest

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
		user.FirstName = reqBody.FirstName
		user.LastName = reqBody.LastName
		user.Password, _ = helpers.HashPassword(reqBody.Password)
		user.VerifyToken, _ = helpers.GenerateVerifyToken()
		user.SessionToken, _ = helpers.GenerateSessionToken()
	} else {
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

		user = models.User{
			FirstName:    reqBody.FirstName,
			LastName:     reqBody.LastName,
			Email:        reqBody.Email,
			Password:     hashedPassword,
			SessionToken: sessionToken,
			IsVerified:   false,
			VerifyToken:  verifyToken,
		}
	}

	if err := utils.SendEmail(utils.EmailDetails{
		From:    "ankushsingh.dev@gmail.com",
		To:      reqBody.Email,
		Subject: "Verify your email",
		Body:    "Your verification token is " + user.VerifyToken,
	}); err != nil {
		c.JSON(500, gin.H{"message": "Failed to send email"})
		return
	}

	if err := db.Save(&user).Error; err != nil {
		c.JSON(500, gin.H{"message": "Failed to save user"})
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
	var reqBody types.SignInRequest

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(500, gin.H{"error": "Request Body is invalid!!"})
		return
	}

	var db = config.GetDB()

	var user models.User

	if err := db.Where("email = ?", &reqBody.Email).First(&user).Error; err != nil {
		c.JSON(404, gin.H{"message": "User does not exist, please signup!!"})
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

func sendResetMail(c *gin.Context) {
	var reqBody struct {
		Email string `json:"email" binding:"required"`
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

	otp, err := helpers.GenerateVerifyToken()
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to generate OTP"})
		return
	}

	if err := utils.SendEmail(utils.EmailDetails{
		From:    "ankushsingh.dev@gmail.com",
		To:      reqBody.Email,
		Subject: "Reset your password",
		Body:    "Your OTP is " + otp,
	}); err != nil {

		c.JSON(500, gin.H{"message": "Failed to send email"})
		return
	}

	user.VerifyToken = otp
	db.Save(&user)

	c.JSON(200, gin.H{"message": "OTP sent successfully"})
}

func verifyOTP(c *gin.Context) {
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
		c.JSON(401, gin.H{"message": "Invalid OTP!!"})
		return
	}

	c.JSON(200, gin.H{"message": "OTP verified successfully"})
}

func changePassword(c *gin.Context) {
	var reqBody struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
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

	hashedPassword, err := helpers.HashPassword(reqBody.Password)
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to hash password"})
		return
	}

	user.Password = hashedPassword
	db.Save(&user)

	c.JSON(200, gin.H{"message": "Password changed successfully"})
}
