package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jpg013/ratelimit"
)

var rateLimiterTypes = []string{"crawlera"}

type Server struct {
	rateLimitManagers map[string]*ratelimit.Manager
}

type RateLimitResponse struct {
	Token string `json:"token"`
}

func (s *Server) HandleReleaseRateLimit(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	err := json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	rateLimitType, ok := body["rate_limit_type"]

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token, ok := body["token"]

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	manager, ok := s.rateLimitManagers[rateLimitType]

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Release the token
	manager.Release(token)

	// Return OK status
	w.WriteHeader(http.StatusOK)
}

func (s *Server) HandleAcquireRateLimit(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	err := json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	rateLimitType, ok := body["rate_limit_type"]

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	manager, ok := s.rateLimitManagers[rateLimitType]

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Acquire a rate limit resource token
	token := manager.Acquire()

	jsonBytes, err := json.Marshal(RateLimitResponse{
		Token: token,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}

func fireRequests() {
	doWork := func(id int) {
		// Simulate some work
		n := rand.Intn(5)
		fmt.Printf("Worker %d Sleeping %d seconds...\n", id, n)
		time.Sleep(time.Duration(n) * time.Second)
		fmt.Printf("Worker %d Done\n", id)
	}

	acquireToken := func() string {
		requestBody, err := json.Marshal(map[string]string{
			"rate_limit_type": "crawlera",
		})

		if err != nil {
			panic(err)
		}

		resp, err := http.Post("http://localhost:8080/acquire_rate_limit", "application/json", bytes.NewBuffer(requestBody))

		if err != nil {
			panic(err)
		}

		defer resp.Body.Close()

		var respBody map[string]string

		err = json.NewDecoder(resp.Body).Decode(&respBody)

		if err != nil {
			panic(err)
		}

		token, ok := respBody["token"]

		if !ok {
			panic("invalid token")
		}

		return token
	}

	releaseToken := func(token string) {
		requestBody, err := json.Marshal(map[string]string{
			"rate_limit_type": "crawlera",
			"token":           token,
		})

		if err != nil {
			panic(err)
		}

		resp, err := http.Post("http://localhost:8080/release_rate_limit", "application/json", bytes.NewBuffer(requestBody))

		if err != nil {
			panic(err)
		}

		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			panic("Error releasing token")
		}
	}

	for i := 0; i < 100; i++ {
		go func(id int) {
			token := acquireToken()
			doWork(id)
			releaseToken(token)
		}(i)
	}
}

func main() {
	server := Server{
		rateLimitManagers: make(map[string]*ratelimit.Manager),
	}

	// init each rate limiter
	for _, n := range rateLimiterTypes {
		m, err := ratelimit.NewManager(n)

		if err != nil {
			panic(err)
		}

		server.rateLimitManagers[n] = m
	}

	r := mux.NewRouter()
	r.HandleFunc("/acquire_rate_limit", server.HandleAcquireRateLimit).Methods(http.MethodPost)
	r.HandleFunc("/release_rate_limit", server.HandleReleaseRateLimit).Methods(http.MethodPost)

	go fireRequests()

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
