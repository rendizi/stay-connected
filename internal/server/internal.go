package server

import (
	"encoding/json"
	"net/http"
)

func Internal(data map[string]interface{}, w http.ResponseWriter) {
	jsonResponse, err := json.Marshal(data)
	if err != nil {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_, err = w.Write(jsonResponse)
	return
}
