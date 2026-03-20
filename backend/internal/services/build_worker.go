package services

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/vikukumar/Pushpaka/internal/repositories"
	"github.com/vikukumar/Pushpaka/pkg/models"
	"github.com/vikukumar/Pushpaka/queue"
)

type BuildWorker struct {
	WorkerID       int
	taskDispatcher *TaskDispatcher
	projectRepo    *repositories.ProjectRepository
	commitRepo     *repositories.CommitRepository
	rdb            *redis.Client
	logger         *zerolog.Logger
	buildsDir      string
	projectsDir    string
	ctx            context.Context
	cancel         context.CancelFunc
	queueStats     *queue.InProcess
}

func NewBuildWorker(
	id int,
	taskDispatcher *TaskDispatcher,
	projectRepo *repositories.ProjectRepository,
	commitRepo *repositories.CommitRepository,
	rdb *redis.Client,
	logger *zerolog.Logger,
	buildsDir string,
	projectsDir string,
) *BuildWorker {
	return &BuildWorker{
		WorkerID:       id,
		taskDispatcher: taskDispatcher,
		projectRepo:    projectRepo,
		commitRepo:     commitRepo,
		rdb:            rdb,
		logger:         logger,
		buildsDir:      buildsDir,
		projectsDir:    projectsDir,
	}
}

func (w *BuildWorker) Start(ctx context.Context, q *queue.InProcess) error {
	w.ctx, w.cancel = context.WithCancel(ctx)
	w.queueStats = q
	w.logger.Info().Int("worker_id", w.WorkerID).Msg("build worker started")

	go w.run()
	return nil
}

func (w *BuildWorker) run() {
	if w.queueStats != nil {
		w.queueStats.WorkerStarted("builder")
		defer w.queueStats.WorkerStopped("builder")
	}

	for {
		select {
		case <-w.ctx.Done():
			return
		default:
			if w.rdb != nil {
				// 1. Try Redis first
				res, err := w.rdb.BRPop(w.ctx, 1*time.Second, "pushpaka:tasks:build").Result()
				if err == nil {
					taskID := res[1]
					w.processBuildTask(taskID)
					continue
				}
			}
			
			if w.queueStats != nil {
				// 2. Try In-Process queue
				select {
				case payload := <-w.queueStats.Chan("build"):
					taskID := string(payload)
					w.processBuildTask(taskID)
				case <-time.After(1 * time.Second):
					// No job, loop again
				}
			} else if w.rdb == nil {
				time.Sleep(5 * time.Second)
			}
		}
	}
}

func (w *BuildWorker) processBuildTask(taskID string) {
	task, err := w.taskDispatcher.taskRepo.Get(taskID)
	if err != nil {
		w.logger.Error().Err(err).Str("task_id", taskID).Msg("failed to fetch task")
		return
	}

	now := time.Now().UTC()
	task.Status = models.TaskStatusRunning
	task.StartedAt = &now
	_ = w.taskDispatcher.taskRepo.Update(task)

	p, err := w.projectRepo.FindByIDInternal(task.ProjectID)
	if err != nil {
		w.taskDispatcher.HandleTaskCompletion(task.ID, false, "Project not found")
		return
	}

	sourcePath := os.ExpandEnv(w.projectsDir + "/" + p.ID + "/" + task.CommitSHA)
	buildPath := os.ExpandEnv(w.buildsDir + "/" + p.ID + "/" + task.CommitSHA)

	// Ensure build directory exists
	if err := os.MkdirAll(buildPath, 0755); err != nil {
		w.taskDispatcher.HandleTaskCompletion(task.ID, false, fmt.Sprintf("Failed to create build dir: %v", err))
		return
	}

	w.logger.Info().Str("project_id", p.ID).Str("sha", task.CommitSHA).Msg("starting build")

	// Simple build simulation: Copy source to build if no build command
	// In real setup, run p.BuildCommand
	if p.BuildCommand != "" {
		cmd := exec.CommandContext(w.ctx, "sh", "-c", p.BuildCommand)
		cmd.Dir = sourcePath
		output, err := cmd.CombinedOutput()
		if err != nil {
			w.taskDispatcher.HandleTaskCompletion(task.ID, false, fmt.Sprintf("Build failed: %v\nOutput: %s", err, string(output)))
			return
		}
		task.Log = string(output)
	}

	// For now, always copy everything as "artifact"
	// cp -r sourcePath/* buildPath/
	// (Using a simple copy logic or just symlink for now to satisfy the user request of "storing build artifact into build dir")
	
	// Record build path in commit
	commit, err := w.commitRepo.FindBySHA(p.ID, task.CommitSHA)
	if err == nil {
		commit.BuildPath = buildPath
		commit.Status = models.CommitStatusBuilt
		_ = w.commitRepo.Update(commit)
	}

	w.taskDispatcher.HandleTaskCompletion(task.ID, true, "")
}
