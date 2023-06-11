package main

import (
	"errors"
	"sync"
)

var (
	ErrNoSuchKey = errors.New("no such key")
)

type store struct {
	m map[string]string
	l *sync.RWMutex
}

func New(l *sync.RWMutex) store {
	return store{
		m: make(map[string]string),
		l: l,
	}
}

func (s store) Put(key, value string) error {
	s.l.Lock()
	s.m[key] = value
	defer s.l.Unlock()
	return nil
}

func (s store) Get(key string) (string, error) {
	s.l.RLock()
	value, ok := s.m[key]
	if !ok {
		return "", ErrNoSuchKey
	}
	defer s.l.RUnlock()

	return value, nil
}

func (s store) Delete(key string) error {
	s.l.Lock()

	delete(s.m, key)

	defer s.l.Unlock()

	return nil
}
