package rabbitmq

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"
)

const (
	ErrorPrefix = "_error"
)

// Todo : Move go to Configs.
// Todo: And if you dont use other package start with lower letter like "a" not "A"
var (
	ConcurrentLimit = 1
	RetryCount      = 0
	PrefetchCount   = 1
	ReconnectDelay  = 2 * time.Second
)

type Client struct {
	context          context.Context
	shutDownFunction context.CancelFunc
	consumerBuilder  ConsumerBuilder
}

func (c *Client) RunConsumers(ctx context.Context) error {

	sendSystemNotification("READY=1")

	c.checkConsumerConnection(ctx)

	for _, consumer := range c.consumerBuilder.consumers {

		consumer := consumer
		c.childRoutines.Go(func() error {

			for {
				select {
				case isConnected := <-consumer.startConsumerCn:

					if isConnected {

						var (
							err           error
							brokerChannel *Channel
						)
						if brokerChannel, err = c.consumerBuilder.messageBroker.CreateChannel(); err != nil {
							panic(err)
						}

						if consumer.consistentExchangeType == ConsistentHashing {

							brokerChannel.createQueue(consumer.queueName).
								exchangeToQueueBind(consumer.consistentExchangeName, consumer.queueName, consumer.consistentRoutingKey, consumer.consistentExchangeType).
								createErrorQueueAndBind(consumer.errorExchangeName, consumer.errorQueueName)

						} else {

							brokerChannel.createQueue(consumer.queueName).
								createErrorQueueAndBind(consumer.errorExchangeName, consumer.errorQueueName)

						}

						for _, item := range consumer.exchanges {

							if consumer.consistentExchangeType == ConsistentHashing {

								brokerChannel.
									createExchange(item.exchangeName, item.exchangeType, item.args).
									exchangeToConsistentExchangeBind(consumer.consistentExchangeName, item.exchangeName, item.routingKey, item.exchangeType)

							} else {
								brokerChannel.
									createExchange(item.exchangeName, item.exchangeType, item.args).
									exchangeToQueueBind(item.exchangeName, consumer.queueName, item.routingKey, item.exchangeType)

							}

						}

						brokerChannel.rabbitChannel.Qos(c.parameters.PrefetchCount, 0, false)

						delivery, _ := brokerChannel.listenToQueue(consumer.queueName)

						if consumer.singleGoroutine {

							c.deliver(brokerChannel, consumer, delivery)

						} else {

							for i := 0; i < c.parameters.PrefetchCount; i++ {
								go func() {
									c.deliver(brokerChannel, consumer, delivery)
								}()
							}
						}
					}
				}
			}

			return nil
		})
	}

	return c.childRoutines.Wait()
}

func (c *Client) AddConsumerWithConsistentHashExchange(queueName string, routingKey string, exchangeName string) *Consumer {

	consumer := &Consumer{
		queueName:                  queueName,
		consistentHashRoutingKey:   routingKey,
		consistentHashExchangeType: ConsistentHashExchange,
		consistentExchangeName:     exchangeName,
		errorQueueName:             queueName + ErrorPrefix,
		errorExchangeName:          queueName + ErrorPrefix,
		startConsumerChannel:       make(chan bool),
		singleGoroutine:            false,
	}

	checkIsEmptyRabbitComponentValues(routingKey, queueName, exchangeName)

	if c.checkIsAlreadyDeclaredQueue(queueName) == false {
		c.consumerBuilder.consumers = append(c.consumerBuilder.consumers, consumer)
	}

	return consumer
}

func (c *Client) checkIsAlreadyDeclaredQueue(queueName string) bool {

	var isAlreadyDeclaredQueue bool

	for _, consumer := range c.consumerBuilder.consumers {
		if consumer.queueName == queueName {
			isAlreadyDeclaredQueue = true
		}
	}

	return isAlreadyDeclaredQueue
}

func checkIsEmptyRabbitComponentValues(routingKey string, queueName string, exchangeName string) {

	if routingKey == "" || queueName == "" || exchangeName == "" {
		panic("RabbitMq component can not be empty values")
	}
}

func sendSystemNotification(state string) error {
	//Todo: Refactor Here
	notifySocket := os.Getenv("NOTIFY_SOCKET")
	if notifySocket == "" {
		return fmt.Errorf("NOTIFY_SOCKET environment variable empty or unset.")
	}
	socketAddr := &net.UnixAddr{
		Name: notifySocket,
		Net:  "unixgram",
	}
	conn, err := net.DialUnix(socketAddr.Net, nil, socketAddr)
	if err != nil {
		return err
	}

	_, err = conn.Write([]byte(state))
	conn.Close()
	return err

}
