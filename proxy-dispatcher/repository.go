package main

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

type Project struct {
	ID string `bson:"_id"`
}

type Proxy struct {
	ID            string `bson:"_id"`
	TransportType string `bson:"transportType"`
	UserAgent     string `bson:"useragent"`
	Config        struct {
		Address struct {
			Hostname string `bson:"hostname"`
			Port     int    `bson:"port"`
		} `bson:"address"`
		Certificate struct {
			Cert string `bson:"cert"`
			Key  string `bson:"key"`
		} `bson:"certificate"`
	} `bson:"config"`
	Status string `bson:"status"`
}

type Repository interface {
	GetProjectByToken(token string) (*Project, error)
	GetProxy(project Project) (*Proxy, error)
}

type MongoRepository struct {
	client   *mongo.Client
	database string
}

func NewMongoRepository(client *mongo.Client, database string) *MongoRepository {
	return &MongoRepository{client: client, database: database}
}

func (r *MongoRepository) GetProxy(project Project) (*Proxy, error) {
	filter := bson.D{
		{"$and",
			bson.A{
				bson.D{{"projectId", project.ID}},
				bson.D{{"status", "STARTED"}},
				bson.D{{"fingerprint", bson.D{{"$ne", nil}}}},
				bson.D{{"removing", false}},
			}},
	}
	opts := options.FindOne().SetSort(bson.D{{"lastConnectionTs", 1}})
	var proxy Proxy

	coll := r.client.Database(r.database).Collection("proxies")
	err := coll.FindOne(context.TODO(), filter, opts).Decode(&proxy)

	return &proxy, err
}

func (r *MongoRepository) GetProjectByToken(token string) (*Project, error) {
	filter := bson.D{{"token", token}}
	opts := options.FindOne().SetProjection(bson.D{{"_id", 1}})

	coll := r.client.Database(r.database).Collection("projects")
	var project Project
	err := coll.FindOne(context.TODO(), filter, opts).Decode(&project)
	return &project, err
}

func (r *MongoRepository) Ping() error {
	ctxPing, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return r.client.Ping(ctxPing, readpref.Primary())
}
