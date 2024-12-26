// package handler

// import (
// 	"context"
// 	"crypto/rand"
// 	"crypto/subtle"
// 	"encoding/base64"
// 	"errors"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"sync"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"golang.org/x/crypto/bcrypt"
// 	"gorm.io/gorm"
// )

// const (
// 	maxWorkers       = 10
// 	workerQueueSize  = 100
// 	contextTimeout   = 10 * time.Second
// 	maxRequestSize   = 1 << 20
// 	bcryptCost      = 12
// 	sessionDuration = 24 * time.Hour
// 	cookieDomain    = "localhost"
// 	secureCookie    = true
// 	httpOnlyCookie  = true
// 	sameSiteCookie  = http.SameSiteStrictMode
// 	tokenLength     = 32
// 	maxEmailRetries = 3
// )

// type SignInRequest struct {
// 	Email    string `json:"email" binding:"required,email,max=255"`
// 	Password string `json:"password" binding:"required,min=8,max=72"`
// }

// type SignUpRequest struct {
// 	FirstName string `json:"first_name" binding:"required,max=50"`
// 	LastName  string `json:"last_name" binding:"required,max=50"`
// 	Email     string `json:"email" binding:"required,email,max=255"`
// 	Password  string `json:"password" binding:"required,min=8,max=72"`
// }

// type VerifyEmailRequest struct {
// 	Email       string `json:"email" binding:"required,email,max=255"`
// 	VerifyToken string `json:"verify_token" binding:"required,len=44"`
// }

// type User struct {
// 	ID           uint      `gorm:"primarykey"`
// 	FirstName    string    `gorm:"size:50;not null"`
// 	LastName     string    `gorm:"size:50;not null"`
// 	Email        string    `gorm:"size:255;uniqueIndex;not null"`
// 	Password     string    `gorm:"size:72;not null"`
// 	SessionToken string    `gorm:"size:44"`
// 	VerifyToken  string    `gorm:"size:44"`
// 	IsVerified   bool      `gorm:"default:false"`
// 	CreatedAt    time.Time
// 	UpdatedAt    time.Time
// }

// type UserTask struct {
// 	Context  context.Context
// 	Request  interface{}
// 	Response chan Response
// }

// type Response struct {
// 	Status  int
// 	Message string
// 	Data    interface{}
// 	Error   error
// }

// type WorkerPool struct {
// 	tasks    chan UserTask
// 	db       *gorm.DB
// 	wg       sync.WaitGroup
// 	shutdown chan struct{}
// }

// func NewWorkerPool(db *gorm.DB) *WorkerPool {
// 	wp := &WorkerPool{
// 		tasks:    make(chan UserTask, workerQueueSize),
// 		db:       db,
// 		shutdown: make(chan struct{}),
// 	}
// 	for i := 0; i < maxWorkers; i++ {
// 		wp.wg.Add(1)
// 		go wp.worker()
// 	}
// 	return wp
// }

// func (wp *WorkerPool) worker() {
// 	defer wp.wg.Done()
// 	for {
// 		select {
// 		case task := <-wp.tasks:
// 			select {
// 			case <-task.Context.Done():
// 				task.Response <- Response{Status: http.StatusGatewayTimeout, Error: errors.New("operation timed out")}
// 			default:
// 				wp.handleTask(task)
// 			}
// 		case <-wp.shutdown:
// 			return
// 		}
// 	}
// }

// func (wp *WorkerPool) handleTask(task UserTask) {
// 	switch req := task.Request.(type) {
// 	case *SignUpRequest:
// 		user := &User{
// 			FirstName: req.FirstName,
// 			LastName:  req.LastName,
// 			Email:     req.Email,
// 			Password:  req.Password,
// 		}
// 		err := wp.db.WithContext(task.Context).Create(user).Error
// 		if err != nil {
// 			task.Response <- Response{Status: http.StatusInternalServerError, Error: err}
// 			return
// 		}
// 		task.Response <- Response{Status: http.StatusOK, Data: user}
// 	case *VerifyEmailRequest:
// 		var user User
// 		err := wp.db.WithContext(task.Context).Where("email = ?", req.Email).First(&user).Error
// 		if err != nil {
// 			task.Response <- Response{Status: http.StatusNotFound, Error: errors.New("user not found")}
// 			return
// 		}
// 		task.Response <- Response{Status: http.StatusOK, Data: user}
// 	}
// }

// func (wp *WorkerPool) Shutdown() {
// 	close(wp.shutdown)
// 	wp.wg.Wait()
// }

// func rateLimiter() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		c.Next()
// 	}
// }

// func securityHeaders() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		c.Header("X-Content-Type-Options", "nosniff")
// 		c.Header("X-Frame-Options", "DENY")
// 		c.Header("X-XSS-Protection", "1; mode=block")
// 		c.Header("Content-Security-Policy", "default-src 'self'")
// 		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
// 		c.Next()
// 	}
// }

// func RegisterAuthRoutes(r *gin.Engine, wp *WorkerPool) {
// 	r.Use(gin.Recovery())
// 	r.Use(securityHeaders())
// 	r.Use(rateLimiter())
// 	r.MaxMultipartMemory = maxRequestSize

// 	auth := r.Group("/auth")
// 	{
// 		auth.POST("/signup", handleSignup(wp))
// 		auth.POST("/verify-email", handleVerifyEmail(wp))
// 		auth.POST("/signin", handleSignin(wp))
// 		auth.GET("/verify", handleVerify(wp))
// 		auth.POST("/signout", handleSignout(wp))
// 	}
// }

// func handleSignup(wp *WorkerPool) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		ctx, cancel := context.WithTimeout(c.Request.Context(), contextTimeout)
// 		defer cancel()

