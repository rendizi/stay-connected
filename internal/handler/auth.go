package handler

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"os"
	"regexp"
	"stay-connected/internal/server"
	"stay-connected/internal/services/db"
	stories "stay-connected/internal/services/inst"
	"time"
)

func Register(w http.ResponseWriter, r *http.Request) {
	var signUpCredentials SignCredentials
	err := json.NewDecoder(r.Body).Decode(&signUpCredentials)
	if err != nil {
		server.Bad(map[string]interface{}{"message": err.Error()}, w)
		return
	}
	if !isValidEmail(signUpCredentials.Email) {
		server.Bad(map[string]interface{}{"message": "invalid email format"}, w)
		return
	}
	if signUpCredentials.Password == "" {
		server.Bad(map[string]interface{}{"message": "password is not provided"}, w)
		return
	}
	err = db.Register(signUpCredentials.Email, signUpCredentials.Password)
	if err != nil {
		server.Bad(map[string]interface{}{"message": err.Error()}, w)
		return
	}
	server.Ok(map[string]interface{}{"message": "Registered successfully"}, w)
}

func Login(w http.ResponseWriter, r *http.Request) {
	var signInCredentials SignCredentials
	err := json.NewDecoder(r.Body).Decode(&signInCredentials)
	if err != nil {
		server.Bad(map[string]interface{}{"message": err.Error()}, w)
		return
	}
	id, err := db.Login(signInCredentials.Email, signInCredentials.Password)
	if err != nil {
		server.Bad(map[string]interface{}{"message": err.Error()}, w)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  id,
		"exp": time.Now().Add(3 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(os.Getenv(("JWT_SECRET"))))
	if err != nil {
		server.Internal(map[string]interface{}{"message": "Error generating token", "status": 400}, w)
		return
	}

	server.Ok(map[string]interface{}{"message": "Login successful", "token": tokenString}, w)
}

func UpdateCredentials(w http.ResponseWriter, r *http.Request) {
	var updateCredentials InstagramCredentials
	err := json.NewDecoder(r.Body).Decode(&updateCredentials)
	if err != nil {
		server.Bad(map[string]interface{}{"message": err.Error()}, w)
		return
	}
	if updateCredentials.Password == "" || updateCredentials.Username == "" {
		server.Bad(map[string]interface{}{"message": "username or password is not provided"}, w)
		return
	}

	_, err = stories.LoginToInst(updateCredentials.Username, updateCredentials.Password)
	if err != nil {
		server.Bad(map[string]interface{}{"message": err.Error()}, w)
		return
	}

	id := server.GetId(w, r)
	if id == 0 {
		return
	}

	err = db.LinkInstagram(updateCredentials.Username, updateCredentials.Password, id)
	if err != nil {
		server.Bad(map[string]interface{}{"message": err.Error()}, w)
		return
	}
	server.Ok(map[string]interface{}{"message": "Updated credentials successfully"}, w)
}

func DeleteCredentials(w http.ResponseWriter, r *http.Request) {
	id := server.GetId(w, r)
	if id == 0 {
		return
	}
	err := db.UnlinkInstagram(id)
	if err != nil {
		server.Bad(map[string]interface{}{"message": err.Error()}, w)
		return
	}
	server.Ok(map[string]interface{}{"message": "Deleted credentials successfully"}, w)
}

func isValidEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(emailRegex, email)
	return match
}
