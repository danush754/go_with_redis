package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

var client *redis.Client

func init() {

	client = redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	err := client.Ping(context.Background()).Err()
	if err != nil {
		log.Fatalf("failed to connect to the redis server: %v", err)
	}
}

func main() {
	r := mux.NewRouter()

	// r.HandleFunc("/send", send).Methods(http.MethodPost)

	log.Println("started HTTP server...")
	log.Fatal(http.ListenAndServe(":8080", r))
}
