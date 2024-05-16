# Go Raft Message Queue

This project is an implementation of a distributed message queue using the Raft consensus algorithm in Go. It is inspired by the following projects:

- [GopherCon 2023: Philip O'Toole - Build Your Own Distributed System Using Go](https://youtu.be/8XbxQ1Epi5w?si=pwj8mIM4gzpvvyTZ)
- [Raft Consensus Algorithm](https://raft.github.io/)
- [Golang Implementation of the Raft Consensus Protocol](https://github.com/hashicorp/raft)
- [Raft backend implementation using BoltDB](https://github.com/hashicorp/raft-boltdb)
- [A reference use of Hashicorp's Raft implementation](https://github.com/otoolep/hraftd)

## Getting Started

These instructions will get you up and running on your local machine for development and testing purposes.

### Prerequisites

- [Go](https://go.dev/) (currently running: `go version go1.22.3 darwin/arm64`)
- [Make](https://www.gnu.org/software/make/) (currently running: `GNU Make 3.81`)

### Building

```sh
git clone https://github.com/yourusername/go-raft-message-queue.git
cd go-raft-message-queue
make build
```

### Running the Leader Node

```sh
./queue -leader -id=node01 -raddr=localhost:3001 -dir=./tmp/node01 -haddr=localhost:3000
```

### Running the Follower Nodes

```sh
./queue -id=node02 -raddr=localhost:3003 -dir=./tmp/node02 -paddr=localhost:3000 -haddr=localhost:3002
./queue -id=node03 -raddr=localhost:3005 -dir=./tmp/node03 -paddr=localhost:3000 -haddr=localhost:3004
```
