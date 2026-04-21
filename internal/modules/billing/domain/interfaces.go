package domain

import (
	"context"
	"errors"
	"time"
)

// Domain Errors
var (
	ErrInvoiceNotFound = errors.New("invoice not found")
	ErrInvalidPatient  = errors.New("invalid patient reference")
)

// Entity
type Invoice struct {
	ID          string
	PatientID   string
	Amount      float64
	Status      string
	Description string
	CreatedAt   time.Time
}

// Repository Interface
type InvoiceRepository interface {
	GetByID(ctx context.Context, id string) (*Invoice, error)
	CreateWithOutbox(ctx context.Context, inv *Invoice) error
}

// Infrastructure Interfaces
type CacheManager interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
}

type EventPublisher interface {
	PublishEventAsync(ctx context.Context, topic string, payload []byte)
}

// Service Interface (Public Contract)
type BillingService interface {
	GetInvoice(ctx context.Context, id string) (*Invoice, error)
	GenerateInvoice(ctx context.Context, patientID string, amount float64, description string) (*Invoice, error)
}
