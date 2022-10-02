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
	client *mongo.Client
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
	return &TasksMongoStorage{client: client}, nil
}

func (tms *TasksMongoStorage) Get(taskName string) (task core.Task, err error) {
	coll := tms.client.Database("Time_Manager").Collection("Tasks")
	filter := bson.D{{"_id", taskName}}
	res := coll.FindOne(context.TODO(), filter, nil)
	err = res.Decode(&task)
	fmt.Printf("task: %v, err: %v\n", res, err)

	if err != nil {
		return
	}
	//fmt.Printf("task: %v", task)
	return
}
func (tms *TasksMongoStorage) Set(taskName string, task core.Task) error {
	coll := tms.client.Database("Time_Manager").Collection("Tasks")
	doc, err := bson.Marshal(task)
	doc, err = bson.MarshalAppend(doc, map[string]string{"_id": taskName})
	if err != nil {
		//fmt.Println(err)
		return err
	}
	res, err := coll.InsertOne(context.TODO(), doc)
	if err != nil {
		//fmt.Println(err)
		return err
	} else {
		fmt.Println(res)
	}
	return nil
}

func (tms *TasksMongoStorage) Delete(taskName string) error {
	return nil
}
func (tms *TasksMongoStorage) Disconnect() error {
	return tms.client.Disconnect(context.TODO())
}
