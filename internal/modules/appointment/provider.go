package appointment

import (
	"github.com/goforj/wire"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/appointment/handler"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/appointment/repository"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/appointment/service"
)

var AppointmentSet = wire.NewSet(
	repository.NewAppointmentRepo,
	service.NewAppointmentService,
	handler.NewAppointmentHandler,
)
