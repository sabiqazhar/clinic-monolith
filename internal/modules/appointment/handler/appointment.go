package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/appointment/domain"
	"go.uber.org/zap"
)

type AppointmentHandler struct {
	svc domain.AppointmentService
	log *zap.Logger
}

func NewAppointmentHandler(svc domain.AppointmentService, log *zap.Logger) *AppointmentHandler {
	return &AppointmentHandler{svc: svc, log: log}
}

func (h *AppointmentHandler) RegisterRoutes(g *gin.RouterGroup) {
	g.GET("/:id", h.GetAppointment)
	g.POST("/", h.Schedule)
	g.DELETE("/:id", h.Cancel)
}

func (h *AppointmentHandler) GetAppointment(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	appt, err := h.svc.GetAppointment(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrAppointmentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": appt})
}

func (h *AppointmentHandler) Schedule(c *gin.Context) {
	var req struct {
		PatientID   string    `json:"patient_id" binding:"required"`
		DoctorID    string    `json:"doctor_id" binding:"required"`
		ScheduledAt time.Time `json:"scheduled_at" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	appt, err := h.svc.Schedule(ctx, req.PatientID, req.DoctorID, req.ScheduledAt)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidPatient) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid patient"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to schedule"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": appt})
}

func (h *AppointmentHandler) Cancel(c *gin.Context) {
	id := c.Param("id")
	err := h.svc.CancelAppointment(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "cancelled"})
}
