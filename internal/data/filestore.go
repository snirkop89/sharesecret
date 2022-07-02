package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"sync"
)

type FileStore struct {
	FilePath string
	Store    map[string]string
	mu       sync.Mutex
}

// initialize creates a file if it does not exist
func initialize(file string) error {
	if _, err := os.Stat(file); err != nil && errors.Is(err, fs.ErrNotExist) {
		_, err := os.Create(file)
		if err != nil {
			return fmt.Errorf("failed creating file: %w", err)
		}
	}

	return nil
}

// NewFileStore returns a file-based store which implements the store interface
func NewFileStore(file string) (*FileStore, error) {
	if err := initialize(file); err != nil {
		return nil, err
	}

	f := &FileStore{
		mu:       sync.Mutex{},
		FilePath: file,
		Store:    make(map[string]string),
	}

	return f, nil
}

func (store *FileStore) Has(key string) bool {
	if _, found := store.Store[key]; found {
		return true
	}
	return false
}

func (store *FileStore) Add(key, val string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	// if store already has the value, return
	if store.Has(key) {
		return nil
	}

	// save the value in-memory store
	store.Store[key] = val

	// update the file
	if err := store.save(); err != nil {
		return err
	}

	return nil
}

func (store *FileStore) Get(key string) (string, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	// update the store with latest
	err := store.load()
	if err != nil {
		return "", err
	}

	if !store.Has(key) {
		return "", ErrNotFound
	}

	// extract the value and delete if from the store
	val := store.Store[key]
	delete(store.Store, key)

	err = store.save()
	if err != nil {
		return "", err
	}

	return val, nil
}

// save dumps the store content into the file
func (store *FileStore) save() error {
	f, err := os.Create(store.FilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(store.Store)
}

// load reads the file content into the in-memory store
func (store *FileStore) load() error {
	f, err := os.Open(store.FilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	// unmarshal the file's content
	return json.Unmarshal(data, &store.Store)
}
