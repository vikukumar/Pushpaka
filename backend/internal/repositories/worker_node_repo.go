package repositories

import (
	"errors"
	"time"

	"gorm.io/gorm"

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
	return r.db.Create(worker).Error
}

func (r *WorkerNodeRepository) Update(worker *models.WorkerNode) error {
	return r.db.Save(worker).Error
}

func (r *WorkerNodeRepository) FindByID(id string) (*models.WorkerNode, error) {
	var worker models.WorkerNode
	err := r.db.First(&worker, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkerNotFound
		}
		return nil, err
	}
	return &worker, nil
}

func (r *WorkerNodeRepository) FindByAuthToken(token string) (*models.WorkerNode, error) {
	var worker models.WorkerNode
	err := r.db.First(&worker, "auth_token = ?", token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkerNotFound
		}
		return nil, err
	}
	return &worker, nil
}

func (r *WorkerNodeRepository) ListAll() ([]models.WorkerNode, error) {
	var workers []models.WorkerNode
	err := r.db.Order("created_at desc").Find(&workers).Error
	return workers, err
}

func (r *WorkerNodeRepository) Delete(id string) error {
	return r.db.Delete(&models.WorkerNode{}, "id = ?", id).Error
}

func (r *WorkerNodeRepository) UpdateLastSeen(id string) error {
	now := time.Now().UTC()
	return r.db.Model(&models.WorkerNode{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_seen_at": now,
			"updated_at":   now,
			"status":       models.WorkerStatusActive,
		}).Error
}

func (r *WorkerNodeRepository) MarkStaleWorkersOffline(timeout time.Duration) (int64, error) {
	threshold := time.Now().UTC().Add(-timeout)
	result := r.db.Model(&models.WorkerNode{}).
		Where("last_seen_at < ? AND status = ?", threshold, models.WorkerStatusActive).
		Updates(map[string]interface{}{
			"status":     models.WorkerStatusOffline,
			"updated_at": time.Now().UTC(),
		})
	return result.RowsAffected, result.Error
}
