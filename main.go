package main

import (
	"github.com/hashicorp/raft"
	raftmdb "github.com/hashicorp/raft-mdb"
)

func main() {

	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID("node1")

	store, err := raftmdb.NewMDBStore("./")
	if err != nil {
		return err
	}

	logStore, stableStore := store, store

}
