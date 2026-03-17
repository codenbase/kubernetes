package app

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Comment represents a user's comment on an article.
type Comment struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int       `gorm:"not null;index" json:"user_id"`
	ArticleID int       `gorm:"not null" json:"article_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type CommentHandler struct {
	db *gorm.DB
}

func NewCommentHandler(db *gorm.DB) *CommentHandler {
	return &CommentHandler{db: db}
}

// CreateRequest defines the expected payload for posting a comment.
type CreateCommentRequest struct {
	UserID    int    `json:"user_id" binding:"required"`
	ArticleID int    `json:"article_id" binding:"required"`
	Content   string `json:"content" binding:"required"`
}

// CreateComment synchronously writes a new comment to the database.
func (h *CommentHandler) CreateComment(c *gin.Context) {
	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json body or missing fields"})
		return
	}

	comment := Comment{
		UserID:    req.UserID,
		ArticleID: req.ArticleID,
		Content:   req.Content,
	}

	// Synchronous blocking database write operation
	result := h.db.Create(&comment)
	if result.Error != nil {
		log.Printf("Database write error: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create comment"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "comment created success", "id": comment.ID})
}
