package main

import (
	// "fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

/*
Should have it's own storage/path to SQL Lite.
*/

var upgrader = websocket.Upgrader{}

type Node struct {
	name string
	conn *websocket.Conn
}

// There will no writebacks for now. I only want to read.
// After reading, I want to send that message to other nodes.
// I also want to store it in SQL Lite DB.
func (n *Node) recieveMessage() {
	// Close the connection
	defer n.conn.Close()

	for {
		_, message, err := n.conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		log.Printf("Recieved: %s", message)
	}

	n.conn = nil
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
	if err != nil {
		log.Println(err)
		return
	}

	n.conn = conn

	go n.recieveMessage()
}
