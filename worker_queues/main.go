package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

var client *redis.Client

const Key = "mailProcess"

func init() {

	client = redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	err := client.Ping(context.Background()).Err()
	if err != nil {
		log.Fatalf("failed to connect to the redis server: %v", err)
	}
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Worker Queue sample server started successfully")
	}).Methods(http.MethodGet)

	r.HandleFunc("/send", SendMail).Methods(http.MethodPost)

	log.Println("started HTTP server...")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func SendMail(w http.ResponseWriter, r *http.Request) {

	var email Email

	err := json.NewDecoder(r.Body).Decode(&email)
	if err != nil {
		log.Println("Error while decoding the body", err)
	}

	log.Println("email", email)

	var processId = strconv.Itoa(rand.Intn(100))
	var processInfo = ProcessInfo{Email: email, ProcessId: processId}

	log.Println("processData", processInfo)

	process, err := json.Marshal(processInfo)
	if err != nil {
		log.Println("error matshaling json", err)
	}

	err = client.LPush(context.Background(), Key, process).Err()
	if err != nil {
		log.Println("error while pushing data to the redis queue", err)
	}

	w.Header().Add("processId", processId)
}

type Email struct {
	ToWhom      string `json:"to"`
	WhatMessage string `json:"message"`
}

type ProcessInfo struct {
	Email     Email  `json:"email"`
	ProcessId string `json:"id"`
}
