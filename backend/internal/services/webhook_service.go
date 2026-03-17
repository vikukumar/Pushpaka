package services

import (
"crypto/hmac"
"crypto/rand"
"crypto/sha256"
"encoding/hex"
"encoding/json"
"errors"
"fmt"
"strings"

"github.com/google/uuid"

"github.com/vikukumar/Pushpaka/internal/config"
"github.com/vikukumar/Pushpaka/internal/models"
"github.com/vikukumar/Pushpaka/internal/repositories"
)

var ErrWebhookNotFound = errors.New("webhook not found")
var ErrWebhookSignatureInvalid = errors.New("webhook signature invalid")

type WebhookService struct {
webhookRepo   *repositories.WebhookRepository
projectRepo   *repositories.ProjectRepository
deploymentSvc *DeploymentService
cfg           *config.Config
}

func NewWebhookService(
webhookRepo *repositories.WebhookRepository,
projectRepo *repositories.ProjectRepository,
deploymentSvc *DeploymentService,
cfg *config.Config,
) *WebhookService {
return &WebhookService{
webhookRepo:   webhookRepo,
projectRepo:   projectRepo,
deploymentSvc: deploymentSvc,
cfg:           cfg,
}
}

func (s *WebhookService) Create(userID string, req *models.CreateWebhookRequest) (*models.WebhookConfigResponse, error) {
if _, err := s.projectRepo.FindByID(req.ProjectID, userID); err != nil {
return nil, ErrProjectNotFound
}
provider := req.Provider
if provider == "" {
provider = "github"
}
b := make([]byte, 20)
rand.Read(b) //nolint:errcheck
secret := hex.EncodeToString(b)
now := models.NowUTC()
wh := &models.WebhookConfig{
ID:        uuid.New().String(),
ProjectID: req.ProjectID,
UserID:    userID,
Secret:    secret,
Provider:  provider,
Branch:    req.Branch,
CreatedAt: now,
UpdatedAt: now,
}
if err := s.webhookRepo.Create(wh); err != nil {
return nil, fmt.Errorf("creating webhook: %w", err)
}
return &models.WebhookConfigResponse{
ID:         wh.ID,
ProjectID:  wh.ProjectID,
Provider:   wh.Provider,
Branch:     wh.Branch,
WebhookURL: s.cfg.BaseURL + "/api/v1/webhooks/" + wh.ID + "/receive",
CreatedAt:  wh.CreatedAt,
}, nil
}

func (s *WebhookService) List(userID string) ([]models.WebhookConfigResponse, error) {
webhooks, err := s.webhookRepo.ListByUserID(userID)
if err != nil {
return nil, err
}
result := make([]models.WebhookConfigResponse, len(webhooks))
for i, wh := range webhooks {
result[i] = models.WebhookConfigResponse{
ID:         wh.ID,
ProjectID:  wh.ProjectID,
Provider:   wh.Provider,
Branch:     wh.Branch,
WebhookURL: s.cfg.BaseURL + "/api/v1/webhooks/" + wh.ID + "/receive",
CreatedAt:  wh.CreatedAt,
}
}
return result, nil
}

func (s *WebhookService) Delete(id, userID string) error {
return s.webhookRepo.Delete(id, userID)
}

func (s *WebhookService) Receive(webhookID string, payload []byte, sigHeader string) error {
wh, err := s.webhookRepo.FindByID(webhookID)
if err != nil {
return ErrWebhookNotFound
}
if !verifyHMAC(payload, sigHeader, wh.Secret, wh.Provider) {
return ErrWebhookSignatureInvalid
}
branch := parsePushBranch(payload)
if branch == "" {
return nil
}
if wh.Branch != "" && wh.Branch != branch {
return nil
}
req := &models.DeployRequest{
ProjectID: wh.ProjectID,
Branch:    branch,
}
_, err = s.deploymentSvc.Trigger(wh.UserID, req)
return err
}

func verifyHMAC(payload []byte, sig, secret, provider string) bool {
mac := hmac.New(sha256.New, []byte(secret))
mac.Write(payload)
expectedHex := hex.EncodeToString(mac.Sum(nil))
switch provider {
case "gitlab":
if sig == secret {
return true
}
return hmac.Equal([]byte(sig), []byte("sha256="+expectedHex)) ||
hmac.Equal([]byte(sig), []byte(expectedHex))
default:
expected := "sha256=" + expectedHex
return hmac.Equal([]byte(sig), []byte(expected))
}
}

func parsePushBranch(payload []byte) string {
var p struct {
Ref string `json:"ref"`
}
if err := json.Unmarshal(payload, &p); err != nil || p.Ref == "" {
return ""
}
return strings.TrimPrefix(p.Ref, "refs/heads/")
}
