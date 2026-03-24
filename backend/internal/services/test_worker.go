package services

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/vikukumar/Pushpaka/internal/repositories"
	"github.com/vikukumar/Pushpaka/pkg/models"
	"github.com/vikukumar/Pushpaka/queue"
)

type TestWorker struct {
	WorkerID       int
	taskDispatcher *TaskDispatcher
	projectRepo    *repositories.ProjectRepository
	commitRepo     *repositories.CommitRepository
	rdb            *redis.Client
	logger         *zerolog.Logger
	testsDir       string
	ctx            context.Context
	cancel         context.CancelFunc
	queueStats     *queue.InProcess
}

func NewTestWorker(
	id int,
	taskDispatcher *TaskDispatcher,
	projectRepo *repositories.ProjectRepository,
	commitRepo *repositories.CommitRepository,
	rdb *redis.Client,
	logger *zerolog.Logger,
	testsDir string,
) *TestWorker {
	return &TestWorker{
		WorkerID:       id,
		taskDispatcher: taskDispatcher,
		projectRepo:    projectRepo,
		commitRepo:     commitRepo,
		rdb:            rdb,
		logger:         logger,
		testsDir:       testsDir,
	}
}

func (w *TestWorker) Start(ctx context.Context, q *queue.InProcess) error {
	w.ctx, w.cancel = context.WithCancel(ctx)
	w.queueStats = q
	w.logger.Info().Int("worker_id", w.WorkerID).Str("role", "tester").Msgf("test worker [%d] started", w.WorkerID)

	go w.run()
	return nil
}

func (w *TestWorker) run() {
	if w.queueStats != nil {
		w.queueStats.WorkerStarted("tester")
		defer w.queueStats.WorkerStopped("tester")
	}

	for {
		select {
		case <-w.ctx.Done():
			return
		default:
			if w.rdb != nil {
				// 1. Try Redis first
				res, err := w.rdb.BRPop(w.ctx, 1*time.Second, "pushpaka:tasks:test").Result()
				if err == nil {
					taskID := res[1]
					w.processTestTask(taskID)
					continue
				}
			}

			if w.queueStats != nil {
				// 2. Try In-Process queue
				select {
				case payload := <-w.queueStats.Chan("test"):
					taskID := string(payload)
					w.processTestTask(taskID)
				case <-time.After(1 * time.Second):
					// No job, loop again
				}
			} else if w.rdb == nil {
				time.Sleep(5 * time.Second)
			}
		}
	}
}

func (w *TestWorker) processTestTask(taskID string) {
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

	w.logger.Info().Str("project_id", p.ID).Str("sha", task.CommitSHA).Msg("running tests")

	// Placeholder test logic: simulate test run
	time.Sleep(3 * time.Second)

	// In real setup, copy build artifact to testsDir and run p.TestCommand

	w.taskDispatcher.HandleTaskCompletion(task.ID, true, "")
}
