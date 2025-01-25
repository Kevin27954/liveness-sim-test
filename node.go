package main

import (
	// "fmt"
	"log"
	"net/http"
	"strings"
	// "time"

	"github.com/Kevin27954/liveness-sim-test/assert"
	"github.com/gorilla/websocket"
)

/*
Should have it's own storage/path to SQL Lite.
*/

const (
	READLIMIT = 512
	PONGTIME  = 30
	PINGTIME  = (PONGTIME * 9) / 10
)

var upgrader = websocket.Upgrader{}

type Node struct {
	name string
	conn *websocket.Conn
	db   DB
}

func (n *Node) Start(w http.ResponseWriter, r *http.Request) {
	if n.conn != nil {
		// There should only be 1 non node connection at once to the node.
		return
	}

	upgrader.CheckOrigin = func(r *http.Request) bool {
		// In production, make this your origin (URL to your server)
		return true
	}

	// Should start it's own websocket server.
	conn, err := upgrader.Upgrade(w, r, nil)
	assert.NoError(err, "Unable to upgrade TCP")

	n.conn = conn

	go n.recieveMessage()
}

// There will no writebacks for now. I only want to read.
// After reading, I want to send that message to other nodes.
// I also want to store it in SQL Lite DB.
func (n *Node) recieveMessage() {
	// Close the connection when the function finishes
	defer n.conn.Close()

	n.conn.SetReadLimit(READLIMIT)
	// n.conn.SetReadDeadline(time.Now().Add(PONGTIME))
	// Look into Ping Pong?
	// n.conn.SetPingHandler

	for {
		_, message, err := n.conn.ReadMessage()
		if websocket.IsCloseError(err, websocket.CloseGoingAway) {
			break
		} else {
			assert.NoError(err, "Unable to read message")
		}

		assert.Assert(len(message) <= 512, len(message), 512, "greater than 512 bytes")

		n.db.AddMessage(string(message))

		log.Printf("Recieved: %s", message)
	}

	n.conn = nil
}

func (n *Node) GetMessages() (string, error) {
	messages, err := n.db.GetMessages()

	return strings.Join(messages, "\n"), err
}
