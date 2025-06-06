package client

import (
	"errors"
	"fmt"
	"log"
	"syscall"
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

	reconnCounter int
}

func Init(rand rand.Randomizer) Client {
	c := Client{randomizer: rand, reconnCounter: 0}

	return c
}

func (c *Client) Start(numMsg int) {
	log.Println("Client: Starting Sending Messsages")
	cpNumMsg := numMsg

	for numMsg > 0 {
		// c.WriteMsg(fmt.Sprint("Msg #", cpNumMsg-numMsg, " - From Client to ", c.connUrl))
		c.WriteMsg(fmt.Sprint(cpNumMsg - numMsg))
		numMsg -= 1
		time.Sleep(time.Duration(c.randomizer.GetIntN(1000)) * time.Millisecond)

		if c.reconnCounter > 10 {
			log.Println("Stopping Client")
			return
		}
	}

	log.Println("Finished Sending Messages to ", c.connUrl)
}

func (c *Client) Connect(connUrl string) *websocket.Conn {
	c.connUrl = connUrl

	conn, _, err := websocket.DefaultDialer.Dial(connUrl, nil)
	if err != nil {
		log.Println("Unable to connect to :", connUrl, " err:", err)
		return nil
	}

	c.conn = conn
	return conn
}

func (c *Client) WriteMsg(msg string) {

	c.conn.SetWriteDeadline(time.Now().Add(WRITE_WAIT * time.Second))
	err := c.conn.WriteMessage(websocket.TextMessage, []byte(msg))
	if err != nil {
		if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) || errors.Is(err, syscall.EPIPE) {
			// Attempt reconnect
			for {
				c.reconnCounter += 1
				time.Sleep(5 * time.Second)
				log.Println("Attemping to Reconnect Client...")
				if c.Connect(c.connUrl) != nil {
					break
				}
			}
		} else {
			// log.Println("Unable to write message: ", msg, "\n", "err: ", err)
			log.Printf("\033[%dm%s\033[0m", 200+31, err)
		}
	}
}
