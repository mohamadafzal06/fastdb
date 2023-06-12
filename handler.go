package main

import (
	"errors"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

type Handler struct {
	store  *store
	logger TransactionLogger
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
	h.logger.WritePut(key, string(value))

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

func (h Handler) keyValueDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	err := h.store.Delete(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.logger.WriteDelete(key)

	w.WriteHeader(http.StatusCreated)
}
