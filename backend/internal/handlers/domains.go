package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/vikukumar/Pushpaka/internal/middleware"
	"github.com/vikukumar/Pushpaka/internal/services"
	"github.com/vikukumar/Pushpaka/pkg/models"
)

type DomainHandler struct {
	domainSvc *services.DomainService
}

func NewDomainHandler(domainSvc *services.DomainService) *DomainHandler {
	return &DomainHandler{domainSvc: domainSvc}
}

func (h *DomainHandler) Add(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req models.AddDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	domain, err := h.domainSvc.Add(userID, &req)
	if err != nil {
		if errors.Is(err, services.ErrDomainExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, services.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add domain"})
		return
	}
	c.JSON(http.StatusCreated, domain)
}

func (h *DomainHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	domains, err := h.domainSvc.List(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch domains"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": domains})
}

func (h *DomainHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := c.Param("id")

	if err := h.domainSvc.Delete(id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete domain"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "domain removed"})
}
