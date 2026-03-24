package repositories

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/vikukumar/Pushpaka/pkg/basemodel"
	"github.com/vikukumar/Pushpaka/pkg/models"
)

var ErrWorkerNotFound = errors.New("worker node not found")

type WorkerNodeRepository struct {
	db *gorm.DB
}

func NewWorkerNodeRepository(db *gorm.DB) *WorkerNodeRepository {
	return &WorkerNodeRepository{db: db}
}

func (r *WorkerNodeRepository) Create(worker *models.WorkerNode) error {
	return basemodel.Add(r.db, worker)
}

func (r *WorkerNodeRepository) Update(worker *models.WorkerNode) error {
	return basemodel.Modify(r.db, worker)
}

func (r *WorkerNodeRepository) FindByID(id string) (*models.WorkerNode, error) {
	worker, err := basemodel.Get[models.WorkerNode](r.db, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkerNotFound
		}
		return nil, err
	}
	return worker, nil
}

func (r *WorkerNodeRepository) FindByAuthToken(token string) (*models.WorkerNode, error) {
	worker, err := basemodel.First[models.WorkerNode](r.db, "auth_token = ?", token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkerNotFound
		}
		return nil, err
	}
	return worker, nil
}

func (r *WorkerNodeRepository) ListAll() ([]models.WorkerNode, error) {
	return basemodel.Query[models.WorkerNode](r.db, "1=1")
}

func (r *WorkerNodeRepository) Delete(id string) error {
	return basemodel.Delete[models.WorkerNode](r.db, id)
}

func (r *WorkerNodeRepository) UpdateLastSeen(id string) error {
	now := time.Now().UTC()
	return basemodel.Update[models.WorkerNode](r.db, id, map[string]interface{}{
		"last_seen_at": now,
		"updated_at":   now,
		"status":       models.WorkerStatusActive,
	})
}

func (r *WorkerNodeRepository) MarkStaleWorkersOffline(timeout time.Duration) (int64, error) {
	basemodel.EnsureSynced[models.WorkerNode](r.db)
	threshold := time.Now().UTC().Add(-timeout)
	result := r.db.Model(&models.WorkerNode{}).
		Where("last_seen_at < ? AND status = ?", threshold, models.WorkerStatusActive).
		Updates(map[string]interface{}{
			"status":     models.WorkerStatusOffline,
			"updated_at": time.Now().UTC(),
		})
	return result.RowsAffected, result.Error
}
