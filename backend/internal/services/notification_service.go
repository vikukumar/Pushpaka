package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/vikukumar/Pushpaka/internal/config"
	"github.com/vikukumar/Pushpaka/internal/models"
	"github.com/vikukumar/Pushpaka/internal/repositories"
)

type NotificationService struct {
	repo *repositories.NotificationRepository
	cfg  *config.Config
}

func NewNotificationService(repo *repositories.NotificationRepository, cfg *config.Config) *NotificationService {
	return &NotificationService{repo: repo, cfg: cfg}
}

func (s *NotificationService) GetConfig(userID string) (*models.NotificationConfig, error) {
	cfg, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		// Return a default (empty) config rather than nil
		now := models.Time{Time: time.Now().UTC()}
		cfg = &models.NotificationConfig{
			UserID:          userID,
			SMTPPort:        587,
			NotifyOnSuccess: true,
			NotifyOnFailure: true,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
	}
	return cfg, nil
}

func (s *NotificationService) UpsertConfig(userID string, req *models.UpsertNotificationConfigRequest) (*models.NotificationConfig, error) {
	existing, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	now := models.Time{Time: time.Now().UTC()}
	var cfg models.NotificationConfig

	if existing != nil {
		cfg = *existing
		cfg.UpdatedAt = now
	} else {
		cfg = models.NotificationConfig{
			ID:              uuid.New().String(),
			UserID:          userID,
			NotifyOnSuccess: true,
			NotifyOnFailure: true,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
	}

	cfg.SlackWebhookURL = req.SlackWebhookURL
	cfg.DiscordWebhookURL = req.DiscordWebhookURL
	cfg.SMTPHost = req.SMTPHost
	if req.SMTPPort > 0 {
		cfg.SMTPPort = req.SMTPPort
	} else if cfg.SMTPPort == 0 {
		cfg.SMTPPort = 587
	}
	cfg.SMTPUsername = req.SMTPUsername
	// Only update password if a new one is provided (non-empty)
	if req.SMTPPassword != "" {
		cfg.SMTPPassword = req.SMTPPassword
	}
	cfg.SMTPFrom = req.SMTPFrom
	cfg.SMTPTo = req.SMTPTo
	if req.NotifyOnSuccess != nil {
		cfg.NotifyOnSuccess = *req.NotifyOnSuccess
	}
	if req.NotifyOnFailure != nil {
		cfg.NotifyOnFailure = *req.NotifyOnFailure
	}

	if err := s.repo.Upsert(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Dispatch sends notifications for the given deployment event to all enabled
// channels configured for the deployment owner. It runs asynchronously.
func (s *NotificationService) Dispatch(userID string, event *models.NotificationEvent) {
	go func() {
		cfg, err := s.repo.FindByUserID(userID)
		if err != nil || cfg == nil {
			return
		}

		isSuccess := event.Status == string(models.DeploymentRunning)
		if isSuccess && !cfg.NotifyOnSuccess {
			return
		}
		if !isSuccess && !cfg.NotifyOnFailure {
			return
		}

		title := fmt.Sprintf("Pushpaka: %s deployed %s", event.ProjectName, statusEmoji(event.Status))
		body := formatMessage(event)

		if cfg.SlackWebhookURL != "" {
			_ = sendSlack(cfg.SlackWebhookURL, title, body, event)
		}
		if cfg.DiscordWebhookURL != "" {
			_ = sendDiscord(cfg.DiscordWebhookURL, title, body, event)
		}
		if cfg.SMTPHost != "" && cfg.SMTPTo != "" {
			_ = sendEmail(cfg, title, body)
		}
	}()
}

// DispatchInternal is called by the internal notification callback endpoint
// from the worker. It looks up user config and fires all channels.
func (s *NotificationService) DispatchInternal(event *models.NotificationEvent, userID string) {
	s.Dispatch(userID, event)
}

func statusEmoji(status string) string {
	if status == string(models.DeploymentRunning) {
		return "succeeded"
	}
	return "failed"
}

func formatMessage(e *models.NotificationEvent) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Project: %s\n", e.ProjectName))
	sb.WriteString(fmt.Sprintf("Status:  %s\n", e.Status))
	sb.WriteString(fmt.Sprintf("Branch:  %s\n", e.Branch))
	if e.CommitSHA != "" {
		sha := e.CommitSHA
		if len(sha) > 8 {
			sha = sha[:8]
		}
		sb.WriteString(fmt.Sprintf("Commit:  %s\n", sha))
	}
	if e.URL != "" {
		sb.WriteString(fmt.Sprintf("URL:     %s\n", e.URL))
	}
	if e.ErrorMsg != "" {
		sb.WriteString(fmt.Sprintf("Error:   %s\n", e.ErrorMsg))
	}
	return sb.String()
}

func sendSlack(webhookURL, title, body string, e *models.NotificationEvent) error {
	color := "#36a64f" // green
	if e.Status != string(models.DeploymentRunning) {
		color = "#d9534f" // red
	}
	payload := map[string]any{
		"attachments": []map[string]any{
			{
				"color":  color,
				"title":  title,
				"text":   body,
				"ts":     time.Now().Unix(),
				"footer": "Pushpaka",
			},
		},
	}
	return postJSON(webhookURL, payload)
}

func sendDiscord(webhookURL, title, body string, e *models.NotificationEvent) error {
	color := 3066993 // green
	if e.Status != string(models.DeploymentRunning) {
		color = 15158332 // red
	}
	payload := map[string]any{
		"embeds": []map[string]any{
			{
				"title":       title,
				"description": body,
				"color":       color,
				"timestamp":   time.Now().UTC().Format(time.RFC3339),
			},
		},
	}
	return postJSON(webhookURL, payload)
}

func postJSON(url string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook responded with %d", resp.StatusCode)
	}
	return nil
}

func sendEmail(cfg *models.NotificationConfig, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)
	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		cfg.SMTPFrom, cfg.SMTPTo, subject, body))

	var auth smtp.Auth
	if cfg.SMTPUsername != "" {
		auth = smtp.PlainAuth("", cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPHost)
	}

	return smtp.SendMail(addr, auth, cfg.SMTPFrom, []string{cfg.SMTPTo}, msg)
}
