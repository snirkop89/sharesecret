package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/snirkop89/sharesecret/internal/data"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	payload := struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	}
	app.writeJSON(w, payload, http.StatusOK)
}

func (app *application) secretHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		app.getSecret(w, r)
	case http.MethodPost:
		app.saveSecret(w, r)
	default:
		http.Error(w, "not supported", http.StatusMethodNotAllowed)
		return
	}
}

type secretResponse struct {
	Data string `json:"data"`
}

func (app *application) getSecret(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/")
	if id == "" {
		app.writeJSON(w, secretResponse{}, http.StatusBadRequest)
		return
	}

	secret, err := app.store.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrNotFound):
			app.writeJSON(w, secretResponse{}, http.StatusNotFound)
		default:
			app.writeError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	app.writeJSON(w, secretResponse{Data: secret}, http.StatusOK)
}

func (app *application) saveSecret(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "" {
		if r.Header.Get("Content-Type") != "application/json" {
			msg := "Content-Type header is not application/json"
			app.writeError(w, msg, http.StatusBadRequest)
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
		app.writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hash := md5hash(secret.PlainText)

	err = app.store.Add(hash, secret.PlainText)
	if err != nil {
		app.writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// return the id for the user
	idResponse := struct {
		ID string `json:"id"`
	}{
		ID: hash,
	}

	app.writeJSON(w, idResponse, http.StatusCreated)
}
