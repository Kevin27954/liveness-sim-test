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
	"github.com/gorilla/websocket"
)

func main() {
	log.SetOutput(os.Stdout)
	rand.New(rand.NewSource(69))

	args := os.Args[1:]
	addr := args[0]

	var upgrader = websocket.Upgrader{}

	var portList string
	if len(args) > 1 {
		portList = args[1]
	}

	size := 3

	addrAsInt, err := strconv.Atoi(addr)
	assert.NoError(err, "Unable to covert to int")
	id := addrAsInt % 8000
	name := "node" + strconv.Itoa(id)

	// Init all things
	sqlDb := db.Init(name)
	raftState := r.Init(name, id, sqlDb, portList, size)
	serverNode := node.Node{Conn: nil, Status: 0, Hub: node.Hub{}}

	srv := &http.Server{Addr: ":" + addr}

	http.HandleFunc("/ws", serverNode.Start)
	http.HandleFunc("/internal/{id}", func(w http.ResponseWriter, r *http.Request) {
		// Checks for internal node key in prod if there is a prod.
		upgrader.CheckOrigin = func(r *http.Request) bool {
			// In production, make this your origin (URL to your server)
			return true
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		assert.NoError(err, "Unable to upgrade internal nodes socket conn")

		urlId := r.PathValue("id")
		id, err := strconv.Atoi(urlId)
		if err != nil {
			log.Fatal("Unable to get nodeID")
		}

		raftState.AddConn(conn, id)
	})

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

	http.HandleFunc("/quit", func(w http.ResponseWriter, r *http.Request) {
		serverNode.Close()
		ctx, ctxClose := context.WithTimeout(context.Background(), 10*time.Second)
		err := srv.Shutdown(ctx)
		if err != nil {
			ctxClose()
			log.Fatal("Unable to close server: ", err)
		}

		defer ctxClose()
	})

	log.Printf("Starting %s on \"localhost:%s\"", serverNode.Hub.Name, addr)

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// unexpected error. port in use?
		log.Fatalf("ListenAndServe(): %v", err)
	}

	log.Println("Server closed")
}
