package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

var client *redis.Client

const Key = "mailProcess"
const TempProcess = "tempProcess"

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
		result, err := client.BLMove(context.Background(), Key, TempProcess, "Right", "Left", 2*time.Second).Result()

		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
			log.Println("error getting data from the list", err)
		}

		log.Println("result", result)
		if len(result) == 0 {
			continue
		}

		err = json.Unmarshal([]byte(result), &process)
		if err != nil {
			log.Println("job info unmarshal issue issue", err)
		}

		log.Println("recieved new process", process)

		go func() {
			err = client.LRem(context.Background(), TempProcess, 0, result).Err()
			if err != nil {
				log.Println("error while removing process from temp process")
			}
		}()
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
