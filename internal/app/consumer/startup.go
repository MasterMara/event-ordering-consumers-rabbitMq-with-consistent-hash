package consumer

import (
	"context"
	"event-ordering-consumers-rabbitmq-with-consistent-hash/internal/app/consumer/repository"
	"event-ordering-consumers-rabbitmq-with-consistent-hash/internal/app/consumer/service"
	"event-ordering-consumers-rabbitmq-with-consistent-hash/pkg/mongo"
	"event-ordering-consumers-rabbitmq-with-consistent-hash/pkg/rabbitmq"
	"log"
)

// Run creates objects via constructors.
func Run() {

	//Todo : Create Logger

	//Initialize Broker Todo: Read From Config
	var nodes = []string{"127.0.0.1"}
	var messageBroker = rabbitmq.NewRabbitMqClient(nodes,
		"MasterMara",
		"123456",
		"",
		log.Logger{},
		rabbitmq.RetryCount(0),
		rabbitmq.PrefetchCount(1),
	)

	// Initialize DB
	db, err := mongo.NewDatabase("mongodb://root:root@localhost:27017/admin", "") // Manage this From Environment
	if err != nil {
		//return err Todo: Handle This !!
	}

	//Init Repository
	var repository = repository.Repository(db)

	//Init Services
	var service = service.NewService(repository)

	// Main Consumer
	var consumer = NewConsumer(log.Logger{}, service, messageBroker, 3)

	consumer.RunConsumer(context.Background())
}
