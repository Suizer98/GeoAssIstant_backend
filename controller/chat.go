package controller

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"geoai-app/model"
	"geoai-app/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

type ChatController struct {
	SystemPrompt map[string]string
	ChatHistory  []map[string]string
	mu           sync.Mutex
	DB           *sql.DB
}

// NewChatController creates a new instance of ChatController
func NewChatController(db *sql.DB) *ChatController {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: No .env file found")
	}

	// Initialize with system prompt
	systemPrompt := map[string]string{
		"role": "system",
		"content": `You are a helpful assistant named GeoAI. Respond concisely to the user's queries in the following format:
		{
		"locations": "comma-separated list of key locations",
		"messages": "detailed response to the user's query"
		}`,
	}

	return &ChatController{
		SystemPrompt: systemPrompt,
		ChatHistory:  []map[string]string{systemPrompt},
		DB:           db,
	}
}

// @Summary Handle chat requests
// @Description Process chat requests and generate responses using the Groq API. Optionally, provide a `user_id` query parameter to associate the chat with a specific user.
// @Tags chat
// @Accept json
// @Produce json
// @Param user_id query string false "Optional User ID to associate the chat with a user"
// @Param requestBody body model.ChatRequest true "Chat request body containing the user's input"
// @Success 200 {object} map[string]interface{} "Assistant's response, with locations and messages"
// @Failure 400 {object} map[string]interface{} "Bad request - missing or invalid input"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Server error - issue with Groq API or database operations"
// @Router /chat [post]
func (cc *ChatController) HandleChatRequest(c *gin.Context) {
	userID := c.Query("user_id")
	var user *model.User
	var err error

	// Validate user if user_id is provided
	if userID != "" {
		userRepo := repository.NewUserRepository(cc.DB)
		user, err = userRepo.GetUserByID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
			return
		}
		if user == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id: user not found"})
			return
		}
	}

	var requestBody model.ChatRequest
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add the user's input to chat history
	cc.mu.Lock()
	cc.ChatHistory = append(cc.ChatHistory, map[string]string{
		"role":    "user",
		"content": requestBody.Content,
	})
	cc.mu.Unlock()

	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GROQ_API_KEY is not set"})
		return
	}

	// Make request to the Groq API
	payload := map[string]interface{}{
		"model":    "llama-3.3-70b-versatile",
		"messages": cc.ChatHistory,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request payload"})
		return
	}

	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(payloadBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make the request"})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read the response"})
		return
	}

	var apiResponse struct {
		Choices []struct {
			Message map[string]string `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &apiResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse API response"})
		return
	}

	assistantResponse := apiResponse.Choices[0].Message
	cc.mu.Lock()
	cc.ChatHistory = append(cc.ChatHistory, assistantResponse)
	cc.mu.Unlock()

	// Save conversation if user_id is provided and valid
	if user != nil {
		conversationRepo := repository.NewConversationRepository(cc.DB)

		// Check for existing conversation
		existingConversation, err := conversationRepo.GetLatestConversationByUserID(int(user.ID))
		if err != nil && err != sql.ErrNoRows {
			fmt.Printf("Error fetching latest conversation: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch conversation"})
			return
		}

		if existingConversation == nil {
			// Create a new conversation
			conversation := model.Conversation{
				UserID:         user.ID,
				ConversationID: uuid.New().String(),
				ChatHistory:    cc.ChatHistory, // Pass raw Go type
			}

			if err := conversationRepo.CreateConversation(&conversation); err != nil {
				fmt.Printf("Error saving conversation: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save conversation"})
				return
			}
		} else {
			// Update the existing conversation
			existingConversation.ChatHistory = cc.ChatHistory

			if err := conversationRepo.UpdateConversation(existingConversation); err != nil {
				fmt.Printf("Error updating conversation: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update conversation"})
				return
			}
		}
	}

	// Return the assistant's response
	c.JSON(http.StatusOK, gin.H{
		"role":    "assistant",
		"content": assistantResponse,
	})
}
