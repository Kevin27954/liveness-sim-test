package node

import (
	// "fmt"
	"log"
	"net/http"
	"strings"

	// "time"

	"github.com/Kevin27954/liveness-sim-test/Db"
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
	Name string
	Conn *websocket.Conn
	Db   db.DB
}

func (n *Node) Start(w http.ResponseWriter, r *http.Request) {
	if n.Conn != nil {
		// There should only be 1 non node Connection at once to the node.
		return
	}

	upgrader.CheckOrigin = func(r *http.Request) bool {
		// In production, make this your origin (URL to your server)
		return true
	}

	// Should start it's own websocket server.
	conn, err := upgrader.Upgrade(w, r, nil)
	assert.NoError(err, "Unable to upgrade TCP")

	n.Conn = conn

	go n.recieveMessage()
}

// There will no writebacks for now. I only want to read.
// After reading, I want to send that message to other nodes.
// I also want to store it in SQL Lite Db.
func (n *Node) recieveMessage() {
	// Close the Connection when the function finishes
	defer n.Conn.Close()

	n.Conn.SetReadLimit(READLIMIT)
	// n.Conn.SetReadDeadline(time.Now().Add(PONGTIME))
	// Look into Ping Pong?
	// n.Conn.SetPingHandler

	for {
		_, message, err := n.Conn.ReadMessage()
		if websocket.IsCloseError(err, websocket.CloseGoingAway) {
			break
		} else {
			assert.NoError(err, "Unable to read message")
		}

		assert.Assert(len(message) <= 512, len(message), 512, "greater than 512 bytes")

		n.Db.AddMessage(string(message))

		log.Printf("Recieved: %s", message)
	}

	n.Conn = nil
}

func (n *Node) GetMessages() (string, error) {
	messages, err := n.Db.GetMessages()

	return strings.Join(messages, "\n"), err
}
