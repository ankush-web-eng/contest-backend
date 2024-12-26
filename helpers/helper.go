package helpers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/ankush-web-eng/contest-backend/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func GenerateVerifyToken() (string, error) {
	bytes := make([]byte, 2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	code := int(bytes[0])<<8 + int(bytes[1])
	return fmt.Sprintf("%04d", code%10000), nil
}

func InvalidatePreviousSessions(db *gorm.DB, userID uint) error {
	return db.Model(&models.User{}).Where("id = ?", userID).Update("session_token", nil).Error
}
