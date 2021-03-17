/*
This Package is written to update MongoDB Collection if any single tiny url is not being accessed
in last 30 days. It will expire it automatically. 
*/
package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectDB() *mongo.Collection {
	clientOpt := options.Client().ApplyURI("mongodb+srv://tinyurl:hDupohejntPcH6LH@tinyurl.sefky.mongodb.net/myFirstDatabase?retryWrites=true&w=majority")
	client, err := mongo.Connect(context.TODO(), clientOpt)

	fmt.Println("Connected to MongoDB!")

	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	collection := client.Database("tinyurl").Collection("tinyurl")
	mod := mongo.IndexModel{
		Keys: bson.M{
			"updated_at": 1,
		}, Options: options.Index().SetExpireAfterSeconds(2592000),
	}
	_, err = collection.Indexes().CreateOne(ctx, mod)
	if err != nil {
		log.Fatal(err)
	}
	return collection
}
