package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

var client *redis.Client

const Key = "mailProcess"

func init() {

	client = redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	err := client.Ping(context.Background()).Err()
	if err != nil {
		log.Fatalf("failed to connect to the redis server: %v", err)
	}

	log.Println("consumer ready")
}

func main() {

	for {

		var process ProcessInfo
		result, err := client.BRPop(context.Background(), 2*time.Second, Key).Result()

		if err != nil {
			log.Println("error getting data from the list", err)
		}

		log.Println("result", result)
		if len(result) == 0 {
			continue
		}

		data := result[1]

		err = json.Unmarshal([]byte(data), &process)
		if err != nil {
			log.Println("job info unmarshal issue issue", err)
		}

		log.Println("recieved new process", process)
	}

}

type Email struct {
	ToWhom      string `json:"to"`
	WhatMessage string `json:"message"`
}

type ProcessInfo struct {
	Email     Email  `json:"email"`
	ProcessId string `json:"id"`
}
