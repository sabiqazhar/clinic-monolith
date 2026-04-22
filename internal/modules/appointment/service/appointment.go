package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	v1 "github.com/sabiqazhar/clinic-monolith/contracts/events/v1"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/appointment/domain"
	patientdomain "github.com/sabiqazhar/clinic-monolith/internal/modules/patient/domain"
)

type appointmentService struct {
	repo       domain.AppointmentRepository
	patientSvc patientdomain.PatientService // 🔵 Sync Comm: Validasi Patient
	cache      domain.CacheManager          // Asumsi ada interface cache di domain atau infra
	pub        domain.EventPublisher
	log        *zap.Logger
}

func NewAppointmentService(
	repo domain.AppointmentRepository,
	patientSvc patientdomain.PatientService,
	cache domain.CacheManager,
	pub domain.EventPublisher,
	log *zap.Logger,
) domain.AppointmentService {
	return &appointmentService{
		repo:       repo,
		patientSvc: patientSvc,
		cache:      cache,
		pub:        pub,
		log:        log,
	}
}

func (s *appointmentService) GetAppointment(ctx context.Context, id string) (*domain.Appointment, error) {
	// Simple DB lookup for now (bisa tambah cache-aside nanti)
	appt, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get appointment: %w", err)
	}
	return appt, nil
}

func (s *appointmentService) Schedule(ctx context.Context, patientID, doctorID string, scheduledAt time.Time) (*domain.Appointment, error) {
	// 🔵 1. SYNC COMM: Validasi Patient
	if _, err := s.patientSvc.GetProfile(ctx, patientID); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidPatient, err)
	}

	// TODO: Validasi Doctor & Check Time Slot Availability (Business Logic)

	appt := &domain.Appointment{
		ID:          uuid.New().String(),
		PatientID:   patientID,
		DoctorID:    doctorID,
		ScheduledAt: scheduledAt,
		Status:      "scheduled",
		CreatedAt:   time.Now(),
	}

	// 🔵 2. ATOMIC SAVE: Insert appointment + outbox
	if err := s.repo.CreateWithOutbox(ctx, appt); err != nil {
		return nil, fmt.Errorf("failed to schedule: %w", err)
	}

	// 🔵 3. ASYNC PUBLISH: Event akan dikirim OutboxRelay
	payload, _ := json.Marshal(v1.AppointmentScheduledV1{
		AppointmentID: appt.ID,
		PatientID:     appt.PatientID,
		DoctorID:      appt.DoctorID,
		ScheduledAt:   appt.ScheduledAt,
	})
	s.pub.PublishEventAsync(ctx, "app.appointment.scheduled.v1", payload)

	return appt, nil
}

func (s *appointmentService) CancelAppointment(ctx context.Context, id string) error {
	return s.repo.Cancel(ctx, id)
}
