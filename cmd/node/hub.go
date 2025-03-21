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
	hasLeader *websocket.Conn
}

func (h *Hub) ConnectConns() {
	assert.Assert(len(h.ConnStr) > 0, len(h.ConnStr), 1, "Expected length greater than 0")

	log.Println(h.Name, "Connecting to other nodes")

	conns := strings.Split(h.ConnStr, ",")
	for _, connStr := range conns {
		conn, _, err := websocket.DefaultDialer.Dial(connStr, nil)
		assert.NoError(err, "Unable to connect to other nodes")

		log.Printf("Connected to %s", connStr)

		h.AddConn(conn)
	}
}

func (h *Hub) Run(addr int) {
	if len(h.ConnStr) > 0 {
		h.ConnectConns()
	}

	interval := (addr % 8000) * 5

	newTicker := h.newTimerEveryMin(interval)
	defer newTicker.Stop()

	for {
		select {
		case <-newTicker.C:
			// Send a rquest to all conn, if conn is up and recieve a single no then we sa it is reject
			// if all conn is yes and no err, then we proceed with request.

			log.Println(h.Name, "I am leader: ", h.IsLeader)
			if h.IsLeader {
				// Leader Ping
				h.Ping()
			} else if h.hasLeader != nil {
				// Idk what yet
			} else {
				h.InitiateElection()
			}

		}
	}
}

func (h *Hub) RecieveMessage(conn *websocket.Conn) {
	log.Println("Listening For Messages")

	conn.SetPingHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(PONGTIME))
		conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))
		err := conn.WriteMessage(websocket.PongMessage, []byte("PONG"))
		if err != nil {
			log.Println(h.Name, "Ping Err:", err)

		}

		log.Println(h.Name, "PING")
		return nil
	})

	for {
		// The idea is that we don't know WHEN we might receive a message. So we just want to wait and be on
		// the lookout for any potential messages that might come.
		// Removes the deadline

		// This does not work if it s a Ping Message, as if there is a PingHandler, it won't be counted towards
		// the message and thus all readmessage is not ran. E.g. this line below does not work as conn.RecieveMessage()
		// will never actually run unless it is a non Ping message.
		// ```err := conn.SetReadDeadline(time.Now().Add(PONGTIME))```

		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(h.Name, "Read Err: ", err)

			// Reset the Status
			h.Lock.Lock()
			h.removeConn(conn)
			h.Lock.Unlock()
			return
		}

		if string(msg) == ELECTION {
			conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))
			if h.IsLeader {
				conn.WriteMessage(websocket.TextMessage, []byte(AM_LEADER))
			} else if h.hasLeader != nil {
				conn.WriteMessage(websocket.TextMessage, []byte(VOTE_NO))
			} else {
				conn.WriteMessage(websocket.TextMessage, []byte(VOTE_YES))
				h.Lock.Lock()
				h.hasLeader = conn
				h.Lock.Unlock()
			}

		} else if string(msg) == VOTE_YES {
			if h.hasLeader == nil {
				h.Lock.Lock()
				h.IsLeader = true
				h.Lock.Unlock()
				log.Println(h.Name, "I became leader")
			}

		} else if string(msg) == AM_LEADER {
			h.Lock.Lock()
			h.hasLeader = conn
			h.Lock.Unlock()

		} else if string(msg) == NEW_MSG_ADD {
			// If is leader, send it to other conns
			// if is non leader, send it to the leader

			// Respond with vote yes or vote no?

			// Pretend this is leader sending:
			// leader sends new_msg_aadd request to all nodes
			// alll nodes sends back a yes or no
			// if yes then leader confirms and tells others to init request
			// else leaders tells others to say no
			// sends response message back to the node/user

			// To do this, I'm pretty sure ou need to store
			// What it is that is being voted for right now. Thus far there is only one.
			// But now there ist two, so there isa  need to have a way of what is being voted on.
			// Wellactually there is still only one, since election is done to itself rather than
			// being shred with other nodes.

			// we then also need to know what it is we are adding to the db.

			// Pretend this is node sending to leader:
			// Gets msg request
			// sends to the leader
			// proceed with above scenario
			// recieve result of consensus
			// send backs error message
		} else {
			log.Println(h.Name, "I got msg func:", string(msg))
		}
	}

}

func (h *Hub) StoreMessage() {
	// Logic should be encompassed here such that ti can be called
	// when someone adds a message.

	if h.IsLeader {
		// sends conn to everyone

		for _, conn := range h.connList {
			conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))
			err := conn.WriteMessage(websocket.TextMessage, []byte(NEW_MSG_ADD))
			if err != nil {
				log.Println(h.Name, "Error", err)
			}
		}

	} else {
		//sends conn to leader
	}
}

func (h *Hub) InitiateElection() {
	log.Println(h.Name, "Election started")

	for _, conn := range h.connList {
		conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))
		err := conn.WriteMessage(websocket.TextMessage, []byte(ELECTION))
		if err != nil {
			log.Println(h.Name, "Error", err)
		}
	}

}

func (h *Hub) removeConn(conn *websocket.Conn) {
	var removeIdx = -1
	for i, c := range h.connList {
		if c == conn {
			removeIdx = i
			break
		}
	}

	if removeIdx == -1 {
		return
	}

	if conn == h.hasLeader {
		h.hasLeader = nil
	}

	assert.Assert(removeIdx != -1, removeIdx, 0, "Remove index should not be -1")

	h.connList[removeIdx] = h.connList[len(h.connList)-1]
	h.connList = h.connList[:len(h.connList)-1]
}

func (h *Hub) AddConn(conn *websocket.Conn) {
	h.Lock.Lock()
	h.connList = append(h.connList, conn)
	h.Lock.Unlock()
	go h.RecieveMessage(conn)
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

func (h *Hub) newTimerEveryMin(sec int) *time.Ticker {
	currTime := time.Now()
	timeLeft := time.Duration(((ONE_MIN-currTime.Second())+sec)%ONE_MIN) * time.Second

	time.Sleep(timeLeft)

	return time.NewTicker(time.Minute)
}
