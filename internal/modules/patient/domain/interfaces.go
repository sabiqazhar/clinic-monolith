package domain

import (
	"context"
	"errors"
	"time"
)

// Domain Errors
var (
	ErrPatientNotFound = errors.New("patient not found")
)

// Entity
type Patient struct {
	ID        string
	FullName  string
	Email     string
	CreatedAt time.Time
}

// Repository Interface
// Implementasi ada di /repository. Hanya definisi kontrak.
type PatientRepository interface {
	FindByID(ctx context.Context, id string) (*Patient, error)
	SaveWithOutbox(ctx context.Context, p *Patient) error
}

// Infrastructure Interfaces
// Didefinisikan di domain agar modul tidak depend ke infrastructure langsung
type CacheManager interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
}

type EventPublisher interface {
	PublishEventAsync(ctx context.Context, topic string, payload []byte)
}

// Service Interface (Public Contract)
// Modul lain (appointment/billing) hanya boleh import ini via DI (Dependency Injection)
type PatientService interface {
	GetProfile(ctx context.Context, id string) (*Patient, error)
	Register(ctx context.Context, fullName, email string) (*Patient, error)
}
