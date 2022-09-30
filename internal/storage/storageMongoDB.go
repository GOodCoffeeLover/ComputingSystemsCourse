package storage

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const mongo_url = "mongodb://root:password@mongoDB:27017/"

type TasksMongoStorage struct {
	ctx context.Context
}

func NewTasksMongoStorage() *TasksMapStorage {
	uri := mongo_url
	//uri := os.Getenv("MONGODB_URI")
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	client.Disconnect(context.TODO())
	return nil
}
