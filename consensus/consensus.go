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

// Consensus is the consensus module
type Consensus struct {
	Node *raft.Raft
}

// Config is the configuration for the consensus module
type Config struct {
	// IsLeader is a flag that indicates if the server is the leader
	IsLeader bool

	// ServerID is the unique identifier for the server
	ServerID string

	// BaseDirectory is the directory where the raft data will be stored
	BaseDirectory string

	// Address is the address at which the server will be listening
	Address string
}

// NewConsensusConfig creates a new consensus config
func NewConsensusConfig() *Config {
	return &Config{}
}

// NewConsensus creates a new instance of the consensus module
func NewConsensus(fsm raft.FSM, conf *Config) (*Consensus, error) {
	// Create the raft configuration
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(conf.ServerID)

	// Set the snapshot interval to 1 second and the snapshot threshold to 1
	// so a snapshot is taken after every log entry for testing
	// config.SnapshotInterval = 1 * time.Second
	// config.SnapshotThreshold = 1

	// Create the raft store
	store, err := raftboltdb.NewBoltStore(filepath.Join(conf.BaseDirectory, "raft.db"))
	if err != nil {
		return nil, err
	}
	logStore, stableStore := store, store

	// Create the snapshot store
	snapshotStore, err := raft.NewFileSnapshotStore(conf.BaseDirectory, 2, os.Stderr)
	if err != nil {
		return nil, err
	}

	// Create the transport
	address, err := net.ResolveTCPAddr("tcp", conf.Address)
	if err != nil {
		return nil, err
	}
	transport, err := raft.NewTCPTransport(conf.Address, address, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, err
	}

	node, err := raft.NewRaft(config, fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		return nil, err
	}

	if conf.IsLeader {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raft.ServerID(conf.ServerID),
					Address: raft.ServerAddress(conf.Address),
				},
			},
		}
		node.BootstrapCluster(configuration)
	}

	return &Consensus{Node: node}, nil
}

// Join joins the raft cluster
func (c *Consensus) Join(nodeID, address string) error {
	configFuture := c.Node.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		return err
	}

	for _, server := range configFuture.Configuration().Servers {
		// The node is already part of the cluster
		if server.ID == raft.ServerID(nodeID) && server.Address == raft.ServerAddress(address) {
			return nil
		}

		// There's a node with the same ID or address, remove it first
		if server.ID == raft.ServerID(nodeID) || server.Address == raft.ServerAddress(address) {
			future := c.Node.RemoveServer(server.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("error removing existing node %s at %s: %s", nodeID, address, err)
			}
		}
	}

	// Add the new node as a voter
	f := c.Node.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(address), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}

	return nil
}
