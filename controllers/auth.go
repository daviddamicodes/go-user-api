package controllers

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"time"

	"github.com/absagar/go-bcrypt"
	"github.com/daviddamicodes/go-user-api/models"
	redisClient "github.com/daviddamicodes/go-user-api/redisclient"
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

func (uc UserController) RequestPasswordReset(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid UserID", http.StatusBadRequest)
		return
	}

	var u models.User

	if err := uc.session.FindOne(context.TODO(), primitive.D{{Key: "_id", Value: oid}}).Decode(&u); err != nil {
		http.Error(w, "User does not exist", http.StatusNotFound)
		return
	}

	key := fmt.Sprintf("reset_code_%v", u.Username)

	// Define the minimum and maximum values for a 4-digit integer
	minValue := big.NewInt(1000) // 4-digit minimum
	maxValue := big.NewInt(9999) // 4-digit maximum

	// Calculate the range (maxValue - minValue)
	rangeValue := new(big.Int).Sub(maxValue, minValue)

	// Generate a random value within the specified range
	randomValue, err := randRange(minValue, rangeValue)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Add minValue to the random value to get a 4-digit random integer
	randomValue.Add(randomValue, minValue)

	// Convert the random big.Int to a string with leading zeros if necessary
	randomString := fmt.Sprintf("%04s", randomValue.String())

	fmt.Printf("CODE IS %v\n", randomString)

	redisCache, err := redisClient.RedisClient()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
		
	err = redisCache.Set(context.Background(), key, randomString, time.Second*60*60).Err()
	if err != nil {
		panic(err)
	}
	
	w.WriteHeader(201)
	fmt.Fprintf(w, "Reset Request Sent\n")
}

func randRange(min, rangeValue *big.Int) (*big.Int, error) {
	randomValue, err := rand.Int(rand.Reader, rangeValue)
	if err != nil {
		return nil, err
	}
	return randomValue, nil
}

func (uc UserController) ResetPassword(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid ObjectID", http.StatusBadRequest)
		return
	}

	var u models.User

	if err := uc.session.FindOne(context.TODO(), primitive.D{{Key: "_id", Value: oid}}).Decode(&u); err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	key := fmt.Sprintf("reset_code_%v", u.Username)

	redisCache, err := redisClient.RedisClient()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	val, err := redisCache.Get(context.Background(), key).Result()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// fmt.Printf("RETRIEVED CODE IS %v \n", val)
	
	type RequestStruct struct {
		Password string `json:"password" bson:"password"`
		Code string `json:"code" bson:"code"`
	}

	var eu RequestStruct
	
	if err := json.NewDecoder(r.Body).Decode(&eu); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	
	if eu.Code != val {
		http.Error(w, "Incorrect Code", http.StatusBadRequest)
		return
	}

	salt, _ := bcrypt.Salt(10)
	hashedPassword, err := bcrypt.Hash(eu.Password, salt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	updateResult := uc.session.FindOneAndUpdate(context.TODO(), primitive.D{{Key: "_id", Value: oid}}, primitive.D{{Key: "$set", Value: primitive.D{{Key: "password", Value: hashedPassword}}}})

	if updateResult.Err() != nil {
		http.Error(w, updateResult.Err().Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"message": "Password reset successful"}
	uj, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(uj)
}