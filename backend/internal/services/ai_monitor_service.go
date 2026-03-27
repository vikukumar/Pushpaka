package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/vikukumar/pushpaka/internal/repositories"
	"github.com/vikukumar/pushpaka/pkg/models"
)

type AIMonitorService struct {
	aiSvc        *AIService
	aiRepo       *repositories.AIConfigRepository
	deployRepo   *repositories.DeploymentRepository
	logRepo      *repositories.LogRepository
	pollInterval time.Duration
}

func NewAIMonitorService(
	aiSvc *AIService,
	aiRepo *repositories.AIConfigRepository,
	deployRepo *repositories.DeploymentRepository,
	logRepo *repositories.LogRepository,
) *AIMonitorService {
	return &AIMonitorService{
		aiSvc:        aiSvc,
		aiRepo:       aiRepo,
		deployRepo:   deployRepo,
		logRepo:      logRepo,
		pollInterval: 5 * time.Minute,
	}
}

func (s *AIMonitorService) Start(ctx context.Context) {
	log.Info().Msg("AI Monitoring Service started")
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("AI Monitoring Service stopping")
			return
		case <-ticker.C:
			s.runCheck(ctx)
		}
	}
}

func (s *AIMonitorService) runCheck(ctx context.Context) {
	userIDs, err := s.aiRepo.ListUsersWithMonitoring()
	if err != nil {
		log.Error().Err(err).Msg("failed to list users for AI monitoring")
		return
	}

	for _, userID := range userIDs {
		s.checkUserDeployments(ctx, userID)
	}
}

func (s *AIMonitorService) checkUserDeployments(ctx context.Context, userID string) {
	// Load user config
	userCfg, err := s.aiRepo.GetByUserID(userID)
	if err != nil || userCfg == nil || !userCfg.MonitoringEnabled {
		return
	}

	// Fetch recent failed or unhealthy deployments for this user
	deployments, err := s.deployRepo.ListFailedRecent(userID, 10)
	if err != nil {
		return
	}

	for _, d := range deployments {
		// Skip if alert already exists and is unresolved
		exists, _ := s.aiRepo.AlertExistsForDeployment(d.ID)
		if exists {
			continue
		}

		log.Info().Str("deployment_id", d.ID).Str("user_id", userID).Msg("monitoring: analyzing failed deployment")

		// Get logs
		logEntries, err := s.logRepo.FindByDeploymentID(d.ID)
		if err != nil || len(logEntries) == 0 {
			continue
		}

		var sb strings.Builder
		for _, l := range logEntries {
			sb.WriteString(l.Message + "\n")
		}

		// Load RAG
		ragDocs, _ := s.aiRepo.ListRAG(userID)

		// Analyze
		analysis, err := s.aiSvc.AnalyzeLogsWithConfig(userCfg, ragDocs, sb.String())
		if err != nil {
			log.Warn().Err(err).Str("deployment_id", d.ID).Msg("monitoring: AI analysis failed")
			continue
		}

		// Create Alert
		alert := &models.AIMonitorAlert{
			UserID:       userID,
			DeploymentID: d.ID,
			Severity:     "error",
			Title:        fmt.Sprintf("Deployment failure in branch %s", d.Branch),
			Message:      analysis,
			Resolved:     false,
		}

		if err := s.aiRepo.CreateAlert(alert); err != nil {
			log.Error().Err(err).Msg("failed to create AI monitor alert")
		} else {
			log.Info().Str("alert_id", alert.ID).Msg("AI monitor alert created")
		}
	}
}
