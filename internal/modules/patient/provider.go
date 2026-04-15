package patient

import (
	"github.com/goforj/wire"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/patient/handler"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/patient/repository"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/patient/service"
)

var PatientSet = wire.NewSet(
	repository.NewPatientRepo,
	service.NewPatientService,
	handler.NewPatientHandler,
)
