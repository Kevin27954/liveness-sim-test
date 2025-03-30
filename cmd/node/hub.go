package node

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/Kevin27954/liveness-sim-test/assert"
	"github.com/Kevin27954/liveness-sim-test/db"
	"github.com/gorilla/websocket"
)

const (
	WRITEWAIT = 10 * time.Second
	PONGTIME  = 70 * time.Second
	PINGTIME  = (PONGTIME * 9) / 10
	ONE_MIN   = 60

	SYNC_TIME = 3 * time.Second
)

type Hub struct {
	Name    string
	ConnStr string
	DB      db.DB

	connList         []*websocket.Conn
	Lock             sync.Mutex
	IsLeader         bool
	hasLeader        *websocket.Conn
	currentConsensus int
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

		// Get latest log info.
		// Send the latest log over: time, operation, id?, msg, etc
		log.Println(h.Name, "Should be request: ", appData)

		err := conn.WriteMessage(websocket.PongMessage, []byte("Latest is at: XX:XX"))
		if err != nil {
			log.Println(h.Name, "Ping Err:", err)
		}

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

		switch string(msg) {
		case CONSENSUS_YES:
			assert.Assert(true, h.IsLeader, true, "Only Leaders can recieve votes")
			h.Lock.Lock()
			h.currentConsensus += 1
			h.Lock.Unlock()
			log.Println("I got consensus")

			// Potential Scenario where you don't receive all the CONSENSUS
			if h.currentConsensus == len(h.connList) {
				// for _, conn := range h.connList {
				// 	conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))
				// 	// Current Operation at HEAD of Queue
				// 	err := conn.WriteMessage(websocket.TextMessage, []byte(NEW_MSG_ADD))
				// 	if err != nil {
				// 		log.Println(h.Name, "Error", err)
				// 	}
				// }

				h.Lock.Lock()
				h.currentConsensus = 0
				// Pop Current Operation from Queue

				// Current Operation at HEAD of Queue
				// The message also needs to be kept
				// It will be added to nodes during the SYNC phase
				h.DB.AddOperation(NEW_MSG_ADD, "This is a Test")

				h.Lock.Unlock()

				// We need to implement the SYNC immediately. It can be an extension of Ping Pong perhaps?
				// or like replacing the Ping and Pong that exists.
			}

		case CONSENSUS_NO:
			assert.Assert(true, h.IsLeader, true, "Only Leaders can recieve votes")

		case VOTE_NO:
			// Nothing happens
		case VOTE_YES:

			if h.hasLeader == nil {
				h.Lock.Lock()
				h.IsLeader = true
				h.Lock.Unlock()
				log.Println(h.Name, "I became leader")
			}

		case ELECTION:

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

		case AM_LEADER:

			h.Lock.Lock()
			h.hasLeader = conn
			h.Lock.Unlock()

		case NEW_MSG_ADD:
			// If is leader, send it to other conns

			// Respond with vote yes or vote no?

			if h.IsLeader {
				// Sends to all other nodes.
				h.StoreMessage()

			} else {
				// Check if I can store it or not
				// if yes, send "consensus_agree"
				// else, send "consensus_disagree"

				// "consensus_agree" & "consensus_disagree" are going to be the
				// general code for all consensus related. Each send is related to
				// current consensus_topic in queue I believe.

				conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))
				h.hasLeader.WriteMessage(websocket.TextMessage, []byte(CONSENSUS_YES))
			}

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
		default:
			log.Println(h.Name, "I got msg func:", string(msg))
		}

	}
}

// Can probably pass the code in the input params and able to deal with both
// add msg and delete msg.
func (h *Hub) StoreMessage() {
	// Logic should be encompassed here such that ti can be called
	// when someone adds a message.

	if h.IsLeader {

		for _, conn := range h.connList {
			conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))
			err := conn.WriteMessage(websocket.TextMessage, []byte(NEW_MSG_ADD))
			if err != nil {
				log.Println(h.Name, "Error", err)
			}

			// Probably store the thing in a queue?
			// When they send msg back, we do stuff to it (e.g. remove from queue?).
		}

	} else {
		//sends conn to leader
		h.hasLeader.SetWriteDeadline(time.Now().Add(WRITEWAIT))
		err := h.hasLeader.WriteMessage(websocket.TextMessage, []byte(NEW_MSG_ADD))
		if err != nil {
			log.Println(h.Name, "Error", err)
		}
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

// It should be rebranded to SYNC
func (h *Hub) Ping() {
	for _, conn := range h.connList {
		go func(conn *websocket.Conn) {

			conn.SetPongHandler(func(appData string) error {
				conn.SetReadDeadline(time.Now().Add(PONGTIME))
				conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))

				// Get the missing logs starting fro their latest logs.
				log.Println(h.Name, "Should be telling me when(everything about the log): ", appData)
				// get logs (will be in string format.

				err := conn.WriteMessage(websocket.TextMessage, []byte("Should be sending missing logs if any"))
				if err != nil {
					log.Println(h.Name, "Pong Err:", err)
				}

				return nil
			})

			conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))
			err := conn.WriteMessage(websocket.PingMessage, []byte("Request Latest Log"))
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
