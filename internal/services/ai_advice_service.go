package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/zayyadi/finance-tracker/internal/models"
)

// OpenRouterRequest defines the structure for the request to OpenRouter API.
type OpenRouterRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

// ChatMessage defines the structure for a message in the OpenRouter request.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenRouterResponseChoice defines a single choice in the OpenRouter response.
type OpenRouterResponseChoice struct {
	Message ChatMessage `json:"message"`
}

// OpenRouterResponse defines the structure for the response from OpenRouter API.
type OpenRouterResponse struct {
	Choices []OpenRouterResponseChoice `json:"choices"`
	Error   *OpenRouterError           `json:"error,omitempty"`
}

// OpenRouterError defines the structure for an error object in the OpenRouter response.
type OpenRouterError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
}

// AIAdviceService provides methods for getting AI-based financial advice.
type AIAdviceService struct{}

// NewAIAdviceService creates a new AIAdviceService.
func NewAIAdviceService() *AIAdviceService {
	return &AIAdviceService{}
}

// GetFinancialAdvice contacts OpenRouter API to get financial advice based on the summary.
func (s *AIAdviceService) GetFinancialAdvice(summary *models.FinancialSummary) (string, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Println("Warning: OPENROUTER_API_KEY environment variable not set. Returning predefined message.")
		return "AI features are currently unavailable as the API key is not configured.", nil
		// Alternatively, return an error:
		// return "", fmt.Errorf("OPENROUTER_API_KEY is not set")
	}
	if apiKey == "YOUR_DUMMY_OPENROUTER_API_KEY_FOR_TESTING" { // Check for dummy key
		log.Println("Warning: Using dummy OPENROUTER_API_KEY. Returning predefined test message.")
		return "This is a test advice message because a dummy API key is being used. Your net balance is good!", nil
	}

	prompt := fmt.Sprintf(
		"Given this financial summary: Total Income %.2f, Total Expenses %.2f, Net Balance %.2f for the period %s to %s, provide concise financial advice in 2-3 short sentences.",
		summary.TotalIncome,
		summary.TotalExpenses,
		summary.NetBalance,
		summary.PeriodStartDate.Format("January 2, 2006"),
		summary.PeriodEndDate.Format("January 2, 2006"),
	)

	requestPayload := OpenRouterRequest{
		Model: "mistralai/mistral-7b-instruct:free", // Example free model
		Messages: []ChatMessage{
			{Role: "user", Content: prompt},
		},
	}

	payloadBytes, err := json.Marshal(requestPayload)
	if err != nil {
		log.Printf("Error marshalling OpenRouter request: %v", err)
		return "", fmt.Errorf("error preparing request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Error creating new HTTP request: %v", err)
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	// Optional: Add other headers as recommended by OpenRouter, e.g., HTTP-Referer
	req.Header.Set("HTTP-Referer", "http://localhost:8080") // Replace with your actual app URL
	req.Header.Set("X-Title", "Finance Tracker Go")         // Replace with your app name

	client := &http.Client{Timeout: time.Second * 30} // Set a timeout
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making request to OpenRouter: %v", err)
		return "", fmt.Errorf("error contacting AI service: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading OpenRouter response body: %v", err)
		return "", fmt.Errorf("error reading AI response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("OpenRouter API responded with status %d: %s", resp.StatusCode, string(responseBody))
		// Try to parse OpenRouterError if present
		var errResp OpenRouterResponse
		if json.Unmarshal(responseBody, &errResp) == nil && errResp.Error != nil {
			return "", fmt.Errorf("AI service error (%s): %s", resp.Status, errResp.Error.Message)
		}
		return "", fmt.Errorf("AI service responded with status: %s", resp.Status)
	}

	var openRouterResp OpenRouterResponse
	if err := json.Unmarshal(responseBody, &openRouterResp); err != nil {
		log.Printf("Error unmarshalling OpenRouter response: %v. Body: %s", err, string(responseBody))
		return "", fmt.Errorf("error parsing AI response: %w", err)
	}

	if len(openRouterResp.Choices) == 0 || openRouterResp.Choices[0].Message.Content == "" {
		log.Printf("OpenRouter response contained no advice. Body: %s", string(responseBody))
		return "", fmt.Errorf("AI service returned empty advice")
	}

	return openRouterResp.Choices[0].Message.Content, nil
}
