//go:build wireinject
// +build wireinject

package main

import (
	"context"
	"time"

	"github.com/goforj/wire"
	"go.uber.org/zap"

	"github.com/sabiqazhar/clinic-monolith/internal/infrastructure/broker"
	"github.com/sabiqazhar/clinic-monolith/internal/infrastructure/cache"
	"github.com/sabiqazhar/clinic-monolith/internal/infrastructure/db"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/appointment"
	appointmentdomain "github.com/sabiqazhar/clinic-monolith/internal/modules/appointment/domain"
	appointmenthandler "github.com/sabiqazhar/clinic-monolith/internal/modules/appointment/handler"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/billing"
	billingdomain "github.com/sabiqazhar/clinic-monolith/internal/modules/billing/domain"
	billinghandler "github.com/sabiqazhar/clinic-monolith/internal/modules/billing/handler"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/patient"
	patientdomain "github.com/sabiqazhar/clinic-monolith/internal/modules/patient/domain"
	patienthandler "github.com/sabiqazhar/clinic-monolith/internal/modules/patient/handler"
)

// Adapters: wrap concrete types to domain interfaces

type cacheAdapter struct{ *cache.RedisCache }

func (a *cacheAdapter) Get(ctx context.Context, key string) ([]byte, error) {
	return a.RedisCache.Get(ctx, key)
}

func (a *cacheAdapter) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return a.RedisCache.Set(ctx, key, value, ttl)
}

type publisherAdapter struct{ *broker.RabbitMQ }

func (p *publisherAdapter) PublishEventAsync(ctx context.Context, topic string, payload []byte) {
	_ = p.RabbitMQ.Publish(ctx, topic, payload)
}

// App adalah struct yang menampung semua handler yang sudah dirakit.
// Wire akan mengisi field-field ini secara otomatis.
type App struct {
	PatientHandler     *patienthandler.PatientHandler
	BillingHandler     *billinghandler.BillingHandler
	AppointmentHandler *appointmenthandler.AppointmentHandler
}

// InitializeApp is the INJECTOR FUNCTION.
// All config values are passed as parameters - Wire provides them from injector's args
func InitializeApp(
	pgDsn db.PGDsn,
	_ db.MySQLDsn, // TODO: uncomment when appointment module ready
	redisAddr cache.RedisAddr,
	rabbitURL broker.RabbitURL,
	logger *zap.Logger,
) (*App, error) {
	wire.Build(
		// Base providers
		db.NewPostgresPool,
		db.NewMySQLDB, // TODO: uncomment when appointment module ready
		cache.NewRedisClient,
		broker.NewRabbitMQ,

		// Adapters - must be explicitly provided to wire
		newCacheAdapter,
		newPublisherAdapter,

		// Interface adapters - bind domain interfaces to concrete types
		// Patient domain interfaces (both are identical, binding one set is sufficient)
		wire.Bind(new(patientdomain.CacheManager), new(*cacheAdapter)),
		wire.Bind(new(patientdomain.EventPublisher), new(*publisherAdapter)),
		// Billing domain interfaces
		wire.Bind(new(billingdomain.CacheManager), new(*cacheAdapter)),
		wire.Bind(new(billingdomain.EventPublisher), new(*publisherAdapter)),
		// Appointment domain interfaces
		wire.Bind(new(appointmentdomain.CacheManager), new(*cacheAdapter)),
		wire.Bind(new(appointmentdomain.EventPublisher), new(*publisherAdapter)),

		// Patient module provider set
		patient.PatientSet,
		billing.BillingSet,
		appointment.AppointmentSet,

		// App struct
		wire.Struct(new(App), "*"),
	)

	return nil, nil
}

// Provider functions for adapters
func newCacheAdapter(redis *cache.RedisCache) *cacheAdapter {
	return &cacheAdapter{redis}
}

func newPublisherAdapter(rabbit *broker.RabbitMQ) *publisherAdapter {
	return &publisherAdapter{rabbit}
}
