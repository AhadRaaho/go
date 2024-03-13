package database

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectToMongoDB() (*mongo.Client, error) {
	return mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
}
func Sum() int {
	return 2
}
