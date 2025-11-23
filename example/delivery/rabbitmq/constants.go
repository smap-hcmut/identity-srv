package rabbitmq

import "gitlab.com/gma-vietnam/tanca-connect/pkg/rabbitmq"

const (
	CreateSystemEventExchangeName = "connect_create_system_event_exc"
	CreateSystemEventQueueName    = "connect_create_system_event"

	CreateNotificationExchangeName = "notification_create_exc"
	CreateNotificationQueueName    = "notification_create"

	UpdateRequestEventIDExcName   = "request_update_request_event_id_exc"
	UpdateRequestEventIDQueueName = "request_update_request_event_id"

	DeleteSystemEventExchangeName = "connect_delete_system_event_exc"
	DeleteSystemEventQueueName    = "connect_delete_system_event"

	UpdateTaskEventIDExchangeName = "hiring_task_update_event_id_exc"
	UpdateTaskEventIDQueueName    = "hiring_task_update_event_id"

	UpdateSystemEventExchangeName = "connect_update_system_event_exc"
	UpdateSystemEventQueueName    = "connect_update_system_event"
)

var (
	CreateSystemEventExchange = rabbitmq.ExchangeArgs{
		Name:       CreateSystemEventExchangeName,
		Type:       rabbitmq.ExchangeTypeFanout,
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
	}

	CreateNotificationExchange = rabbitmq.ExchangeArgs{
		Name:       CreateNotificationExchangeName,
		Type:       rabbitmq.ExchangeTypeFanout,
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
	}

	UpdateRequestEventIDExchange = rabbitmq.ExchangeArgs{
		Name:       UpdateRequestEventIDExcName,
		Type:       rabbitmq.ExchangeTypeFanout,
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
	}

	DeleteSystemEventExchange = rabbitmq.ExchangeArgs{
		Name:       DeleteSystemEventExchangeName,
		Type:       rabbitmq.ExchangeTypeFanout,
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
	}

	UpdateTaskEventIDExchange = rabbitmq.ExchangeArgs{
		Name:       UpdateTaskEventIDExchangeName,
		Type:       rabbitmq.ExchangeTypeFanout,
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
	}

	UpdateSystemEventExchange = rabbitmq.ExchangeArgs{
		Name:       UpdateSystemEventExchangeName,
		Type:       rabbitmq.ExchangeTypeFanout,
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
	}
)
