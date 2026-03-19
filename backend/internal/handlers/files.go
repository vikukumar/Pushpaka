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
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "no deployment found for this project — trigger a deployment first"})
			return "", false
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "source files missing from server. deployDir: " + h.deployDir})
	return "", false
}

// safePath resolves a relative user-supplied path within the project directory,
// preventing path traversal attacks.
func SafePath(root, rel string) (string, bool) {
	// 1. Clean the root and get its absolute path.
	absRoot, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return "", false
	}

	// 2. Clean the relative path and join it with the root.
	// filepath.Join calls filepath.Clean on the result.
	clean := filepath.Join(absRoot, filepath.FromSlash(rel))

	// 3. Get the absolute version of the target path.
	absTarget, err := filepath.Abs(clean)
	if err != nil {
		return "", false
	}

	// 4. Ensure the target is actually under the root.
	// Use string comparison with a trailing separator to avoid "root-matching"
	// (e.g., /app/data-secret matching /app/data).
	rootPrefix := absRoot
	if !strings.HasSuffix(rootPrefix, string(filepath.Separator)) {
		rootPrefix += string(filepath.Separator)
	}

	if !strings.HasPrefix(absTarget, absRoot) {
		return "", false
	}

	// Re-verify after prefix check for edge cases.
	if !strings.HasPrefix(absTarget+string(filepath.Separator), rootPrefix) && absTarget != absRoot {
		return "", false
	}

	return absTarget, true
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
	rel := strings.TrimPrefix(dir, root)
	if rel != "" && (rel[0] == '/' || rel[0] == '\\') {
		rel = rel[1:]
	}
	node := &FileEntry{
		Name:  filepath.Base(dir),
		Path:  "/" + filepath.ToSlash(rel),
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
	abs, safe := SafePath(root, rel)
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
	if info.Size() > 5*1024*1024 {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file too large for editor (>5 MB)"})
		return
	}

	content, err := os.ReadFile(abs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Allow common image extensions to bypass strict null-byte check
	ext := strings.ToLower(filepath.Ext(rel))
	isImage := ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".ico" || ext == ".webp"

	if !isImage {
		// Detect if it looks binary (contains null bytes in first 512 bytes)
		probe := content
		if len(probe) > 512 {
			probe = probe[:512]
		}
		for _, b := range probe {
			if b == 0 {
				c.JSON(http.StatusUnsupportedMediaType, gin.H{
					"error":     "binary file — cannot open in editor",
					"is_binary": true,
				})
				return
			}
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
	abs, safe := SafePath(root, rel)
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

// CreateFile   POST /api/v1/projects/:id/files/*path
func (h *FileHandler) CreateFile(c *gin.Context) {
	root, ok := h.projectDir(c)
	if !ok {
		return
	}
	rel := c.Param("path")
	abs, safe := SafePath(root, rel)
	if !safe {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid path"})
		return
	}

	// Ensure parent exists
	if err := os.MkdirAll(filepath.Dir(abs), fs.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create directory: " + err.Error()})
		return
	}

	// Check if exists
	if _, err := os.Stat(abs); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "file already exists"})
		return
	}

	if err := os.WriteFile(abs, []byte(""), 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"created": true, "path": rel})
}

// CreateDirectory POST /api/v1/projects/:id/directories/*path
func (h *FileHandler) CreateDirectory(c *gin.Context) {
	root, ok := h.projectDir(c)
	if !ok {
		return
	}
	rel := c.Param("path")
	abs, safe := SafePath(root, rel)
	if !safe {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid path"})
		return
	}

	if err := os.MkdirAll(abs, fs.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "mkdir failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"created": true, "path": rel})
}

// DeleteFile   DELETE /api/v1/projects/:id/files/*path
func (h *FileHandler) DeleteFile(c *gin.Context) {
	root, ok := h.projectDir(c)
	if !ok {
		return
	}
	rel := c.Param("path")
	abs, safe := SafePath(root, rel)
	if !safe {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid path"})
		return
	}

	if err := os.RemoveAll(abs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true, "path": rel})
}

