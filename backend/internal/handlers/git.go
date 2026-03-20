package handlers

import (
	"context"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vikukumar/Pushpaka/internal/config"
	"github.com/vikukumar/Pushpaka/internal/middleware"
	"github.com/vikukumar/Pushpaka/internal/repositories"
)

type GitHandler struct {
	projectRepo    *repositories.ProjectRepository
	deploymentRepo *repositories.DeploymentRepository
	deployDir      string
	cloneDir       string
}

func NewGitHandler(projectRepo *repositories.ProjectRepository, deploymentRepo *repositories.DeploymentRepository, cfg *config.Config) *GitHandler {
	return &GitHandler{
		projectRepo:    projectRepo,
		deploymentRepo: deploymentRepo,
		deployDir:      cfg.DeploysDir,
		cloneDir:       cfg.CloneDir,
	}
}

func (h *GitHandler) getWorkDir(c *gin.Context) (string, bool) {
	userID := middleware.GetUserID(c)
	projectID := c.Param("id")

	proj, err := h.projectRepo.FindByID(projectID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return "", false
	}

	destDir := filepath.Join(h.deployDir, proj.ID[:8])
	return destDir, true
}

// Status GET /api/v1/projects/:id/git/status
func (h *GitHandler) Status(c *gin.Context) {
	dir, ok := h.getWorkDir(c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "git status failed: " + err.Error(), "output": string(out)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": string(out)})
}

// Commit POST /api/v1/projects/:id/git/commit
func (h *GitHandler) Commit(c *gin.Context) {
	dir, ok := h.getWorkDir(c)
	if !ok {
		return
	}

	var body struct {
		Message string `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 1*time.Minute)
	defer cancel()

	// 1. Git add .
	addCmd := exec.CommandContext(ctx, "git", "add", ".")
	addCmd.Dir = dir
	if out, err := addCmd.CombinedOutput(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "git add failed", "output": string(out)})
		return
	}

	// 2. Git commit
	commitCmd := exec.CommandContext(ctx, "git", "commit", "-m", body.Message)
	commitCmd.Dir = dir
	if out, err := commitCmd.CombinedOutput(); err != nil {
		// If nothing to commit, output might contain "nothing to commit"
		if strings.Contains(string(out), "nothing to commit") {
			c.JSON(http.StatusOK, gin.H{"message": "nothing to commit"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "git commit failed", "output": string(out)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "committed"})
}

// Push POST /api/v1/projects/:id/git/push
func (h *GitHandler) Push(c *gin.Context) {
	userID := middleware.GetUserID(c)
	projectID := c.Param("id")

	proj, err := h.projectRepo.FindByID(projectID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	dir := filepath.Join(h.deployDir, proj.ID[:8])

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()

	// Push often needs auth. We might need to use the token in the URL if it's not already set.
	// But usually the repo was cloned with the token, so it should be in the remote URL.
	cmd := exec.CommandContext(ctx, "git", "push")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "git push failed", "output": string(out)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "pushed"})
}

// Pull POST /api/v1/projects/:id/git/pull
func (h *GitHandler) Pull(c *gin.Context) {
	dir, ok := h.getWorkDir(c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "pull")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "git pull failed", "output": string(out)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "pulled"})
}
