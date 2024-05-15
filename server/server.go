package server

import (
	"context"
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
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

}

func (s *Server) handleRecieve(w http.ResponseWriter, r *http.Request) {
	// handle recieve
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	// handle health
}
