package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vikukumar/pushpaka/internal/services"
)

type TaskHandler struct {
	dispatcher *services.TaskDispatcher
}

func NewTaskHandler(dispatcher *services.TaskDispatcher) *TaskHandler {
	return &TaskHandler{dispatcher: dispatcher}
}

type TaskCompleteRequest struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func (h *TaskHandler) InternalComplete(c *gin.Context) {
	taskID := c.Param("id")
	var req TaskCompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	err := h.dispatcher.HandleTaskCompletion(taskID, req.Success, req.Error)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *TaskHandler) List(c *gin.Context) {
	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
		return
	}

	tasks, err := h.dispatcher.GetProjectTasks(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tasks})
}

func (h *TaskHandler) Get(c *gin.Context) {
	id := c.Param("id")
	task, err := h.dispatcher.GetTask(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) Restart(c *gin.Context) {
	id := c.Param("id")
	err := h.dispatcher.RestartTask(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
