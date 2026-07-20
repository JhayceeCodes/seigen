package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func Health(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "ok",
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode((response))
}

func main() {
	http.HandleFunc("/health", Health)

	fmt.Println("Server running on: 8080")

	http.ListenAndServe(":8080", nil)
}
