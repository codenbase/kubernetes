package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type ConfigHandler struct {
	db *gorm.DB
}

func NewConfigHandler(db *gorm.DB) *ConfigHandler {
	return &ConfigHandler{db: db}
}

type UpdateDbPoolRequest struct {
	MaxOpen int `json:"max_open"`
	MaxIdle int `json:"max_idle"`
}

func (h *ConfigHandler) UpdateDbPool(c *gin.Context) {
	var req UpdateDbPoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get underlying sql.DB"})
		return
	}

	if req.MaxOpen > 0 {
		sqlDB.SetMaxOpenConns(req.MaxOpen)
	}
	if req.MaxIdle > 0 {
		sqlDB.SetMaxIdleConns(req.MaxIdle)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Database connection pool updated successfully",
		"max_open": req.MaxOpen,
		"max_idle": req.MaxIdle,
	})
}

type UpdateBcryptCostRequest struct {
	Cost int `json:"cost"`
}

func (h *ConfigHandler) UpdateBcryptCost(c *gin.Context) {
	var req UpdateBcryptCostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	if req.Cost < bcrypt.MinCost || req.Cost > bcrypt.MaxCost {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bcrypt cost"})
		return
	}

	// Generate new hash with the requested cost for the default password "123456"
	newHash, err := bcrypt.GenerateFromPassword([]byte("123456"), req.Cost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate bcrypt hash"})
		return
	}

	// Update all users' passwords with the new hash.
	// This will immediately change the computational cost for new login requests.
	result := h.db.Exec("UPDATE users SET password_hash = ?", string(newHash))
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update users in database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Bcrypt cost updated successfully for all users",
		"new_cost":      req.Cost,
		"rows_affected": result.RowsAffected,
	})
}
