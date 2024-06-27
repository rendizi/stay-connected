package server

import (
	"encoding/json"
	"net/http"
)

func Bad(data map[string]interface{}, w http.ResponseWriter) {
	jsonResponse, err := json.Marshal(data)
	if err != nil {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_, err = w.Write(jsonResponse)
	return
}
