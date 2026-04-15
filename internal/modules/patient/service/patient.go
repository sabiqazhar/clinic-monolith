package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/patient/domain"
	"go.uber.org/zap"
)

type patientService struct {
	repo  domain.PatientRepository
	cache domain.CacheManager
	pub   domain.EventPublisher
	log   *zap.Logger
}

// ini adalah provider yg nanti dipake sama wire
func NewPatientService(
	repo domain.PatientRepository,
	cache domain.CacheManager,
	pub domain.EventPublisher,
	log *zap.Logger,
) domain.PatientService {
	return &patientService{
		repo:  repo,
		cache: cache,
		pub:   pub,
		log:   log,
	}
}

func (s *patientService) GetProfile(ctx context.Context, id string) (*domain.Patient, error) {
	cacheKey := fmt.Sprintf("app:patient:profile:%s", id)

	// 1. cache lookup
	if data, err := s.cache.Get(ctx, cacheKey); err == nil {
		var p domain.Patient
		if err := json.Unmarshal(data, &p); err == nil {
			return &p, nil
		}
	}

	// 2. cache miss -> Fallback ke DB
	patient, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("patient lookup failed:%w", err)
	}

	// 3. Async Cache Warming (Fire & Forget Goroutine)
	// Blueprint: "go func() { ... }() secara asinkron menggunakan goroutine"
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		data, marshalErr := json.Marshal(patient)
		if marshalErr != nil {
			s.log.Error("fail to marshall patient for cache", zap.Error(marshalErr))
			return
		}

		// set ttl 15 menit
		_ = s.cache.Set(bgCtx, cacheKey, data, 15*time.Minute)
	}()

	return patient, nil
}

// Register mengorchestrasi penyimpanan + outbox
func (s *patientService) Register(ctx context.Context, fullName, email string) (*domain.Patient, error) {
	p := &domain.Patient{
		ID:        uuid.New().String(),
		FullName:  fullName,
		Email:     email,
		CreatedAt: time.Now(),
	}

	// Repo menangani INSERT patient + INSERT outbox_events dalam 1 transaksi atomik.
	// Service hanya memicu, tidak tahu detail SQL-nya (isolasi ketat).
	if err := s.repo.SaveWithOutbox(ctx, p); err != nil {
		return nil, fmt.Errorf("failed to register patient: %w", err)
	}

	// Catatan: Publish event sebenarnya sudah di-handle oleh Outbox Relay Worker.
	// Jika ingin publish langsung ke broker (skip outbox), panggil:
	// s.pub.PublishEventAsync(ctx, "app.patient.registered.v1", payload)

	return p, nil
}
