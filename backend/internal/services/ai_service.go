package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/vikukumar/Pushpaka/internal/config"
)

// AIService provides a provider-agnostic interface to large-language models.
// Supported providers: openai, openrouter, gemini, anthropic, ollama.
// All providers except Anthropic use the OpenAI-compatible chat/completions API.
type AIService struct {
	cfg *config.Config
}

func NewAIService(cfg *config.Config) *AIService {
	return &AIService{cfg: cfg}
}

// Available reports whether an AI API key is configured.
func (s *AIService) Available() bool {
	return s.cfg.AIAPIKey != ""
}

// AnalyzeLogs sends deployment logs to the configured AI provider and returns
// a natural-language root-cause analysis with suggested fixes.
func (s *AIService) AnalyzeLogs(logs string) (string, error) {
	if !s.Available() {
		return "", errors.New("AI integration not configured: set AI_API_KEY")
	}

	systemPrompt := `You are an expert DevOps engineer. The user will give you deployment build logs.
Analyze them, identify the root cause of any failures, and provide a concise explanation with
actionable fix suggestions. Format your response in plain text with clear sections:
1. Root Cause  2. Explanation  3. Suggested Fix`

	userPrompt := fmt.Sprintf("Analyze the following deployment logs and identify any errors:\n\n---\n%s\n---", truncateLogs(logs, 8000))

	return s.complete(systemPrompt, userPrompt)
}

// Ask sends an arbitrary prompt to the AI and returns the response.
func (s *AIService) Ask(systemPrompt, userPrompt string) (string, error) {
	if !s.Available() {
		return "", errors.New("AI integration not configured: set AI_API_KEY")
	}
	return s.complete(systemPrompt, userPrompt)
}

func (s *AIService) complete(system, user string) (string, error) {
	switch strings.ToLower(s.cfg.AIProvider) {
	case "anthropic":
		return s.anthropicComplete(system, user)
	default:
		// OpenAI-compatible: openai, openrouter, gemini (via compatibility endpoint), ollama
		return s.openAIComplete(system, user)
	}
}

// ─── OpenAI-compatible ────────────────────────────────────────────────────────

type openAIRequest struct {
	Model    string              `json:"model"`
	Messages []openAIMessage     `json:"messages"`
	MaxTokens int               `json:"max_tokens,omitempty"`
}
type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type openAIResponse struct {
	Choices []struct {
		Message openAIMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (s *AIService) openAIComplete(system, user string) (string, error) {
	endpoint := s.resolveEndpoint()
	model := s.cfg.AIModel
	if model == "" {
		model = "gpt-4o-mini"
	}

	body := openAIRequest{
		Model: model,
		Messages: []openAIMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		MaxTokens: 1024,
	}
	data, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", endpoint+"/chat/completions", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.cfg.AIAPIKey)
	// OpenRouter requires a site header for tracking
	if strings.Contains(endpoint, "openrouter") {
		req.Header.Set("HTTP-Referer", "https://pushpaka")
		req.Header.Set("X-Title", "Pushpaka")
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("AI request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result openAIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parsing AI response: %w", err)
	}
	if result.Error != nil {
		return "", fmt.Errorf("AI error: %s", result.Error.Message)
	}
	if len(result.Choices) == 0 {
		return "", errors.New("AI returned no choices")
	}
	return result.Choices[0].Message.Content, nil
}

func (s *AIService) resolveEndpoint() string {
	if s.cfg.AIBaseURL != "" {
		return strings.TrimRight(s.cfg.AIBaseURL, "/")
	}
	switch strings.ToLower(s.cfg.AIProvider) {
	case "openrouter":
		return "https://openrouter.ai/api/v1"
	case "gemini":
		return "https://generativelanguage.googleapis.com/v1beta/openai"
	case "ollama":
		return "http://localhost:11434/v1"
	default: // openai
		return "https://api.openai.com/v1"
	}
}

// ─── Anthropic ────────────────────────────────────────────────────────────────

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system"`
	Messages  []anthropicMessage `json:"messages"`
}
type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (s *AIService) anthropicComplete(system, user string) (string, error) {
	model := s.cfg.AIModel
	if model == "" {
		model = "claude-3-haiku-20240307"
	}

	body := anthropicRequest{
		Model:     model,
		MaxTokens: 1024,
		System:    system,
		Messages:  []anthropicMessage{{Role: "user", Content: user}},
	}
	data, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", s.cfg.AIAPIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Anthropic request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result anthropicResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parsing Anthropic response: %w", err)
	}
	if result.Error != nil {
		return "", fmt.Errorf("Anthropic error: %s", result.Error.Message)
	}
	if len(result.Content) == 0 {
		return "", errors.New("Anthropic returned empty content")
	}
	return result.Content[0].Text, nil
}

// truncateLogs trims logs to approximately maxBytes characters, keeping the
// beginning and end (most useful for diagnosing failures).
func truncateLogs(logs string, maxBytes int) string {
	if len(logs) <= maxBytes {
		return logs
	}
	half := maxBytes / 2
	return logs[:half] + "\n\n... [truncated] ...\n\n" + logs[len(logs)-half:]
}
