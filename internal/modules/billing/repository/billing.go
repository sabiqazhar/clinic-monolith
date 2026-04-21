package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	v1 "github.com/sabiqazhar/clinic-monolith/contracts/events/v1"
	"github.com/sabiqazhar/clinic-monolith/helper"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/billing/domain"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/billing/repository/query"
	"go.uber.org/zap"
)

type pgBillingRepo struct {
	db  *pgxpool.Pool
	q   *query.Queries
	log *zap.Logger
}

func NewBillingRepo(db *pgxpool.Pool, log *zap.Logger) domain.InvoiceRepository {
	return &pgBillingRepo{
		db:  db,
		q:   query.New(db),
		log: log,
	}
}

func (r *pgBillingRepo) GetByID(ctx context.Context, id string) (*domain.Invoice, error) {
	row, err := r.q.GetInvoiceByID(ctx, id)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, domain.ErrInvoiceNotFound
		}
		return nil, fmt.Errorf("query failed: %w", err)
	}

	amount, _ := row.Amount.Float64Value()

	return &domain.Invoice{
		ID:          row.ID,
		PatientID:   row.PatientID,
		Amount:      amount.Float64,
		Status:      row.Status.String,
		Description: row.Description.String,
		CreatedAt:   row.CreatedAt.Time,
	}, nil
}

func (r *pgBillingRepo) CreateWithOutbox(ctx context.Context, inv *domain.Invoice) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx failed: %w", err)
	}
	defer tx.Rollback(ctx)

	qTx := query.New(tx)

	// 1. Insert invoice
	if err := qTx.CreateInvoice(ctx, query.CreateInvoiceParams{
		ID:          inv.ID,
		PatientID:   inv.PatientID,
		Amount:      helper.ToPgNumeric(inv.Amount),
		Status:      helper.ToPgText(inv.Status),
		Description: helper.ToPgText(inv.Description),
	}); err != nil {
		return fmt.Errorf("insert invoice failed: %w", err)
	}

	// 2. Insert outbox event (ATOMIC dalam transaksi yang sama)
	payload, err := json.Marshal(v1.InvoiceCreatedV1{
		InvoiceID: inv.ID,
		PatientID: inv.PatientID,
		Amount:    helper.ToPgNumeric(inv.Amount),
		CreatedAt: helper.ToPgTime(time.Now()),
	})
	if err != nil {
		return fmt.Errorf("marshal event failed: %w", err)
	}

	if err := qTx.InsertOutboxEvent(ctx, query.InsertOutboxEventParams{
		ID:      uuid.New().String(),
		Topic:   "app.billing.invoice.created.v1",
		Payload: payload,
	}); err != nil {
		return fmt.Errorf("insert outbox failed: %w", err)
	}

	return tx.Commit(ctx)
}
