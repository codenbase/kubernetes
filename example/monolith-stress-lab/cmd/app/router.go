package app

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB, jwtSecret []byte) *gin.Engine {
	// Disable Gin's debug output to reduce overhead during load tests
	gin.SetMode(gin.ReleaseMode)

	r := gin.New() // Use New instead of Default to avoid default logger middleware
	r.Use(gin.Recovery())

	authHandler := NewAuthHandler(db, jwtSecret)

	r.POST("/login", authHandler.Login)
	r.GET("/verify", authHandler.Verify)

	return r
}
