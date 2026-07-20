package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func health(w http.ResponseWriter, r *http.Request) {

	response := map[string]string{
		"message": "Server running fine",
		"status":  "ok",
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)

}

func main() {

	http.HandleFunc("/health", health)

	fmt.Println("Server running on 8080.")

	http.ListenAndServe(":8080", nil)

}
