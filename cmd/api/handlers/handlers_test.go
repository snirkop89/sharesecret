package handlers_test

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSecretHandler(t *testing.T) {
	testCases := []struct {
		name     string
		method   string
		URL      string
		body     string
		expected int
		err      error
	}{
		{name: "test put method to root", method: http.MethodPut, URL: "/", body: "", expected: http.StatusMethodNotAllowed, err: nil},
		{name: "test get method to root", method: http.MethodGet, URL: "/hkjdhfjkas", body: "", expected: http.StatusNotFound, err: nil},
		{name: "test get with no id", method: http.MethodGet, URL: "/", body: "", expected: http.StatusBadRequest, err: nil},
		{name: "test good post method to root", method: http.MethodPost, URL: "/", body: `{"plain_text": "secret123"}`, expected: http.StatusCreated, err: nil},
		{name: "test post method with no body", method: http.MethodPost, URL: "/", body: ``, expected: http.StatusBadRequest, err: nil},
		{name: "test post method with no bad body", method: http.MethodPost, URL: "/", body: `{"not_good": "secret123"}`, expected: http.StatusBadRequest, err: nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// make request
			content := strings.NewReader(tc.body)
			req := httptest.NewRequest(tc.method, tc.URL, bufio.NewReader(content))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			resp := w.Result()

			if resp.StatusCode != tc.expected {
				t.Errorf("expected code %d, got %d", tc.expected, w.Code)
			}
		})
	}
}

func TestSaveSecret(t *testing.T) {
	req, _ := http.NewRequest("POST", "/", strings.NewReader(`{"plain_text": "verysecret"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	rr.Result()

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}
}

func TestGetSecret(t *testing.T) {
	// add key to store
	hash := md5.Sum([]byte("verysecret"))
	ms.Add(fmt.Sprintf("%x", hash), "verysecret")

	req, _ := http.NewRequest("GET", fmt.Sprintf("/%x", hash), nil)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("exptected status %d, got %d", rr.Code, http.StatusOK)
	}

	if !strings.Contains(rr.Body.String(), "verysecret") {
		t.Errorf("exptected to find 'verysecret' did not find it in body")
	}

	// try again, body should be empty
	req, _ = http.NewRequest("GET", fmt.Sprintf("/%x", hash), nil)
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if strings.Contains(rr.Body.String(), "verysecret") {
		t.Errorf("did not expect to find 'verysecret' in body, but did")
	}

}

func TestHealthcheck(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthcheck", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, w.Code)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "ok") {
		t.Errorf("exptected 'ok' in response, got %v", string(body))
	}
}
