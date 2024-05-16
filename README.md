# Go Raft Message Queue

This project is an implementation of a distributed message queue using the Raft consensus algorithm in Go. It is heavily inspired by the following projects / talks / resources:

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

## Command Line Arguments

- The `-leader` flag is used to specify that the node is the leader node.
- The `-id` flag is used to specify a unique string identifying the server.
- The `-raddr` flag is used to specify the Raft address of the server.
- The `-dir` flag is used to specify the directory where the server's data will be stored.
- The `-paddr` flag is used to specify the host and port of the leader node to join the cluster.
- The `-haddr` flag is used to specify the host and port of the server for the client to interact with.

- The host and port of the leader node can be specified using the `-haddr` flag. The host and port of the node can be specified using the `-paddr` flag. The Raft address of the node can be specified using the `-raddr` flag. The directory where the node's data will be stored can be specified using the `-dir` flag.

### Running the Leader Node

```sh
./queue -leader -id=node01 -raddr=localhost:3001 -dir=./tmp/node01 -haddr=localhost:3000
```

### Running the Follower Nodes

```sh
./queue -id=node02 -raddr=localhost:3003 -dir=./tmp/node02 -paddr=localhost:3000 -haddr=localhost:3002
./queue -id=node03 -raddr=localhost:3005 -dir=./tmp/node03 -paddr=localhost:3000 -haddr=localhost:3004
```

## API Endpoints

- `POST /send` - Push a message to the queue
- `GET /recieve` - Pop a message from the queue
- `GET /stats` - Get the status of the raft node
- `POST /join` - Join a node to the cluster

### Pushing a message to the queue

The model of the message is as follows:

```go
type Comment struct {
	Timestamp *time.Time `json:"timestamp,omitempty"`
	Author    string     `json:"author,omitempty"`
	Content   string     `json:"content,omitempty"`
}
```

This can be pushed to the queue using the following command:

```sh
curl -X POST -d '{"timestamp": "2022-05-15T17:19:09Z", "author": "John Doe", "content": "This is a sample comment."}' http://localhost:3000/send
```

### Popping a message from the queue

This can be done using the following command:

```sh
curl -X GET http://localhost:3000/recieve
```

Will return the following response:

```json
{
  "Data": {
    "timestamp": "2022-05-15T17:19:09Z",
    "author": "John Doe",
    "content": "This is a sample comment."
  }
}
```

If the queue is empty, the following response will be returned:

```json
{
  "Data": {}
}
```

### Getting the stats of the raft node

Can be used for debugging purposes. Will return the following [Raft.Stats](https://pkg.go.dev/github.com/hashicorp/raft#Raft.Stats) map.

```sh
curl -X GET http://localhost:3000/stats
```
