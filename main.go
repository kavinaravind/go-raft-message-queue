package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/kavinaravind/go-raft-message-queue/consensus"
	"github.com/kavinaravind/go-raft-message-queue/model"
	"github.com/kavinaravind/go-raft-message-queue/server"
	"github.com/kavinaravind/go-raft-message-queue/store"
)

type config struct {
	Concensus *consensus.Config
	Server    *server.Config
}

func newConfig() *config {
	return &config{
		Concensus: consensus.NewConsensusConfig(),
		Server:    server.NewServerConfig(),
	}
}

var conf *config

func init() {
	conf = newConfig()

	// Concensus Specific Flags
	flag.StringVar(&conf.Concensus.ServerID, "id", "node01", "")
	flag.StringVar(&conf.Concensus.Address, "raddr", "localhost:3001", "")
	flag.StringVar(&conf.Concensus.BaseDirectory, "dir", "/tmp", "")

	// Server Specific Flags
	flag.StringVar(&conf.Server.Address, "haddr", "localhost:3000", "")

	// Set Usage Details
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	logger := slog.Default()

	if conf.Concensus.ServerID == "" {
		logger.Error("The -id flag is required")
		os.Exit(2)
	}

	// Create a new store instance with the given logger
	store := store.NewStore[model.Comment](logger)

	// Initialize the store
	err := store.Initialize(conf.Concensus)
	if err != nil {
		logger.Error("Failed to initialize store", "error", err)
		os.Exit(1)
	}

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Create a new instance of the server
	server := server.NewServer(store, logger)

	// Intitiliaze the server
	server.Intitiliaze(ctx, conf.Server)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs
	switch sig {
	case syscall.SIGINT:
		logger.Info("Received SIGINT, shutting down")
	case syscall.SIGTERM:
		logger.Info("Received SIGTERM, shutting down")
	}

	logger.Info("Shutting down server")

	// Cancel the context to stop the HTTP server
	cancel()

}
