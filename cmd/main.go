package main

import (
	// "fmt"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	// "time"

	"github.com/Kevin27954/liveness-sim-test/assert"
	"github.com/Kevin27954/liveness-sim-test/cmd/node"
	"github.com/Kevin27954/liveness-sim-test/db"
)

func createNode(num int, connList string) node.Node {
	name := "node" + strconv.Itoa(num)
	sql_db := db.Init(name)

	return node.Node{Conn: nil, Status: 0, Hub: node.Hub{ConnStr: connList, DB: sql_db, Name: name}}
}

func main() {
	log.SetOutput(os.Stdout)
	rand.New(rand.NewSource(69))

	args := os.Args[1:]

	addr := args[0]

	var connList string
	if len(args) > 1 {
		connList = args[1]
	}

	addrAsInt, err := strconv.Atoi(addr)
	assert.NoError(err, "Unable to covert to int")
	serverNode := createNode(addrAsInt%8000, connList)

	go func() {
		time.Sleep(10 * time.Second)
		serverNode.Hub.Run(addrAsInt)
	}()

	http.HandleFunc("/ws", serverNode.Start)
	http.HandleFunc("/internal", serverNode.Internal)
	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		messages, err := serverNode.GetMessages()
		if err != nil {
			http.Error(w, "Error getting messages", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, messages)
	})
	http.HandleFunc("/get/logs", func(w http.ResponseWriter, r *http.Request) {
		messages, err := serverNode.GetLogs()
		if err != nil {
			http.Error(w, "Error getting messages", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, messages)
	})

	log.Printf("Starting %s on \"localhost:%s\"", serverNode.Hub.Name, addr)

	err = http.ListenAndServe("localhost:"+addr, nil)
	assert.NoError(err, "Unable to start server")
}
