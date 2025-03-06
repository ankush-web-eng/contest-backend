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
	VerifyToken  string
	IsVerified   bool
	IsAdmin      bool `gorm:"default:false"`
	LastLogin    time.Time

	CurrentRating     int `gorm:"default:1000;index"`
	MaxRating         int `gorm:"default:1500"`
	MinRating         int `gorm:"default:1500"`
	GlobalRank        int `gorm:"index"`
	TotalContests     int
	ContestsWon       int
	TotalSubmissions  int
	TotalSolved       int
	CurrentStreak     int
	MaxStreak         int
	LastProblemSolved time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Contests       []Contest `gorm:"many2many:user_contests;"`
	Submissions    []Submission
	ContestCreated []Contest `gorm:"foreignKey:CreatorID"`
	RatingHistory  []RatingChange
}

type Contest struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null"`
	Description string
	StartTime   time.Time `gorm:"not null;index"`
	EndTime     time.Time `gorm:"not null;index"`

	IsPublic    bool `gorm:"default:true"`
	MaxDuration int  // in minutes, go for 0 for no limit
	CreatorID   uint `gorm:"not null"`

	Status      string `gorm:"default:'pending'"` // pending, active, completed
	RatingFloor int
	RatingCeil  int

	IsRated       bool   `gorm:"default:true"`
	RatingType    string `gorm:"default:'standard'"` // standard, performance, random
	RatingKFactor int    `gorm:"default:32"`         // Rating change magnitude factor

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Problems []Problem `gorm:"constraint:OnDelete:CASCADE;"`
	Users    []User    `gorm:"many2many:user_contests;"`
	Creator  User      `gorm:"foreignKey:CreatorID"`
}

type RatingChange struct {
	ID          uint      `gorm:"primaryKey"`
	UserID      uint      `gorm:"not null;index"`
	ContestID   uint      `gorm:"not null;index"`
	OldRating   int       `gorm:"not null"`
	NewRating   int       `gorm:"not null"`
	Rank        int       `gorm:"not null"`
	Performance int       `gorm:"not null"` // Performance rating in the contest
	Volatility  float64   // Rating volatility/uncertainty
	ChangeTime  time.Time `gorm:"not null;index"`

	User    User    `gorm:"foreignKey:UserID"`
	Contest Contest `gorm:"foreignKey:ContestID"`
}

type UserContest struct {
	UserID    uint `gorm:"primaryKey"`
	ContestID uint `gorm:"primaryKey"`
	Score     float64
	Rank      int
	StartTime time.Time
	EndTime   time.Time
	Status    string // registered, started, finished

	InitialRating int
	RatingChange  int
	Performance   int
	ExpectedRank  float64
	Volatility    float64 // Rating volatility/uncertainty

	CreatedAt time.Time
	UpdatedAt time.Time

	User    User
	Contest Contest
}

type Problem struct {
	ID          uint   `gorm:"primaryKey"`
	ContestID   uint   `gorm:"not null;index"`
	Title       string `gorm:"not null"`
	Description string `gorm:"not null"`

	TimeLimit   int    `gorm:"not null"` // in milliseconds
	MemoryLimit int    `gorm:"not null"` // in MB
	Difficulty  string // easy, medium, hard
	Score       int    `gorm:"not null"`
	Rating      int    `gorm:"not null"` // Problem difficulty rating

	SampleInput    string
	SampleOutput   string
	TestCasesCount int `gorm:"not null"`

	TotalSubmissions      int
	SuccessfulSubmissions int

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Contest     Contest      `gorm:"foreignKey:ContestID"`
	Submissions []Submission `gorm:"constraint:OnDelete:CASCADE;"`
	TestCases   []TestCase   `gorm:"constraint:OnDelete:CASCADE;"`
}

type TestCase struct {
	ID          uint   `gorm:"primaryKey"`
	ProblemID   uint   `gorm:"not null;index"`
	Input       string `gorm:"not null"`
	Output      string `gorm:"not null"`
	IsHidden    bool   `gorm:"default:true"`
	TimeLimit   int    // in milliseconds
	MemoryLimit int    // in KB

	CreatedAt time.Time
	UpdatedAt time.Time

	Problem Problem
}

type Submission struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"not null;index"`
	ProblemID uint   `gorm:"not null;index"`
	Language  string `gorm:"not null"`
	Code      string `gorm:"not null"`
	Status    string `gorm:"not null"` // accepted, wrong_answer, time_limit_exceeded, etc.
	Score     float64
	Runtime   int // in milliseconds
	Memory    int // in KB

	SubmittedAt time.Time `gorm:"not null;index"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	User    User    `gorm:"foreignKey:UserID"`
	Problem Problem `gorm:"foreignKey:ProblemID"`
}
