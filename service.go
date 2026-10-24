package main

import (
	"codealong/store"
	"encoding/base64"
)

func SetKeyWithEncryption(store store.Storer, key, val string) (string, error) {
	encoded := base64.StdEncoding.EncodeToString([]byte(val))
	if err := store.Set(key, encoded); err != nil {
		return "", err
	}

	return store.Get(key)
}