// RenameFile   PATCH /api/v1/projects/:id/files/*path
func (h *FileHandler) RenameFile(c *gin.Context) {
	root, ok := h.projectDir(c)
	if !ok {
		return
	}
	rel := c.Param("path")
	abs, safe := SafePath(root, rel)
	if !safe {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid path"})
		return
	}

	var body struct {
		NewPath string `json:"newPath" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "newPath required"})
		return
	}

	newAbs, safe := SafePath(root, body.NewPath)
	if !safe {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid destination path"})
		return
	}

	if err := os.Rename(abs, newAbs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "rename failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"renamed": true, "oldPath": rel, "newPath": body.NewPath})
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

// --- System-wide File Management ---

// systemRoot resolves the global deploy directory.
func (h *FileHandler) systemRoot(c *gin.Context) (string, bool) {
	// For now, allow all authenticated users (or restrict to admins if role is available)
	return h.deployDir, true
}

// ListSystemFiles GET /api/v1/system/files
func (h *FileHandler) ListSystemFiles(c *gin.Context) {
	root, ok := h.systemRoot(c)
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

// ReadSystemFile GET /api/v1/system/files/*path
func (h *FileHandler) ReadSystemFile(c *gin.Context) {
	root, ok := h.systemRoot(c)
	if !ok {
		return
	}
	rel := c.Param("path")
	abs, safe := SafePath(root, rel)
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
	if info.Size() > 10*1024*1024 { // 10MB
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file too large (>10 MB)"})
		return
	}

	content, err := os.ReadFile(abs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"path":    rel,
		"content": string(content),
		"size":    info.Size(),
	})
}

// SaveSystemFile PUT /api/v1/system/files/*path
func (h *FileHandler) SaveSystemFile(c *gin.Context) {
	root, ok := h.systemRoot(c)
	if !ok {
		return
	}
	rel := c.Param("path")
	abs, safe := SafePath(root, rel)
	if !safe {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid path"})
		return
	}

	var body struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content required"})
		return
	}

	if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "mkdir failed"})
		return
	}

	if err := os.WriteFile(abs, []byte(body.Content), 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "write failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"saved": true})
}

// CreateSystemFile POST /api/v1/system/files/*path
func (h *FileHandler) CreateSystemFile(c *gin.Context) {
	root, ok := h.systemRoot(c)
	if !ok {
		return
	}
	rel := c.Param("path")
	abs, safe := SafePath(root, rel)
	if !safe {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid path"})
		return
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "mkdir failed"})
		return
	}
	if err := os.WriteFile(abs, []byte(""), 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create failed"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"created": true})
}

// CreateSystemDirectory POST /api/v1/system/directories/*path
func (h *FileHandler) CreateSystemDirectory(c *gin.Context) {
	root, ok := h.systemRoot(c)
	if !ok {
		return
	}
	rel := c.Param("path")
	abs, safe := SafePath(root, rel)
	if !safe {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid path"})
		return
	}
	if err := os.MkdirAll(abs, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "mkdir failed"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"created": true})
}

// DeleteSystemFile DELETE /api/v1/system/files/*path
func (h *FileHandler) DeleteSystemFile(c *gin.Context) {
	root, ok := h.systemRoot(c)
	if !ok {
		return
	}
	rel := c.Param("path")
	abs, safe := SafePath(root, rel)
	if !safe {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid path"})
		return
	}
	if err := os.RemoveAll(abs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

// RenameSystemFile PATCH /api/v1/system/files/*path
func (h *FileHandler) RenameSystemFile(c *gin.Context) {
	root, ok := h.systemRoot(c)
	if !ok {
		return
	}
	rel := c.Param("path")
	abs, safe := SafePath(root, rel)
	if !safe {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid path"})
		return
	}
	var body struct {
		NewPath string `json:"newPath" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "newPath required"})
		return
	}
	newAbs, safe := SafePath(root, body.NewPath)
	if !safe {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid destination"})
		return
	}
	if err := os.Rename(abs, newAbs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "rename failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"renamed": true})
}
