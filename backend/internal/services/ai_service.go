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
	"github.com/vikukumar/Pushpaka/internal/models"
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

// Available reports whether an AI API key is configured (server-wide or user-specific).
func (s *AIService) Available() bool {
	return s.cfg.AIAPIKey != ""
}

// AvailableWithUserConfig reports whether the user or global config has an API key.
func (s *AIService) AvailableWithUserConfig(userCfg *models.AIConfig) bool {
	if userCfg != nil && userCfg.APIKey != "" {
		return true
	}
	return s.cfg.AIAPIKey != ""
}

// resolvedProvider returns the effective provider for a user (user config takes precedence).
func (s *AIService) resolvedProvider(userCfg *models.AIConfig) string {
	if userCfg != nil && userCfg.Provider != "" {
		return userCfg.Provider
	}
	if s.cfg.AIProvider != "" {
		return s.cfg.AIProvider
	}
	return "openai"
}

func (s *AIService) resolvedAPIKey(userCfg *models.AIConfig) string {
	if userCfg != nil && userCfg.APIKey != "" {
		return userCfg.APIKey
	}
	return s.cfg.AIAPIKey
}

func (s *AIService) resolvedModel(userCfg *models.AIConfig) string {
	if userCfg != nil && userCfg.Model != "" {
		return userCfg.Model
	}
	if s.cfg.AIModel != "" {
		return s.cfg.AIModel
	}
	return "gpt-4o-mini"
}

func (s *AIService) resolvedBaseURL(userCfg *models.AIConfig) string {
	if userCfg != nil && userCfg.BaseURL != "" {
		return userCfg.BaseURL
	}
	return s.cfg.AIBaseURL
}

// buildRAGContext prepends RAG documents as context to the system prompt.
func buildRAGContext(ragDocs []models.RAGDocument, systemPrompt string) string {
	if len(ragDocs) == 0 {
		return systemPrompt
	}
	var sb strings.Builder
	sb.WriteString("## Custom Knowledge Base (use this reference information to inform your answers)\n\n")
	for i, doc := range ragDocs {
		sb.WriteString(fmt.Sprintf("### Document %d: %s\n%s\n\n", i+1, doc.Title, doc.Content))
	}
	sb.WriteString("---\n\n")
	sb.WriteString(systemPrompt)
	return sb.String()
}

// AnalyzeLogs sends deployment logs to the configured AI provider and returns
// a natural-language root-cause analysis with suggested fixes.
func (s *AIService) AnalyzeLogs(logs string) (string, error) {
	return s.AnalyzeLogsWithConfig(nil, nil, logs)
}

// AnalyzeLogsWithConfig is the user-config-aware version of AnalyzeLogs.
func (s *AIService) AnalyzeLogsWithConfig(userCfg *models.AIConfig, ragDocs []models.RAGDocument, logs string) (string, error) {
	if !s.AvailableWithUserConfig(userCfg) {
		return "", errors.New("AI integration not configured: set AI_API_KEY or configure your AI settings")
	}

	baseSystem := `You are an expert DevOps engineer. The user will give you deployment build logs.
Analyze them, identify the root cause of any failures, and provide a concise explanation with
actionable fix suggestions. Format your response in plain text with clear sections:
1. Root Cause  2. Explanation  3. Suggested Fix`

	// Use user's custom system prompt if set
	if userCfg != nil && userCfg.SystemPrompt != "" {
		baseSystem = userCfg.SystemPrompt
	}

	systemPrompt := buildRAGContext(ragDocs, baseSystem)
	userPrompt := fmt.Sprintf("Analyze the following deployment logs and identify any errors:\n\n---\n%s\n---", truncateLogs(logs, 8000))

	return s.completeWithConfig(userCfg, systemPrompt, userPrompt)
}

// Ask sends an arbitrary prompt to the AI and returns the response.
func (s *AIService) Ask(systemPrompt, userPrompt string) (string, error) {
	return s.AskWithConfig(nil, nil, systemPrompt, userPrompt)
}

// AskWithConfig is the user-config-aware version of Ask.
func (s *AIService) AskWithConfig(userCfg *models.AIConfig, ragDocs []models.RAGDocument, systemPrompt, userPrompt string) (string, error) {
	if !s.AvailableWithUserConfig(userCfg) {
		return "", errors.New("AI integration not configured: set AI_API_KEY or configure your AI settings")
	}

	// Merge custom system prompt + RAG
	if userCfg != nil && userCfg.SystemPrompt != "" {
		systemPrompt = userCfg.SystemPrompt + "\n\n" + systemPrompt
	}
	fullSystem := buildRAGContext(ragDocs, systemPrompt)

	return s.completeWithConfig(userCfg, fullSystem, userPrompt)
}

func (s *AIService) complete(system, user string) (string, error) {
	return s.completeWithConfig(nil, system, user)
}

func (s *AIService) completeWithConfig(userCfg *models.AIConfig, system, user string) (string, error) {
	provider := s.resolvedProvider(userCfg)
	switch strings.ToLower(provider) {
	case "anthropic":
		return s.anthropicCompleteWithConfig(userCfg, system, user)
	default:
		return s.openAICompleteWithConfig(userCfg, system, user)
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

func (s *AIService) openAICompleteWithConfig(userCfg *models.AIConfig, system, user string) (string, error) {
	endpoint := s.resolveEndpointWithConfig(userCfg)
	model := s.resolvedModel(userCfg)
	apiKey := s.resolvedAPIKey(userCfg)

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
	req.Header.Set("Authorization", "Bearer "+apiKey)
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

func (s *AIService) resolveEndpointWithConfig(userCfg *models.AIConfig) string {
	baseURL := s.resolvedBaseURL(userCfg)
	if baseURL != "" {
		return strings.TrimRight(baseURL, "/")
	}
	provider := s.resolvedProvider(userCfg)
	switch strings.ToLower(provider) {
	case "openrouter":
		return "https://openrouter.ai/api/v1"
	case "gemini":
		return "https://generativelanguage.googleapis.com/v1beta/openai"
	case "ollama":
		return "http://localhost:11434/v1"
	default:
		return "https://api.openai.com/v1"
	}
}

// Keep legacy wrappers for backward compat
func (s *AIService) openAIComplete(system, user string) (string, error) {
	return s.openAICompleteWithConfig(nil, system, user)
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

func (s *AIService) anthropicCompleteWithConfig(userCfg *models.AIConfig, system, user string) (string, error) {
	model := s.resolvedModel(userCfg)
	if model == "" {
		model = "claude-3-haiku-20240307"
	}
	apiKey := s.resolvedAPIKey(userCfg)

	body := anthropicRequest{
		Model:     model,
		MaxTokens: 1024,
		System:    system,
		Messages:  []anthropicMessage{{Role: "user", Content: user}},
	}
	data, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
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
