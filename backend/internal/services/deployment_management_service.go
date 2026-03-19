package services

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	
	"github.com/vikukumar/Pushpaka/pkg/basemodel"
	"github.com/vikukumar/Pushpaka/pkg/models"
	"github.com/vikukumar/Pushpaka/internal/repositories"
)

type DeploymentManagementService struct {
	dmRepo         *repositories.DeploymentManagementRepository
	projectRepo    *repositories.ProjectRepository
	deploymentRepo *repositories.DeploymentRepository
	baseDir        string
}

func NewDeploymentManagementService(
	dmRepo *repositories.DeploymentManagementRepository,
	projectRepo *repositories.ProjectRepository,
	deploymentRepo *repositories.DeploymentRepository,
	baseDir string,
) *DeploymentManagementService {
	return &DeploymentManagementService{
		dmRepo:         dmRepo,
		projectRepo:    projectRepo,
		deploymentRepo: deploymentRepo,
		baseDir:        baseDir,
	}
}

// ============= Project Repository Cloning =============

// InitializeProjectClone clones repository for a project and stores in projectId folder
func (s *DeploymentManagementService) InitializeProjectClone(ctx context.Context, project *models.Project) (string, error) {
	// Create base clone directory
	projectClonePath := filepath.Join(s.baseDir, project.ID)

	// Remove existing directory if it exists
	if err := os.RemoveAll(projectClonePath); err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to remove existing clone directory: %w", err)
	}

	// Create project directory
	if err := os.MkdirAll(projectClonePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create project clone directory: %w", err)
	}

	// Clone repository
	branch := project.Branch
	if branch == "" {
		branch = "main"
	}

	cmd := exec.CommandContext(ctx, "git", "clone", "-b", branch, "--depth", "1")
	if project.IsPrivate && project.GitToken != "" {
		// Use token for private repo
		repoURL := project.RepoURL
		// Insert token: https://token@github.com/user/repo.git
		cmd.Args = append(cmd.Args, fmt.Sprintf("https://%s@%s", project.GitToken, repoURL[8:]))
	} else {
		cmd.Args = append(cmd.Args, project.RepoURL)
	}
	cmd.Args = append(cmd.Args, projectClonePath)

	if output, err := cmd.CombinedOutput(); err != nil {
		os.RemoveAll(projectClonePath)
		return "", fmt.Errorf("git clone failed: %v: %s", err, string(output))
	}

	// Update project with clone path
	project.GitClonePath = projectClonePath
	project.CloneDirectory = s.baseDir

	if err := s.projectRepo.Update(project); err != nil {
		return "", fmt.Errorf("failed to update project with clone path: %w", err)
	}

	return projectClonePath, nil
}

// RefreshProjectClone reclones the repository (when repo or branch changed)
func (s *DeploymentManagementService) RefreshProjectClone(ctx context.Context, project *models.Project, oldBranch string) error {
	// Delete existing clone
	if project.GitClonePath != "" {
		if err := os.RemoveAll(project.GitClonePath); err != nil && !os.IsNotExist(err) {
			log.Printf("warning: failed to remove old clone: %v", err)
		}
	}

	// Reclone with new URL/branch
	newPath, err := s.InitializeProjectClone(ctx, project)
	if err != nil {
		return err
	}

	project.GitClonePath = newPath
	return s.projectRepo.Update(project)
}

// ============= Code Signature Creation =============

// CaptureCodeSignature creates a signature of current code state
func (s *DeploymentManagementService) CaptureCodeSignature(
	deployment *models.Deployment,
	project *models.Project,
	sourcePath string,
) (*models.DeploymentCodeSignature, error) {
	if sourcePath == "" {
		sourcePath = project.GitClonePath
	}

	// Calculate code hash
	codeHash, fileCount, err := s.calculateCodeHash(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate code hash: %w", err)
	}

	// Create signature record
	sig := &models.DeploymentCodeSignature{
		BaseModel: basemodel.BaseModel{
			ID:        uuid.New().String(),
			CreatedAt: time.Now().UTC(),
		},
		DeploymentID:  deployment.ID,
		ProjectID:     project.ID,
		CommitSHA:     deployment.CommitSHA,
		CommitMessage: deployment.CommitMsg,
		Branch:        deployment.Branch,
		CodeHash:      codeHash,
		FileCount:     fileCount,
		DirectoryPath: sourcePath,
	}

	if err := s.dmRepo.CreateCodeSignature(sig); err != nil {
		return nil, fmt.Errorf("failed to create code signature: %w", err)
	}

	return sig, nil
}

// ============= Deployment Instance Management =============

