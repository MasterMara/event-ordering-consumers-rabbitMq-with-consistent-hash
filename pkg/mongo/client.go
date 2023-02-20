package mongo

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func newDatabase(uri, databaseName string, dbOpts ...*options.DatabaseOptions) (db *mongo.Database, err error) {

	//logger := log.NewLogger()

	//logger.Info(context.TODO(), fmt.Sprintf("Mongo:Connection Uri:%s", uri))
	clientOptions := options.
		Client().
		ApplyURI(uri)

	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		return db, errors.New("") // Handle All Error
	}

	ctxWithTimeout, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctxWithTimeout)
	if err != nil {
		//logger.Exception(context.TODO(), "Mongo: mongo client couldn't connect with background context: %v", err)
		return db, errors.New("")
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		return db, errors.New("")
	}

	db = client.Database(databaseName, dbOpts...)

	return db, err
}

func NewDatabase(uri, databaseName string) (db *mongo.Database, err error) {
	return newDatabase(uri, databaseName)
}
