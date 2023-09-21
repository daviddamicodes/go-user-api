package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"

	"github.com/daviddamicodes/go-user-api/controllers"
	"github.com/daviddamicodes/go-user-api/middleware"

	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {
	router := httprouter.New()
	// user session
	us := controllers.NewUserController(getUserCollection())
	//USERS
	router.GET("/user/:id", us.GetUser)
	router.GET("/user", middleware.AuthMiddleware(us.GetUsers))
	router.POST("/user", us.CreateUser)
	router.PATCH("/user/:id", us.UpdateUser)
	router.DELETE("/user/:id", us.DeleteUser)
	//AUTH
	router.GET("/auth/login", us.Login)
	router.GET("/auth/request-reset/:id", us.RequestPasswordReset)
	router.GET("/auth/reset/:id", us.ResetPassword)
	http.ListenAndServe("localhost:8080", router)
}

func getUserCollection() *mongo.Collection {
	
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal(err)
	}

	uri := os.Getenv("MONGODB_URI")
	// Use the SetServerAPIOptions() method to set the Stable API version to 1
  client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	} else {
		log.Print("Db connected successfully")
	}

	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}

	userCollection := client.Database(os.Getenv("DB_COLLECTION")).Collection("users")

	return userCollection
}
