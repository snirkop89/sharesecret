package handlers_test

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/snirkop89/sharesecret/cmd/api/handlers"
	"github.com/snirkop89/sharesecret/internal/data"
)

var mux *http.ServeMux

type mockStore struct {
	store map[string]string
}

var ms *mockStore

func (m *mockStore) Has(key string) bool {
	if _, found := m.store[key]; found {
		return true
	}
	return false
}

func (m *mockStore) Add(key, val string) error {
	m.store[key] = val
	return nil
}

func (m *mockStore) Get(key string) (string, error) {
	val, found := m.store[key]
	if !found {
		return "", data.ErrNotFound
	}
	fmt.Println("found", val)
	delete(m.store, key)
	return val, nil
}

func TestMain(m *testing.M) {
	setupTest()
	os.Exit(m.Run())
}

func setupTest() {
	l := log.New(os.Stdout, "", log.LstdFlags)
	ms = &mockStore{
		store: make(map[string]string),
	}
	sh := handlers.NewSecretsHandler(l, ms)
	mux = http.NewServeMux()
	mux.HandleFunc("/", sh.Secret)
	mux.HandleFunc("/healthcheck", sh.Healthcheck)
}
