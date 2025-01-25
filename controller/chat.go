package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type ChatController struct {
	SystemPrompt map[string]string
	ChatHistory  []map[string]string
	mu           sync.Mutex // To handle concurrent access to ChatHistory
}

// NewChatController creates a new instance of ChatController
func NewChatController() *ChatController {
	// Load the .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: No .env file found")
	}

	// Initialize with system prompt
	systemPrompt := map[string]string{
		"role": "system",
		"content": "You are a helpful assistant named GeoAI. Respond concisely to the user's queries.",
	}

	return &ChatController{
		SystemPrompt: systemPrompt,
		ChatHistory:  []map[string]string{systemPrompt},
	}
}

// HandleChatRequest processes requests for the /chat route
func (cc *ChatController) HandleChatRequest(c *gin.Context) {
	var requestBody struct {
		Content string `json:"content" binding:"required"`
	}

	// Parse the incoming request
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add the user input to the chat history
	cc.mu.Lock()
	cc.ChatHistory = append(cc.ChatHistory, map[string]string{
		"role":    "user",
		"content": requestBody.Content,
	})
	cc.mu.Unlock()

	// Fetch the Groq API key from the environment
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GROQ_API_KEY is not set"})
		return
	}

	// Create the payload for the Groq API
	payload := map[string]interface{}{
		"model":    "llama-3.3-70b-versatile",
		"messages": cc.ChatHistory,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request payload"})
		return
	}

	// Send the request to the Groq API
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

	// Read the response body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read the response"})
		return
	}

	// Parse and append the assistant's response to the chat history
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

	// Return the assistant's response to the client
	c.JSON(http.StatusOK, assistantResponse)
}
