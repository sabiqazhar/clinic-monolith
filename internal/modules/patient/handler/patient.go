package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/patient/domain"
	"go.uber.org/zap"
)

type PatientHandler struct {
	svc domain.PatientService
	log *zap.Logger
}

// NewPatientHandler adalah provider untuk Wire
func NewPatientHandler(svc domain.PatientService, log *zap.Logger) *PatientHandler {
	return &PatientHandler{svc: svc, log: log}
}

// RegisterRoutes mendaftarkan endpoint ke Gin RouterGroup
func (h *PatientHandler) RegisterRoutes(g *gin.RouterGroup) {
	g.GET("/:id", h.GetProfile)
	g.POST("/", h.Register)
}

// GET /api/v1/patients/:id
func (h *PatientHandler) GetProfile(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing patient id"})
		return
	}

	ctx := c.Request.Context()
	patient, err := h.svc.GetProfile(ctx, id)
	if err != nil {
		// Mapping error domain ke HTTP status
		if errors.Is(err, domain.ErrPatientNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "patient not found"})
			return
		}
		h.log.Error("failed to get patient profile",
			zap.String("id", id),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": patient})
}

// POST /api/v1/patients
func (h *PatientHandler) Register(c *gin.Context) {
	var req struct {
		FullName string `json:"full_name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
	}

	// Validasi payload HTTP
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload", "details": err.Error()})
		return
	}

	ctx := c.Request.Context()
	patient, err := h.svc.Register(ctx, req.FullName, req.Email)
	if err != nil {
		h.log.Error("failed to register patient",
			zap.String("email", req.Email),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register patient"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": patient})
}
