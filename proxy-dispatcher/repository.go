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
	Status           string `bson:"status"`
	Requests         int    `bson:"requests"`
	LastConnectionTs int    `bson:"lastConnectionTs"`
}

type ProxyMetrics struct {
	Status   map[string]int64
	Removing map[bool]int64
}

type Repository interface {
	GetProjectByToken(token string) (*Project, error)
	//GetProxy(project Project) (*Proxy, error)
	GetProxyAndUpdateConnection(project Project) (*Proxy, error)

	GetProjectCount() int64
	GetConnectorCount() int64
	GetProxyCount() int64
	GetProxyCountByStatus() ProxyMetrics
}

type MongoRepository struct {
	client   *mongo.Client
	database string
}

func NewMongoRepository(client *mongo.Client, database string) *MongoRepository {
	return &MongoRepository{client: client, database: database}
}

func (r *MongoRepository) GetProjectCount() int64 {
	coll := r.client.Database(r.database).Collection("projects")
	opts := options.Count().SetHint("_id_")
	count, err := coll.CountDocuments(context.TODO(), bson.D{}, opts)
	if err != nil {
		return 0
	}
	return count
}

func (r *MongoRepository) GetConnectorCount() int64 {
	coll := r.client.Database(r.database).Collection("connectors")
	opts := options.Count().SetHint("_id_")
	count, err := coll.CountDocuments(context.TODO(), bson.D{}, opts)
	if err != nil {
		return 0
	}
	return count
}

func (r *MongoRepository) GetProxyCount() int64 {
	coll := r.client.Database(r.database).Collection("proxies")
	opts := options.Count().SetHint("_id_")
	count, err := coll.CountDocuments(context.TODO(), bson.D{}, opts)
	if err != nil {
		return 0
	}
	return count
}

func (r *MongoRepository) GetProxyCountByStatus() ProxyMetrics {
	m := ProxyMetrics{
		Status:   make(map[string]int64),
		Removing: make(map[bool]int64),
	}
	aggStage := bson.D{
		{"$group", bson.D{
			{"_id", bson.D{{"status", "$status"}, {"removing", "$removing"}}},
			{"count", bson.D{
				{"$count", bson.D{}},
			}},
		}},
	}
	coll := r.client.Database(r.database).Collection("proxies")
	cursor, err := coll.Aggregate(context.TODO(), mongo.Pipeline{aggStage})
	if err != nil {
		return m
	}

	type result struct {
		ID struct {
			Status   string `bson:"status"`
			Removing bool   `bson:"removing"`
		} `bson:"_id"`
		Count int64 `bson:"count"`
	}
	var results []result

	if err = cursor.All(context.TODO(), &results); err != nil {
		return m
	}

	for _, result := range results {
		m.Status[result.ID.Status] = m.Status[result.ID.Status] + result.Count
		m.Removing[result.ID.Removing] = m.Removing[result.ID.Removing] + result.Count
	}
	return m
}

//
//func (r *MongoRepository) GetProxy(project Project) (*Proxy, error) {
//	filter := bson.D{
//		{"$and",
//			bson.A{
//				bson.D{{"projectId", project.ID}},
//				bson.D{{"status", "STARTED"}},
//				bson.D{{"fingerprint", bson.D{{"$ne", nil}}}},
//				bson.D{{"removing", false}},
//			}},
//	}
//	opts := options.FindOne().SetSort(bson.D{{"lastConnectionTs", 1}})
//	var proxy Proxy
//
//	coll := r.client.Database(r.database).Collection("proxies")
//	err := coll.FindOne(context.TODO(), filter, opts).Decode(&proxy)
//
//	return &proxy, err
//}

func (r *MongoRepository) GetProxyAndUpdateConnection(project Project) (*Proxy, error) {
	filter := bson.D{
		{"$and",
			bson.A{
				bson.D{{"projectId", project.ID}},
				bson.D{{"status", "STARTED"}},
				bson.D{{"fingerprint", bson.D{{"$ne", nil}}}},
				bson.D{{"removing", false}},
			}},
	}
	update := bson.M{
		"$inc": bson.M{"requests": 1},
		"$set": bson.M{"lastConnectionTs": time.Now().Unix()},
	}

	sort := bson.D{{"lastConnectionTs", 1}, {"requests", -1}}
	opts := options.FindOneAndUpdate().SetSort(sort).SetUpsert(false)
	var proxy Proxy

	coll := r.client.Database(r.database).Collection("proxies")
	err := coll.FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(&proxy)

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
