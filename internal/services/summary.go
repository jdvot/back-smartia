package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// SummaryService handles AI summarization
type SummaryService struct {
	openaiURL string
	openaiKey string
	geminiURL string
	geminiKey string
}

// NewSummaryService creates a new summary service
func NewSummaryService() *SummaryService {
	return &SummaryService{
		openaiURL: os.Getenv("OPENAI_API_URL"),
		openaiKey: os.Getenv("OPENAI_API_KEY"),
		geminiURL: os.Getenv("GEMINI_API_URL"),
		geminiKey: os.Getenv("GEMINI_API_KEY"),
	}
}

// GenerateSummary generates a summary of the provided text
func (s *SummaryService) GenerateSummary(ctx context.Context, text string) (string, error) {
	// Try OpenAI first
	if s.openaiURL != "" && s.openaiKey != "" {
		return s.generateWithOpenAI(ctx, text)
	}

	// Try Gemini
	if s.geminiURL != "" && s.geminiKey != "" {
		return s.generateWithGemini(ctx, text)
	}

	// Fallback to mock summary
	return s.generateMockSummary(ctx, text)
}

// generateWithOpenAI uses OpenAI GPT for summarization
func (s *SummaryService) generateWithOpenAI(ctx context.Context, text string) (string, error) {
	// Truncate text if too long (OpenAI has token limits)
	if len(text) > 4000 {
		text = text[:4000] + "..."
	}

	requestBody := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a helpful assistant that creates concise summaries of documents. Provide a clear, well-structured summary in 2-3 sentences.",
			},
			{
				"role":    "user",
				"content": fmt.Sprintf("Please summarize the following document text:\n\n%s", text),
			},
		},
		"max_tokens": 150,
		"temperature": 0.3,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.openaiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.openaiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var openaiResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&openaiResponse); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if openaiResponse.Error.Message != "" {
		return "", fmt.Errorf("OpenAI error: %s", openaiResponse.Error.Message)
	}

	if len(openaiResponse.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return strings.TrimSpace(openaiResponse.Choices[0].Message.Content), nil
}

// generateWithGemini uses Google Gemini for summarization
func (s *SummaryService) generateWithGemini(ctx context.Context, text string) (string, error) {
	// Truncate text if too long
	if len(text) > 30000 {
		text = text[:30000] + "..."
	}

	requestBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{
						"text": fmt.Sprintf("Please provide a concise summary of the following document in 2-3 sentences:\n\n%s", text),
					},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"maxOutputTokens": 150,
			"temperature":     0.3,
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.geminiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var geminiResponse struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&geminiResponse); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if geminiResponse.Error.Message != "" {
		return "", fmt.Errorf("Gemini error: %s", geminiResponse.Error.Message)
	}

	if len(geminiResponse.Candidates) == 0 || len(geminiResponse.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from Gemini")
	}

	return strings.TrimSpace(geminiResponse.Candidates[0].Content.Parts[0].Text), nil
}

// generateMockSummary returns a mock summary for testing
func (s *SummaryService) generateMockSummary(ctx context.Context, text string) (string, error) {
	// Create a simple mock summary based on text length
	wordCount := len(strings.Fields(text))
	
	if wordCount < 10 {
		return "This is a short document with minimal content.", nil
	} else if wordCount < 50 {
		return "This document contains moderate content that has been processed for summarization.", nil
	} else {
		return "This is a comprehensive document with substantial content that has been analyzed and summarized for easy understanding.", nil
	}
} 