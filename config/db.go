package config

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	var err error
	DB, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}))

	if err != nil {
		log.Fatal("Failed to connect to database!!!", err)
	}

	sqlDB, err := DB.DB()

	if err != nil {
		log.Fatal("Failed to get database instance:", err)
	}

	sqlDB.SetMaxIdleConns(100)
	sqlDB.SetMaxOpenConns(10000)
	sqlDB.SetConnMaxLifetime(time.Hour)
}

func CloseDB() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			log.Println("Error getting database instance:", err)
			return
		}
		err = sqlDB.Close()
		if err != nil {
			log.Println("Error closing database connection:", err)
		}
	}
}
