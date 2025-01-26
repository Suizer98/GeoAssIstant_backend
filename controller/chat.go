package controller

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
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
// @Description Start a new conversation or continue an existing one
// @Tags chat
// @Accept json
// @Produce json
// @Param user_id query string true "User ID to associate the chat"
// @Param uuid query string false "UUID of the existing conversation"
// @Param requestBody body model.ChatRequest true "Chat request body"
// @Success 200 {object} map[string]interface{} "Chat response"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Conversation not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /chat [post]
func (cc *ChatController) HandleChatRequest(c *gin.Context) {
	userIDStr := c.Query("user_id")
	conversationUUID := c.Query("uuid")

	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	// Parse user ID to uint
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id format"})
		return
	}

	// Check if the user exists
	userRepo := repository.NewUserRepository(cc.DB)
	user, err := userRepo.GetUserByID(strconv.Itoa(int(userID)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking user existence"})
		return
	}
	if user == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id: user does not exist"})
		return
	}

	// Initialize repositories
	conversationRepo := repository.NewConversationRepository(cc.DB)

	var conversation *model.Conversation

	if conversationUUID == "" {
		// Create a new conversation
		conversation = &model.Conversation{
			UserID:         uint(userID),
			ConversationID: uuid.New().String(),
			ChatHistory:    []map[string]string{cc.SystemPrompt},
		}
		err = conversationRepo.CreateConversation(conversation)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create a new conversation"})
			return
		}
		cc.ChatHistory = conversation.ChatHistory
	} else {
		// Continue an existing conversation
		conversation, err = conversationRepo.GetConversationByUUID(conversationUUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch conversation"})
			return
		}
		if conversation == nil || conversation.UserID != uint(userID) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
			return
		}
		cc.ChatHistory = conversation.ChatHistory
	}

	var requestBody model.ChatRequest
	err = c.ShouldBindJSON(&requestBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Append user input
	cc.mu.Lock()
	cc.ChatHistory = append(cc.ChatHistory, map[string]string{
		"role":    "user",
		"content": requestBody.Content,
	})
	cc.mu.Unlock()

	// Send to Groq API
	responseMessage, err := sendToGroqAPI(cc.ChatHistory)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch response from Groq API"})
		return
	}

	cc.mu.Lock()
	cc.ChatHistory = append(cc.ChatHistory, responseMessage)
	cc.mu.Unlock()

	// Update conversation
	conversation.ChatHistory = cc.ChatHistory
	err = conversationRepo.UpdateConversation(conversation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save conversation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": responseMessage})
}

// Helper function to interact with the Groq API
func sendToGroqAPI(chatHistory []map[string]string) (map[string]string, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		fmt.Println("GROQ_API_KEY is not set")
		return nil, fmt.Errorf("GROQ_API_KEY not set")
	}

	payload := map[string]interface{}{
		"model":    "llama-3.3-70b-versatile",
		"messages": chatHistory,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshaling payload: %v\n", err)
		return nil, err
	}
	fmt.Printf("Payload: %s\n", string(payloadBytes))

	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(payloadBytes))
	if err != nil {
		fmt.Printf("Error creating HTTP request: %v\n", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making HTTP request: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return nil, err
	}
	fmt.Printf("API Response Status: %d, Body: %s\n", resp.StatusCode, string(respBody))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Groq API returned status: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var apiResponse struct {
		Choices []struct {
			Message map[string]string `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &apiResponse); err != nil {
		fmt.Printf("Error unmarshaling response: %v\n", err)
		return nil, err
	}

	if len(apiResponse.Choices) == 0 {
		return nil, fmt.Errorf("No choices returned in response")
	}

	return apiResponse.Choices[0].Message, nil
}
