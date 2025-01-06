package main

import (
	"time"

	"github.com/ankush-web-eng/contest-backend/config"
	"github.com/ankush-web-eng/contest-backend/handler"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	config.InitDB()
	// if err := config.DB.AutoMigrate(
	// 	&models.User{},
	// 	&models.Contest{},
	// 	&models.Problem{},
	// 	&models.TestCase{},
	// 	&models.Submission{},
	// 	&models.UserContest{},
	// 	&models.RatingChange{}); err != nil {
	// 	panic("Failed to migrate database: " + err.Error())
	// }
	// gin.SetMode(gin.ReleaseMode)

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	handler.RegisterAuthRoutes(r)
	handler.RegisterUserRoutes(r)
	handler.RegisterContestRoutes(r)
	handler.RegisterCodeRoutes(r)
	handler.RegisterLiveRoutes(r)
	if err := r.Run(":8080"); err != nil {
		panic("Failed to start server: " + err.Error())
	}
}
