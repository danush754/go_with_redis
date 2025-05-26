package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

var client *redis.Client

const (
	userHashnamePrefix = "user:"
	userIdCounter      = "userid_counter"
	userSetKey         = "users"
	gameLeaderBoard    = "leaderboard"
)

func init() {
	client = redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancelCtx()
	err := client.Ping(ctx).Err()
	if err != nil {
		log.Fatalf("failed to connect to redis server.. - %v", err)
	} else {
		for i := 3; i >= 1; i-- {
			fmt.Println(i)
			time.Sleep(1 * time.Second)

		}
		fmt.Println("sucessfully connected to the local redis server")
	}

	err = client.Del(context.Background(), userSetKey).Err()
	if err != nil {
		log.Println("could not delete set", userSetKey, err)
	}

	err = client.Del(context.Background(), gameLeaderBoard).Err()
	if err != nil {
		log.Println("could not delete sorted set", gameLeaderBoard, err)
	}

	for i := 1; i <= 10; i++ {
		err = client.SAdd(context.Background(), userSetKey, "user-"+strconv.Itoa(i)).Err()
		if err != nil {
			log.Println("could not able to add user to set", err)
		}
	}

	// users, err := client.SMembers(context.Background(), userSet).Result()
	// if err != nil {
	// 	log.Println("could not able to fetch users from the set", err)
	// }

	// log.Println("user list ", users)
}

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		for i := 1; i < 4; i++ {
			fmt.Println(i)
			time.Sleep(1 * time.Second)

		}
		fmt.Println("Application is ready... ")
	}).Methods(http.MethodGet)

	// routes for user-management
	// r.HandleFunc("/", add).Methods(http.MethodPost)

	// r.HandleFunc("/{id}", get).Methods(http.MethodGet)

	// r.HandleFunc("/del-field/{id}", delField).Methods(http.MethodPatch)

	// routes for leaderboard

	r.HandleFunc("/add-user", adduser).Methods(http.MethodPost)
	r.HandleFunc("/play", play).Methods(http.MethodGet)
	r.HandleFunc("/leaderboard/{n}", leaderboard).Methods(http.MethodGet)
	r.HandleFunc()

	log.Println("HTTP server has been started..")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func add(w http.ResponseWriter, r *http.Request) {
	var user map[string]string

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Println("failed to decode the json payload", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("user", user)

	id, err := client.Incr(context.Background(), userIdCounter).Result()
	if err != nil {
		log.Println("failed to generate userid", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("id:", id)

	userHashName := userHashnamePrefix + strconv.Itoa(int(id))

	err = client.HSet(r.Context(), userHashName, user).Err()
	if err != nil {
		log.Println("failed top save user: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Location", "http://"+r.Host+"/"+strconv.Itoa(int(id)))
	w.WriteHeader(http.StatusCreated)

	log.Println("User added successfully: ", userHashName)

}

func get(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	id := vars["id"]

	log.Println("searching for id", id)

	userHash := userHashnamePrefix + id

	user, err := client.HGetAll(r.Context(), userHash).Result()
	if err != nil {
		log.Println("error fetching user", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(user) == 0 {
		log.Println("user with id", id, "not found")
		http.Error(w, "user does not exit", http.StatusNotFound)
		return
	}

	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		log.Println("failed to encode user data", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func delField(w http.ResponseWriter, r *http.Request) {

	variables := mux.Vars(r)

	id := variables["id"]

	userHash := userHashnamePrefix + id

	var userData map[string]string
	err := json.NewDecoder(r.Body).Decode(&userData)
	if err != nil {
		log.Println("failed to decode the json payload", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = client.HDel(r.Context(), userHash, "city").Err()
	if err != nil {
		log.Println("failed to delete field from the user", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func adduser(w http.ResponseWriter, r *http.Request) {

	newUser, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("failed to read payload", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userExists, err := client.SIsMember(context.Background(), userSetKey, string(newUser)).Result()
	if err != nil {
		log.Println("failed to check user in the set", string(newUser), err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !userExists {
		err = client.SAdd(context.Background(), userSetKey, string(newUser)).Err()
		if err != nil {
			log.Println("failed to add user in the set", string(newUser), err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Println("user added successfully", string(newUser))

	} else {
		log.Println("user already exists in the set", string(newUser))
		w.WriteHeader(http.StatusConflict)
		return
	}

}

func play(w http.ResponseWriter, r *http.Request) {
	go func() {

		log.Println("game simulation has been started...")

		userList, err := client.SMembers(context.Background(), userSetKey).Result()
		if err != nil {
			log.Println("unable to get the userlist	")
			return
		}

		for _, user := range userList {
			_, err := client.ZIncrBy(context.Background(), gameLeaderBoard, float64(rand.Intn(20)+1), user).Result()
			if err != nil {
				log.Println("couldn't able to increment score for the player", err)
				return
			}
		}

		time.Sleep(5 * time.Second)

	}()

	w.WriteHeader(http.StatusAccepted)
}

func leaderboard(w http.ResponseWriter, r *http.Request) {

	queryParams := mux.Vars(r)
	n := queryParams["n"]

	log.Println("will fetch top ", n, "players")

	num, _ := strconv.Atoi(n)

	topPlayers, err := client.ZRevRangeWithScores(context.Background(), gameLeaderBoard, 0, int64(num-1)).Result()
	if err != nil {
		log.Println("failed to get the top ", num, " players")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(topPlayers)
	if err != nil {
		log.Println("failed to encode leaderboard info")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("sucessfully fetched leaderboard info")
}
