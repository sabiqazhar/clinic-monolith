-- name: FindPatientByID :one
SELECT id, full_name, email, created_at 
FROM patients 
WHERE id = $1 AND deleted_at IS NULL;

-- name: InsertPatient :exec
INSERT INTO patients (id, full_name, email, created_at) 
VALUES ($1, $2, $3, NOW());

-- name: InsertOutboxEvent :exec
INSERT INTO outbox_events (id, topic, payload, status, created_at) 
VALUES ($1, $2, $3, 'pending', NOW());
