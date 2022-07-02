package main

import "net/http"

func (app *application) writeError(w http.ResponseWriter, err string, status ...int) {
	respStatus := http.StatusInternalServerError
	if len(status) > 0 {
		respStatus = status[0]
	}

	payload := struct {
		Error string `json:"error"`
	}{
		Error: err,
	}
	app.writeJSON(w, payload, respStatus)
}
