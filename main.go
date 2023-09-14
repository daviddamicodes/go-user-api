package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"

	"github.com/daviddamicodes/go-user-api/controllers"

	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {
	router := httprouter.New()
	session := controllers.NewUserController(getUserCollection())
	router.GET("/user/:id", session.GetUser)
	router.GET("/user", session.GetUsers)
	router.POST("/user", session.CreateUser)
	router.DELETE("/user/:id", session.DeleteUser)
	http.ListenAndServe("localhost:8080", router)
}

func getUserCollection() *mongo.Collection {
	_err := godotenv.Load(".env")

	if _err != nil {
		log.Fatal(_err)
	}

	uri := os.Getenv("MONGODB_URI")
	// Use the SetServerAPIOptions() method to set the Stable API version to 1
  client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}

	userCollection := client.Database("go-user-api").Collection("users")

	return userCollection
}