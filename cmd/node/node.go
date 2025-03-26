package node

import (
	// "fmt"
	"log"
	"net/http"

	"github.com/Kevin27954/liveness-sim-test/assert"
	"github.com/gorilla/websocket"
)

/*
Should have it's own storage/path to SQL Lite.
*/

const (
	READLIMIT = 512
)

var upgrader = websocket.Upgrader{}

type Node struct {
	Conn   *websocket.Conn
	Status int // 1 = Leader, 0 = member
	Hub    Hub
}

func (n *Node) Start(w http.ResponseWriter, r *http.Request) {
	defer n.Conn.Close()

	upgrader.CheckOrigin = func(r *http.Request) bool {
		// In production, make this your origin (URL to your server)
		return true
	}

	// Should start it's own websocket server.
	conn, err := upgrader.Upgrade(w, r, nil)
	assert.NoError(err, "Unable to upgrade TCP")

	n.Conn = conn

	n.recieveMessage()
}

func (n *Node) Internal(w http.ResponseWriter, r *http.Request) {
	// Checks for internal node key in prod if there is a prod.

	upgrader.CheckOrigin = func(r *http.Request) bool {
		// In production, make this your origin (URL to your server)
		return true
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	assert.NoError(err, "Unable to upgrade internal nodes socket conn")

	n.Hub.AddConn(conn)
	// n.Hub.Pong(conn)
}

// There will no writebacks for now. I only want to read.
// After reading, I want to send that message to other nodes.
// I also want to store it in SQL Lite Db.
func (n *Node) recieveMessage() {
	// Close the Connection when the function finishes
	defer n.Conn.Close()

	n.Conn.SetReadLimit(READLIMIT)

	for {
		_, message, err := n.Conn.ReadMessage()
		if websocket.IsCloseError(err, websocket.CloseGoingAway) {
			break
		} else {
			assert.NoError(err, "Unable to read message")
		}

		assert.Assert(len(message) <= 512, len(message), 512, "greater than 512 bytes")

		n.Hub.StoreMessage()
		// n.Hub.DB.AddMessage(string(message))

		log.Printf("Recieved: %s", message)
	}

	n.Conn = nil
}

func (n *Node) GetMessages() (string, error) {
	messages, err := n.Hub.DB.GetMessages()
	if err != nil {
		return "", err
	}

	var allMessages string

	for _, message := range messages {
		allMessages += message.Msg + "\n"
	}

	return allMessages, nil
}
