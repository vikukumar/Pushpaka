package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/vikukumar/pushpaka/internal/services"
)

type OAuthHandler struct {
	oauthSvc *services.OAuthService
	cfg      interface{ GithubEnabled() bool }
}

func NewOAuthHandler(oauthSvc *services.OAuthService) *OAuthHandler {
	return &OAuthHandler{oauthSvc: oauthSvc}
}

// GithubRedirect initiates the GitHub OAuth flow.
// GET /api/v1/auth/github
func (h *OAuthHandler) GithubRedirect(c *gin.Context) {
	state, err := h.oauthSvc.GenerateState("github", c.Query("redirect"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate state"})
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, h.oauthSvc.GithubAuthURL(state))
}

// GithubCallback handles the GitHub OAuth callback code exchange.
// GET /api/v1/auth/github/callback
func (h *OAuthHandler) GithubCallback(c *gin.Context) {
	if errParam := c.Query("error"); errParam != "" {
		c.Redirect(http.StatusTemporaryRedirect, "/login?error="+errParam)
		return
	}

	state := c.Query("state")
	if err := h.oauthSvc.ValidateState(state); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired OAuth state"})
		return
	}

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}

	resp, err := h.oauthSvc.ExchangeGithub(code)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/login?error=oauth_failed")
		return
	}

	// Redirect to frontend with token in query string.
	// The frontend reads it and stores it, then redirects to /dashboard.
	c.Redirect(http.StatusTemporaryRedirect,
		"/login?token="+resp.Token+"&oauth=1")
}

// GitlabRedirect initiates the GitLab OAuth flow.
// GET /api/v1/auth/gitlab
func (h *OAuthHandler) GitlabRedirect(c *gin.Context) {
	state, err := h.oauthSvc.GenerateState("gitlab", c.Query("redirect"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate state"})
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, h.oauthSvc.GitlabAuthURL(state))
}

// GitlabCallback handles the GitLab OAuth callback.
// GET /api/v1/auth/gitlab/callback
func (h *OAuthHandler) GitlabCallback(c *gin.Context) {
	if errParam := c.Query("error"); errParam != "" {
		c.Redirect(http.StatusTemporaryRedirect, "/login?error="+errParam)
		return
	}

	state := c.Query("state")
	if err := h.oauthSvc.ValidateState(state); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired OAuth state"})
		return
	}

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}

	resp, err := h.oauthSvc.ExchangeGitlab(code)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/login?error=oauth_failed")
		return
	}

	c.Redirect(http.StatusTemporaryRedirect,
		"/login?token="+resp.Token+"&oauth=1")
}
