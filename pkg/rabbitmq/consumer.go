package rabbitmq

import (
	"github.com/streadway/amqp"
	"time"
)

const (
	DirectExchange = iota + 1
	FanoutExchange
	TopicExchange
	ConsistentHashExchange
)

type ConsumerBuilder struct {
	consumers []*Consumer
}

type Exchange struct {
	exchangeName string
	routingKey   string
	exchangeType int
	args         amqp.Table
}

type Message struct {
	Payload       []byte
	CorrelationId string
	MessageId     string
	TimeStamp     time.Time
}

type Consumer struct {
	queueName                  string
	errorQueueName             string
	errorExchangeName          string
	startConsumerChannel       chan bool
	singleGoroutine            bool
	exchanges                  []Exchange
	consistentHashRoutingKey   string
	consistentHashExchangeType int
	consistentExchangeName     string
	handleConsumer             handleConsumer
}

type handleConsumer func(message Message) error

func (c *Consumer) SubscribeExchange(routingKey string, exchangeType int, exchangeName string) *Consumer {

	if c.checkIsAlreadyDeclaredExchange(exchangeName, routingKey) {
		return c
	}

	c.exchanges = append(c.exchanges, Exchange{
		exchangeName: exchangeName,
		exchangeType: exchangeType,
		routingKey:   routingKey,
	})

	return c
}

func (c *Consumer) WithSingleGoroutine(value bool) *Consumer {
	c.singleGoroutine = value
	return c
}

func (c *Consumer) HandleConsumer(consumer handleConsumer) *Consumer {
	c.handleConsumer = consumer
	return c
}

func (c *Consumer) checkIsAlreadyDeclaredExchange(exchangeName string, routingKey string) bool {

	var isAlreadyDeclaredExchange bool

	for _, exchange := range c.exchanges {
		if exchange.exchangeName == exchangeName && exchange.routingKey == routingKey {
			isAlreadyDeclaredExchange = true
		}
	}

	return isAlreadyDeclaredExchange
}
