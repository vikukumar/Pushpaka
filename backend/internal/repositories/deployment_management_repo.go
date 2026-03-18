package repositories

import (
	"github.com/jmoiron/sqlx"
	"github.com/vikukumar/Pushpaka/internal/models"
)

type DeploymentManagementRepository struct {
	db *sqlx.DB
}

func NewDeploymentManagementRepository(db *sqlx.DB) *DeploymentManagementRepository {
	return &DeploymentManagementRepository{db: db}
}

// ============= Code Signature Operations =============

// CreateCodeSignature creates a new code signature record
func (r *DeploymentManagementRepository) CreateCodeSignature(sig *models.DeploymentCodeSignature) error {
	query := `
		INSERT INTO deployment_code_signatures (
			id, deployment_id, project_id, commit_sha, commit_message, commit_author,
			branch, code_hash, file_count, directory_path, created_at
		) VALUES (
			:id, :deployment_id, :project_id, :commit_sha, :commit_message, :commit_author,
			:branch, :code_hash, :file_count, :directory_path, :created_at
		)`
	_, err := r.db.NamedExec(query, sig)
	return err
}

// GetCodeSignature retrieves a code signature by ID
func (r *DeploymentManagementRepository) GetCodeSignature(id string) (*models.DeploymentCodeSignature, error) {
	var sig models.DeploymentCodeSignature
	err := r.db.Get(&sig, r.db.Rebind("SELECT * FROM deployment_code_signatures WHERE id = ?"), id)
	if err != nil {
		return nil, err
	}
	return &sig, nil
}

// GetCodeSignatureByDeployment retrieves latest code signature for a deployment
func (r *DeploymentManagementRepository) GetCodeSignatureByDeployment(deploymentID string) (*models.DeploymentCodeSignature, error) {
	var sig models.DeploymentCodeSignature
	err := r.db.Get(&sig,
		r.db.Rebind("SELECT * FROM deployment_code_signatures WHERE deployment_id = ? ORDER BY created_at DESC LIMIT 1"),
		deploymentID)
	if err != nil {
		return nil, err
	}
	return &sig, nil
}

// ============= Deployment Instance Operations =============

// CreateDeploymentInstance creates a new deployment instance
func (r *DeploymentManagementRepository) CreateDeploymentInstance(inst *models.DeploymentInstance) error {
	query := `
		INSERT INTO deployment_instances (
			id, deployment_id, project_id, role, status, container_id, process_id,
			port, code_signature_id, instance_dir, started_at, last_health_check,
			health_status, restart_count, error_log, created_at, updated_at
		) VALUES (
			:id, :deployment_id, :project_id, :role, :status, :container_id, :process_id,
			:port, :code_signature_id, :instance_dir, :started_at, :last_health_check,
			:health_status, :restart_count, :error_log, :created_at, :updated_at
		)`
	_, err := r.db.NamedExec(query, inst)
	return err
}

// GetDeploymentInstance retrieves deployment instance by ID
func (r *DeploymentManagementRepository) GetDeploymentInstance(id string) (*models.DeploymentInstance, error) {
	var inst models.DeploymentInstance
	err := r.db.Get(&inst, r.db.Rebind("SELECT * FROM deployment_instances WHERE id = ?"), id)
	if err != nil {
		return nil, err
	}
	return &inst, nil
}

// GetDeploymentInstances retrieves all instances for a deployment
func (r *DeploymentManagementRepository) GetDeploymentInstances(deploymentID string) ([]models.DeploymentInstance, error) {
	var instances []models.DeploymentInstance
	err := r.db.Select(&instances,
		r.db.Rebind("SELECT * FROM deployment_instances WHERE deployment_id = ? ORDER BY role, created_at"),
		deploymentID)
	return instances, err
}

// GetMainDeploymentInstance retrieves the main (production) instance
func (r *DeploymentManagementRepository) GetMainDeploymentInstance(projectID string) (*models.DeploymentInstance, error) {
	var inst models.DeploymentInstance
	err := r.db.Get(&inst,
		r.db.Rebind("SELECT * FROM deployment_instances WHERE project_id = ? AND role = ? AND status = 'running' LIMIT 1"),
		projectID, models.DeploymentRoleMain)
	if err != nil {
		return nil, err
	}
	return &inst, nil
}

