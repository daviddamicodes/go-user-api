package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/absagar/go-bcrypt"
	"github.com/daviddamicodes/go-user-api/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (uc UserController) Login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal(err)
	}

	var loginRequest struct {
		UsernameOrEmail string 	`json:"usernameOrEmail" bson:"usernameOrEmail"`
		Password				string	`json:"password" bson:"password"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var user models.User
	if err := uc.session.FindOne(context.TODO(), primitive.M{
		"$or": []primitive.M{
			{"username": loginRequest.UsernameOrEmail},
			{"email": loginRequest.UsernameOrEmail},
		},
	}).Decode(&user); err != nil {
		http.Error(w, "Invalid Username or Email", http.StatusUnauthorized)
		return
	}

	if !bcrypt.Match(loginRequest.Password, user.Password) {
		http.Error(w, "Invalid Password", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.Id.Hex(),
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SIGNATURE")))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err.Error())
		return
	}
	
	response := struct {
		User models.User `json:"user"`
		AccessToken string `json:"accessToken"`
	} {
		User: user,
		AccessToken: tokenString,
	}

	uj, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(uj)
}