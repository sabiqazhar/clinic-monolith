package subscriber

import (
	"context"
	"encoding/json"

	v1 "github.com/sabiqazhar/clinic-monolith/contracts/events/v1"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/billing/domain"
	"go.uber.org/zap"
)

type PatientSubscriber struct {
	svc   domain.BillingService
	log   *zap.Logger
	topic string
}

func NewPatientSubscriber(svc domain.BillingService, log *zap.Logger) *PatientSubscriber {
	return &PatientSubscriber{
		svc:   svc,
		log:   log,
		topic: "app.patient.registered.v1",
	}
}

func (s *PatientSubscriber) Topic() string {
	return s.topic
}

func (s *PatientSubscriber) HandleEvent(ctx context.Context, payload []byte) error {
	var evt v1.PatientRegisteredV1
	if err := json.Unmarshal(payload, &evt); err != nil {
		s.log.Error("failed to unmarshal patient event", zap.Error(err))
		return err
	}

	// Auto-generate welcome invoice (contoh: amount 0 untuk welcome)
	_, err := s.svc.GenerateInvoice(ctx, evt.PatientID, 0.0, "Welcome invoice - auto generated")
	if err != nil {
		s.log.Error("failed to generate welcome invoice",
			zap.String("patient_id", evt.PatientID),
			zap.Error(err),
		)
		return err
	}

	s.log.Info("welcome invoice generated", zap.String("patient_id", evt.PatientID))
	return nil
}
