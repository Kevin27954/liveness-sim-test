package raft

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Kevin27954/liveness-sim-test/db"
	p "github.com/Kevin27954/liveness-sim-test/pkg"
	task "github.com/Kevin27954/liveness-sim-test/pkg/task"
	"github.com/Kevin27954/liveness-sim-test/pkg/transponder"
	"github.com/Kevin27954/liveness-sim-test/test/randomizer"
	"github.com/gorilla/websocket"
)

const (
	HEARTBEAT_TIME = 3 * time.Second
	WAIT_TIME      = 7

	SEPERATOR = "≡" // A Hambuger menu looking thing
)

type BinarySearch struct {
	Low  int
	Mid  int
	High int
}

type Raft struct {
	name string
	id   int // port % 8000

	syncMap map[int]BinarySearch // The thing that keeps track of binary serach idx

	hasVoted bool
	term     int
	isLeader bool

	eventCh         chan any // We only want specific types (structs); will be filtered.
	heartBeatTicker *time.Ticker
	close           bool

	closeChan chan int

	task        task.TaskManager
	Db          db.DB
	transponder transponder.Transponder
}

func Init(name string, id int, db db.DB, portList string, size int) *Raft {

	r := Raft{
		name: name,
		id:   id,

		eventCh: make(chan any),
		syncMap: make(map[int]BinarySearch),

		hasVoted: false,
		term:     -1,
		isLeader: false,

		task:        task.Init(),
		Db:          db,
		transponder: transponder.Init(id, size),
	}

	r.transponder.OnRecv(func(msg any) {
		go func() { r.eventCh <- msg }()
	})

	go r.transponder.StartConns(portList)
	go r.Run()

	return &r
}

func (r *Raft) Run() {
	defer r.Db.Close()

	r.heartBeatTicker = r.newTimerEveryMin(WAIT_TIME)
	// r.heartBeatTicker = r.newTimerEveryMin(totalWait)
	defer r.heartBeatTicker.Stop()

	for {
		// Time of every 3 seconds
		select {
		case event := <-r.eventCh:
			switch event.(type) {
			case p.MessageEvent:
				r.handleEvent(event.(p.MessageEvent))
			case p.TickerEvent:
				log.Println("I should not run")
			default:
				log.Println("Unknown Type")
			}
		case <-r.heartBeatTicker.C:
			if !r.isLeader {
				r.InitiateElection()
			}
		case <-r.closeChan:
			return

		}
	}
}

func (r *Raft) InitiateElection() {
	log.Println(r.name, "Election started")

	r.hasVoted = true
	r.term += 1

	taskId := r.task.AddTask(p.ELECTION, "")

	msg := r.transponder.CreateMsg(p.ELECTION, r.id, taskId, r.term)
	r.transponder.Write(msg)
}

// Write to the connection with id
func (r *Raft) startHeartBeat(id int) {
	heartBeatTicker := time.NewTicker(3 * time.Second)

	if r.task.NumMsg() > 0 {
		msgs := r.task.GetQueuedMsg()
		for _, msg := range msgs {
			// r.transponder.WriteTo(id, r.transponder.CreateMsg(p.APPEND_ENTRIES, r.id, msg))
			r.handleNewOp(id, msg)
		}
	}

	for {
		if !r.isLeader {
			log.Println("I was not the leader")
			return
		}

		select {
		case <-heartBeatTicker.C:
			err := r.transponder.WriteTo(id, r.transponder.CreateMsg(p.APPEND_ENTRIES, r.id))
			if err != nil {
				log.Println("From: ", r.id, "Error", err)
				return
			}
			r.task.SetMaxNodes(r.transponder.GetTotalConns())

		}
	}
}

func (r *Raft) SendNewOp(operation string, msg string) {
	// log.Println(r.name, " Value of isleader is: ", r.isLeader)
	if !r.isLeader {
		if r.hasVoted {
			// If a new term started and leader is elected, send to leader
			r.transponder.Write(r.transponder.CreateMsg(p.NEW_OP, r.id, operation, msg))
		} else {
			// store in task manager, send when they finished voting.
			// TODO Store the operatin and msg in this.
			r.task.AddQueueTask(operation, msg)
		}

		return
	}

	err := r.Db.AddOperation(operation, r.term, msg)
	if err != nil {
		log.Println("Unable to add operation: ", err)
	}

	taskId := r.task.AddTask(operation, msg)

	r.transponder.Write(r.transponder.CreateMsg(p.APPEND_ENTRIES, r.id, taskId, operation, msg))
}

func (r *Raft) GetState(w http.ResponseWriter, req *http.Request) {
	var sb strings.Builder

	sb.WriteString("Raft State:\n")
	sb.WriteString(fmt.Sprintf("  Name: %s\n", r.name))
	sb.WriteString(fmt.Sprintf("  ID: %d\n", r.id))
	sb.WriteString(fmt.Sprintf("  Term: %d\n", r.term))
	sb.WriteString(fmt.Sprintf("  IsLeader: %t\n", r.isLeader))
	sb.WriteString(fmt.Sprintf("  HasVoted: %t\n", r.hasVoted))

	sb.WriteString("  SyncMap Keys: [")
	first := true
	for k := range r.syncMap {
		if !first {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%d", k))
		first = false
	}
	sb.WriteString("]\n")

	sb.WriteString(fmt.Sprintf("  TaskManager: %s\n", r.task.String()))
	// sb.WriteString(fmt.Sprintf("  DB: %T\n", r.Db))
	sb.WriteString(fmt.Sprintf("  Transponder: %s\n", r.transponder.String()))

	fmt.Fprintf(w, sb.String())
}

func (r *Raft) newTimerEveryMin(wait int) *time.Ticker {
	time.Sleep(time.Duration(wait) * time.Second)

	myRand := randomizer.Init(420)
	for range r.id {
		myRand.GetIntN(11)
	}

	electionWait := myRand.GetIntRange(5000, 12345)
	log.Println(electionWait)
	timeout := time.Duration(electionWait) * time.Millisecond
	return time.NewTicker(timeout)
}

func (r *Raft) AddConn(conn *websocket.Conn, id int) {
	r.transponder.AddConn(conn, id)
}

func (r *Raft) IsLeader() bool {
	return r.isLeader
}

func (r *Raft) Quit() bool {
	r.Db.Close()
	go func() {
		r.closeChan <- 1
		log.Println("Idk did I run?")
	}()
	return true
}
