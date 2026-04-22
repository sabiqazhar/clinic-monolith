package v1

import "time"

type AppointmentScheduledV1 struct {
	AppointmentID string    `json:"appointment_id"`
	PatientID     string    `json:"patient_id"`
	DoctorID      string    `json:"doctor_id"`
	ScheduledAt   time.Time `json:"scheduled_at"`
}
