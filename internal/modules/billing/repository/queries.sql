-- name: GetInvoiceByID :one
SELECT id, patient_id, amount, status, description, created_at 
FROM billing_invoices 
WHERE id = $1;

-- name: CreateInvoice :exec
INSERT INTO billing_invoices (id, patient_id, amount, status, description, created_at) 
VALUES ($1, $2, $3, $4, $5, NOW());

-- name: InsertOutboxEvent :exec
INSERT INTO outbox_events (id, topic, payload, status, created_at) 
VALUES ($1, $2, $3, 'pending', NOW());
