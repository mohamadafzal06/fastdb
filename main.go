package main

import (
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

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

type Handler struct {
	store store
}

func NewHandler(s store) Handler {
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
	s := New()
	r := mux.NewRouter()
	h := NewHandler(s)

	r.HandleFunc("/v1/{key}", h.keyValuePutHandler).Methods("PUT")
	r.HandleFunc("/v1/{key}", h.keyValueGetHandler).Methods("GET")

	log.Fatal(http.ListenAndServe(":8888", r))

}
