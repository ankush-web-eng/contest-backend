package handler

import (
	"io"
	"net/http"
	"os"

	"github.com/ankush-web-eng/contest-backend/config"
	"github.com/ankush-web-eng/contest-backend/helpers"
	"github.com/ankush-web-eng/contest-backend/models"
	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(r *gin.Engine) {
	userRouter := r.Group("/user")
	{
		userRouter.POST("/image-upload", updateProfilePicture)
		userRouter.POST("/update-details", UpdateUserDetails)
	}
}

func updateProfilePicture(c *gin.Context) {
	err := c.Request.ParseMultipartForm(10 << 20)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unable to parse form"})
		return
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unable to read file"})
		return
	}
	defer file.Close()

	email := c.Request.FormValue("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Email is required"})
		return
	}

	db := config.GetDB()

	var user models.User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	if err := os.MkdirAll("temp-images", os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Unable to create temp directory"})
		return
	}

	tempFile, err := os.CreateTemp("temp-images", "upload-*.png")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Unable to create temp file"})
		return
	}
	defer tempFile.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Unable to read file bytes"})
		return
	}

	if _, err := tempFile.Write(fileBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Unable to write to temp file"})
		return
	}

	url, err := config.UploadFileToCloudinary(tempFile.Name())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to upload to Cloudinary"})
		return
	}

	if url == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to upload to Cloudinary"})
		return
	}

	defer os.Remove(tempFile.Name())

	user.Image = url
	db.Save(&user)

	c.JSON(200, gin.H{
		"message": "Profile image uploaded successfully",
		"url":     url,
	})
}

func UpdateUserDetails(c *gin.Context) {
	var reqBody struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email" binding:"required,email"`
		Password  string `json:"password"`
		Phone     string `json:"phone"`
		Gender    string `json:"gender"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	db := config.GetDB()
	var user models.User

	if err := db.Where("email = ?", reqBody.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	user.FirstName = reqBody.FirstName
	user.LastName = reqBody.LastName
	user.Phone = reqBody.Phone
	user.Gender = reqBody.Gender
	if reqBody.Password != "" {
		user.Password, _ = helpers.HashPassword(reqBody.Password)
	}

	db.Save(&user)

	c.JSON(http.StatusOK, gin.H{"message": "User details updated successfully"})
}
