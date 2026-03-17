package handlers

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vikukumar/Pushpaka/internal/config"
	"github.com/vikukumar/Pushpaka/internal/middleware"
	"github.com/vikukumar/Pushpaka/internal/repositories"
)

// FileHandler serves project file browsing for the in-browser editor.
// Files are served from cfg.DeployDir/<projectID[:8]>/ (permanent deploy dir);
// if that directory does not exist, falls back to the latest deployment's
// clone directory so the editor works before/without a permanent deploy copy.
type FileHandler struct {
	projectRepo    *repositories.ProjectRepository
	deploymentRepo *repositories.DeploymentRepository
	deployDir      string
	cloneDir       string
}

func NewFileHandler(projectRepo *repositories.ProjectRepository, deploymentRepo *repositories.DeploymentRepository, cfg *config.Config) *FileHandler {
	return &FileHandler{
		projectRepo:    projectRepo,
		deploymentRepo: deploymentRepo,
		deployDir:      cfg.DeployDir,
		cloneDir:       cfg.CloneDir,
	}
}

// projectDir resolves the working directory for a project, verifying ownership.
// Priority: deployDir/<projectID[:8]> → cloneDir/<latestDeploymentID>.
// Returns "" and writes an error JSON if no directory can be resolved.
func (h *FileHandler) projectDir(c *gin.Context) (string, bool) {
	userID := middleware.GetUserID(c)
	projectID := c.Param("id")

	proj, err := h.projectRepo.FindByID(projectID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return "", false
	}

	// 1. Try the permanent deploy directory (Linux direct-deploy copies files here).
	deployPath := filepath.Join(h.deployDir, proj.ID[:8])
	if _, err := os.Stat(deployPath); err == nil {
		return deployPath, true
	}

	// 2. Fall back to the latest deployment's clone/build directory.
	//    On Windows (and when the copy step is skipped) files remain here.
	if h.cloneDir != "" {
		deployment, err := h.deploymentRepo.FindLatestByProjectID(proj.ID)
		if err == nil {
			clonePath := filepath.Join(h.cloneDir, deployment.ID)
			if _, err := os.Stat(clonePath); err == nil {
				return clonePath, true
			}
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "no source files found — trigger a deployment first"})
	return "", false
}

// safePath resolves a relative user-supplied path within the project directory,
// preventing path traversal attacks.
func safePath(root, rel string) (string, bool) {
	// Clean and join
	clean := filepath.Join(root, filepath.FromSlash(rel))
	// Ensure the result is still under root
	if !strings.HasPrefix(clean+string(filepath.Separator), root+string(filepath.Separator)) {
		return "", false
	}
	return clean, true
}

// FileEntry is a single entry in the file tree.
type FileEntry struct {
	Name     string       `json:"name"`
	Path     string       `json:"path"`
	IsDir    bool         `json:"is_dir"`
	Size     int64        `json:"size,omitempty"`
	Children []*FileEntry `json:"children,omitempty"`
}

// ListFiles   GET /api/v1/projects/:id/files
// Returns a recursive file tree (max depth 6, skips hidden dirs and node_modules).
func (h *FileHandler) ListFiles(c *gin.Context) {
	root, ok := h.projectDir(c)
	if !ok {
		return
	}
	tree, err := buildTree(root, root, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"files": tree.Children})
}

