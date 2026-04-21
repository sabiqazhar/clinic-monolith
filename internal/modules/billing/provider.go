package billing

import (
	"github.com/goforj/wire"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/billing/handler"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/billing/repository"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/billing/service"
)

var BillingSet = wire.NewSet(
	repository.NewBillingRepo,
	service.NewBillingService,
	handler.NewBillingHandler,
)
