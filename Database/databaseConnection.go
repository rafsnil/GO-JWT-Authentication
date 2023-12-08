package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DbInstance() *mongo.Client {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading.env file")
		// fmt.Println("Error loading.env file")
	}
	//Getting mongodb url from .env
	MongoDb := os.Getenv("MONGODB_URL")

	// Creating Mongo Context with a timer
	cntxt, cancelContext := context.WithTimeout(context.Background(), 10*time.Second)

	//Cancelling the context once the value from this func is return
	defer cancelContext()

	/*
		Connecting to the mongo database:
		This code creates a new options.ClientOptions object, which
		allows you to configure various connection parameters like credentials,
		timeouts, and pool sizes. It then sets the connection URI to the
		provided mongoDbUrl using the ApplyURI() method and returns
		the updated options object
	*/
	client, err := mongo.Connect(cntxt, options.Client().ApplyURI(MongoDb))
	if err != nil {
		fmt.Println("Could Not Connect to DB")
		log.Fatal(err)
	}

	return client
}

var Client *mongo.Client = DbInstance()

/*
This function takes in a pointer to a mongo.Client and a
collectionName string as arguments and return a collection pointer.
Inside the function I accessed the database with the
client.Database(db name) function and then I got the
collection name using the .Collection(collectionName)
*/
func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	var collection *mongo.Collection = client.Database("cluster0").Collection(collectionName)
	return collection
}