// UpdateDeploymentInstance updates a deployment instance
func (r *DeploymentManagementRepository) UpdateDeploymentInstance(inst *models.DeploymentInstance) error {
	query := `
		UPDATE deployment_instances SET
			status = :status, container_id = :container_id, process_id = :process_id,
			port = :port, code_signature_id = :code_signature_id, started_at = :started_at,
			stopped_at = :stopped_at, last_health_check = :last_health_check,
			health_status = :health_status, restart_count = :restart_count,
			error_log = :error_log, updated_at = :updated_at
		WHERE id = :id`
	_, err := r.db.NamedExec(query, inst)
	return err
}

// DeleteDeploymentInstance deletes a deployment instance
func (r *DeploymentManagementRepository) DeleteDeploymentInstance(id string) error {
	_, err := r.db.Exec(r.db.Rebind("DELETE FROM deployment_instances WHERE id = ?"), id)
	return err
}

// ============= Backup Operations =============

// CreateBackup creates a new backup record
func (r *DeploymentManagementRepository) CreateBackup(backup *models.DeploymentBackup) error {
	query := `
		INSERT INTO deployment_backups (
			id, deployment_id, project_id, code_signature_id, instance_id,
			backup_path, size, reason, is_restored, restored_at, created_at
		) VALUES (
			:id, :deployment_id, :project_id, :code_signature_id, :instance_id,
			:backup_path, :size, :reason, :is_restored, :restored_at, :created_at
		)`
	_, err := r.db.NamedExec(query, backup)
	return err
}

// GetBackup retrieves a backup by ID
func (r *DeploymentManagementRepository) GetBackup(id string) (*models.DeploymentBackup, error) {
	var backup models.DeploymentBackup
	err := r.db.Get(&backup, r.db.Rebind("SELECT * FROM deployment_backups WHERE id = ?"), id)
	if err != nil {
		return nil, err
	}
	return &backup, nil
}

// GetBackupsByDeployment retrieves all backups for a deployment (limited)
func (r *DeploymentManagementRepository) GetBackupsByDeployment(deploymentID string, limit int) ([]models.DeploymentBackup, error) {
	var backups []models.DeploymentBackup
	err := r.db.Select(&backups,
		r.db.Rebind("SELECT * FROM deployment_backups WHERE deployment_id = ? ORDER BY created_at DESC LIMIT ?"),
		deploymentID, limit)
	return backups, err
}

// GetOldestBackups returns oldest backups for cleanup
func (r *DeploymentManagementRepository) GetOldestBackups(projectID string, keepCount int) ([]models.DeploymentBackup, error) {
	var backups []models.DeploymentBackup
	err := r.db.Select(&backups,
		r.db.Rebind(`
			SELECT * FROM deployment_backups 
			WHERE project_id = ? AND is_restored = false
			ORDER BY created_at ASC
			LIMIT (SELECT COUNT(*) FROM deployment_backups WHERE project_id = ?) - ?
		`),
		projectID, projectID, keepCount)
	return backups, err
}

// UpdateBackup updates a backup record
func (r *DeploymentManagementRepository) UpdateBackup(backup *models.DeploymentBackup) error {
	query := `
		UPDATE deployment_backups SET
			is_restored = :is_restored, restored_at = :restored_at
		WHERE id = :id`
	_, err := r.db.NamedExec(query, backup)
	return err
}

// DeleteBackup deletes a backup record
func (r *DeploymentManagementRepository) DeleteBackup(id string) error {
	_, err := r.db.Exec(r.db.Rebind("DELETE FROM deployment_backups WHERE id = ?"), id)
	return err
}

// ============= Deployment Action Operations =============

// CreateDeploymentAction records a deployment action
func (r *DeploymentManagementRepository) CreateDeploymentAction(action *models.DeploymentAction) error {
	query := `
		INSERT INTO deployment_actions (
			id, deployment_id, instance_id, project_id, user_id, action,
			status, result, created_at, updated_at
		) VALUES (
			:id, :deployment_id, :instance_id, :project_id, :user_id, :action,
			:status, :result, :created_at, :updated_at
		)`
	_, err := r.db.NamedExec(query, action)
	return err
}

// UpdateDeploymentAction updates an action record
func (r *DeploymentManagementRepository) UpdateDeploymentAction(action *models.DeploymentAction) error {
	query := `
		UPDATE deployment_actions SET
			status = :status, result = :result, updated_at = :updated_at
		WHERE id = :id`
	_, err := r.db.NamedExec(query, action)
	return err
}

