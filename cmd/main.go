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
	addr := args[0]

	var portList string
	if len(args) > 1 {
		portList = args[1]
	}

	size := 3 // Change this to either recieve from args or dynamically figure it out

	addrAsInt, err := strconv.Atoi(addr)
	assert.NoError(err, "Unable to covert to int")
	id := addrAsInt % 8000
	name := "node" + strconv.Itoa(id)

	// Init all things
	sqlDb := db.Init(name)
	raftState := r.Init(name, id, sqlDb, portList, size)
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
