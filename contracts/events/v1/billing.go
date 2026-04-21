package v1

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// InvoiceCreatedV1 adalah kontrak event saat invoice baru dibuat.
// Event ini akan dipublikasikan ke RabbitMQ dengan topic: "app.billing.invoice.created.v1"
type InvoiceV1 struct {
	InvoiceID   string    `json:"invoice_id"`
	PatientID   string    `json:"patient_id"`
	Amount      float64   `json:"amount"`
	Status      string    `json:"status"` // e.g., "pending", "paid"
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type InvoiceCreatedV1 struct {
	InvoiceID   string             `json:"invoice_id"`
	PatientID   string             `json:"patient_id"`
	Amount      pgtype.Numeric     `json:"amount"`
	Status      string             `json:"status"` // e.g., "pending", "paid"
	Description string             `json:"description"`
	CreatedAt   pgtype.Timestamptz `json:"created_at"`
}
