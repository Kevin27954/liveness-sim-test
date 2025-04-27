package node

import (
	"fmt"
	"log"
	"strconv"
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

	SEPERATOR = "≡" // A Hambuger menu looking thing
)

type Task struct {
	operation string
	data      string
}

type Hub struct {
	Name     string
	ConnStr  string
	DB       db.DB
	connList []*websocket.Conn

	Lock sync.Mutex

	IsLeader         bool
	hasVoted         bool
	hasLeader        *websocket.Conn
	currentConsensus int
	term             int

	syncMap map[*websocket.Conn]int // The thing that keeps track of binary serach idx

	consensusMap  [20]int // This will just be a random number for now.
	taskQueue     []Task
	taskCompleted int
	taskStarted   int
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
	h.syncMap = make(map[*websocket.Conn]int)
	h.term = -1

	newTicker := h.newTimerEveryMin(interval)
	defer newTicker.Stop()

	for {
		select {
		case <-newTicker.C:
			// Send a rquest to all conn, if conn is up and recieve a single no then we sa it is reject
			// if all conn is yes and no err, then we proceed with request.

			log.Println(h.Name, "I am leader: ", h.IsLeader)
			if h.IsLeader {
				// Leader Ping SYNC REQ
				h.SyncReqInit()
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
	for {
		// The idea is that we don't know WHEN we might receive a message. So we just want to wait and be on
		// the lookout for any potential messages that might come.

		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(h.Name, "Read Err: ", err)

			// Reset the Status
			h.Lock.Lock()
			h.removeConn(conn)
			h.Lock.Unlock()
			return
		}

		splitMsg := strings.Split(string(msg), SEPERATOR)

		// Piece it back together?
		code, data := splitMsg[0], strings.Join(splitMsg[1:], SEPERATOR)

		// This is basically the index / ID in db
		low := 0
		high, err := h.DB.GetNumLogs()
		if err != nil {
			log.Println(h.Name, "DB Query err: ", err)

			h.Lock.Lock()
			h.removeConn(conn)
			h.Lock.Unlock()
			return
		}

		h.syncMap[conn] = (high + low) / 2

		switch string(code) {
		case SYNC_REQ_ASK:
			assert.Assert(h.IsLeader == false, h.IsLeader, false, "Only NON Leaders can recieve SYNC_REQ_ASK")
			log.Println("Data gotten for sync: ", data)

			has, err := h.DB.HasLog(data)
			if err != nil {
				log.Println("Unable to check for log")
			}
			//check if I have log that was received.
			h.hasLeader.SetWriteDeadline(time.Now().Add(WRITEWAIT))
			if has {
				h.hasLeader.WriteMessage(websocket.TextMessage, []byte(SYNC_REQ_HAS))
			} else {
				h.hasLeader.WriteMessage(websocket.TextMessage, []byte(SYNC_REQ_NO_HAS))
			}

		case SYNC_REQ_HAS, SYNC_REQ_NO_HAS:
			assert.Assert(h.IsLeader == true, h.IsLeader, true, "Only Leaders can recieve SYNC_REQ_HAS")

			switch code {
			case SYNC_REQ_HAS:
				low = h.syncMap[conn] + 1
				h.Lock.Lock()
				h.syncMap[conn] = (high + low) / 2
				h.Lock.Unlock()

			case SYNC_REQ_NO_HAS:
				high = h.syncMap[conn] - 1
				h.Lock.Lock()
				h.syncMap[conn] = (high + low) / 2
				h.Lock.Unlock()

			}

			conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))
			if low <= high {
				syncInitMsg := SYNC_REQ_ASK

				nextLog, err := h.DB.GetLogByID(h.syncMap[conn])
				if err != nil {
					log.Println("Unable to get Log")
				}

				// This would just be one log
				syncInitMsg += SEPERATOR + nextLog.String()
				log.Println(syncInitMsg)

				err = conn.WriteMessage(websocket.TextMessage, []byte(syncInitMsg))
				if err != nil {
					log.Println(h.Name, "Sync Init", err)
				}
			} else { // It found the matching log
				syncReqCommitMsg := SYNC_REQ_COMMIT

				opsArr, err := h.DB.GetLogsById(high)
				if err != nil {
					log.Println("Unable to get Log")
				}

				for _, ops := range opsArr {
					syncReqCommitMsg += SEPERATOR + ops.String()
				}
				conn.WriteMessage(websocket.TextMessage, []byte(syncReqCommitMsg))
			}

		case SYNC_REQ_COMMIT:
			assert.Assert(h.IsLeader == false, h.IsLeader, false, "Only NON Leaders can recieve SYNC_REQ_COMMIT")

			low = 0
			high, err = h.DB.GetNumLogs()
			if err != nil {
				log.Println(h.Name, "DB Query err: ", err)

				h.Lock.Lock()
				h.removeConn(conn)
				h.Lock.Unlock()
				return
			}

			log.Println(h.Name, "Operations: ", data)

		case CONSENSUS_YES:
			assert.Assert(true, h.IsLeader, true, "Only Leaders can recieve votes")

			log.Println("I got consensus")

			idx, err := strconv.Atoi(data)
			if err != nil {
				log.Println("Error parsing num from consensus yes: ", err)
			}

			if idx < h.taskCompleted {
				continue // IGNORE VOTE
			}

			h.consensusMap[idx%len(h.consensusMap)] += 1

			if h.consensusMap[idx%len(h.consensusMap)] > len(h.connList)/2 {
				h.Lock.Lock()
				h.taskCompleted += 1
				h.consensusMap[idx%len(h.consensusMap)] = 0

				// POP from taskQueue
				operation := h.taskQueue[0]
				h.taskQueue = h.taskQueue[1:]

				h.DB.AddOperation(operation.operation, h.term, operation.data)
				h.Lock.Unlock()
			}

		case CONSENSUS_NO:
			assert.Assert(true, h.IsLeader, true, "Only Leaders can recieve votes")

		case VOTE_NO:
			// Nothing happens
		case VOTE_YES:

			if !h.nodeAllAgree() && !h.IsLeader {
				h.Lock.Lock()
				h.currentConsensus += 1
				h.Lock.Unlock()
			} else if h.nodeAllAgree() {
				h.Lock.Lock()
				h.IsLeader = true
				h.currentConsensus = 0 // Reset
				h.Lock.Unlock()

				for _, conn := range h.connList {
					conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))
					// Perhaps change this to be a start of the log request??
					err := conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%s%s%d", AM_LEADER, SEPERATOR, h.term)))
					if err != nil {
						log.Println("Unable to send 'Leader Ping' to nodes")
						return
					}
				}

				log.Println(h.Name, "I became leader")
			}

		case ELECTION:
			conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))

			if h.IsLeader {
				// If election is send to leader node on accident
				conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%s%s%d", AM_LEADER, SEPERATOR, h.term)))
			} else if !h.hasVoted {
				if h.hasLeader != nil {
					conn.WriteMessage(websocket.TextMessage, []byte(VOTE_NO))
				} else {
					conn.WriteMessage(websocket.TextMessage, []byte(VOTE_YES))
				}

				h.Lock.Lock()
				h.hasVoted = true
				h.term += 1
				h.Lock.Unlock()
			}

		case AM_LEADER:

			currTerm, err := strconv.Atoi(data)
			assert.NoError(err, "Unable to convert term from string to int from AM_LEADER ping")

			h.Lock.Lock()
			h.hasLeader = conn
			if h.term == -1 {
				h.term = currTerm
			}
			h.Lock.Unlock()

		case NEW_MSG_ADD:
			// If is leader, send it to other conns

			if h.IsLeader {
				// Sends to all other nodes.

				h.StoreMessage("")
				// This is leader so it can be blank
				h.Lock.Lock()
				// I actually need to set up the code bruh
				h.taskQueue = append(h.taskQueue, Task{operation: code, data: data})
				h.Lock.Unlock()

			} else {
				// Check if I can store it or not
				// if yes, send "consensus_agree"
				// else, send "consensus_disagree"

				// Data here is the taskStarted number
				conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))
				err := h.hasLeader.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%s%s%s", CONSENSUS_YES, SEPERATOR, data)))
				if err != nil {
					log.Println("Error writng back to leader, new msg add: ", err)
				}
			}

			// Pretend this is leader sending:
			// leader sends new_msg_aadd request to all nodes
			// all nodes sends back a yes or no
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
// Leader doesn't need an actual message
func (h *Hub) StoreMessage(message string) {

	if h.IsLeader {

		for _, conn := range h.connList {
			conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))
			err := conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%s%s%d", NEW_MSG_ADD, SEPERATOR, h.taskStarted)))
			if err != nil {
				log.Println(h.Name, "Error", err)
			}

			// This might be needed here
			h.Lock.Lock()
			h.taskStarted += 1
			h.Lock.Unlock()

			// Probably store the thing in a queue?
			// When they send msg back, we do stuff to it (e.g. remove from queue?).
		}

	} else {
		//sends conn to leader
		h.hasLeader.SetWriteDeadline(time.Now().Add(WRITEWAIT))
		err := h.hasLeader.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%s%s%s", NEW_MSG_ADD, SEPERATOR, message)))
		if err != nil {
			log.Println(h.Name, "Error", err)
		}
	}
}

