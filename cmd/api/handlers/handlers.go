package handlers

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/snirkop89/sharesecret/internal/data"
)

type Secrets struct {
	l     *log.Logger
	store data.Store
}

func NewSecretsHandler(l *log.Logger, store data.Store) *Secrets {
	return &Secrets{
		l:     l,
		store: store,
	}
}

func (s *Secrets) writeJSON(w http.ResponseWriter, data any, status int, headers ...http.Header) {
	if len(headers) > 0 {
		for k, v := range headers[0] {
			w.Header()[k] = v
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	content, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(content)
}

func (s *Secrets) Healthcheck(w http.ResponseWriter, r *http.Request) {
	payload := struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	}
	s.writeJSON(w, payload, http.StatusOK)
}

func (s *Secrets) Secret(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getSecret(w, r)
	case http.MethodPost:
		s.saveSecret(w, r)
	default:
		http.Error(w, "not supported", http.StatusMethodNotAllowed)
		return
	}
}

type secretResponse struct {
	Data string `json:"data"`
}

func (s *Secrets) getSecret(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/")
	if id == "" {
		s.writeJSON(w, secretResponse{}, http.StatusBadRequest)
		return
	}

	secret, err := s.store.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNotFound):
			s.writeJSON(w, secretResponse{}, http.StatusNotFound)
		default:
			s.writeJSON(w, secretResponse{}, http.StatusInternalServerError)
		}
		return
	}

	s.writeJSON(w, secretResponse{Data: secret}, http.StatusOK)
}

func (s *Secrets) saveSecret(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "" {
		if r.Header.Get("Content-Type") != "application/json" {
			msg := "Content-Type header is not application/json"
			s.writeJSON(w, msg, http.StatusBadRequest)
			return
		}
	}

	// limit the body size to 1MB
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	secret := struct {
		PlainText string `json:"plain_text"`
	}{}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&secret)
	if err != nil {
		var syntaxError *json.SyntaxError
		var typeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &typeError):
			s.writeJSON(w, "request body contatins an invalid value", http.StatusBadRequest)
		case errors.As(err, &syntaxError):
			s.writeJSON(w, "badly formed json", http.StatusBadRequest)
		case errors.Is(err, io.ErrUnexpectedEOF):
			s.writeJSON(w, "badly formed json", http.StatusBadRequest)
		case errors.Is(err, io.EOF):
			s.writeJSON(w, "Request body must not be empty", http.StatusBadRequest)
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			s.writeJSON(w, msg, http.StatusBadRequest)
		default:
			s.writeJSON(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	hash := md5hash(secret.PlainText)

	err = s.store.Add(hash, secret.PlainText)
	if err != nil {
		s.writeJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// return the id for the user
	idResponse := struct {
		ID string `json:"id"`
	}{
		ID: hash,
	}

	s.writeJSON(w, idResponse, http.StatusCreated)
}

// hash claculates and returns the md5 hash value of a plain text
func md5hash(plaintext string) string {
	// hash the secret using md5
	hash := md5.Sum([]byte(plaintext))
	return fmt.Sprintf("%x", hash)
}
