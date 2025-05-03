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
	ONE_MIN   = 30

	SYNC_TIME = 3 * time.Second

	SEPERATOR = "≡" // A Hambuger menu looking thing
)

type Task struct {
	operation string
	data      string
}

type BinarySearch struct {
	Low  int
	Mid  int
	High int
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

	syncMap map[*websocket.Conn]BinarySearch // The thing that keeps track of binary serach idx

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
	h.syncMap = make(map[*websocket.Conn]BinarySearch)
	h.term = -1

	defer h.DB.Close()

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

			h.Lock.Lock()
			h.removeConn(conn)
			h.Lock.Unlock()
			return
		}

		splitMsg := strings.Split(string(msg), SEPERATOR)

		// Piece it back together?
		code, data := splitMsg[0], strings.Join(splitMsg[1:], SEPERATOR)

		switch string(code) {
		case SYNC_REQ_ASK:
			assert.Assert(h.IsLeader == false, h.IsLeader, false, "Only NON Leaders can recieve SYNC_REQ_ASK")
			log.Println("Data gotten for sync: ", data)

			ops, err := db.ParseLog(data)
			if err != nil {
				log.Println(h.Name, "Unable to parse Data")
				continue
			}

			has, err := h.DB.HasLog(ops)
			if err != nil {
				log.Println(h.Name, "Unable to check for log")
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

			cpy := h.syncMap[conn]

			switch code {
			case SYNC_REQ_HAS:
				cpy.Low = cpy.Mid + 1
				cpy.Mid = (cpy.High + cpy.Low) / 2

			case SYNC_REQ_NO_HAS:
				cpy.High = cpy.Mid - 1
				cpy.Mid = (cpy.High + cpy.Low) / 2
			}

			conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))
			if cpy.Low <= cpy.High {
				syncInitMsg := SYNC_REQ_ASK

				nextLog, err := h.DB.GetLogByID(cpy.Mid)
				if err != nil {
					log.Println(h.Name, "Unable to get Log")
					continue
				}

				// This would just be one log
				syncInitMsg += SEPERATOR + nextLog.String()

				err = conn.WriteMessage(websocket.TextMessage, []byte(syncInitMsg))
				if err != nil {
					log.Println(h.Name, "Sync Init", err)
				}
			} else { // It found the matching log
				syncReqCommitMsg := SYNC_REQ_COMMIT

				opsArr, err := h.DB.GetLogsById(cpy.High)
				if err != nil {
					log.Println(h.Name, "Unable to get Log")
					continue
				}

				for _, ops := range opsArr {
					syncReqCommitMsg += SEPERATOR + ops.String()
				}

				err = conn.WriteMessage(websocket.TextMessage, []byte(syncReqCommitMsg))
				if err != nil {
					log.Println(h.Name, "Unable to send message")
				}

			}

			h.Lock.Lock()
			h.syncMap[conn] = cpy
			h.Lock.Unlock()

		case SYNC_REQ_COMMIT:
			assert.Assert(h.IsLeader == false, h.IsLeader, false, "Only NON Leaders can recieve SYNC_REQ_COMMIT")

			opsStr := strings.Split(data, SEPERATOR)

			var opsArr []db.Operation

			for _, ops := range opsStr {
				if ops == "" {
					continue
				}

				ops, err := db.ParseLog(ops)
				if err != nil {
					// DO I SKIP THIS LOG?
					log.Println("Unable to parse OPS, ", err)
					continue
				}

				opsArr = append(opsArr, ops)
			}

			if len(opsArr) == 0 { // Meaning it was empty (no SEPERATOR)
				log.Println(h.Name, "Commit Arr Was empty")
				continue
			}

			_, err = h.DB.CommitLogs(opsArr)
			if err != nil {
				log.Println(h.Name, "error commiting logs")
			}

		case CONSENSUS_YES:
			assert.Assert(true, h.IsLeader, true, "Only Leaders can recieve votes")

			idx, err := strconv.Atoi(data)
			if err != nil {
				log.Println("Error parsing num from consensus yes: ", err)
				continue
			}

			if idx < h.taskCompleted {
				log.Println("IGNORED ", idx, " ", h.taskCompleted)
				continue // IGNORE VOTE
			}

			h.Lock.Lock()
			h.consensusMap[idx%len(h.consensusMap)] += 1
			h.Lock.Unlock()

			if h.consensusMap[idx%len(h.consensusMap)] > len(h.connList)/2 {
				h.Lock.Lock()
				h.taskCompleted += 1
				h.consensusMap[idx%len(h.consensusMap)] = 0

				// POP from taskQueue
				operation := h.taskQueue[0]
				h.taskQueue = h.taskQueue[1:]
				h.Lock.Unlock()

				h.DB.AddOperation(operation.operation, h.term, operation.data)
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
						log.Println(h.Name, "Unable to send 'Leader Ping' to nodes: ", err)
						continue
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
			h.hasVoted = false //reset
			h.Lock.Unlock()

		case NEW_MSG_ADD:

			if h.IsLeader {
				h.StoreMessage(data)

			} else {
				// Data here is the taskStarted number
				conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))
				err := h.hasLeader.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%s%s%s", CONSENSUS_YES, SEPERATOR, data)))
				if err != nil {
					log.Println(h.Name, "Error writing back to leader, NEW_MSG_ADD: ", err)
				}
			}

		default:
			log.Println(h.Name, "I got msg: ", string(msg))
		}

	}
}

