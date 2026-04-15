package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	v1 "github.com/sabiqazhar/clinic-monolith/contracts/events/v1"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/patient/domain"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/patient/repository/query"
	"go.uber.org/zap"
)

type pgRepo struct {
	db  *pgxpool.Pool
	q   query.Querier
	log *zap.Logger
}

func NewPatientRepo(db *pgxpool.Pool, log *zap.Logger) domain.PatientRepository {
	return &pgRepo{
		db:  db,
		q:   query.New(db),
		log: log,
	}
}

func (r *pgRepo) FindByID(ctx context.Context, id string) (*domain.Patient, error) {
	row, err := r.q.FindPatientByID(ctx, id)
	if err != nil {
		if err.Error() == "no rows in result set" { // pgx pakai error string ini
			return nil, domain.ErrPatientNotFound
		}
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return &domain.Patient{
		ID:       row.ID,
		FullName: row.FullName,
		Email:    row.Email,
	}, nil
}

func (r *pgRepo) SaveWithOutbox(ctx context.Context, p *domain.Patient) error {
	// pgxpool.Begin() langsung return *pgx.Tx yang implement pgx.DBTX
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx failed: %w", err)
	}
	defer tx.Rollback(ctx)

	qTx := query.New(tx)

	if err := qTx.InsertPatient(ctx, query.InsertPatientParams{
		ID:       p.ID,
		FullName: p.FullName,
		Email:    p.Email,
	}); err != nil {
		return fmt.Errorf("insert patient failed: %w", err)
	}

	payload, _ := json.Marshal(v1.PatientRegisteredV1{
		PatientID: p.ID,
		FullName:  p.FullName,
		Email:     p.Email,
	})

	if err := qTx.InsertOutboxEvent(ctx, query.InsertOutboxEventParams{
		ID:      uuid.New().String(),
		Topic:   "app.patient.registered.v1",
		Payload: payload,
	}); err != nil {
		return fmt.Errorf("insert outbox failed: %w", err)
	}

	return tx.Commit(ctx) // Commit pakai context
}
