package app

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Article represents a blog post or news article.
type Article struct {
	ID       int    `gorm:"primaryKey;autoIncrement" json:"id"`
	Title    string `gorm:"size:255;not null" json:"title"`
	Content  string `gorm:"type:text;not null" json:"content"`
	SizeType string `gorm:"size:20;not null" json:"size_type"`
}

type ArticleHandler struct {
	db *gorm.DB
}

func NewArticleHandler(db *gorm.DB) *ArticleHandler {
	return &ArticleHandler{db: db}
}

// GetArticle fetches a single article by ID direct from the database.
func (h *ArticleHandler) GetArticle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article ID format"})
		return
	}

	var article Article
	result := h.db.First(&article, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		} else {
			log.Printf("Database read error: %v", result.Error)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, article)
}
