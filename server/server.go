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

type Server struct {
	httpServer *http.Server
	store      *store.Store[model.Comment]
	logger     *slog.Logger
}

func NewServer(store *store.Store[model.Comment], logger *slog.Logger) *Server {
	return &Server{
		store:  store,
		logger: logger,
	}
}

func (s *Server) StartServer(ctx context.Context, conf *Config) {
	mux := http.NewServeMux()

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