// 		var req SignUpRequest
// 		if err := c.ShouldBindJSON(&req); err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
// 			return
// 		}

// 		var wg sync.WaitGroup
// 		var hashedPassword string
// 		var verifyToken string
// 		var sessionToken string
// 		var errorsChan = make(chan error, 3)

// 		wg.Add(3)

// 		go func() {
// 			defer wg.Done()
// 			hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)
// 			if err != nil {
// 				errorsChan <- err
// 				return
// 			}
// 			hashedPassword = string(hash)
// 		}()

// 		go func() {
// 			defer wg.Done()
// 			token, err := generateSecureToken()
// 			if err != nil {
// 				errorsChan <- err
// 				return
// 			}
// 			verifyToken = token
// 		}()

// 		go func() {
// 			defer wg.Done()
// 			token, err := generateSecureToken()
// 			if err != nil {
// 				errorsChan <- err
// 				return
// 			}
// 			sessionToken = token
// 		}()

// 		wg.Wait()
// 		close(errorsChan)

// 		for err := range errorsChan {
// 			if err != nil {
// 				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
// 				return
// 			}
// 		}

// 		respChan := make(chan Response, 1)
// 		wp.tasks <- UserTask{
// 			Context: ctx,
// 			Request: &SignUpRequest{
// 				FirstName: req.FirstName,
// 				LastName:  req.LastName,
// 				Email:     req.Email,
// 				Password:  hashedPassword,
// 			},
// 			Response: respChan,
// 		}

// 		select {
// 		case resp := <-respChan:
// 			if resp.Error != nil {
// 				c.JSON(resp.Status, gin.H{"error": resp.Error.Error()})
// 				return
// 			}
// 			c.JSON(resp.Status, resp.Data)
// 		case <-ctx.Done():
// 			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "Request timed out"})
// 		}
// 	}
// }

// func handleVerifyEmail(wp *WorkerPool) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		ctx, cancel := context.WithTimeout(c.Request.Context(), contextTimeout)
// 		defer cancel()

// 		var req VerifyEmailRequest
// 		if err := c.ShouldBindJSON(&req); err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
// 			return
// 		}

// 		respChan := make(chan Response, 1)
// 		wp.tasks <- UserTask{
// 			Context:  ctx,
// 			Request:  &req,
// 			Response: respChan,
// 		}

// 		select {
// 		case resp := <-respChan:
// 			if resp.Error != nil {
// 				c.JSON(resp.Status, gin.H{"error": resp.Error.Error()})
// 				return
// 			}
// 			user := resp.Data.(*User)
// 			if !compareTokens(user.VerifyToken, req.VerifyToken) {
// 				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid verification token"})
// 				return
// 			}
// 			user.IsVerified = true
// 			user.VerifyToken = ""
// 			if err := wp.db.Save(user).Error; err != nil {
// 				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify user"})
// 				return
// 			}
// 			c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
// 		case <-ctx.Done():
// 			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "Request timed out"})
// 		}
// 	}
// }

// func handleSignin(wp *WorkerPool) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		ctx, cancel := context.WithTimeout(c.Request.Context(), contextTimeout)
// 		defer cancel()

// 		var req SignInRequest
// 		if err := c.ShouldBindJSON(&req); err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
// 			return
// 		}

// 		var user User
// 		if err := wp.db.WithContext(ctx).Where("email = ?", req.Email).First(&user).Error; err != nil {
// 			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
// 			return
// 		}

// 		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
// 			return
// 		}

// 		if !user.IsVerified {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "Email not verified"})
// 			return
// 		}

// 		sessionToken, err := generateSecureToken()
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate session"})
// 			return
// 		}

// 		user.SessionToken = sessionToken
// 		if err := wp.db.Save(&user).Error; err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
// 			return
// 		}

// 		c.SetCookie("session_token", sessionToken, int(sessionDuration.Seconds()), "/", cookieDomain, secureCookie, httpOnlyCookie)
// 		c.JSON(http.StatusOK, gin.H{"message": "Signed in successfully", "user": user})
// 	}
// }

// func handleVerify(wp *WorkerPool) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		ctx, cancel := context.WithTimeout(c.Request.Context(), contextTimeout)
// 		defer cancel()

// 		sessionToken, err := c.Cookie("session_token")
// 		if err != nil {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "No session token"})
// 			return
// 		}

// 		var user User
// 		if err := wp.db.WithContext(ctx).Where("session_token = ?", sessionToken).First(&user).Error; err != nil {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
// 			return
// 		}

// 		c.JSON(http.StatusOK, gin.H{"user": user})
// 	}
// }

// func handleSignout(wp *WorkerPool) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		ctx, cancel := context.WithTimeout(c.Request.Context(), contextTimeout)
// 		defer cancel()

// 		sessionToken, err := c.Cookie("session_token")
// 		if err != nil {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "No session token"})
// 			return
// 		}

// 		var user User
// 		if err := wp.db.WithContext(ctx).Where("session_token = ?", sessionToken).First(&user).Error; err != nil {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
// 			return
// 		}

// 		user.SessionToken = ""
// 		if err := wp.db.Save(&user).Error; err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign out"})
// 			return
// 		}

// 		c.SetCookie("session_token", "", -1, "/", cookieDomain, secureCookie, httpOnlyCookie)
// 		c.JSON(http.StatusOK, gin.H{"message": "Signed out successfully"})
// 	}
// }

// func compareTokens(a, b string) bool {
// 	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
// }

//	func generateSecureToken() (string, error) {
//		b := make([]byte, tokenLength)
//		if _, err := rand.Read(b); err != nil {
//			return "", err
//		}
//		return base64.URLEncoding.EncodeToString(b), nil
//	}
package handler

import "log"

func State() {
	log.Printf("Test")
}
