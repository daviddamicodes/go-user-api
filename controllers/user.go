package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/daviddamicodes/go-user-api/models"
	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserController struct {
	session *mongo.Collection
}

func NewUserController(s *mongo.Collection) *UserController{
	return &UserController{s}
}

func (uc UserController) GetUser(w http.ResponseWriter, r *http.Request, p httprouter.Params ) {
	id := p.ByName("id")

	// Check if the ID is a valid ObjectId
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid ObjectId", http.StatusBadRequest)
    return
	}

	var u = models.User{}

	if err := uc.session.FindOne(context.TODO(), primitive.D{{Key: "_id", Value: oid}}).Decode(&u); err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	uj, err := json.Marshal(u)
	if err != nil {
		fmt.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(uj)
}

func (uc UserController) CreateUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	u := models.User{}

	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	u.Id = primitive.NewObjectID()
	
	_, err := uc.session.InsertOne(context.TODO(), u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	uj, err := json.Marshal(u)

	if err != nil {
		fmt.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	// fmt.Fprintf(w, "%s\n", uj)
	w.Write(uj)
}

func (uc UserController) GetUsers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// findOptions := uc.session.Find()
	// findOptions.SetL

	// Here's an array in which you can store the decoded documents
	var users []*models.User

	// Passing bson.D{{}} as the filter matches all documents in the collection
	cursor, err := uc.session.Find(context.TODO(), primitive.D{{}})
	if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

	for cursor.Next(context.TODO()) {
		// Finding multiple documents returns a cursor
		// Iterating through the cursor allows us to decode documents one at a time
		var elem models.User
		err := cursor.Decode(&elem)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		users = append(users, &elem)
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	userJSON, err := json.Marshal(users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Close the cursor once finished
	cursor.Close(context.TODO())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(userJSON)
}

func (uc UserController) UpdateUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid Object ID", http.StatusBadRequest)
	}

	var u map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	update := primitive.D{}

	if age, ok := u["age"].(float64); ok {
		update = append(update, primitive.E{Key: "$set", Value: primitive.D{{Key:"age", Value: age}}})
	}
	if name, ok := u["name"].(string); ok {
		update = append(update, primitive.E{Key: "$set", Value: primitive.D{{Key: "name", Value: name}}})
	}
	if gender, ok := u["gender"].(string); ok {
		update = append(update, primitive.E{Key: "$set", Value: primitive.D{{Key: "gender", Value: gender}}})
	}

	// Check if there are any fields to update
	if len(update) == 0 {
		http.Error(w, "No fields to update", http.StatusBadRequest)
		return
	}

  // Update the user in the database
  filter := primitive.D{{Key: "_id", Value: oid}} // Filter by the user's ObjectId

	_, err = uc.session.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "User updated successfully")
}

func (uc UserController) DeleteUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid Object ID", http.StatusBadRequest)
	}

	_, err = uc.session.DeleteOne(context.TODO(), primitive.D{{Key: "_id", Value: oid}})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "User with ID %v has been deleted\n", id)
}