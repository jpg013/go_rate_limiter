package main

import (
	"encoding/json"
	"log"
	"net/http"

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

	if type, ok := body["rate_limit_type"]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if token, ok := body["token"]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if manager, ok := s.rateLimitManagers[type]; !ok {
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

	if type, ok := body["rate_limit_type"]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if manager, ok := s.rateLimitManagers[type]; !ok {
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

func main() {
	server := Server{
		rateLimitManagers: make(map[string]*ratelimit.Manager),
	}

	// init each rate limiter
	for _, n := range rateLimitTypes {
		m, err := ratelimit.NewManager(n)

		if err != nil {
			panic(err)
		}

		server.rateLimitManagers[n] = m
	}

	r := mux.NewRouter()
	r.HandleFunc("/acquire_rate_limit", server.HandleAcquireRateLimit).Methods(http.MethodPost)
	r.HandleFunc("/release_rate_limit", server.HandleReleaseRateLimit).Methods(http.MethodDelete)

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