func (h *Hub) InitiateElection() {
	if h.hasLeader != nil {
		log.Println("I ran and and there was a leader")
		return
	}

	log.Println(h.Name, "Election started")

	h.Lock.Lock()
	h.hasVoted = true // Voted for himself / can't participate in votes
	h.term += 1
	h.Lock.Unlock()

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
func (h *Hub) SyncReqInit() {
	for _, conn := range h.connList {
		go func(conn *websocket.Conn) {
			syncInitMsg := SYNC_REQ_ASK

			// We would append the message to the SyncInitMsg
			nextLog, err := h.DB.GetLogByID(h.syncMap[conn])
			if err != nil {
				log.Println("Unable to get Log")
			}

			syncInitMsg += SEPERATOR + nextLog.String()

			conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))
			err = conn.WriteMessage(websocket.TextMessage, []byte(syncInitMsg))
			if err != nil {
				log.Println(h.Name, "Sync Init", err)
			}

			// Store it in a way that can be easily indexed else where.
			h.Lock.Lock()
			h.syncMap[conn] = 0
			h.Lock.Unlock()

		}(conn)
	}
}

func (h *Hub) newTimerEveryMin(sec int) *time.Ticker {
	currTime := time.Now()
	timeLeft := time.Duration(((ONE_MIN-currTime.Second())+sec)%ONE_MIN) * time.Second

	time.Sleep(timeLeft)

	return time.NewTicker(time.Minute)
}

func (h *Hub) nodeAllAgree() bool {
	minConsensus := len(h.connList) / 2
	// minConsensus := len(h.connList)

	return h.currentConsensus >= minConsensus
}