func buildTree(root, dir string, depth int) (*FileEntry, error) {
	if depth > 6 {
		return nil, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	node := &FileEntry{
		Name:  filepath.Base(dir),
		Path:  "/" + filepath.ToSlash(strings.TrimPrefix(dir, root+string(filepath.Separator))),
		IsDir: true,
	}
	for _, e := range entries {
		name := e.Name()
		// Skip hidden entries and known large/irrelevant directories
		if strings.HasPrefix(name, ".") {
			continue
		}
		skip := map[string]bool{
			"node_modules": true, "target": true, ".git": true,
			"vendor": true, "__pycache__": true, ".next": true,
			"dist": true, "build": true,
		}
		if e.IsDir() && skip[name] {
			continue
		}
		childPath := filepath.Join(dir, name)
		if e.IsDir() {
			child, err := buildTree(root, childPath, depth+1)
			if err != nil || child == nil {
				continue
			}
			node.Children = append(node.Children, child)
		} else {
			info, _ := e.Info()
			var sz int64
			if info != nil {
				sz = info.Size()
			}
			rel := strings.TrimPrefix(childPath, root)
			if rel != "" && (rel[0] == '/' || rel[0] == '\\') {
				rel = rel[1:]
			}
			node.Children = append(node.Children, &FileEntry{
				Name: name,
				Path: "/" + filepath.ToSlash(rel),
				Size: sz,
			})
		}
	}
	return node, nil
}

// ReadFile   GET /api/v1/projects/:id/files/*path
// Returns the file content (text). Refuses files > 512 KB.
func (h *FileHandler) ReadFile(c *gin.Context) {
	root, ok := h.projectDir(c)
	if !ok {
		return
	}
	rel := c.Param("path") // already cleaned by Gin, starts with "/"
	abs, safe := safePath(root, rel)
	if !safe {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid path"})
		return
	}

	info, err := os.Stat(abs)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}
	if info.IsDir() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path is a directory"})
		return
	}
	if info.Size() > 512*1024 {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file too large for editor (>512 KB)"})
		return
	}

	content, err := os.ReadFile(abs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Detect if it looks binary (contains null bytes in first 512 bytes)
	probe := content
	if len(probe) > 512 {
		probe = probe[:512]
	}
	for _, b := range probe {
		if b == 0 {
			c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "binary file — cannot open in editor"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"path":    rel,
		"content": string(content),
		"size":    info.Size(),
	})
}

// SaveFile   PUT /api/v1/projects/:id/files/*path
// Writes updated file content back to the deploy directory.
func (h *FileHandler) SaveFile(c *gin.Context) {
	root, ok := h.projectDir(c)
	if !ok {
		return
	}
	rel := c.Param("path")
	abs, safe := safePath(root, rel)
	if !safe {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid path"})
		return
	}

	var body struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content field required"})
		return
	}

	// Reject if content is too large (2 MB limit)
	if len(body.Content) > 2*1024*1024 {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "content exceeds 2 MB"})
		return
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(abs), fs.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create directory: " + err.Error()})
		return
	}

	if err := os.WriteFile(abs, []byte(body.Content), 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "write failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"saved": true, "path": rel})
}

// SyncFiles   POST /api/v1/projects/:id/files/sync
// Re-clones the project repo into deployDir/<projectID[:8]> so the in-browser
// editor can access the latest source code without needing a full deployment.
// If the directory already exists it is removed and re-cloned (fresh sync).
func (h *FileHandler) SyncFiles(c *gin.Context) {
	userID := middleware.GetUserID(c)
	projectID := c.Param("id")

	proj, err := h.projectRepo.FindByID(projectID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	if proj.RepoURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project has no repository URL"})
		return
	}

	destDir := filepath.Join(h.deployDir, proj.ID[:8])

	// Remove stale copy if present.
	if _, statErr := os.Stat(destDir); statErr == nil {
		if err := os.RemoveAll(destDir); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove old directory: " + err.Error()})
			return
		}
	}

	// Ensure parent directory exists.
	if err := os.MkdirAll(filepath.Dir(destDir), fs.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create parent directory: " + err.Error()})
		return
	}

	branch := proj.Branch
	if branch == "" {
		branch = "main"
	}

	// Build the clone URL — inject token for private repos.
	cloneURL := proj.RepoURL
	if proj.IsPrivate && proj.GitToken != "" {
		// e.g. https://token@github.com/user/repo.git
		cloneURL = injectToken(cloneURL, proj.GitToken)
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()

	args := []string{"clone", "--depth=1", "--branch", branch, cloneURL, destDir}
	cmd := exec.CommandContext(ctx, "git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "git clone failed: " + err.Error(),
			"output": string(out),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "synced",
		"branch":  branch,
	})
}

// injectToken embeds a personal access token into an HTTPS clone URL.
// e.g. https://github.com/user/repo → https://TOKEN@github.com/user/repo
func injectToken(repoURL, token string) string {
	// Only modify HTTPS URLs.
	for _, prefix := range []string{"https://", "http://"} {
		if strings.HasPrefix(repoURL, prefix) {
			return fmt.Sprintf("%s%s@%s", prefix, token, strings.TrimPrefix(repoURL, prefix))
		}
	}
	return repoURL
}
