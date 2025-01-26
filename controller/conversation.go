package controller

import (
	"database/sql"
	"net/http"

	// "geoai-app/model"
	"geoai-app/repository"
	"github.com/gin-gonic/gin"
)

type ConversationController struct {
	DB *sql.DB
}

func NewConversationController(db *sql.DB) *ConversationController {
	return &ConversationController{DB: db}
}

// @Summary List conversations
// @Description List all conversations for a user
// @Tags conversations
// @Produce json
// @Param user_id query string true "User ID to fetch conversations"
// @Success 200 {object} map[string]interface{} "List of conversations"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "No conversations found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /conversations [get]
func (cc *ConversationController) GetConversations(ctx *gin.Context) {
	userID := ctx.Query("user_id")
	if userID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	repo := repository.NewConversationRepository(cc.DB)
	conversations, err := repo.GetConversationsByUserID(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch conversations"})
		return
	}

	if len(conversations) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "No conversations found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"conversations": conversations})
}
