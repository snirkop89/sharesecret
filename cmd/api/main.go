package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/snirkop89/sharesecret/cmd/api/handlers"
	"github.com/snirkop89/sharesecret/internal/data"
)

var port int

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	flag.IntVar(&port, "port", 8080, "port to listen on")
	flag.Parse()

	if os.Getenv("DATA_FILE_PATH") == "" {
		return errors.New("a file location is missing. specify it using DATA_FILE_PATH environment variable")
	}

	store, err := data.NewFileStore(os.Getenv("DATA_FILE_PATH"))
	if err != nil {
		fmt.Printf("unexpected error: %v", err)
		os.Exit(1)
	}

	logger := log.New(os.Stdout, "secret-app\t", log.LstdFlags)

	sh := handlers.NewSecretsHandler(logger, store)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthcheck", sh.Healthcheck)
	mux.HandleFunc("/", sh.Secret)

	srv := http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
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
