package client

import (
	"fmt"
	"log"
	"time"

	rand "github.com/Kevin27954/liveness-sim-test/test/randomizer"
	"github.com/gorilla/websocket"
)

const (
	WRITE_WAIT = 10
)

type Client struct {
	connUrl    string
	conn       *websocket.Conn
	randomizer rand.Randomizer
}

func Init(rand rand.Randomizer) Client {
	c := Client{randomizer: rand}

	return c
}

func (c *Client) Start(numMsg int) {
	log.Println("Client: Starting Sending Messsages")
	cpNumMsg := numMsg

	for numMsg > 0 {
		c.WriteMsg(fmt.Sprint("Msg #", cpNumMsg-numMsg, " - From Client to ", c.connUrl))
		numMsg -= 1
		time.Sleep(time.Duration(c.randomizer.GetIntN(1000)) * time.Millisecond)
	}

	log.Println("Finished Sending Messages to ", c.connUrl)
}

func (c *Client) Connect(connUrl string) {
	c.connUrl = connUrl

	conn, _, err := websocket.DefaultDialer.Dial(connUrl, nil)
	if err != nil {
		log.Println("Unable to connect to :", connUrl, " err:", err)
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
		if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) {
			// Attempt reconnect
			for {
				time.Sleep(5 * time.Second)
				log.Println("Attemping to Reconnect Client...")
				c.Connect(c.connUrl)
			}
		} else {
			// log.Println("Unable to write message: ", msg, "\n", "err: ", err)
		}
	}
}
