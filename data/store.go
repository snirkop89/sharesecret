package data

type Store interface {
	Has(key string) bool
	Add(val string) (string, error)
	Get(key string) (string, error)
}
