package consensus

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/raft"
)

// MockFSM is a mock finite state machine for testing
type MockFSM struct{}

func (m *MockFSM) Apply(l *raft.Log) interface{} {
	return nil
}

func (m *MockFSM) Snapshot() (raft.FSMSnapshot, error) {
	return nil, nil
}

func (m *MockFSM) Restore(rc io.ReadCloser) error {
	return nil
}

func TestNodes(t *testing.T) {
	fsm := &MockFSM{}

	tmpDir1, err := os.MkdirTemp("", "node1")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tmpDir1)
	})

	conf1 := &Config{
		IsLeader:      true,
		ServerID:      "node1",
		BaseDirectory: tmpDir1,
		Address:       "localhost:8000",
	}

	node1, err := NewConsensus(fsm, conf1)
	t.Cleanup(func() {
		future := node1.Node.Shutdown()
		if err := future.Error(); err != nil {
			t.Logf("Failed to shutdown node1: %v", err)
		}
	})
	t.Run("TestNewConsensusLeader", func(t *testing.T) {
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("TestNode1BecomesLeader", func(t *testing.T) {
		if err := node1.WaitForNodeToBeLeader(5 * time.Second); err != nil {
			t.Fatalf("expected node1 to be leader, got: %v", err)
		}
	})

	tmpDir2, err := os.MkdirTemp("", "node2")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tmpDir1)
	})

	conf2 := &Config{
		IsLeader:      false,
		ServerID:      "node2",
		BaseDirectory: tmpDir2,
		Address:       "localhost:8001",
	}

	node2, err := NewConsensus(fsm, conf2)
	t.Cleanup(func() {
		future := node2.Node.Shutdown()
		if err := future.Error(); err != nil {
			t.Logf("Failed to shutdown node2: %v", err)
		}
	})
	t.Run("TestNewConsensusFollower", func(t *testing.T) {
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("TestNode2IsFollower", func(t *testing.T) {
		stats := node2.Node.Stats()
		if stats["state"] != "Follower" {
			t.Errorf("expected node2 to be a follower, got: %v", stats["state"])
		}
	})

	t.Run("TestJoin", func(t *testing.T) {
		err = node1.Join("node2", "localhost:8001")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})
}
