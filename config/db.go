package config

import (
	"context"
	"mofe64/playlistGen/util"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var DATABASE *mongo.Client = ConnectDB()

func ConnectDB() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Create Mongo Client and connect
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(EnvMongoURI()))
	if err != nil {
		util.ErrorLog.Fatalln("Error connecting to Mongodb", err)
	}
	// Ping Mongo db Database
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		util.ErrorLog.Fatalln("Could not ping Mongodb database", err)
	}
	util.InfoLog.Println("Connected to MongoDB ....")
	return client
}

// GetCollection - helper function to retrieve collection from database
func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	collection := client.Database(EnvDatabaseName()).Collection(collectionName)
	return collection
}
