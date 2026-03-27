package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"github.com/vikukumar/pushpaka/internal/config"
	"github.com/vikukumar/pushpaka/internal/middleware"
	"github.com/vikukumar/pushpaka/internal/repositories"
)

var editorUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// EditorWSMessage defines the schema for editor websocket events.
type EditorWSMessage struct {
	Type      string      `json:"type"`
	Workspace string      `json:"workspace"` // "system" or project_id
	Path      string      `json:"path,omitempty"`
	Content   string      `json:"content,omitempty"`
	User      string      `json:"user,omitempty"`
	Cursor    interface{} `json:"cursor,omitempty"`
}

type EditorWSHandler struct {
	projectRepo    *repositories.ProjectRepository
	deploymentRepo *repositories.DeploymentRepository
	deployDir      string
	cloneDir       string

	// Presence tracking: workspaceID -> map[conn]userID
	rooms sync.Map
}

func NewEditorWSHandler(projectRepo *repositories.ProjectRepository, deploymentRepo *repositories.DeploymentRepository, cfg *config.Config) *EditorWSHandler {
	return &EditorWSHandler{
		projectRepo:    projectRepo,
		deploymentRepo: deploymentRepo,
		deployDir:      cfg.DeploysDir,
		cloneDir:       cfg.CloneDir,
	}
}

func (h *EditorWSHandler) Connect(c *gin.Context) {
	userID := middleware.GetUserID(c)

	conn, err := editorUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error().Err(err).Msg("editor websocket upgrade failed")
		return
	}
	defer conn.Close()

	var currentWorkspace string

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var event EditorWSMessage
		if err := json.Unmarshal(msg, &event); err != nil {
			continue
		}

		switch event.Type {
		case "join":
			// Handle joining a workspace
			currentWorkspace = event.Workspace
			h.joinRoom(currentWorkspace, conn, userID)
			log.Debug().Str("user", userID).Str("workspace", currentWorkspace).Msg("user joined editor workspace")

		case "file:write":
			// Real-time file write and sync
			if currentWorkspace == "" {
				continue
			}
			h.handleFileWrite(conn, userID, currentWorkspace, event)

		case "cursor:move":
			// Broadcast cursor presence
			if currentWorkspace == "" {
				continue
			}
			h.broadcastToRoom(currentWorkspace, conn, event)
		}
	}

	if currentWorkspace != "" {
		h.leaveRoom(currentWorkspace, conn)
	}
}

func (h *EditorWSHandler) joinRoom(workspaceID string, conn *websocket.Conn, userID string) {
	room, _ := h.rooms.LoadOrStore(workspaceID, &sync.Map{})
	room.(*sync.Map).Store(conn, userID)
}

func (h *EditorWSHandler) leaveRoom(workspaceID string, conn *websocket.Conn) {
	if room, ok := h.rooms.Load(workspaceID); ok {
		room.(*sync.Map).Delete(conn)
	}
}

func (h *EditorWSHandler) broadcastToRoom(workspaceID string, sender *websocket.Conn, msg EditorWSMessage) {
	if room, ok := h.rooms.Load(workspaceID); ok {
		room.(*sync.Map).Range(func(key, value interface{}) bool {
			targetConn := key.(*websocket.Conn)
			if targetConn != sender {
				_ = targetConn.WriteJSON(msg)
			}
			return true
		})
	}
}

func (h *EditorWSHandler) handleFileWrite(conn *websocket.Conn, userID, workspaceID string, msg EditorWSMessage) {
	// Resolve root directory
	var root string
	if workspaceID == "system" {
		root = h.deployDir
	} else {
		// Resolve project dir (simplified check for WS)
		proj, err := h.projectRepo.FindByID(workspaceID, userID)
		if err != nil {
			_ = conn.WriteJSON(gin.H{"type": "error", "message": "unauthorized"})
			return
		}

		// Priority: deployDir/<projectID[:8]> → cloneDir/<latestDeploymentID>
		deployPath := filepath.Join(h.deployDir, proj.ID[:8])
		if _, err := os.Stat(deployPath); err == nil {
			root = deployPath
		} else if h.cloneDir != "" {
			deployment, err := h.deploymentRepo.FindLatestByProjectID(proj.ID)
			if err == nil {
				clonePath := filepath.Join(h.cloneDir, deployment.ID)
				if _, err := os.Stat(clonePath); err == nil {
					root = clonePath
				}
			}
		}

		if root == "" {
			_ = conn.WriteJSON(gin.H{"type": "error", "message": "workspace root not found — deploy the project first"})
			return
		}
	}

	// Safety check for path
	absTarget, ok := SafePath(root, msg.Path)
	if !ok {
		_ = conn.WriteJSON(gin.H{"type": "error", "message": "invalid path"})
		return
	}

	// Perform write
	if err := os.WriteFile(absTarget, []byte(msg.Content), 0644); err != nil {
		_ = conn.WriteJSON(gin.H{"type": "error", "message": "write failed: " + err.Error()})
		return
	}

	// Broadcast update to other users in the same workspace
	updateMsg := EditorWSMessage{
		Type:      "file:update",
		Workspace: workspaceID,
		Path:      msg.Path,
		Content:   msg.Content,
		User:      userID,
	}
	h.broadcastToRoom(workspaceID, conn, updateMsg)
}
