package server

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"os"
)

func GetId(w http.ResponseWriter, r *http.Request) int8 {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		Bad(map[string]interface{}{"message": "auth token is not provided", "status": 400}, w)
		return 0
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !token.Valid {
		Bad(map[string]interface{}{"message": "Token is not valid", "status": 400}, w)
		return 0
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		Bad(map[string]interface{}{"message": "Error extracting token claims", "status": 400}, w)
		return 0
	}
	id, ok := claims["id"].(float64)
	if !ok {
		Bad(map[string]interface{}{"message": "Error extracting id from token", "status": 400}, w)
		return 0
	}

	return int8(id)
}
