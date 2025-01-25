package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"geoai-app/model"
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
		"content": `You are a helpful assistant named GeoAI. Respond concisely to the user's queries in the following format:
		{
		"locations": "comma-separated list of key locations",
		"messages": "detailed response to the user's query"
		}`,
	}

	return &ChatController{
		SystemPrompt: systemPrompt,
		ChatHistory:  []map[string]string{systemPrompt},
	}
}

// @Summary Handle chat requests
// @Description Process chat requests and generate responses using the Groq API
// @Tags chat
// @Accept json
// @Produce json
// @Param requestBody body model.ChatRequest true "Chat request body"
// @Success 200 {object} map[string]string "Assistant's response"
// @Failure 400 {object} map[string]interface{} "Bad request - missing or invalid input"
// @Failure 500 {object} map[string]interface{} "Server error - issue with Groq API or environment variables"
// @Router /chat [post]
func (cc *ChatController) HandleChatRequest(c *gin.Context) {
	var requestBody model.ChatRequest

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
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read the response"})
		return
	}

	// Parse the API response
	var apiResponse struct {
		Choices []struct {
			Message map[string]string `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &apiResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse API response"})
		return
	}

	// Extract the assistant's response
	assistantResponse := apiResponse.Choices[0].Message
	content := assistantResponse["content"]

	// Parse the response content to extract `locations` and `messages`
	locations, messages := extractStructuredContent(content)

	// Add the assistant's response to the chat history
	cc.mu.Lock()
	cc.ChatHistory = append(cc.ChatHistory, assistantResponse)
	cc.mu.Unlock()

	// Return the transformed structured response
	c.JSON(http.StatusOK, gin.H{
		"role": "assistant",
		"content": map[string]string{
			"locations": locations,
			"messages":  messages,
		},
	})
}

// extractStructuredContent extracts `locations` and `messages` from the JSON content
func extractStructuredContent(content string) (string, string) {
	// Define a structure to match the expected response
	var parsedContent struct {
		Locations string `json:"locations"`
		Messages  string `json:"messages"`
	}

	// Try to parse the content as JSON
	if err := json.Unmarshal([]byte(content), &parsedContent); err != nil {
		// If parsing fails, fall back to returning the original content as `messages`
		return "", content
	}

	// Return the extracted locations and messages
	return parsedContent.Locations, parsedContent.Messages
}
