package consensus

import (
	"net"
	"os"
	"time"

	"github.com/hashicorp/raft"
	raftmdb "github.com/hashicorp/raft-mdb"
)

type consensusConfig struct {
	localID string
	base    string
	address string
}

func NewRaftConfig(localID, base, address string) *consensusConfig {
	return &consensusConfig{
		localID: localID,
		base:    base,
		address: address,
	}
}

func NewRaft(fsm raft.FSM, c *consensusConfig) (*raft.Raft, error) {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(c.localID)

	store, err := raftmdb.NewMDBStore(c.base)
	if err != nil {
		return nil, err
	}
	logStore, stableStore := store, store

	snapshotStore, err := newRaftFileSnapshotStore(c.base)
	if err != nil {
		return nil, err
	}

	transport, err := newRaftTCPTransport("", c.address)
	if err != nil {
		return nil, err
	}

	return raft.NewRaft(config, fsm, logStore, stableStore, snapshotStore, transport)
}

func newRaftFileSnapshotStore(base string) (*raft.FileSnapshotStore, error) {
	return raft.NewFileSnapshotStore(base, 2, os.Stderr)
}

func newRaftTCPTransport(address, bindAddr string) (*raft.NetworkTransport, error) {
	advertise, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}

	return raft.NewTCPTransport(bindAddr, advertise, 3, 5*time.Second, os.Stderr)
}
