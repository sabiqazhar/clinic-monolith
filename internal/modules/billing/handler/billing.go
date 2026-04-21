package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/billing/domain"
	"go.uber.org/zap"
)

type BillingHandler struct {
	svc domain.BillingService
	log *zap.Logger
}

func NewBillingHandler(svc domain.BillingService, log *zap.Logger) *BillingHandler {
	return &BillingHandler{svc: svc, log: log}
}

func (h *BillingHandler) RegisterRoutes(g *gin.RouterGroup) {
	g.GET("/:id", h.GetInvoice)
	g.POST("/", h.GenerateInvoice)
}

// GET /api/v1/billing/:id
func (h *BillingHandler) GetInvoice(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing invoice id"})
		return
	}

	ctx := c.Request.Context()
	invoice, err := h.svc.GetInvoice(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrInvoiceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
			return
		}
		h.log.Error("failed to get invoice",
			zap.String("id", id),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": invoice})
}

// POST /api/v1/billing
func (h *BillingHandler) GenerateInvoice(c *gin.Context) {
	var req struct {
		PatientID   string  `json:"patient_id" binding:"required"`
		Amount      float64 `json:"amount" binding:"required,min=0"`
		Description string  `json:"description" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload", "details": err.Error()})
		return
	}

	ctx := c.Request.Context()
	invoice, err := h.svc.GenerateInvoice(ctx, req.PatientID, req.Amount, req.Description)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidPatient) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid patient", "details": err.Error()})
			return
		}
		h.log.Error("failed to generate invoice",
			zap.String("patient_id", req.PatientID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate invoice"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": invoice})
}
