package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"

	"github.com/vikukumar/Pushpaka/internal/middleware"
	"github.com/vikukumar/Pushpaka/internal/repositories"
)

var termUpgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type TerminalHandler struct {
	deploymentRepo *repositories.DeploymentRepository
}

func NewTerminalHandler(deploymentRepo *repositories.DeploymentRepository) *TerminalHandler {
	return &TerminalHandler{deploymentRepo: deploymentRepo}
}

// Connect opens an interactive terminal into a running deployment's container.
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

	if deployment.ContainerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no container running for this deployment"})
		return
	}

	conn, err := termUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error().Err(err).Msg("terminal websocket upgrade failed")
		return
	}
	defer conn.Close()

	containerID := deployment.ContainerID
	shell := detectShell(containerID)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	args := []string{"exec", "-it", containerID, shell}
	cmd := exec.CommandContext(ctx, "docker", args...)

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
		writeWSError(conn, fmt.Sprintf("docker exec failed: %v", err))
		return
	}
	defer cmd.Process.Kill() //nolint:errcheck

	// Forward container stdout -> WebSocket
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

	// Forward WebSocket input -> container stdin
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

// detectShell tries /bin/bash then /bin/sh to find an available shell.
func detectShell(containerID string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "docker", "exec", containerID, "which", "bash").Output()
	if err == nil && strings.TrimSpace(string(out)) != "" {
		return "/bin/bash"
	}
	return "/bin/sh"
}

func writeWSError(conn *websocket.Conn, msg string) {
	conn.WriteMessage(websocket.TextMessage, []byte("\r\n\x1b[31mError: "+msg+"\x1b[0m\r\n")) //nolint:errcheck
}
