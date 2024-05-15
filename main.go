package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	consensus "github.com/kavinaravind/go-raft-message-queue/concensus"
)

type config struct {
	concensus *consensus.Config
	httpAddr  *
}

var config *consensus.Config

func init() {
	config = consensus.NewConsensusConfig()

	// Raft Specific Flags
	flag.StringVar(&config.ServerID, "id", "", "")
	flag.StringVar(&config.Address, "raddr", "localhost", "")
	flag.StringVar(&config.Base, "join", "", "")

	// Server Specific Flags
	flag.StringVar(&httpAddr, "haddr", "", "")

	// Set usage details
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

}

func main() {
	flag.Parse()

	if config.ServerID == "" {
		slog.Error("The -id flag is required")
		os.Exit(2)
	}

	s := store.NewStore()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-sigs:
		switch sig {
		case syscall.SIGINT:
			slog.Info("Received SIGINT, shutting down")
		case syscall.SIGTERM:
			slog.Info("Received SIGTERM, shutting down")
		}
	}
}
