package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	v1 "github.com/sabiqazhar/clinic-monolith/contracts/events/v1"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/appointment/domain"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/appointment/repository/query"
	"go.uber.org/zap"
)

type mysqlRepo struct {
	db  *sql.DB
	q   *query.Queries
	log *zap.Logger
}

func NewAppointmentRepo(db *sql.DB, log *zap.Logger) domain.AppointmentRepository {
	return &mysqlRepo{
		db:  db,
		q:   query.New(db),
		log: log,
	}
}

func (r *mysqlRepo) GetByID(ctx context.Context, id string) (*domain.Appointment, error) {
	row, err := r.q.GetAppointmentByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrAppointmentNotFound
		}
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return &domain.Appointment{
		ID:          row.ID,
		PatientID:   row.PatientID,
		DoctorID:    row.DoctorID,
		ScheduledAt: row.ScheduledAt,
	}, nil
}

func (r *mysqlRepo) CreateWithOutbox(ctx context.Context, appt *domain.Appointment) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx failed: %w", err)
	}
	defer tx.Rollback()

	qTx := query.New(tx)

	// 1. Insert appointment
	if err := qTx.InsertAppointment(ctx, query.InsertAppointmentParams{
		ID:          appt.ID,
		PatientID:   appt.PatientID,
		DoctorID:    appt.DoctorID,
		ScheduledAt: appt.ScheduledAt,
	}); err != nil {
		return fmt.Errorf("insert appointment failed: %w", err)
	}

	// 2. Insert outbox event (ATOMIC)
	payload, err := json.Marshal(v1.AppointmentScheduledV1{
		AppointmentID: appt.ID,
		PatientID:     appt.PatientID,
		DoctorID:      appt.DoctorID,
		ScheduledAt:   appt.ScheduledAt,
	})
	if err != nil {
		return fmt.Errorf("marshal event failed: %w", err)
	}

	if err := qTx.InsertOutboxEvent(ctx, query.InsertOutboxEventParams{
		ID:      uuid.New().String(),
		Topic:   "app.appointment.scheduled.v1",
		Payload: payload,
	}); err != nil {
		return fmt.Errorf("insert outbox failed: %w", err)
	}

	return tx.Commit()
}

func (r *mysqlRepo) Cancel(ctx context.Context, id string) error {
	return r.q.CancelAppointment(ctx, id)
}
