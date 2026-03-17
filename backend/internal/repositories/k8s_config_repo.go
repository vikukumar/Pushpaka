package repositories

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/vikukumar/Pushpaka/internal/models"
)

// K8sConfigRepository manages per-user Kubernetes cluster configurations.
type K8sConfigRepository struct {
	db *sqlx.DB
}

func NewK8sConfigRepository(db *sqlx.DB) *K8sConfigRepository {
	return &K8sConfigRepository{db: db}
}

func (r *K8sConfigRepository) GetByUserID(userID string) (*models.K8sConfig, error) {
	var cfg models.K8sConfig
	err := r.db.Get(&cfg, `SELECT * FROM k8s_configs WHERE user_id = ?`, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &cfg, nil
}

func (r *K8sConfigRepository) Upsert(cfg *models.K8sConfig) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if cfg.ID == "" {
		cfg.ID = uuid.New().String()
		cfg.CreatedAt = models.Time{Time: time.Now().UTC()}
	}
	_, err := r.db.Exec(`
		INSERT INTO k8s_configs
			(id, user_id, name, server_url, token, namespace, kubeconfig, registry_url, enabled, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(user_id) DO UPDATE SET
			name=excluded.name,
			server_url=excluded.server_url,
			token=CASE WHEN excluded.token='' THEN k8s_configs.token ELSE excluded.token END,
			namespace=excluded.namespace,
			kubeconfig=CASE WHEN excluded.kubeconfig='' THEN k8s_configs.kubeconfig ELSE excluded.kubeconfig END,
			registry_url=excluded.registry_url,
			enabled=excluded.enabled,
			updated_at=excluded.updated_at
	`, cfg.ID, cfg.UserID, cfg.Name, cfg.ServerURL, cfg.Token, cfg.Namespace,
		cfg.Kubeconfig, cfg.RegistryURL, cfg.Enabled, now, now)
	return err
}

func (r *K8sConfigRepository) GetByUserIDInternal(userID string) (*models.K8sConfig, error) {
	return r.GetByUserID(userID)
}

// GetEnabledByUserID returns the K8s config only if it is enabled.
func (r *K8sConfigRepository) GetEnabledByUserID(userID string) (*models.K8sConfig, error) {
	var cfg models.K8sConfig
	err := r.db.Get(&cfg, `SELECT * FROM k8s_configs WHERE user_id = ? AND enabled = 1`, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &cfg, nil
}
