package domain

import (
	"context"
	"errors"
	"time"
)

// ── Domain Errors ──
var (
	ErrAppointmentNotFound = errors.New("appointment not found")
	ErrInvalidPatient      = errors.New("invalid patient reference")
	ErrInvalidDoctor       = errors.New("doctor not available")
	ErrTimeSlotTaken       = errors.New("time slot already booked")
)

// ── Entity ──
type Appointment struct {
	ID          string
	PatientID   string
	DoctorID    string
	ScheduledAt time.Time
	Status      string // "scheduled", "completed", "cancelled"
	CreatedAt   time.Time
}

// ── Repository Interface ──
type AppointmentRepository interface {
	GetByID(ctx context.Context, id string) (*Appointment, error)
	CreateWithOutbox(ctx context.Context, appt *Appointment) error
	Cancel(ctx context.Context, id string) error
}

// ── Infrastructure Interfaces (Shared) ──
type CacheManager interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
}

type EventPublisher interface {
	PublishEventAsync(ctx context.Context, topic string, payload []byte)
}

// ── Cross-Module Interface (Sync Call) ──
// Kita import interface dari patient domain untuk validasi
type PatientService interface {
	GetProfile(ctx context.Context, id string) (*struct{ ID string }, error) // Simplified for demo
}

// ── Service Interface ──
type AppointmentService interface {
	GetAppointment(ctx context.Context, id string) (*Appointment, error)
	Schedule(ctx context.Context, patientID, doctorID string, scheduledAt time.Time) (*Appointment, error)
	CancelAppointment(ctx context.Context, id string) error
}
