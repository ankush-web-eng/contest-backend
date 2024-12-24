package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint   `gorm:"primaryKey"`
	FirstName string `gorm:"not null"`
	LastName  string `gorm:"not null"`
	Email     string `gorm:"unique;not null"`
	Password  string `gorm:"not null"`
	Image     string
	Phone     string
	Gender    string

	SessionToken string `gorm:"unique"`
	LastLogin    time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`

	Contests []Contest `gorm:"many2many:user_contests;"`
}

type Contest struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"not null"`
	StartTime time.Time `gorm:"not null"`
	EndTime   time.Time `gorm:"not null"`
	Users     []User    `gorm:"many2many:user_contests;"`
}

type Problem struct {
	ID        uint    `gorm:"primaryKey"`
	ContestID uint    `gorm:"not null"`
	Contest   Contest `gorm:"foreignKey:ContestID"`
}
