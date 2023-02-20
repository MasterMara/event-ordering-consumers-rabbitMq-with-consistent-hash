package rabbitmq

import (
	"context"
	"encoding/json"
	utils "event-ordering-consumers-rabbitmq-with-consistent-hash/pkg/util"
	"github.com/streadway/amqp"
	"reflect"
	"strconv"
	"sync"
	"time"
)

var (
	deliveryMode     uint8 = 2
	headerError            = "Error"
	headerRetryCount       = "RetryCount"
	headerStackTrace       = "StackTrace"
	headerTime             = "Time"
	maxChannelMutex        = &sync.Mutex{}
)

const resendDelay = 1 * time.Millisecond

type (
	PublisherBuilder struct {
		openChannelCount         int
		messageBroker            MessageBroker
		brokerChannel            *Channel
		publishers               []Publisher
		startPublisherCh         chan bool
		isAlreadyStartConnection bool
		isChannelActive          bool
	}
	Publisher struct {
		isAlreadyCreated bool
		exchangeName     string
		exchangeType     int
		payloadTypes     []reflect.Type
	}
)

func errorPublishMessage(correlationId string, payload []byte, retryCount int, err error, stackTracing string) amqp.Publishing {

	headers := make(map[string]interface{})
	headers[headerRetryCount] = strconv.Itoa(retryCount)
	headers[headerError] = err.Error()
	headers[headerStackTrace] = stackTracing
	headers[headerTime] = time.Now().String()

	return amqp.Publishing{
		MessageId:       utils.GetGuid(),
		Body:            payload,
		Headers:         headers,
		CorrelationId:   correlationId,
		Timestamp:       time.Now(),
		DeliveryMode:    deliveryMode,
		ContentEncoding: "UTF-8",
		ContentType:     "application/json",
	}
}

func getBytes(key interface{}) ([]byte, error) {
	return json.Marshal(key)
}

func (p *PublisherBuilder) SubscriberExchange() *PublisherBuilder {

	for index, item := range p.publishers {
		if !item.isAlreadyCreated {
			p.brokerChannel.createExchange(item.exchangeName, item.exchangeType, nil)
			p.publishers[index].isAlreadyCreated = true
		}
	}
	return p
}

func (p *PublisherBuilder) Publish(ctx context.Context, routingKey string, exchangeName string, payload interface{}) error {

	var (
		err     error
		message amqp.Publishing
	)
	for {
		if p.isChannelActive {
			break
		}
	}

	p.SubscriberExchange()

	if message, err = publishMessage("", payload); err != nil {
		return err
	}

	err = p.brokerChannel.RabbitMqChannel.Publish(exchangeName, routingKey, false, false, message)

	select {

	case confirm := <-p.brokerChannel.NotifyConfirmation:
		if confirm.Ack {
			break
		}
	case <-time.After(resendDelay):
	}

	return err
}

func publishMessage(correlationId string, payload interface{}) (amqp.Publishing, error) {

	headers := make(map[string]interface{})
	headers[headerTime] = time.Now().String()

	bodyJson, err := json.Marshal(payload)

	return amqp.Publishing{
		MessageId:       utils.GetGuid(),
		Body:            bodyJson,
		Headers:         headers,
		CorrelationId:   correlationId,
		Timestamp:       time.Now(),
		DeliveryMode:    deliveryMode,
		ContentEncoding: "UTF-8",
		ContentType:     "application/json",
	}, err
}

func (p *PublisherBuilder) CreateChannel() {

	var err error

	go func() {

		for {
			select {

			case isConnected := <-p.startPublisherCh:
				if isConnected {
					p.brokerChannel, err = p.messageBroker.CreateChannel()
					if err != nil {
						panic(err)
					}
					p.isChannelActive = true

					p.brokerChannel.RabbitMqChannel.NotifyPublish(p.brokerChannel.NotifyConfirmation)

				} else {
					p.isChannelActive = false
				}
			}
		}

	}()

}
