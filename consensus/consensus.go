package consensus

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
)

type Config struct {
	ServerID      string
	BaseDirectory string
	Address       string
}

// NewConsensusConfig creates a new consensus config
func NewConsensusConfig() *Config {
	return &Config{}
}

// NewRaft creates a new raft instance with the given fsm and config
func NewRaft(fsm raft.FSM, c *Config) (*raft.Raft, error) {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(c.ServerID)

	store, err := raftboltdb.NewBoltStore(filepath.Join(c.BaseDirectory, "raft.db"))
	if err != nil {
		return nil, err
	}
	logStore, stableStore := store, store

	snapshotStore, err := raft.NewFileSnapshotStore(c.BaseDirectory, 2, os.Stderr)
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

// Join joins the raft cluster
func Join(node *raft.Raft, nodeID, address string) error {
	configFuture := node.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		return err
	}

	for _, srv := range configFuture.Configuration().Servers {
		if srv.ID == raft.ServerID(nodeID) || srv.Address == raft.ServerAddress(address) {
			if srv.Address == raft.ServerAddress(address) && srv.ID == raft.ServerID(nodeID) {
				return nil
			}

			future := node.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("error removing existing node %s at %s: %s", nodeID, address, err)
			}
		}
	}

	f := node.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(address), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}

	return nil
}
