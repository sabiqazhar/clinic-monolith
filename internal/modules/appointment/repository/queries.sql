-- name: GetAppointmentByID :one
SELECT id, patient_id, doctor_id, scheduled_at, status, created_at 
FROM appointments 
WHERE id = ?;

-- name: InsertAppointment :exec
INSERT INTO appointments (id, patient_id, doctor_id, scheduled_at, status, created_at) 
VALUES (?, ?, ?, ?, 'scheduled', NOW());

-- name: InsertOutboxEvent :exec
INSERT INTO outbox_events (id, topic, payload, status, created_at) 
VALUES (?, ?, ?, 'pending', NOW());
