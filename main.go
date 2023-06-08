package main

import "errors"

var (
	ErrNoSuchKey = errors.New("no such key")
)

type store map[string]string

func New() store {
	return store{}
}

func (s store) Put(key, value string) error {
	s[key] = value
	return nil
}

func (s store) Get(key string) (string, error) {
	value, ok := s[key]
	if !ok {
		return "", ErrNoSuchKey
	}

	return value, nil
}

func (s store) Delete(key string) error {
	delete(s, key)

	return nil
}

func main() {
}
