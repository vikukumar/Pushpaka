package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/vikukumar/Pushpaka/internal/repositories"
	"github.com/vikukumar/Pushpaka/pkg/basemodel"
	"github.com/vikukumar/Pushpaka/pkg/models"
)

var (
	ErrInvalidZonePAT = errors.New("invalid zone PAT")
	ErrWorkerNotFound = errors.New("worker not found")
)

type WorkerNodeService struct {
	repo       *repositories.WorkerNodeRepository
	systemRepo *repositories.SystemConfigRepository
}

func NewWorkerNodeService(
	repo *repositories.WorkerNodeRepository,
	systemRepo *repositories.SystemConfigRepository,
) *WorkerNodeService {
	return &WorkerNodeService{
		repo:       repo,
		systemRepo: systemRepo,
	}
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// RegisterWorker authenticates a worker using the ZONE_PAT and registers taking their specs
func (s *WorkerNodeService) RegisterWorker(req *models.RegisterWorkerRequest) (*models.WorkerAuthResponse, error) {
	zonePAT, err := s.systemRepo.Get("ZONE_PAT")
	if err != nil || zonePAT == "" {
		// If Zone PAT not generated, fail closed
		return nil, ErrInvalidZonePAT
	}
	if req.ZonePAT != zonePAT {
		return nil, ErrInvalidZonePAT
	}

	workerID := uuid.New().String()
	authToken := generateToken()
	now := time.Now().UTC()

	worker := &models.WorkerNode{
		BaseModel: basemodel.BaseModel{
			ID:        workerID,
			CreatedAt: now,
			UpdatedAt: now,
		},
		Name:          req.Name,
		Type:          req.Type,
		Status:        models.WorkerStatusActive,
		IPAddress:     req.IPAddress,
		OS:            req.OS,
		Architecture:  req.Architecture,
		GoVersion:     req.GoVersion,
		DockerVersion: req.DockerVersion,
		NodeVersion:   req.NodeVersion,
		MemoryTotal:   req.MemoryTotal,
		CPUCount:      req.CPUCount,
		AuthToken:     authToken,
		LastSeenAt:    &now,
	}

	if err := s.repo.Create(worker); err != nil {
		return nil, err
	}

	return &models.WorkerAuthResponse{
		WorkerID:  workerID,
		AuthToken: authToken,
		ExpiresIn: 86400, // 24 hours standard TTL before mandatory rotation
	}, nil
}

// Authenticate validates a worker token and returns the corresponding node
func (s *WorkerNodeService) Authenticate(token string) (*models.WorkerNode, error) {
	if token == "" {
		return nil, ErrInvalidZonePAT
	}
	worker, err := s.repo.FindByAuthToken(token)
	if err != nil {
		return nil, ErrWorkerNotFound
	}
	// Update last seen heartbeat
	_ = s.repo.UpdateLastSeen(worker.ID)
	return worker, nil
}

// RotateToken generates a new auth token for a worker
func (s *WorkerNodeService) RotateToken(workerID string) (*models.WorkerAuthResponse, error) {
	worker, err := s.repo.FindByID(workerID)
	if err != nil {
		return nil, err
	}

	newToken := generateToken()
	worker.AuthToken = newToken
	worker.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(worker); err != nil {
		return nil, err
	}

	return &models.WorkerAuthResponse{
		WorkerID:  workerID,
		AuthToken: newToken,
		ExpiresIn: 86400,
	}, nil
}

// ListWorkers returns a list of all registered worker nodes
func (s *WorkerNodeService) ListWorkers() ([]models.WorkerNode, error) {
	return s.repo.ListAll()
}

// GetZonePAT retrieves the global installation Zone PAT
func (s *WorkerNodeService) GetZonePAT() (string, error) {
	pat, err := s.systemRepo.Get("ZONE_PAT")
	if err != nil {
		return "", err
	}
	// Note: We might want to obscure parts of this in the future if we build a robust
	// rotate PAT feature. For now, the dashboard config needs to display it for copying.
	return pat, nil
}
