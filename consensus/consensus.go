package consensus

import (
	"net"
	"os"
	"time"

	"github.com/hashicorp/raft"
	raftmdb "github.com/hashicorp/raft-mdb"
)

type Config struct {
	ServerID string
	Base     string
	Address  string
}

// NewConsensusConfig creates a new consensus config
func NewConsensusConfig() *Config {
	return &Config{}
}

// NewRaft creates a new raft instance with the given fsm and config
func NewRaft(fsm raft.FSM, c *Config) (*raft.Raft, error) {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(c.ServerID)

	store, err := raftmdb.NewMDBStore(c.Base)
	if err != nil {
		return nil, err
	}
	logStore, stableStore := store, store

	snapshotStore, err := raft.NewFileSnapshotStore(c.Base, 2, os.Stderr)
	if err != nil {
		return nil, err
	}

	address, err := net.ResolveTCPAddr("tcp", c.Address)
	if err != nil {
		return nil, err
	}
	transport, err := raft.NewTCPTransport(c.Address, address, 3, 5*time.Second, os.Stderr)
	if err != nil {
		return nil, err
	}

	return raft.NewRaft(config, fsm, logStore, stableStore, snapshotStore, transport)
}
