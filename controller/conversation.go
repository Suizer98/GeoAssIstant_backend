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

// @Summary Get conversations by user ID
// @Description Fetch a list of conversations for a specific user by their user ID
// @Tags conversations
// @Produce json
// @Param user_id query string true "User ID to fetch conversations"
// @Success 200 {object} map[string]interface{} "Success response with conversations data"
// @Failure 400 {object} map[string]interface{} "Bad request - user_id query parameter is required"
// @Failure 404 {object} map[string]interface{} "No conversations found for the given user ID"
// @Failure 500 {object} map[string]interface{} "Failed to fetch conversations due to an internal server error"
// @Router /conversations [get]
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

// @Summary Create a new conversation
// @Description Add a new conversation to the database
// @Tags conversations
// @Accept json
// @Produce json
// @Param conversation body model.Conversation true "New conversation data"
// @Success 201 {object} map[string]interface{} "Conversation successfully created"
// @Failure 400 {object} map[string]interface{} "Invalid request data"
// @Failure 500 {object} map[string]interface{} "Failed to create conversation due to an internal server error"
// @Router /conversations [post]
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
