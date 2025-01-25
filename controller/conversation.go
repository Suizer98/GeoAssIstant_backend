package controller

import (
	"database/sql"
	"net/http"

	"geoai-app/model"
	"geoai-app/repository"
	"github.com/gin-gonic/gin"
)

type ConversationController struct {
	DB *sql.DB
}

func NewConversationController(db *sql.DB) *ConversationController {
	return &ConversationController{DB: db}
}

func (c *ConversationController) GetConversations(ctx *gin.Context) {
	userID := ctx.Query("user_id")
	repo := repository.NewConversationRepository(c.DB)

	if userID != "" {
		// Fetch conversations by user ID
		conversations, err := repo.GetConversationsByUserID(userID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Failed to fetch conversations"})
			return
		}
		if len(conversations) == 0 {
			ctx.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": "No conversations found"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": conversations})
		return
	}

	// If no user_id is provided, return a bad request error
	ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "user_id query parameter is required"})
}

func (c *ConversationController) CreateConversation(ctx *gin.Context) {
	var newConversation model.Conversation
	if err := ctx.ShouldBindJSON(&newConversation); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": "Invalid request data"})
		return
	}

	repo := repository.NewConversationRepository(c.DB)
	if err := repo.CreateConversation(&newConversation); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": "Failed to create conversation"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": newConversation})
}
