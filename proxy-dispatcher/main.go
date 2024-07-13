package main

import (
	"context"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"time"
)

func main() {
	viper.SetDefault("proxyManagerPort", ":8080")
	viper.SetDefault("mongodbURI", "mongodb://root:password@localhost:27017")
	viper.SetDefault("mongodbDB", "scrapoxy")

	viper.BindEnv("proxyManagerPort", "PROXY_MANAGER_PORT")
	viper.BindEnv("mongodbURI", "STORAGE_DISTRIBUTED_MONGO_URI")
	viper.BindEnv("mongodbDB", "STORAGE_DISTRIBUTED_MONGO_DB")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(viper.GetString("mongodbURI")))
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	repository := NewMongoRepository(client, viper.GetString("mongodbDB"))
	err = repository.Ping()
	if err != nil {
		log.Fatal(err)
	}

	handler := NewHandler(repository)

	// Create a new HTTP server with the handleRequest function as the handler
	server := http.Server{
		Addr:    viper.GetString("proxyManagerPort"),
		Handler: http.HandlerFunc(handler.handleRequest),
	}

	// Start the server and log any errors
	log.Printf("Starting proxy server on %s\n", viper.GetString("proxyManagerPort"))
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal("Error starting proxy server: ", err)
	}
}
