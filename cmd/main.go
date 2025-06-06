package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Kevin27954/liveness-sim-test/assert"
	"github.com/Kevin27954/liveness-sim-test/cmd/node"
	"github.com/Kevin27954/liveness-sim-test/db"
	r "github.com/Kevin27954/liveness-sim-test/pkg/raft"
)

func main() {
	log.SetOutput(os.Stdout)
	rand.New(rand.NewSource(69))

	args := os.Args[1:]

	var addr string
	var portList string
	numNodes := 3
	for i := range args {

		switch args[i] {
		case "-n":
			num, err := strconv.Atoi(args[i+1])
			if err != nil {
				log.Fatal("Exepected a number")
			}
			numNodes = num

			i += 1
		case "-a":
			addr = args[i+1]
			i += 1
		case "-p":
			portList = args[i+1]
			i += 1

		}
	}

	addrAsInt, err := strconv.Atoi(addr)
	assert.NoError(err, "Unable to covert to int")
	id := addrAsInt % 8000
	name := "node" + strconv.Itoa(id)

	// Init all things
	sqlDb := db.Init(name)
	raftState := r.Init(name, id, sqlDb, portList, numNodes)
	serverNode := node.Node{Conn: nil, Status: 0, Raft: raftState}

	srv := &http.Server{Addr: ":" + addr}

	http.HandleFunc("/ws", serverNode.Start)
	http.HandleFunc("/internal/{id}", serverNode.Internal)
	http.HandleFunc("/get/logs", serverNode.GetLogs)
	http.HandleFunc("/gets", serverNode.GetMessages)
	http.HandleFunc("/raft", raftState.GetState)
	http.HandleFunc("/leader", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, raftState.IsLeader())
	})

	http.HandleFunc("/quit", func(w http.ResponseWriter, r *http.Request) {
		serverNode.Close()
		ctx, ctxClose := context.WithTimeout(context.Background(), 10*time.Second)
		err := srv.Shutdown(ctx)
		if err != nil {
			ctxClose()
			log.Fatal("Unable to close server: ", err)
		}

		defer ctxClose()

		fmt.Fprintf(w, "Closing")
	})

	log.Printf("Starting on \"http://localhost:%s\"", addr)

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe(): %v", err)
	}

	log.Println("Server closed")
}
