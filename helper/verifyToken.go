package helper

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

func VerifyToken(tokenString string) (jwt.MapClaims, error) {
	tokenString = strings.Replace(tokenString, "Bearer ", "", 1)

	if err := godotenv.Load(".env"); err != nil {
		log.Fatal(err)
	}

	key := []byte(os.Getenv("JWT_SIGNATURE"))

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing method")
		}

		return key, nil
	})

	if err != nil {
    return nil, err
  }

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}