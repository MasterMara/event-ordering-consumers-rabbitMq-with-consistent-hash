package rabbitmq

import (
	"context"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"time"
)

type MessageBroker interface {
	CreateChannel() (*Channel, error)
	CreateConnection(ctx context.Context, params MessageBrokerParameters) error
	SignalConnectionStatus(status bool)
	SignalConnection() chan bool
	IsConnected() bool
}

type MessageBrokerParameters struct {
	Nodes             []string
	PrefetchCount     int
	RetryCount        int
	RetryInterval     time.Duration
	ConcurrentLimit   int
	UserName          string
	Password          string
	SelectedNodeIndex int
	VirtualHost       string
}

type Channel struct {
	RabbitMqChannel    *amqp.Channel
	PrefetchCount      int
	RetryCount         int
	ConcurrentLimit    int
	NotifyConfirmation chan amqp.Confirmation
}

type Broker struct {
	Parameters              MessageBrokerParameters
	Connection              *amqp.Connection
	ConnectionNotifyChannel chan bool
	logger                  log.Logger
}

func (b *Broker) CreateChannel() (*Channel, error) {

	rabbitChannel, err := b.Connection.Channel()
	if err != nil {
		return nil, err
	}

	var brokerChannel = Channel{
		RabbitMqChannel:    rabbitChannel,
		PrefetchCount:      b.Parameters.PrefetchCount,
		RetryCount:         b.Parameters.RetryCount,
		ConcurrentLimit:    b.Parameters.ConcurrentLimit,
		NotifyConfirmation: make(chan amqp.Confirmation, 1),
	}

	return &brokerChannel, nil
}

func (b *Broker) CreateConnection(ctx context.Context, parameters MessageBrokerParameters) error {

	var err error

	b.Parameters = parameters

	for {

		b.Connection, err = amqp.Dial(b.chooseNode(ctx)) //amqp.Dial(b.chooseNode(ctx))
		if err != nil {
			time.Sleep(ReconnectDelay)
			//b.logger.Exception(ctx, "Application Retried To Connect RabbitMq", err)
			continue
		}

		b.onClose(ctx)
		//b.logger.Info(ctx, "Application  Connected RabbitMq")
		b.SignalConnectionStatus(true)

		break
	}

	return err
}

func (b *Broker) chooseNode(ctx context.Context) string {

	if b.Parameters.SelectedNodeIndex == len(b.Parameters.Nodes) {
		b.Parameters.SelectedNodeIndex = 0
	}

	var selectedNode = b.Parameters.Nodes[b.Parameters.SelectedNodeIndex]
	b.Parameters.SelectedNodeIndex++
	//b.logger.Info(ctx, fmt.Sprintf("Started To Listen On Node %s", selectedNode))
	return fmt.Sprintf("amqp://%s:%s@%s/%s", b.Parameters.UserName, b.Parameters.Password, selectedNode, b.Parameters.VirtualHost)

}

func (b *Broker) onClose(ctx context.Context) {
	go func() {
		err := <-b.Connection.NotifyClose(make(chan *amqp.Error))
		if err != nil {
			//b.logger.Exception(ctx, "Rabbitmq Connection is Down", err)
			b.SignalConnectionStatus(false)
			return
		}
	}()
}

func (b *Broker) SignalConnectionStatus(status bool) {
	go func() {
		b.ConnectionNotifyChannel <- status
	}()
}

func (c *Channel) createQueue(queueName string) *Channel {

	c.RabbitMqChannel.QueueDeclare(queueName, true, false, false, false, nil)

	return c
}

func (c *Channel) exchangeToQueueBind(exchange string, queueName string, routingKey string, exchangeType int) *Channel {
	c.RabbitMqChannel.ExchangeDeclare(exchange, string(rune(exchangeType)), true, false, false, false, nil)
	c.RabbitMqChannel.QueueBind(queueName, routingKey, exchange, false, nil)
	return c
}

func (c *Channel) createErrorQueueAndBind(errorExchangeName string, errorQueueName string) *Channel {
	c.RabbitMqChannel.ExchangeDeclare(errorExchangeName, "fanout", true, false, false, false, nil)
	q, _ := c.RabbitMqChannel.QueueDeclare(errorQueueName, true, false, false, false, nil)
	c.RabbitMqChannel.QueueBind(q.Name, "", errorExchangeName, false, nil)
	return c

}

func (c *Channel) createExchange(exchange string, exchangeType int, args amqp.Table) *Channel {
	c.RabbitMqChannel.ExchangeDeclare(exchange, string(rune(exchangeType)), true, false, false, false, args)
	return c

}

func (c *Channel) exchangeToConsistentExchangeBind(destinationExchange string, sourceExchange string, routingKey string, exchangeType int) *Channel {
	c.RabbitMqChannel.ExchangeDeclare(sourceExchange, string(rune(exchangeType)), true, false, false, false, nil)
	c.RabbitMqChannel.ExchangeBind(destinationExchange, routingKey, sourceExchange, false, nil)
	return c
}

func (b Channel) listenToQueue(queueName string) (<-chan amqp.Delivery, error) {

	msg, err := b.RabbitMqChannel.Consume(queueName,
		"",
		false,
		false,
		false,
		false,
		nil)

	if err != nil {
		return nil, err
	}

	return msg, nil
}
