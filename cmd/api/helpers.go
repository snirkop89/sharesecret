package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
)

func (app *application) writeJSON(w http.ResponseWriter, data any, status int, headers ...http.Header) {
	if len(headers) > 0 {
		for k, v := range headers[0] {
			w.Header()[k] = v
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// hash claculates and returns the md5 hash value of a plain text
func md5hash(plaintext string) string {
	// hash the secret using md5
	hash := md5.Sum([]byte(plaintext))
	return fmt.Sprintf("%x", hash)
}
