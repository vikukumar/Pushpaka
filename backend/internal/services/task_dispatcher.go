package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/vikukumar/Pushpaka/internal/repositories"
	"github.com/vikukumar/Pushpaka/pkg/basemodel"
	"github.com/vikukumar/Pushpaka/pkg/models"
	"github.com/vikukumar/Pushpaka/queue"
)

type TaskDispatcher struct {
	taskRepo    *repositories.TaskRepository
	projectRepo *repositories.ProjectRepository
	rdb         *redis.Client
	inQueue     *queue.InProcess
	log         *zerolog.Logger
}

func NewTaskDispatcher(
	taskRepo *repositories.TaskRepository,
	projectRepo *repositories.ProjectRepository,
	rdb *redis.Client,
	inQueue *queue.InProcess,
	log *zerolog.Logger,
) *TaskDispatcher {
	return &TaskDispatcher{
		taskRepo:    taskRepo,
		projectRepo: projectRepo,
		rdb:         rdb,
		inQueue:     inQueue,
		log:         log,
	}
}

// CreateTask creates a new task and queues it for the appropriate worker role
func (d *TaskDispatcher) CreateTask(projectID string, taskType models.TaskType, sha string) (*models.ProjectTask, error) {
	task := &models.ProjectTask{
		BaseModel: basemodel.BaseModel{ID: uuid.New().String()},
		ProjectID: projectID,
		Type:      taskType,
		Status:    models.TaskStatusPending,
		CommitSHA: sha,
	}

	if err := d.taskRepo.Create(task); err != nil {
		return nil, err
	}

	// Update project's current task
	d.projectRepo.UpdateTaskStatus(projectID, task.ID, string(taskType))

	// Queue for worker
	d.queueTask(task)

	return task, nil
}

func (d *TaskDispatcher) queueTask(task *models.ProjectTask) {
	// Determine queue based on task type (role-based)
	roleQueue := fmt.Sprintf("pushpaka:tasks:%s", task.Type)
	
	// Payload for worker (simplified version of DeploymentJob if needed, or just TaskID)
	// For now, let's assume workers pick up TaskID and fetch details.
	payload := []byte(task.ID)

	if d.inQueue != nil {
		roleString := string(task.Type)
		_ = d.inQueue.Push(roleString, payload)
	}

	if d.rdb != nil {
		d.rdb.LPush(context.Background(), roleQueue, payload)
	}
}

// HandleTaskCompletion is called by workers when a task finishes.
func (d *TaskDispatcher) HandleTaskCompletion(taskID string, success bool, errStr string) error {
	task, err := d.taskRepo.Get(taskID)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	task.FinishedAt = &now
	if success {
		task.Status = models.TaskStatusCompleted
	} else {
		task.Status = models.TaskStatusFailed
		task.Error = errStr
	}

	if err := d.taskRepo.Update(task); err != nil {
		return err
	}

	if success {
		// Chain next task
		d.triggerNextTask(task)
	}

	return nil
}

// RestartTask resets a task to pending and re-queues it.
func (d *TaskDispatcher) RestartTask(taskID string) error {
	task, err := d.taskRepo.Get(taskID)
	if err != nil {
		return err
	}

	task.Status = models.TaskStatusPending
	task.StartedAt = nil
	task.FinishedAt = nil
	task.Error = ""
	task.Log = "" // Clear logs on restart? Probably better.

	if err := d.taskRepo.Update(task); err != nil {
		return err
	}

	d.queueTask(task)
	return nil
}

func (d *TaskDispatcher) triggerNextTask(task *models.ProjectTask) {
	switch task.Type {
	case models.TaskTypeSync, models.TaskTypeFetch:
		// Success -> Build
		d.CreateTask(task.ProjectID, models.TaskTypeBuild, task.CommitSHA)
	case models.TaskTypeBuild:
		// Success -> Test
		d.CreateTask(task.ProjectID, models.TaskTypeTest, task.CommitSHA)
	case models.TaskTypeTest:
		// Success -> Mark project as ready for deployment
		d.projectRepo.UpdateStatus(task.ProjectID, "ready")
	}
}
func (d *TaskDispatcher) GetProjectTasks(projectID string) ([]models.ProjectTask, error) {
	return d.taskRepo.FindByProjectID(projectID)
}

func (d *TaskDispatcher) GetTask(id string) (*models.ProjectTask, error) {
	return d.taskRepo.Get(id)
}
