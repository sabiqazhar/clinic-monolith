package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	v1 "github.com/sabiqazhar/clinic-monolith/contracts/events/v1"
	"github.com/sabiqazhar/clinic-monolith/helper"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/billing/domain"
	patientdomain "github.com/sabiqazhar/clinic-monolith/internal/modules/patient/domain"
	"go.uber.org/zap"
)

type billingService struct {
	repo       domain.InvoiceRepository
	patientSvc patientdomain.PatientService // 🔵 SYNC COMM: inject interface patient
	cache      domain.CacheManager
	pub        domain.EventPublisher
	log        *zap.Logger
}

func NewBillingService(
	repo domain.InvoiceRepository,
	patientSvc patientdomain.PatientService,
	cache domain.CacheManager,
	pub domain.EventPublisher,
	log *zap.Logger,
) domain.BillingService {
	return &billingService{
		repo:       repo,
		patientSvc: patientSvc,
		cache:      cache,
		pub:        pub,
		log:        log,
	}
}

// GetInvoice → Cache-Aside Pattern (sama seperti patient)
func (s *billingService) GetInvoice(ctx context.Context, id string) (*domain.Invoice, error) {
	cacheKey := fmt.Sprintf("app:billing:invoice:%s", id)

	// 1. Cache Lookup
	if data, err := s.cache.Get(ctx, cacheKey); err == nil {
		var inv domain.Invoice
		if err := json.Unmarshal(data, &inv); err == nil {
			return &inv, nil
		}
	}

	// 2. Cache Miss → Fallback DB
	invoice, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("invoice lookup failed: %w", err)
	}

	// 3. Async Cache Warming (Fire & Forget)
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		data, _ := json.Marshal(invoice)
		_ = s.cache.Set(bgCtx, cacheKey, data, 15*time.Minute)
	}()

	return invoice, nil
}

// GenerateInvoice → Validasi patient via SYNC call + SaveWithOutbox
func (s *billingService) GenerateInvoice(ctx context.Context, patientID string, amount float64, description string) (*domain.Invoice, error) {
	// 🔵 1. SYNC COMM: Validasi patient via interface (bukan import service patient!)
	if _, err := s.patientSvc.GetProfile(ctx, patientID); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidPatient, err)
	}

	inv := &domain.Invoice{
		ID:          uuid.New().String(),
		PatientID:   patientID,
		Amount:      amount,
		Status:      "pending",
		Description: description,
		CreatedAt:   time.Now(),
	}

	// 🔵 2. ATOMIC SAVE: Insert invoice + outbox dalam 1 transaksi
	if err := s.repo.CreateWithOutbox(ctx, inv); err != nil {
		return nil, fmt.Errorf("failed to generate invoice: %w", err)
	}

	// 🔵 3. ASYNC PUBLISH: Event akan dikirim oleh OutboxRelay
	payload, _ := json.Marshal(v1.InvoiceCreatedV1{
		InvoiceID: inv.ID,
		PatientID: inv.PatientID,
		Amount:    helper.ToPgNumeric(inv.Amount),
	})
	s.pub.PublishEventAsync(ctx, "app.billing.invoice.created.v1", payload)

	return inv, nil
}
