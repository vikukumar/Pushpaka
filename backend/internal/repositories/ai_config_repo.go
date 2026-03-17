package repositories

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/vikukumar/Pushpaka/internal/models"
)

// AIConfigRepository manages AI configuration, RAG documents, and monitoring alerts.
type AIConfigRepository struct {
	db *sqlx.DB
}

func NewAIConfigRepository(db *sqlx.DB) *AIConfigRepository {
	return &AIConfigRepository{db: db}
}

// ─── AI Config ────────────────────────────────────────────────────────────────

func (r *AIConfigRepository) GetByUserID(userID string) (*models.AIConfig, error) {
	var cfg models.AIConfig
	err := r.db.Get(&cfg, `SELECT * FROM ai_configs WHERE user_id = ?`, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &cfg, nil
}

func (r *AIConfigRepository) Upsert(cfg *models.AIConfig) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if cfg.ID == "" {
		cfg.ID = uuid.New().String()
		cfg.CreatedAt = models.Time{Time: time.Now().UTC()}
	}
	cfg.UpdatedAt = models.Time{Time: time.Now().UTC()}
	_, err := r.db.Exec(`
		INSERT INTO ai_configs
			(id, user_id, provider, api_key, model, base_url, system_prompt,
			 monitoring_enabled, monitoring_interval, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(user_id) DO UPDATE SET
			provider=excluded.provider,
			api_key=CASE WHEN excluded.api_key='' THEN ai_configs.api_key ELSE excluded.api_key END,
			model=excluded.model,
			base_url=excluded.base_url,
			system_prompt=excluded.system_prompt,
			monitoring_enabled=excluded.monitoring_enabled,
			monitoring_interval=excluded.monitoring_interval,
			updated_at=excluded.updated_at
	`, cfg.ID, cfg.UserID, cfg.Provider, cfg.APIKey, cfg.Model, cfg.BaseURL,
		cfg.SystemPrompt, cfg.MonitoringEnabled, cfg.MonitoringInterval, now, now)
	return err
}

// ─── RAG Documents ────────────────────────────────────────────────────────────

func (r *AIConfigRepository) ListRAG(userID string) ([]models.RAGDocument, error) {
	var docs []models.RAGDocument
	err := r.db.Select(&docs, `SELECT * FROM rag_documents WHERE user_id = ? ORDER BY created_at DESC`, userID)
	return docs, err
}

func (r *AIConfigRepository) CreateRAG(doc *models.RAGDocument) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if doc.ID == "" {
		doc.ID = uuid.New().String()
	}
	_, err := r.db.Exec(`
		INSERT INTO rag_documents (id, user_id, title, content, created_at, updated_at)
		VALUES (?,?,?,?,?,?)
	`, doc.ID, doc.UserID, doc.Title, doc.Content, now, now)
	return err
}

func (r *AIConfigRepository) DeleteRAG(id, userID string) error {
	res, err := r.db.Exec(`DELETE FROM rag_documents WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("document not found")
	}
	return nil
}

// ─── AI Monitor Alerts ────────────────────────────────────────────────────────

func (r *AIConfigRepository) CreateAlert(alert *models.AIMonitorAlert) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if alert.ID == "" {
		alert.ID = uuid.New().String()
	}
	_, err := r.db.Exec(`
		INSERT INTO ai_monitor_alerts
			(id, user_id, deployment_id, severity, title, message, resolved, created_at)
		VALUES (?,?,?,?,?,?,?,?)
	`, alert.ID, alert.UserID, alert.DeploymentID, alert.Severity,
		alert.Title, alert.Message, alert.Resolved, now)
	return err
}

func (r *AIConfigRepository) ListAlerts(userID string, limit int, onlyUnresolved bool) ([]models.AIMonitorAlert, error) {
	var alerts []models.AIMonitorAlert
	if onlyUnresolved {
		err := r.db.Select(&alerts,
			`SELECT * FROM ai_monitor_alerts WHERE user_id=? AND resolved=0 ORDER BY created_at DESC LIMIT ?`,
			userID, limit)
		return alerts, err
	}
	err := r.db.Select(&alerts,
		`SELECT * FROM ai_monitor_alerts WHERE user_id=? ORDER BY created_at DESC LIMIT ?`,
		userID, limit)
	return alerts, err
}

func (r *AIConfigRepository) ResolveAlert(id, userID string) error {
	res, err := r.db.Exec(`UPDATE ai_monitor_alerts SET resolved=1 WHERE id=? AND user_id=?`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("alert not found")
	}
	return nil
}

// AlertExistsForDeployment returns true if there is already an unresolved alert
// for this deployment (avoids duplicate alerts during monitoring sweeps).
func (r *AIConfigRepository) AlertExistsForDeployment(deploymentID string) (bool, error) {
	var count int
	err := r.db.Get(&count,
		`SELECT COUNT(*) FROM ai_monitor_alerts WHERE deployment_id=? AND resolved=0`,
		deploymentID)
	return count > 0, err
}

// ListUsersWithMonitoring returns user IDs of users who have AI monitoring enabled.
func (r *AIConfigRepository) ListUsersWithMonitoring() ([]string, error) {
	var ids []string
	err := r.db.Select(&ids,
		`SELECT user_id FROM ai_configs WHERE monitoring_enabled=1`)
	return ids, err
}

// ─── AI Token Usage / Rate Limiting ──────────────────────────────────────────

// GetOrCreateTodayUsage returns (or creates) the token_usage row for userID for today (UTC).
func (r *AIConfigRepository) GetOrCreateTodayUsage(userID string) (*models.AITokenUsage, error) {
	today := time.Now().UTC().Format("2006-01-02")
	var usage models.AITokenUsage
	err := r.db.Get(&usage, `SELECT * FROM ai_token_usage WHERE user_id=? AND date=?`, userID, today)
	if err == nil {
		return &usage, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	// Create new row
	now := time.Now().UTC().Format(time.RFC3339)
	id := uuid.New().String()
	_, err = r.db.Exec(
		`INSERT OR IGNORE INTO ai_token_usage (id,user_id,date,calls,tokens,updated_at) VALUES (?,?,?,0,0,?)`,
		id, userID, today, now)
	if err != nil {
		return nil, err
	}
	// Re-fetch (handles race where another goroutine inserted first)
	err = r.db.Get(&usage, `SELECT * FROM ai_token_usage WHERE user_id=? AND date=?`, userID, today)
	return &usage, err
}

// IncrementTodayUsage adds deltaCalls and deltaTokens to the user's usage row for today.
func (r *AIConfigRepository) IncrementTodayUsage(userID string, deltaCalls int) error {
	today := time.Now().UTC().Format("2006-01-02")
	now := time.Now().UTC().Format(time.RFC3339)
	id := uuid.New().String()
	_, err := r.db.Exec(
		`INSERT INTO ai_token_usage (id,user_id,date,calls,tokens,updated_at)
		 VALUES (?,?,?,?,0,?)
		 ON CONFLICT(user_id,date) DO UPDATE SET calls=calls+?, updated_at=?`,
		id, userID, today, deltaCalls, now, deltaCalls, now)
	return err
}
