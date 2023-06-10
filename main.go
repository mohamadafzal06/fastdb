package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
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

type Handler struct {
	store *store
}

func NewHandler(s *store) Handler {
	return Handler{
		store: s,
	}
}

func (h Handler) keyValuePutHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	value, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.store.Put(key, string(value))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h Handler) keyValueGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	value, err := h.store.Get(key)
	if errors.Is(err, ErrNoSuchKey) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(value))
}

func main() {
	s := New(&sync.RWMutex{})
	r := mux.NewRouter()
	h := NewHandler(&s)

	r.HandleFunc("/v1/{key}", h.keyValuePutHandler).Methods("PUT")
	r.HandleFunc("/v1/{key}", h.keyValueGetHandler).Methods("GET")

	log.Fatal(http.ListenAndServe(":8888", r))

}
