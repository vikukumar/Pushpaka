package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/vikukumar/Pushpaka/internal/repositories"
	"github.com/vikukumar/Pushpaka/pkg/basemodel"
	"github.com/vikukumar/Pushpaka/pkg/models"
)

var ErrDomainExists = errors.New("domain already registered")

type DomainService struct {
	domainRepo  *repositories.DomainRepository
	projectRepo *repositories.ProjectRepository
}

func NewDomainService(
	domainRepo *repositories.DomainRepository,
	projectRepo *repositories.ProjectRepository,
) *DomainService {
	return &DomainService{
		domainRepo:  domainRepo,
		projectRepo: projectRepo,
	}
}

func (s *DomainService) Add(userID string, req *models.AddDomainRequest) (*models.Domain, error) {
	// Verify project ownership
	if _, err := s.projectRepo.FindByID(req.ProjectID, userID); err != nil {
		return nil, ErrProjectNotFound
	}

	existing, _ := s.domainRepo.FindByDomain(req.Domain)
	if existing != nil {
		return nil, ErrDomainExists
	}

	// Just create it (validation/DNS check happens later, or external to the backend wrapper)
	now := time.Now().UTC() // Changed from models.NowUTC()
	d := &models.Domain{
		BaseModel: basemodel.BaseModel{ // Added BaseModel
			ID:        uuid.New().String(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		ProjectID:  req.ProjectID,
		UserID:     userID,
		Domain:     req.Domain, // Changed from req.Domain to domainName, assuming domainName is a placeholder for req.Domain or will be defined. Reverting to req.Domain as domainName is undefined.
		Verified:   false,
		SSLEnabled: false, // Added SSLEnabled
	}

	if err := s.domainRepo.Create(d); err != nil {
		return nil, fmt.Errorf("creating domain: %w", err)
	}
	return d, nil
}

func (s *DomainService) ListByProject(projectID, userID string) ([]models.Domain, error) {
	if _, err := s.projectRepo.FindByID(projectID, userID); err != nil {
		return nil, ErrProjectNotFound
	}
	return s.domainRepo.FindByProjectID(projectID)
}

func (s *DomainService) List(userID string) ([]models.Domain, error) {
	return s.domainRepo.FindByUserID(userID)
}

func (s *DomainService) Delete(id, userID string) error {
	return s.domainRepo.Delete(id, userID)
}
