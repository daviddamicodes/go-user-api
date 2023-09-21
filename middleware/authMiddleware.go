package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/daviddamicodes/go-user-api/helper"
	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserController struct {
	session *mongo.Collection
}

func NewUserController(s *mongo.Collection) *UserController{
	return &UserController{s}
}

func AuthMiddleware(next httprouter.Handle) httprouter.Handle {
    return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
        // Extract the JWT token from the request header or query parameter.
        tokenString := extractTokenFromRequest(r)

        // Verify the token.
        claims, err := helper.VerifyToken(tokenString)
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        // Add the user claims to the request context for later use.
        r = r.WithContext(context.WithValue(r.Context(), "userClaims", claims))

        // If the token is valid, proceed to the next handler.
        next(w, r, p)
    }
}

// Function to check if the user has the required permission.
// func hasPermission(user models.User) bool {
// 	permissions := getRolePermissions(user.Role)
// 	return contains(permissions, user.Role)
// }

// // Function to get permissions for a role
// // TODO fetch from api
// func getRolePermissions(role string) []string {
// 	switch role {
// 		case "admin":
// 			return []string{"delete_user", "create_user", "promote_user", "create_post", "edit_post"}
// 		case "user":
// 			return []string{"create_post", "edit_post"}
// 		default:
// 			return nil
// 	} 
// }

// // Helper function to check if a slice contains a specific value.
// func contains(slice []string, value string) bool {
// 	for _, item := range slice {
// 		if item == value {
// 			return true
// 		}
// 	}
// 	return false
// }

// func getUserFromContext(ctx mongo.SessionContext) models.User {
// 	var u models.User

// 	if err := ctx.Client().
// }


// Function to extract the JWT token from the request header or query parameter.
func extractTokenFromRequest(r *http.Request) string {
    // Extract the token from the Authorization header (Bearer token).
    authHeader := r.Header.Get("Authorization")
    if strings.HasPrefix(authHeader, "Bearer ") {
      return strings.TrimPrefix(authHeader, "Bearer ")
    }

    // If not found in the header, check the query parameter.
    token := r.URL.Query().Get("token")
    return token
}