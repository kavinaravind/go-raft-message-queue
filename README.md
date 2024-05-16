# go-raft-message-queue

```sh
go run main.go -leader -id=node01 -raddr=localhost:3001 -dir=./tmp/node01 -haddr=localhost:3000
go run main.go -id=node02 -raddr=localhost:3003 -dir=./tmp/node02 -paddr=localhost:3000 -haddr=localhost:3002
go run main.go -id=node03 -raddr=localhost:3005 -dir=./tmp/node03 -paddr=localhost:3000 -haddr=localhost:3004
```
