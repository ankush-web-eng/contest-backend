package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Email        string `gorm:"unique;not null"`
	PasswordHash string `gorm:"not null"`
	SessionToken string `gorm:"unique"`
	LastLogin    time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}
