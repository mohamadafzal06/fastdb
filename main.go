package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

func main() {
	s := New(&sync.RWMutex{})
	r := mux.NewRouter()
	h := NewHandler(&s)

	r.HandleFunc("/v1/{key}", h.keyValuePutHandler).Methods("PUT")
	r.HandleFunc("/v1/{key}", h.keyValueGetHandler).Methods("GET")

	log.Fatal(http.ListenAndServe(":8888", r))

}
