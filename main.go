package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/kavinaravind/go-raft-message-queue/consensus"
	"github.com/kavinaravind/go-raft-message-queue/model"
	"github.com/kavinaravind/go-raft-message-queue/server"
	"github.com/kavinaravind/go-raft-message-queue/store"
)

type config struct {
	JoinAddress string
	Concensus   *consensus.Config
	Server      *server.Config
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
	flag.BoolVar(&conf.Concensus.IsLeader, "leader", false, "Set to true if this node is the leader")
	flag.StringVar(&conf.Concensus.ServerID, "id", "", "The unique identifier for this server")
	flag.StringVar(&conf.Concensus.Address, "raddr", "localhost:3001", "The address that the Raft consensus group should use")
	flag.StringVar(&conf.Concensus.BaseDirectory, "dir", "/tmp", "The base directory for storing Raft data")
	flag.StringVar(&conf.JoinAddress, "paddr", "", "The address of an existing node to join")

	// Server Specific Flags
	flag.StringVar(&conf.Server.Address, "haddr", "localhost:3000", "The address that the HTTP server should use")

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

	// Create the base directory if it does not exist
	if err := os.MkdirAll(conf.Concensus.BaseDirectory, 0755); err != nil {
		logger.Error("Failed to create base directory", "error", err)
		os.Exit(1)
	}

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Create a new store instance with the given logger
	store := store.NewStore[model.Comment](logger)

	// Initialize the store
	nodeShutdownComplete, err := store.Initialize(ctx, conf.Concensus)
	if err != nil {
		logger.Error("Failed to initialize store", "error", err)
		os.Exit(1)
	}

	// Create a new instance of the server
	server := server.NewServer(store, logger)

	// Initialize the server
	serverShutdownComplete := server.Initialize(ctx, conf.Server)

	// If join was specified, make the join request.
	if conf.JoinAddress != "" {
		b, err := json.Marshal(map[string]string{"address": conf.Concensus.Address, "id": conf.Concensus.ServerID})
		if err != nil {
			logger.Error("Failed to marshal join request", "error", err)
			os.Exit(1)
		}
		resp, err := http.Post(fmt.Sprintf("http://%s/join", conf.JoinAddress), "application-type/json", bytes.NewReader(b))
		if err != nil {
			logger.Error("Failed to send join request", "error", err)
			os.Exit(1)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			logger.Error("Received non-OK response to join request", "status", resp.StatusCode)
			os.Exit(1)
		}
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs
	switch sig {
	case syscall.SIGINT:
		logger.Info("Received SIGINT, shutting down")
	case syscall.SIGTERM:
		logger.Info("Received SIGTERM, shutting down")
	}

	// Cancel the context to stop the HTTP server and consensus node
	cancel()

	// Wait for the server and consensus node to finish shutting down
	<-nodeShutdownComplete
	<-serverShutdownComplete
}
