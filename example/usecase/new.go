package usecase

import (
	"gitlab.com/gma-vietnam/tanca-connect/internal/device"
	"gitlab.com/gma-vietnam/tanca-connect/internal/element"
	"gitlab.com/gma-vietnam/tanca-connect/internal/event"
	"gitlab.com/gma-vietnam/tanca-connect/internal/event/delivery/rabbitmq/producer"
	"gitlab.com/gma-vietnam/tanca-connect/internal/event/repository"
	"gitlab.com/gma-vietnam/tanca-connect/internal/eventcategory"
	"gitlab.com/gma-vietnam/tanca-connect/internal/room"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/log"
	"gitlab.com/gma-vietnam/tanca-connect/pkg/microservice"
)

type implUseCase struct {
	l               log.Logger
	repo            repository.Repository
	deviceUC        device.UseCase
	elementUC       element.UseCase
	roomUC          room.UseCase
	shopUC          microservice.ShopUseCase
	eventcategoryUC eventcategory.UseCase
	producer        producer.Producer
}

var _ event.UseCase = implUseCase{}

func New(
	l log.Logger,
	repo repository.Repository,
	deviceUC device.UseCase,
	elementUC element.UseCase,
	roomUC room.UseCase,
	shopUC microservice.ShopUseCase,
	eventcategoryUC eventcategory.UseCase,
	producer producer.Producer,
) event.UseCase {
	return &implUseCase{
		l:               l,
		repo:            repo,
		deviceUC:        deviceUC,
		elementUC:       elementUC,
		roomUC:          roomUC,
		shopUC:          shopUC,
		eventcategoryUC: eventcategoryUC,
		producer:        producer,
	}
}
