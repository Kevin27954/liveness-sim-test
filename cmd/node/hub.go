package node

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/Kevin27954/liveness-sim-test/assert"
	"github.com/gorilla/websocket"
)

const (
	WRITEWAIT = 10 * time.Second
	PONGTIME  = 60 * time.Second
	PINGTIME  = (PONGTIME * 9) / 10
)

type Hub struct {
	ConnStr  string
	connList []*websocket.Conn
	Lock     sync.Mutex
}

func (h *Hub) ConnectConns() {
	hub := Hub{connList: []*websocket.Conn{}}

	assert.Assert(len(h.ConnStr) > 0, len(h.ConnStr), 1, "Expected length greater than 0")

	log.Println("Connecting to other nodes")

	conns := strings.Split(h.ConnStr, ",")
	for _, connStr := range conns {
		conn, _, err := websocket.DefaultDialer.Dial(connStr, nil)
		assert.NoError(err, "Unable to connect to other nodes")

		log.Printf("Connected to %s", connStr)

		hub.connList = append(hub.connList, conn)

		// Starts Pinging
		// h.Ping(conn)
	}
}

func (h *Hub) Ping(conn *websocket.Conn) {
	pingTicker := time.NewTicker(PINGTIME)

	conn.SetPongHandler(func(appData string) error {
		log.Println("PONG")
		return nil
	})

	defer pingTicker.Stop()

	// Reads PONG Message
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				log.Println("Pong: ", err)
				break
			}
		}
	}()

	for {
		select {
		case <-pingTicker.C:
			conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))
			err := conn.WriteMessage(websocket.PingMessage, []byte("PING"))
			if err != nil {
				log.Println("Ping test", err)
				return
			}
			break
		}
	}
}

func (h *Hub) Pong(conn *websocket.Conn) {
	conn.SetReadDeadline(time.Now().Add(PONGTIME))
	conn.SetPingHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(PONGTIME))
		conn.WriteMessage(websocket.PongMessage, []byte("PONG"))
		log.Println("PING")
		return nil
	})

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println("Pong: ", err)
			break
		}
	}
}

func (h *Hub) Run(addr int) {
	if len(h.ConnStr) > 0 {
		h.ConnectConns()
	}

	interval := (addr % 8000) * 5

	newTicker := h.newTimerEveryMin(interval)
	defer newTicker.Stop()

	select {

	case <-newTicker.C:
		// Check if is ledaer
		// If is leader send ping
		// else init election
		log.Println(time.Now())

	}

	// Set up time.
	h.InitiateElection()
}

func (h *Hub) SendMessage(message string) {
}

func (h *Hub) RecieveMessage() {
}

func (h *Hub) InitiateElection() {
}

func (h *Hub) AddConn(conn *websocket.Conn) {
	h.Lock.Lock()
	h.connList = append(h.connList, conn)
	h.Lock.Unlock()
}

func (h *Hub) newTimerEveryMin(sec int) *time.Ticker {
	currTime := time.Now()
	timeLeft := time.Duration(((60-currTime.Second())+sec)%60) * time.Second

	time.Sleep(timeLeft)

	return time.NewTicker(time.Minute)
}
