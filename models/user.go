package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	Id 					primitive.ObjectID 	`json:"id" bson:"_id"`
	FullName 		string							`json:"fullName" bson:"fullName"`
	Username		string							`json:"username" bson:"username"`
	Email				string							`json:"email" bson:"email"`
	Gender			string							`json:"gender" bson:"gender"`
	Age					int									`json:"age" bson:"age"`
	Password		string							`json:"-" bson:"password"`
}