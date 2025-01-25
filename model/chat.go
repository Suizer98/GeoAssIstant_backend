package model

// ChatRequest represents the request body for the /chat endpoint
type ChatRequest struct {
	// Content is the user's input to the chat
	Content string `json:"content" binding:"required"`
}
