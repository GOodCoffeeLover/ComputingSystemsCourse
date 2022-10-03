package storage

import (
	"ComputingSystemsCourse/internal/core"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const mongo_url = "mongodb://root:password@localhost:27017/"

type TasksMongoStorage struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewTasksMongoStorage() (*TasksMongoStorage, error) {
	uri := mongo_url
	//uri := os.Getenv("MONGODB_URI")
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		return nil, err

	}

	return &TasksMongoStorage{
		client:     client,
		collection: client.Database("TasksManager").Collection("Tasks"),
	}, nil
}

func (tms *TasksMongoStorage) Get(taskName string) (task core.Task, err error) {

	filter := bson.D{{"_id", taskName}}
	res := tms.collection.FindOne(context.TODO(), filter)

	if res.Err() != nil {
		err = fmt.Errorf("can't find taks(id:%v) due to %v", taskName, res.Err())
		return
	}

	if err = res.Decode(&task); err != nil {
		err = fmt.Errorf("can't decode res due to %v", err)
		return
	}
	fmt.Printf("task: %v", task)
	return
}
func (tms *TasksMongoStorage) Set(taskName string, task core.Task) error {
	filter := bson.D{{"_id", taskName}}
	update := bson.D{{"$set", task}}
	opts := options.Update().SetUpsert(true)

	res, err := tms.collection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return err
	} else {
		fmt.Println(res)
	}
	return nil
}

func (tms *TasksMongoStorage) Delete(taskName string) error {
	filter := bson.D{{"_id", taskName}}
	res := tms.collection.FindOneAndDelete(context.TODO(), filter)
	if res.Err() != nil {
		return fmt.Errorf("can't delete task(id:%v) due to %v", taskName, res.Err())
	}
	return nil
}
func (tms *TasksMongoStorage) Disconnect() error {
	return tms.client.Disconnect(context.TODO())
}
