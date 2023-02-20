package rabbitmq

import (
	"context"
	"fmt"
	"github.com/streadway/amqp"
	"golang.org/x/sync/errgroup"
	"net"
	"os"
	"runtime"
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
	context            context.Context
	shutdownFn         context.CancelFunc
	childRoutines      *errgroup.Group
	parameters         MessageBrokerParameters
	shutdownReason     string
	shutdownInProgress bool
	consumerBuilder    ConsumerBuilder
	publisherBuilder   *PublisherBuilder
}

func (c *Client) RunConsumers(ctx context.Context) error {

	sendSystemNotification("READY=1")

	c.checkConsumerConnection(ctx)

	for _, consumer := range c.consumerBuilder.consumers {

		consumer := consumer
		c.childRoutines.Go(
			func() error {

				for {
					select {
					case isConnected := <-consumer.startConsumerChannel:

						if isConnected {

							var (
								err           error
								brokerChannel *Channel
							)

							brokerChannel, err = c.consumerBuilder.messageBroker.CreateChannel()
							if err != nil {
								panic(err)
							}

							if consumer.consistentHashExchangeType == ConsistentHashExchange {

								brokerChannel.createQueue(consumer.queueName).
									exchangeToQueueBind(consumer.consistentExchangeName, consumer.queueName, consumer.consistentHashRoutingKey, consumer.consistentHashExchangeType).
									createErrorQueueAndBind(consumer.errorExchangeName, consumer.errorQueueName)

							} else {

								brokerChannel.createQueue(consumer.queueName).
									createErrorQueueAndBind(consumer.errorExchangeName, consumer.errorQueueName)

							}

							for _, item := range consumer.exchanges {

								if consumer.consistentHashExchangeType == ConsistentHashExchange {

									brokerChannel.
										createExchange(item.exchangeName, item.exchangeType, item.args).
										exchangeToConsistentExchangeBind(consumer.consistentExchangeName, item.exchangeName, item.routingKey, item.exchangeType)

								} else {
									brokerChannel.
										createExchange(item.exchangeName, item.exchangeType, item.args).
										exchangeToQueueBind(item.exchangeName, consumer.queueName, item.routingKey, item.exchangeType)

								}

							}

							brokerChannel.RabbitMqChannel.Qos(c.parameters.PrefetchCount, 0, false)

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

func (c *Client) checkConsumerConnection(ctx context.Context) {

	go func() {

		for {
			select {

			case isConnected := <-c.consumerBuilder.messageBroker.SignalConnection():
				if !isConnected {
					c.consumerBuilder.messageBroker.CreateConnection(ctx, c.parameters)
					for _, consumer := range c.consumerBuilder.consumers {
						consumer.startConsumerChannel <- true
					}
				}
			}
		}

	}()
}

func (c *Client) deliver(brokerChannel *Channel, consumer *Consumer, delivery <-chan amqp.Delivery) {
	for d := range delivery {

		Do(func(attempt int) (retry bool, err error) {

			retry = attempt < c.parameters.RetryCount

			defer func() {

				if r := recover(); r != nil {

					if !retry || err == nil {

						err, ok := r.(error)

						if !ok {
							retry = false //Because of panic exception
							err = fmt.Errorf("%v", r)
						}

						stack := make([]byte, 4<<10)
						length := runtime.Stack(stack, false)

						brokerChannel.RabbitMqChannel.Publish(consumer.errorExchangeName, "", false, false, errorPublishMessage(d.CorrelationId, d.Body, c.parameters.RetryCount, err, fmt.Sprintf("[Exception Recover] %v %s\n", err, stack[:length])))

						select {
						case confirm := <-brokerChannel.NotifyConfirmation:
							if confirm.Ack {
								break
							}
						case <-time.After(resendDelay):
						}

						d.Ack(false)
					}
					return
				}
			}()
			err = consumer.handleConsumer(Message{CorrelationId: d.CorrelationId, Payload: d.Body, MessageId: d.MessageId, TimeStamp: d.Timestamp})
			if err != nil {
				if retry {
					time.Sleep(c.parameters.RetryInterval)
				}
				panic(err)
			}
			d.Ack(false)

			return
		})
	}
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
