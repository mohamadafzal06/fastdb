package main

import (
	"errors"
	"fmt"
	"sync"
)

var (
	ErrNoSuchKey = errors.New("no such key")
)

type store struct {
	m map[string]string
	l *sync.RWMutex

	logger TransactionLogger
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

func (s store) initializeTransactionLog() error {
	var err error

	logger, err := NewFileTransactionLogger("transaction.log")
	if err != nil {
		return fmt.Errorf("failed to create event logger: %w", err)
	}
	s.logger = logger

	events, errors := s.logger.ReadEvents()
	e, ok := Event{}, true

	for ok && err == nil {
		select {
		case err, ok = <-errors:
		case e, ok = <-events:
			switch e.EventType {
			case EventDelete:
				err = s.Delete(e.Key)
			case EventPut:
				err = s.Put(e.Key, e.Value)
			}

		}
	}

	s.logger.Run()
	return err
}
