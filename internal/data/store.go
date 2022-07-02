package data

import "errors"

var ErrNotFound = errors.New("id not found")

type Store interface {
	Has(key string) bool
	Add(key, val string) error
	Get(key string) (string, error)
}