// Can probably pass the code in the input params and able to deal with both
// add msg and delete msg.
func (h *Hub) StoreMessage(message string) {
	if h.hasLeader == nil && !h.IsLeader {
		log.Println("Leader is not elected yet")
		return
	}

	if h.IsLeader {

		for _, conn := range h.connList {
			conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))
			err := conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%s%s%d", NEW_MSG_ADD, SEPERATOR, h.taskStarted)))
			if err != nil {
				log.Println(h.Name, "Error Store Message", err)
				return
			}

		}

		h.Lock.Lock()
		h.taskStarted += 1
		h.taskQueue = append(h.taskQueue, Task{operation: NEW_MSG_ADD, data: message})
		h.Lock.Unlock()

	} else {
		//sends conn to leader
		h.hasLeader.SetWriteDeadline(time.Now().Add(WRITEWAIT))
		err := h.hasLeader.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%s%s%s", NEW_MSG_ADD, SEPERATOR, message)))
		if err != nil {
			log.Println(h.Name, "Error Store Message", err)
		}
	}
}

func (h *Hub) InitiateElection() {
	assert.Assert(h.hasLeader == nil, h.hasLeader, nil, "Leader should be nil")

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
			return
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

			low := 1
			high, err := h.DB.GetNumLogs()
			if err != nil {
				// If this breaks, the whole thing doesn't work
				log.Fatal(h.Name, "DB Query err: ", err)
			}

			h.Lock.Lock()
			h.syncMap[conn] = BinarySearch{Low: low, Mid: (low + high) / 2, High: high}
			h.Lock.Unlock()

			if h.syncMap[conn].Mid == 0 {
				log.Println(h.Name, " The idx is 0 meaning there is nothing?")
				return
			}

			syncInitMsg := SYNC_REQ_ASK

			nextLog, err := h.DB.GetLogByID(h.syncMap[conn].Mid)
			if err != nil {
				log.Println(h.Name, "Unable to get Log, ", err)
				return
			}

			syncInitMsg += SEPERATOR + nextLog.String()

			conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))
			err = conn.WriteMessage(websocket.TextMessage, []byte(syncInitMsg))
			if err != nil {
				log.Println(h.Name, "Sync Init", err)
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

func (h *Hub) nodeAllAgree() bool {
	minConsensus := len(h.connList) / 2
	// minConsensus := len(h.connList)

	return h.currentConsensus >= minConsensus
}
