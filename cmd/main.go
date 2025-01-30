package main

import (
	// "fmt"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/Kevin27954/liveness-sim-test/assert"
	"github.com/Kevin27954/liveness-sim-test/cmd/node"
	"github.com/Kevin27954/liveness-sim-test/db"
)

func createNode(num int) node.Node {
	name := "node" + strconv.Itoa(num)
	db := db.Init(name)

	return node.Node{Name: name, Conn: nil, Db: db}
}

func main() {
	args := os.Args[1:]

	addr := args[0]
	nodeNum, err := strconv.Atoi(args[1])
	assert.NoError(err, "Args wasn't a number")

	serverNode := createNode(nodeNum)

	http.HandleFunc("/ws", serverNode.Start)
	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		messages, err := serverNode.GetMessages()
		if err != nil {
			http.Error(w, "Error getting messages", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, messages)
	})

	log.Printf("Starting %s on \"localhost:%s\"", serverNode.Name, addr)
	err = http.ListenAndServe("localhost:"+addr, nil)
	assert.NoError(err, "Unable to start server")
}
