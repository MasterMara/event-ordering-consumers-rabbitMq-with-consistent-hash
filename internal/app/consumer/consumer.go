package consumer

import (
	"context"
	"encoding/json"
	"event-ordering-consumers-rabbitmq-with-consistent-hash/internal/app/consumer/service"
	"event-ordering-consumers-rabbitmq-with-consistent-hash/pkg/constant"
	"event-ordering-consumers-rabbitmq-with-consistent-hash/pkg/rabbitmq"
	"fmt"
	"log"
)

type ConsumerBuilder struct {
	log        log.Logger
	service    service.Service
	messageBus *rabbitmq.Client
	queueCount int //Maybe Change This Configuration
}

func NewConsumer(log log.Logger, service service.Service, messageBus *rabbitmq.Client, queueCount int) *ConsumerBuilder {

	return &ConsumerBuilder{
		log:        log,
		service:    service,
		messageBus: messageBus,
		queueCount: queueCount,
	}
}

func (c *ConsumerBuilder) RunConsumer(ctx context.Context) error {

	for i := 0; i < c.queueCount; i++ {

		consumer := c.messageBus.AddConsumerWithConsistentHashExchange(formatQueueName(constant.QueueName, i), "1", constant.QueueName)

		consumer.SubscribeExchange(constant.RoutingKey, rabbitmq.TopicExchange, constant.OrderLineCreatedExchange)
		consumer.SubscribeExchange(constant.RoutingKey, rabbitmq.TopicExchange, constant.OrderLineInProgressedExchange)
		consumer.SubscribeExchange(constant.RoutingKey, rabbitmq.TopicExchange, constant.OrderLineInTransittedExchange)
		consumer.SubscribeExchange(constant.RoutingKey, rabbitmq.TopicExchange, constant.OrderLineDeliveredExchange)

		consumer.WithSingleGoroutine(true)

		consumer.HandleConsumer(c.ConsumerOrderLineEvents())
	}

	return c.messageBus.RunConsumers(ctx)

}

func (c *ConsumerBuilder) ConsumerOrderLineEvents() func(message rabbitmq.Message) error {

	return func(message rabbitmq.Message) error {

		//Todo: Think This Business Logic with Publishers

		var (
			event OrderLineEvent
			err   error
		)

		err = json.Unmarshal(message.Payload, &event)
		if err != nil {
			return err
		}

		//Todo: Impelement the busuiness logic
		return err
	}

}

func formatQueueName(queueName string, queueOrder int) string {
	return fmt.Sprintf("%s-%d", queueName, queueOrder+1)
}
