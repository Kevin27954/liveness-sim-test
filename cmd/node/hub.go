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
	IsLeader bool
}

func (h *Hub) ConnectConns() {
	assert.Assert(len(h.ConnStr) > 0, len(h.ConnStr), 1, "Expected length greater than 0")

	log.Println("Connecting to other nodes")

	conns := strings.Split(h.ConnStr, ",")
	for _, connStr := range conns {
		conn, _, err := websocket.DefaultDialer.Dial(connStr, nil)
		assert.NoError(err, "Unable to connect to other nodes")

		log.Printf("Connected to %s", connStr)

		h.connList = append(h.connList, conn)
	}
}

func (h *Hub) Ping() {
	wg := sync.WaitGroup{}

	for _, conn := range h.connList {
		go func() {
			wg.Add(1)
			defer wg.Done()
			conn.SetPongHandler(func(appData string) error {
				log.Println("PONG")
				return nil
			})

			conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))
			err := conn.WriteMessage(websocket.PingMessage, []byte("PING"))
			if err != nil {
				log.Println("Ping test", err)
			}
			// Reads PONG Message
			_, _, err = conn.ReadMessage()
			if err != nil {
				log.Println("Pong: ", err)
			}

			println("Goo goo Gag aga")

		}()

	}

	wg.Wait()
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

				log.Println("I am leader:", h.IsLeader)

				if h.IsLeader {
					// Leader Ping
					log.Println("I want to ping")

					h.Ping()

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

func (h *Hub) SendMessage(message string) {
}

func (h *Hub) RecieveMessage(consensus chan int) {
	for _, conn := range h.connList {
		go func() {

			// conn.SetReadDeadline(time.Now().Add(PONGTIME))
			conn.SetPingHandler(func(appData string) error {
				// conn.SetReadDeadline(time.Now().Add(PONGTIME))
				conn.WriteMessage(websocket.PongMessage, []byte("PONG"))
				log.Println("PING")
				return nil
			})

			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					log.Println("Err: ", err)
				}

				if string(msg) == "0" {

					conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))
					if h.IsLeader {
						conn.WriteMessage(websocket.TextMessage, []byte("2"))
					} else {
						conn.WriteMessage(websocket.TextMessage, []byte("1"))
					}

				} else if string(msg) == "1" {
					log.Println("I became leader")
					h.IsLeader = true
				} else {
					log.Println("I got msg func:", string(msg))
				}
			}

		}()
	}
}

func (h *Hub) InitiateElection() {
	log.Println("Election started")

	for _, conn := range h.connList {
		conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))
		err := conn.WriteMessage(websocket.TextMessage, []byte("0"))
		if err != nil {
			log.Println("Error", err)
		}
	}

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