// CreateDeploymentInstance creates a new deployment instance
func (s *DeploymentManagementService) CreateDeploymentInstance(
	deployment *models.Deployment,
	project *models.Project,
	sig *models.DeploymentCodeSignature,
	role models.DeploymentRole,
) (*models.DeploymentInstance, error) {
	// Create instance directory
	instanceDir := filepath.Join(s.baseDir, project.ID, string(role))
	if err := os.MkdirAll(instanceDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create instance directory: %w", err)
	}

	// Copy code from source to instance directory
	if err := s.copyDirectory(sig.DirectoryPath, instanceDir); err != nil {
		os.RemoveAll(instanceDir)
		return nil, fmt.Errorf("failed to copy code to instance: %w", err)
	}

	now := time.Now().UTC()
	instance := &models.DeploymentInstance{
		BaseModel: basemodel.BaseModel{
			ID:        uuid.New().String(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		DeploymentID:    deployment.ID,
		ProjectID:       project.ID,
		Role:            role,
		Status:          "stopped",
		Port:            project.Port,
		CodeSignatureID: sig.ID,
		InstanceDir:     instanceDir,
		HealthStatus:    "unknown",
	}

	if err := s.dmRepo.CreateDeploymentInstance(instance); err != nil {
		os.RemoveAll(instanceDir)
		return nil, fmt.Errorf("failed to create deployment instance: %w", err)
	}

	return instance, nil
}

// ============= Backup Management =============

// CreateBackup creates a backup of the current deployment
func (s *DeploymentManagementService) CreateBackup(
	instance *models.DeploymentInstance,
	project *models.Project,
	reason string,
) (*models.DeploymentBackup, error) {
	// Create backup directory
	backupDir := filepath.Join(s.baseDir, project.ID, "backups", instance.ID+"_"+time.Now().Format("20060102150405"))
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Copy instance directory to backup
	if err := s.copyDirectory(instance.InstanceDir, backupDir); err != nil {
		os.RemoveAll(backupDir)
		return nil, fmt.Errorf("failed to backup deployment: %w", err)
	}

	// Calculate backup size
	size, err := s.getDirectorySize(backupDir)
	if err != nil {
		log.Printf("warning: failed to calculate backup size: %v", err)
		size = 0
	}

	backup := &models.DeploymentBackup{
		BaseModel: basemodel.BaseModel{
			ID:        uuid.New().String(),
			CreatedAt: time.Now().UTC(),
		},
		DeploymentID:    instance.DeploymentID,
		ProjectID:       project.ID,
		CodeSignatureID: instance.CodeSignatureID,
		InstanceID:      instance.ID,
		BackupPath:      backupDir,
		Size:            size,
		Reason:          reason,
	}

	if err := s.dmRepo.CreateBackup(backup); err != nil {
		os.RemoveAll(backupDir)
		return nil, fmt.Errorf("failed to create backup record: %w", err)
	}

	// Clean up old backups if limit exceeded
	if err := s.cleanupOldBackups(project); err != nil {
		log.Printf("warning: failed to cleanup old backups: %v", err)
	}

	return backup, nil
}

// RestoreBackup restores a deployment from backup
func (s *DeploymentManagementService) RestoreBackup(
	backup *models.DeploymentBackup,
	instance *models.DeploymentInstance,
) error {
	// Backup current state first
	tempBackupDir := filepath.Join(s.baseDir, "temp_backups", instance.ID+"_"+time.Now().Format("20060102150405"))
	if err := os.MkdirAll(filepath.Dir(tempBackupDir), 0755); err != nil {
		return fmt.Errorf("failed to create temp backup: %w", err)
	}

	if err := s.copyDirectory(instance.InstanceDir, tempBackupDir); err != nil {
		return fmt.Errorf("failed to backup current instance: %w", err)
	}

	// Clear instance directory
	if err := os.RemoveAll(instance.InstanceDir); err != nil {
		return fmt.Errorf("failed to clear instance directory: %w", err)
	}

	// Restore from backup
	if err := s.copyDirectory(backup.BackupPath, instance.InstanceDir); err != nil {
		// Restore the temp backup
		os.RemoveAll(instance.InstanceDir)
		s.copyDirectory(tempBackupDir, instance.InstanceDir)
		return fmt.Errorf("failed to restore from backup: %w", err)
	}

	// Clean up temp backup
	os.RemoveAll(tempBackupDir)

	restoredTime := time.Now().UTC()
	backup.RestoredAt = &restoredTime

	if err := s.dmRepo.UpdateBackup(backup); err != nil {
		return fmt.Errorf("failed to update backup status: %w", err)
	}

	return nil
}

// ============= Deployment Actions =============

// RecordAction records a deployment action
func (s *DeploymentManagementService) RecordAction(
	deploymentID, instanceID, projectID, userID string,
	action models.DeploymentActionType,
) (*models.DeploymentAction, error) {
	now := time.Now().UTC()
	act := &models.DeploymentAction{
		BaseModel: basemodel.BaseModel{
			ID:        uuid.New().String(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		DeploymentID: deploymentID,
		InstanceID:   instanceID,
		ProjectID:    projectID,
		UserID:       userID,
		Action:       action,
		Status:       "pending",
	}

	if err := s.dmRepo.CreateDeploymentAction(act); err != nil {
		return nil, err
	}

	return act, nil
}

// UpdateActionStatus updates action status
func (s *DeploymentManagementService) UpdateActionStatus(
	actionID string,
	status string,
	result string,
) error {
	act := &models.DeploymentAction{
		BaseModel: basemodel.BaseModel{
			ID:        actionID,
		},
		Status:    status,
		Result:    result,
	}
	act.UpdatedAt = time.Now().UTC()

	return s.dmRepo.UpdateDeploymentAction(act)
}

// ============= Deployment Promotion =============

// PromoteTestingToMain promotes testing deployment to main
func (s *DeploymentManagementService) PromoteTestingToMain(
	project *models.Project,
	testingInstance *models.DeploymentInstance,
	mainInstance *models.DeploymentInstance,
	createBackup bool,
	gracefulShutdownTime time.Duration,
) error {
	if testingInstance.Role != models.DeploymentRoleTesting {
		return fmt.Errorf("can only promote testing deployments")
	}

	if mainInstance.Role != models.DeploymentRoleMain {
		return fmt.Errorf("invalid main deployment")
	}

	// Create backup of current main if requested
	if createBackup {
		_, err := s.CreateBackup(mainInstance, project, "pre_promotion_rollback")
		if err != nil {
			log.Printf("warning: failed to backup main deployment: %v", err)
		}
	}

	// Stop main gracefully
	mainInstance.Status = "stopping"
	mainInstance.UpdatedAt = time.Now().UTC()
	if err := s.dmRepo.UpdateDeploymentInstance(mainInstance); err != nil {
		return fmt.Errorf("failed to update main instance status: %w", err)
	}

	// Wait for graceful shutdown
	time.Sleep(gracefulShutdownTime)

	// Stop main deployment
	mainInstance.Status = "stopped"
	stoppedTime := time.Now().UTC()
	mainInstance.StoppedAt = &stoppedTime
	if err := s.dmRepo.UpdateDeploymentInstance(mainInstance); err != nil {
		return fmt.Errorf("failed to stop main instance: %w", err)
	}

	// Promote testing to main
	testingInstance.Role = models.DeploymentRoleMain
	testingInstance.Status = "running"
	startedTime := time.Now().UTC()
	testingInstance.StartedAt = &startedTime
	testingInstance.UpdatedAt = time.Now().UTC()

	if err := s.dmRepo.UpdateDeploymentInstance(testingInstance); err != nil {
		return fmt.Errorf("failed to promote testing deployment: %w", err)
	}

	// Move old main to backup role
	mainInstance.Role = models.DeploymentRoleBackup
	if err := s.dmRepo.UpdateDeploymentInstance(mainInstance); err != nil {
		return fmt.Errorf("failed to update old main to backup: %w", err)
	}

	// Update project main deployment
	project.MainDeployID = testingInstance.DeploymentID
	if err := s.projectRepo.Update(project); err != nil {
		return fmt.Errorf("failed to update project main deployment: %w", err)
	}

	return nil
}

// ============= Cleanup & Maintenance =============

// cleanupOldBackups removes old backups exceeding max limit
func (s *DeploymentManagementService) cleanupOldBackups(project *models.Project) error {
	if project.MaxBackups <= 0 {
		return nil
	}

	oldBackups, err := s.dmRepo.GetOldestBackups(project.ID, project.MaxBackups)
	if err != nil {
		return err
	}

	for _, backup := range oldBackups {
		if err := os.RemoveAll(backup.BackupPath); err != nil {
			log.Printf("warning: failed to remove backup directory: %v", err)
		}

		if err := s.dmRepo.DeleteBackup(backup.ID); err != nil {
			log.Printf("warning: failed to delete backup record: %v", err)
		}
	}

	return nil
}

// ============= Helper Methods =============

// calculateCodeHash calculates SHA256 hash of directory
func (s *DeploymentManagementService) calculateCodeHash(dirPath string) (string, int, error) {
	hash := sha256.New()
	fileCount := 0

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Skip hidden directories and .git
			if info.Name() == ".git" || (len(info.Name()) > 0 && info.Name()[0] == '.') {
				return filepath.SkipDir
			}
			return nil
		}

		fileCount++

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		if _, err := io.Copy(hash, file); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return "", 0, err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), fileCount, nil
}

// copyDirectory recursively copies directory
func (s *DeploymentManagementService) copyDirectory(src, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return err
		}

		return os.Chmod(dstPath, info.Mode())
	})
}

// getDirectorySize calculates total size of directory
func (s *DeploymentManagementService) getDirectorySize(dirPath string) (int64, error) {
	var size int64

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			size += info.Size()
		}

		return nil
	})

	return size, err
}