// GetDeploymentActions retrieves actions for a deployment (limited)
func (r *DeploymentManagementRepository) GetDeploymentActions(deploymentID string, limit int) ([]models.DeploymentAction, error) {
	var actions []models.DeploymentAction
	err := r.db.Select(&actions,
		r.db.Rebind("SELECT * FROM deployment_actions WHERE deployment_id = ? ORDER BY created_at DESC LIMIT ?"),
		deploymentID, limit)
	return actions, err
}

// GetLastDeploymentAction gets the most recent action
func (r *DeploymentManagementRepository) GetLastDeploymentAction(deploymentID string) (*models.DeploymentAction, error) {
	var action models.DeploymentAction
	err := r.db.Get(&action,
		r.db.Rebind("SELECT * FROM deployment_actions WHERE deployment_id = ? ORDER BY created_at DESC LIMIT 1"),
		deploymentID)
	if err != nil {
		return nil, err
	}
	return &action, nil
}

// ============= Stats Operations =============

// CreateOrUpdateStats creates or updates deployment stats
func (r *DeploymentManagementRepository) CreateOrUpdateStats(stats *models.ProjectDeploymentStats) error {
	// First check if exists
	existing := &models.ProjectDeploymentStats{}
	err := r.db.Get(existing,
		r.db.Rebind("SELECT * FROM project_deployment_stats WHERE project_id = ?"),
		stats.ProjectID)

	if err != nil {
		// Insert new record
		query := `
			INSERT INTO project_deployment_stats (
				id, project_id, main_deployment_id, testing_deployment_id,
				total_deployments, successful_deploys, failed_deploys,
				total_backups, last_deploy_at, avg_deploy_time, updated_at
			) VALUES (
				:id, :project_id, :main_deployment_id, :testing_deployment_id,
				:total_deployments, :successful_deploys, :failed_deploys,
				:total_backups, :last_deploy_at, :avg_deploy_time, :updated_at
			)`
		_, err := r.db.NamedExec(query, stats)
		return err
	}

	// Update existing record
	query := `
		UPDATE project_deployment_stats SET
			main_deployment_id = :main_deployment_id,
			testing_deployment_id = :testing_deployment_id,
			total_deployments = :total_deployments,
			successful_deploys = :successful_deploys,
			failed_deploys = :failed_deploys,
			total_backups = :total_backups,
			last_deploy_at = :last_deploy_at,
			avg_deploy_time = :avg_deploy_time,
			updated_at = :updated_at
		WHERE project_id = :project_id`
	_, err = r.db.NamedExec(query, stats)
	return err
}

// GetStats retrieves deployment stats for a project
func (r *DeploymentManagementRepository) GetStats(projectID string) (*models.ProjectDeploymentStats, error) {
	var stats models.ProjectDeploymentStats
	err := r.db.Get(&stats,
		r.db.Rebind("SELECT * FROM project_deployment_stats WHERE project_id = ?"),
		projectID)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// ============= Helper Methods =============

// CountDeploymentsByProject counts total deployments for project
func (r *DeploymentManagementRepository) CountDeploymentsByProject(projectID string) (int, error) {
	var count int
	err := r.db.Get(&count,
		r.db.Rebind("SELECT COUNT(*) FROM deployment_instances WHERE project_id = ? AND role IN (?, ?)"),
		projectID, models.DeploymentRoleMain, models.DeploymentRoleTesting)
	return count, err
}

// CountBackupsByProject counts total backups for project
func (r *DeploymentManagementRepository) CountBackupsByProject(projectID string) (int, error) {
	var count int
	err := r.db.Get(&count,
		r.db.Rebind("SELECT COUNT(*) FROM deployment_backups WHERE project_id = ?"),
		projectID)
	return count, err
}

// GetProjectDeployments retrieves all deployments for a project
func (r *DeploymentManagementRepository) GetProjectDeployments(projectID string) ([]models.DeploymentInstance, error) {
	var instances []models.DeploymentInstance
	err := r.db.Select(&instances,
		r.db.Rebind("SELECT * FROM deployment_instances WHERE project_id = ? ORDER BY role, created_at"),
		projectID)
	return instances, err
}

// GetBackupByID retrieves a backup by ID
func (r *DeploymentManagementRepository) GetBackupByID(backupID string) (*models.DeploymentBackup, error) {
	var backup models.DeploymentBackup
	err := r.db.Get(&backup,
		r.db.Rebind("SELECT * FROM deployment_backups WHERE id = ?"),
		backupID)
	if err != nil {
		return nil, err
	}
	return &backup, nil
}
