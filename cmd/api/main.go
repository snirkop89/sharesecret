package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/snirkop89/sharesecret/internal/data"
)

type application struct {
	port  int
	mu    *sync.Mutex
	store data.Store
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	if os.Getenv("DATA_FILE_PATH") == "" {
		return errors.New("a file location is missing. specify it using DATA_FILE_PATH environment variable")
	}

	port := 8080
	if os.Getenv("SECRETS_PORT") != "" {
		p, err := strconv.Atoi("SECRETS_PORT")
		if err != nil {
			return fmt.Errorf("not an integer in SECRETS_PORT: %w", err)
		}
		port = p
	}

	store, err := data.NewFileStore(os.Getenv("DATA_FILE_PATH"))
	if err != nil {
		fmt.Printf("unexpected error: %v", err)
		os.Exit(1)
	}

	app := application{
		port:  port,
		mu:    &sync.Mutex{},
		store: store,
	}

	mux := http.ServeMux{}
	mux.HandleFunc("/", app.secretHandler)
	mux.HandleFunc("/healthcheck", app.healthcheckHandler)

	srv := http.Server{
		Addr:         fmt.Sprintf(":%d", app.port),
		Handler:      &mux,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  10 * time.Second,
	}

	err = srv.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}
