package client

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	WRITE_WAIT = 10
)

type Client struct {
	conn *websocket.Conn
}

func Init() Client {
	c := Client{}

	return c
}

func (c *Client) Connect(connUrl string) {
	conn, _, err := websocket.DefaultDialer.Dial(connUrl, nil)
	if err != nil {
		log.Fatal("WTF: ", connUrl, "err:", err)
	}

	c.conn = conn
}

func (c *Client) WriteMsg(msg string) {
	if c.conn == nil {
		log.Fatal("Connection is nil")
	}

	c.conn.SetWriteDeadline(time.Now().Add(WRITE_WAIT * time.Second))
	err := c.conn.WriteMessage(websocket.TextMessage, []byte(msg))
	if err != nil {
		log.Println("Unable to write message: ", msg, "\n", "err: ", err)
	}
}
