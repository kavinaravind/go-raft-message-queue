package consensus

import (
	"fmt"
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

func waitForNodeToBeLeader(node *Consensus, duration time.Duration) error {
	timeout := time.After(duration)
	tick := time.Tick(100 * time.Millisecond)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timed out waiting for node to be leader")
		case <-tick:
			if node.Node.State() == raft.Leader {
				return nil
			}
		}
	}
}

func TestNewConsensus(t *testing.T) {
	fsm := &MockFSM{}

	tmpDir1, err := os.MkdirTemp("", "node1")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir1)

	conf := &Config{
		IsLeader:      true,
		ServerID:      "test",
		BaseDirectory: tmpDir1,
		Address:       "localhost:8000",
	}

	_, err = NewConsensus(fsm, conf)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestJoin(t *testing.T) {
	fsm := &MockFSM{}

	tmpDir1, err := os.MkdirTemp("", "node1")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir1)

	conf1 := &Config{
		IsLeader:      true,
		ServerID:      "node1",
		BaseDirectory: tmpDir1,
		Address:       "localhost:8000",
	}

	node1, err := NewConsensus(fsm, conf1)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	tmpDir2, err := os.MkdirTemp("", "node2")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir2)

	conf2 := &Config{
		IsLeader:      false,
		ServerID:      "node2",
		BaseDirectory: tmpDir2,
		Address:       "localhost:8001",
	}

	_, err = NewConsensus(fsm, conf2)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if err := waitForNodeToBeLeader(node1, 5*time.Second); err != nil {
		t.Fatalf("expected node1 to be leader, got: %v", err)
	}

	err = node1.Join("node2", "localhost:8001")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}
