package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vikukumar/pushpaka/internal/middleware"
	"github.com/vikukumar/pushpaka/internal/repositories"
	"github.com/vikukumar/pushpaka/pkg/models"
)

type EditorStateHandler struct {
	repo *repositories.EditorStateRepository
}

func NewEditorStateHandler(repo *repositories.EditorStateRepository) *EditorStateHandler {
	return &EditorStateHandler{repo: repo}
}

func (h *EditorStateHandler) GetState(c *gin.Context) {
	userID := middleware.GetUserID(c)
	projectID := c.Param("id")

	state, err := h.repo.Get(userID, projectID)
	if err != nil {
		// If not found, return empty state instead of error
		c.JSON(http.StatusOK, gin.H{
			"open_tabs":  "[]",
			"active_tab": "",
			"sidebar":    "{}",
		})
		return
	}

	c.JSON(http.StatusOK, state)
}

func (h *EditorStateHandler) SaveState(c *gin.Context) {
	userID := middleware.GetUserID(c)
	projectID := c.Param("id")

	var body struct {
		OpenTabs  string `json:"open_tabs"`
		ActiveTab string `json:"active_tab"`
		Sidebar   string `json:"sidebar"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	state := &models.UserEditorState{
		UserID:    userID,
		ProjectID: projectID,
		OpenTabs:  body.OpenTabs,
		ActiveTab: body.ActiveTab,
		Sidebar:   body.Sidebar,
	}

	if err := h.repo.Save(state); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save editor state"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"saved": true})
}
