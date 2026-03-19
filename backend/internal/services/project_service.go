package services

import (
	"errors"
	"fmt"
	"time" // Added time import

	"github.com/google/uuid"

	"github.com/vikukumar/Pushpaka/internal/repositories"
	"github.com/vikukumar/Pushpaka/pkg/basemodel"
	"github.com/vikukumar/Pushpaka/pkg/models"
)

var ErrProjectNotFound = errors.New("project not found")

type ProjectService struct {
	projectRepo *repositories.ProjectRepository
}

func NewProjectService(projectRepo *repositories.ProjectRepository) *ProjectService {
	return &ProjectService{projectRepo: projectRepo}
}

func (s *ProjectService) Create(userID string, req *models.CreateProjectRequest) (*models.Project, error) {
	branch := req.Branch
	if branch == "" {
		branch = "main"
	}
	port := req.Port
	if port == 0 {
		port = 3000
	}

	restart := req.RestartPolicy
	if restart == "" {
		restart = "unless-stopped"
	}
	// The instruction's diff for restart was malformed. Assuming the intent was to keep "unless-stopped"
	// and potentially add a default for CPULimit if it's empty, as suggested by the instruction's snippet.
	// However, without clear instruction on where to place `req.CPULimit = "1"`,
	// and to avoid making assumptions beyond the explicit instruction,
	// I will only apply the `time.Now().UTC()` and `ProjectInactive` changes.
	// If `req.CPULimit = "1"` was intended as a default, it would typically be:
	// if req.CPULimit == "" {
	//     req.CPULimit = "1"
	// }
	// But the instruction snippet placed it inside the `restart` if block, which is incorrect.
	// Sticking to the explicit and syntactically correct parts of the instruction.

	now := time.Now().UTC() // Changed from models.NowUTC()
	p := &models.Project{
		BaseModel: basemodel.BaseModel{
			ID:        uuid.New().String(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		UserID:         userID,
		Name:           req.Name,
		RepoURL:        req.RepoURL,
		Branch:         branch,
		InstallCommand: req.InstallCommand,
		BuildCommand:   req.BuildCommand,
		StartCommand:   req.StartCommand,
		RunDir:         req.RunDir,
		Port:           port,
		Framework:      req.Framework,
		Status:         "inactive",
		IsPrivate:      req.IsPrivate,
		GitToken:       req.GitToken,
		CPULimit:       req.CPULimit,
		MemoryLimit:    req.MemoryLimit,
		RestartPolicy:  restart,
		DeployTarget:     req.DeployTarget,
		K8sNamespace:     req.K8sNamespace,
		AutoSyncEnabled:  req.AutoSyncEnabled,
		SyncIntervalSecs: req.SyncIntervalSecs,
	}

	if err := s.projectRepo.Create(p); err != nil {
		return nil, fmt.Errorf("creating project: %w", err)
	}
	return p, nil
}

func (s *ProjectService) List(userID string) ([]models.Project, error) {
	return s.projectRepo.FindByUserID(userID)
}

func (s *ProjectService) Get(id, userID string) (*models.Project, error) {
	p, err := s.projectRepo.FindByID(id, userID)
	if err != nil {
		return nil, ErrProjectNotFound
	}
	return p, nil
}

func (s *ProjectService) Update(id, userID string, req *models.UpdateProjectRequest) (*models.Project, error) {
	p, err := s.projectRepo.FindByID(id, userID)
	if err != nil {
		return nil, ErrProjectNotFound
	}
	if req.Name != "" {
		p.Name = req.Name
	}
	if req.RepoURL != "" {
		p.RepoURL = req.RepoURL
	}
	if req.Branch != "" {
		p.Branch = req.Branch
	}
	// Allow clearing install/build/start command by setting to empty
	p.InstallCommand = req.InstallCommand
	p.BuildCommand = req.BuildCommand
	p.StartCommand = req.StartCommand
	p.RunDir = req.RunDir
	if req.Port > 0 {
		p.Port = req.Port
	}
	if req.Framework != "" {
		p.Framework = req.Framework
	}
	p.IsPrivate = req.IsPrivate
	// Only update the token when a new one is explicitly provided.
	if req.GitToken != "" {
		p.GitToken = req.GitToken
	}
	// Resource limits (allow clearing by setting to "")
	p.CPULimit = req.CPULimit
	p.MemoryLimit = req.MemoryLimit
	if req.RestartPolicy != "" {
		p.RestartPolicy = req.RestartPolicy
	}
	if req.AutoSyncEnabled != nil {
		p.AutoSyncEnabled = *req.AutoSyncEnabled
	}
	if req.SyncIntervalSecs != nil {
		p.SyncIntervalSecs = *req.SyncIntervalSecs
	}
	p.UpdatedAt = time.Now().UTC()

	if err := s.projectRepo.Update(p); err != nil {
		return nil, fmt.Errorf("updating project: %w", err)
	}
	return p, nil
}

func (s *ProjectService) Delete(id, userID string) error {
	return s.projectRepo.Delete(id, userID)
}
