package v1

type PatientRegisteredV1 struct {
	PatientID string `json:"patient_id"`
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
}
