package main

import (
	"time"

	"github.com/ankush-web-eng/contest-backend/handler"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	// gin.SetMode(gin.ReleaseMode)

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "http://localhost:3000"
		},
		MaxAge: 12 * time.Hour,
	}))

	handler.RegisterRoutes(r)
	r.Run(":8080")
}
