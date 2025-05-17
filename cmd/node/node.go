package node

import (
	// "fmt"
	"fmt"
	"log"
	"net/http"

	"github.com/Kevin27954/liveness-sim-test/assert"
	"github.com/Kevin27954/liveness-sim-test/pkg"
	"github.com/Kevin27954/liveness-sim-test/pkg/raft"
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
	Raft   *raft.Raft
}

func (n *Node) Start(w http.ResponseWriter, r *http.Request) {
	defer func() {
		log.Println("Closing Connection with User")
		n.Conn.Close()
		n.Conn = nil
	}()

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
	// Ignore for now, this used to be for connecting internal nodes
}

func (n *Node) GetMessages() (string, error) {
	messages, err := n.Raft.Db.GetMessages()
	if err != nil {
		return "", err
	}

	var allMessages string

	for _, message := range messages {
		allMessages += message.Msg + "\n"
	}

	return allMessages, nil
}

func (n *Node) GetLogs() (string, error) {
	logs, err := n.Raft.Db.GetLogs()
	if err != nil {
		return "", err
	}

	var allLogs string
	for _, log := range logs {
		allLogs += fmt.Sprintf("%s - %d - %s - %s\n", log.Time, log.Id, log.Operation, log.Data)
	}

	return allLogs, nil
}

// There will no writebacks for now. I only want to read.
// After reading, I want to send that message to other nodes.
// I also want to store it in SQL Lite Db.
func (n *Node) recieveMessage() {
	n.Conn.SetReadLimit(READLIMIT)
	log.Println("Listening to Messages from User")

	for {
		_, message, err := n.Conn.ReadMessage()
		if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) {
			if err != nil {
				log.Println(err)
			}
			break
		} else {
			assert.NoError(err, "Unable to read message")
		}

		assert.Assert(len(message) <= 512, len(message), 512, "greater than 512 bytes")

		// n.Hub.StoreMessage(string(message))
		// n.Hub.DB.AddMessage(string(message))

		n.Raft.SendNewOp(pkg.NEW_MSG_ADD, string(message))
		log.Printf("Recieved: %s", message)
	}

}

func (n *Node) Close() {
	// n.Hub.Close()
	if n.Conn != nil {
		err := n.Conn.Close()
		if err != nil {
			log.Fatal("Unable to close conn, node")
		}
	}

	log.Println("Server Connections Closed")
}
