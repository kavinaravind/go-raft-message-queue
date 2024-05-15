package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/kavinaravind/go-raft-message-queue/model"
	"github.com/kavinaravind/go-raft-message-queue/store"
)

type Config struct {
	Address string
}

func NewServerConfig() *Config {
	return &Config{}
}

// Server is a simple HTTP server that listens for incoming requests
type Server struct {
	httpServer *http.Server
	store      *store.Store[model.Comment]
	logger     *slog.Logger
}

// NewServer creates a new instance of the Server
func NewServer(store *store.Store[model.Comment], logger *slog.Logger) *Server {
	return &Server{
		store:  store,
		logger: logger,
	}
}

// Intitiliaze starts the HTTP server
func (s *Server) Intitiliaze(ctx context.Context, conf *Config) {
	s.logger.Info("Initializing server")

	mux := http.NewServeMux()

	mux.HandleFunc("/send", s.handleSend)
	mux.HandleFunc("/recieve", s.handleRecieve)
	mux.HandleFunc("/stats", s.handleStats)
	mux.HandleFunc("/join", s.handleJoin)

	s.httpServer = &http.Server{
		Addr:    conf.Address,
		Handler: mux,
	}

	go func() {
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			s.logger.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	go func() {
		<-ctx.Done()
		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.Error("Failed to shutdown server", "error", err)
		}
	}()
}

func (s *Server) handleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var message model.Comment
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		http.Error(w, "Failed to decode message", http.StatusBadRequest)
		return
	}

	if err := s.store.Send(message); err != nil {
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) handleRecieve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	message, err := s.store.Recieve()
	if err != nil {
		http.Error(w, "Failed to recieve message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(message)
	if err != nil {
		http.Error(w, "Failed to encode message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := s.store.Stats()

	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(stats)
	if err != nil {
		http.Error(w, "Failed to encode message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleJoin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	body := map[string]string{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Failed to decode body", http.StatusBadRequest)
		return
	}

	if len(body) != 2 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	remoteAddress, ok := body["address"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nodeID, ok := body["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := s.store.Join(nodeID, remoteAddress); err != nil {
		http.Error(w, "Failed to join cluster", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
