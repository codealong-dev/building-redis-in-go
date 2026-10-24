package store

import "errors"

var ErrEmptyKey = errors.New("key is mandatory")
var ErrStoreFull = errors.New("store is full")
var ErrNotInteger = errors.New("value is not an integer")

type Storer interface {
	Get(key string) (string, error)
	Set(key, value string) error
	Incr(key string) (int, error)
	Keys() []string
	Delete(key string)
	Len() int
}
