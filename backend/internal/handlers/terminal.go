package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"

	"github.com/vikukumar/pushpaka/internal/config"
	"github.com/vikukumar/pushpaka/internal/middleware"
	"github.com/vikukumar/pushpaka/internal/repositories"
)

var termUpgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		allowed := os.Getenv("ALLOWED_ORIGINS")
		if allowed == "" || allowed == "*" {
			return true
		}
		for _, a := range strings.Split(allowed, ",") {
			if strings.TrimSpace(a) == origin {
				return true
			}
		}
		return false
	},
}

type TerminalHandler struct {
	deploymentRepo *repositories.DeploymentRepository
	deployDir      string
	cloneDir       string
}

func NewTerminalHandler(deploymentRepo *repositories.DeploymentRepository, cfg *config.Config) *TerminalHandler {
	return &TerminalHandler{
		deploymentRepo: deploymentRepo,
		deployDir:      cfg.DeploysDir,
		cloneDir:       cfg.CloneDir,
	}
}

// Connect opens an interactive terminal into a running deployment's container,
// or falls back to a shell in the deployment's local working directory if
// Docker is not available for this deployment.
//
// GET /api/v1/deployments/:id/terminal  (WebSocket upgrade)
//
// Authentication: JWT via ?token= query parameter (WebSocket cannot send headers).
func (h *TerminalHandler) Connect(c *gin.Context) {
	userID := middleware.GetUserID(c)
	deploymentID := c.Param("id")

	deployment, err := h.deploymentRepo.FindByID(deploymentID)
	if err != nil || deployment.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
		return
	}

	conn, err := termUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error().Err(err).Msg("terminal websocket upgrade failed")
		return
	}
	defer conn.Close()

	if deployment.ContainerID != "" {
		// ── Docker path ──────────────────────────────────────────────────────
		h.dockerTerminal(conn, deployment.ContainerID)
	} else {
		// ── Local fallback ────────────────────────────────────────────────────
		h.localTerminal(conn, deployment.ProjectID)
	}
}

// dockerTerminal connects the WebSocket to a running Docker container.
func (h *TerminalHandler) dockerTerminal(conn *websocket.Conn, containerID string) {
	shell := detectShell(containerID)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	args := []string{"exec", "-it", containerID, shell}
	cmd := exec.CommandContext(ctx, "docker", args...)
	h.pipeCmd(conn, cmd)
}

// localTerminal falls back to a shell in the project's running directory.
// This is used when Docker is not available (e.g. direct Node/Python/Go deployments).
func (h *TerminalHandler) localTerminal(conn *websocket.Conn, projectID string) {
	// 1. Try the permanent deploy directory (Worker copies files to deployDir/<projectID[:8]>)
	workDir := filepath.Join(h.deployDir, projectID[:8])
	if _, err := os.Stat(workDir); err != nil {
		// 2. Fall back to the latest deployment's clone directory (common on Windows)
		if h.cloneDir != "" {
			latest, lErr := h.deploymentRepo.FindLatestByProjectID(projectID)
			if lErr == nil {
				clonePath := filepath.Join(h.cloneDir, latest.ID)
				if _, sErr := os.Stat(clonePath); sErr == nil {
					workDir = clonePath
					goto proceed
				}
			}
		}
		writeWSError(conn, "no deployed directory found — deploy the project first")
		return
	}

proceed:
	// Announce the fallback mode to the user
	banner := fmt.Sprintf(
		"\r\n\x1b[33m⚠  No container — connected to local deploy directory: %s\x1b[0m\r\n\r\n",
		workDir,
	)
	conn.WriteMessage(websocket.BinaryMessage, []byte(banner)) //nolint:errcheck

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd.exe")
	} else {
		shell := localShell()
		cmd = exec.CommandContext(ctx, shell)
	}
	cmd.Dir = workDir
	h.pipeCmd(conn, cmd)
}

// pipeCmd wires a command's stdin/stdout to the WebSocket.
func (h *TerminalHandler) pipeCmd(conn *websocket.Conn, cmd *exec.Cmd) {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		writeWSError(conn, "failed to create stdin pipe: "+err.Error())
		return
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		writeWSError(conn, "failed to create stdout pipe: "+err.Error())
		return
	}
	cmd.Stderr = cmd.Stdout // merge stderr into stdout

	if err := cmd.Start(); err != nil {
		writeWSError(conn, fmt.Sprintf("shell start failed: %v", err))
		return
	}
	defer cmd.Process.Kill() //nolint:errcheck

	// Forward stdout -> WebSocket
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				if werr := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); werr != nil {
					return
				}
			}
			if err != nil {
				if err != io.EOF {
					writeWSError(conn, err.Error())
				}
				return
			}
		}
	}()

	// Forward WebSocket input -> stdin
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break // client disconnected
		}
		if _, err := stdin.Write(msg); err != nil {
			break
		}
	}

	stdin.Close()
	cmd.Wait() //nolint:errcheck
}

// detectShell tries /bin/bash then /bin/sh inside a container.
func detectShell(containerID string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "docker", "exec", containerID, "which", "bash").Output()
	if err == nil && strings.TrimSpace(string(out)) != "" {
		return "/bin/bash"
	}
	return "/bin/sh"
}

// localShell returns the best available interactive shell on the host.
func localShell() string {
	for _, sh := range []string{"/bin/bash", "/bin/sh", "/usr/bin/bash", "/usr/bin/sh"} {
		if _, err := os.Stat(sh); err == nil {
			return sh
		}
	}
	return "/bin/sh"
}

func writeWSError(conn *websocket.Conn, msg string) {
	conn.WriteMessage(websocket.TextMessage, []byte("\r\n\x1b[31mError: "+msg+"\x1b[0m\r\n")) //nolint:errcheck
}
