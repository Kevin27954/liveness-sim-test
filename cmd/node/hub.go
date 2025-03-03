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
	WRITEWAIT = 30 * time.Second
	PONGTIME  = 70 * time.Second
	PINGTIME  = (PONGTIME * 9) / 10
	ONE_MIN   = 60
)

type Hub struct {
	Name      string
	ConnStr   string
	connList  []*websocket.Conn
	Lock      sync.Mutex
	IsLeader  bool
	hasLeader bool
}

func (h *Hub) ConnectConns() {
	assert.Assert(len(h.ConnStr) > 0, len(h.ConnStr), 1, "Expected length greater than 0")

	log.Println(h.Name, "Connecting to other nodes")

	conns := strings.Split(h.ConnStr, ",")
	for _, connStr := range conns {
		conn, _, err := websocket.DefaultDialer.Dial(connStr, nil)
		assert.NoError(err, "Unable to connect to other nodes")

		log.Printf("Connected to %s", connStr)

		h.connList = append(h.connList, conn)
	}
}

func (h *Hub) Ping() {

	for _, conn := range h.connList {
		go func(conn *websocket.Conn) {

			conn.SetPongHandler(func(appData string) error {
				log.Println(h.Name, "PONG")
				return nil
			})

			conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))
			err := conn.WriteMessage(websocket.PingMessage, []byte("PING"))
			if err != nil {
				log.Println(h.Name, "Ping test", err)
			}

		}(conn)

	}

}

func (h *Hub) Run(addr int) {
	if len(h.ConnStr) > 0 {
		h.ConnectConns()
	}

	interval := (addr % 8000) * 5

	newTicker := h.newTimerEveryMin(interval)
	defer newTicker.Stop()

	consensus := make(chan int)

	go func() {
		for {
			select {
			case <-newTicker.C:
				// Send a rquest to all conn, if conn is up and recieve a single no then we sa it is reject
				// if all conn is yes and no err, then we proceed with request.

				//  log.Println(h.Name, "I am leader: ", h.IsLeader)

				if h.IsLeader {
					// Leader Ping
					h.Ping()
				} else if h.hasLeader {
				} else {
					h.InitiateElection()
				}

			}
		}
	}()

	h.RecieveMessage(consensus)

	for {
		// Here to keep the functions running.
	}
}

func (h *Hub) RecieveMessage(consensus chan int) {

	for _, conn := range h.connList {
		go func(conn *websocket.Conn) {

			conn.SetPingHandler(func(appData string) error {
				conn.SetReadDeadline(time.Now().Add(PONGTIME))
				conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))
				err := conn.WriteMessage(websocket.PongMessage, []byte("PONG"))
				if err != nil {
					log.Println(h.Name, "Ping:", err)
				}

				log.Println(h.Name, "PING")
				return nil
			})

			for {

				// The idea is that we don't know WHEN we might receive a message. So we just want to wait and be on
				// the lookout for any potential messages that might come.
				// Removes the deadline
				conn.SetReadDeadline(time.Time{})

				_, msg, err := conn.ReadMessage()
				if err != nil {
					log.Println(h.Name, "Err: ", err)
				}

				if string(msg) == "0" {

					conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))
					if h.IsLeader {
						conn.WriteMessage(websocket.TextMessage, []byte("2"))
					} else {
						conn.WriteMessage(websocket.TextMessage, []byte("1"))
						h.Lock.Lock()
						h.hasLeader = true
						h.Lock.Unlock()
					}

				} else if string(msg) == "1" {
					if !h.hasLeader {
						h.Lock.Lock()
						h.IsLeader = true
						h.Lock.Unlock()
						log.Println(h.Name, "I became leader")
					}
				} else {
					log.Println(h.Name, "I got msg func:", string(msg))
				}
			}

		}(conn)
	}
}

func (h *Hub) InitiateElection() {
	log.Println(h.Name, "Election started")

	for _, conn := range h.connList {
		conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))
		err := conn.WriteMessage(websocket.TextMessage, []byte("0"))
		if err != nil {
			log.Println(h.Name, "Error", err)
		}
	}

}

func (h *Hub) SendMessage(message string) {
}

func (h *Hub) AddConn(conn *websocket.Conn) {
	h.Lock.Lock()
	h.connList = append(h.connList, conn)
	h.Lock.Unlock()
}

func (h *Hub) newTimerEveryMin(sec int) *time.Ticker {
	currTime := time.Now()
	timeLeft := time.Duration(((ONE_MIN-currTime.Second())+sec)%ONE_MIN) * time.Second

	time.Sleep(timeLeft)

	return time.NewTicker(time.Minute)
}

func (h *Hub) Pong(conn *websocket.Conn) {
	conn.SetReadDeadline(time.Now().Add(PONGTIME))
	conn.SetPingHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(PONGTIME))
		conn.WriteMessage(websocket.PongMessage, []byte("PONG"))
		log.Println(h.Name, "PING")
		return nil
	})

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println(h.Name, "Pong: ", err)
			break
		}
	}
}
