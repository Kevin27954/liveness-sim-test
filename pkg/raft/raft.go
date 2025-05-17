package raft

import (
	"fmt"
	"log"
	"time"

	"github.com/Kevin27954/liveness-sim-test/db"
	p "github.com/Kevin27954/liveness-sim-test/pkg"
	task "github.com/Kevin27954/liveness-sim-test/pkg/task"
	"github.com/Kevin27954/liveness-sim-test/pkg/transponder"
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

	var totalWait int
	for range r.id {
		totalWait = totalWait + WAIT_TIME
	}
	r.heartBeatTicker = r.newTimerEveryMin(totalWait)
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
				// r.heartBeatTicker.Reset(10 * time.Second)
				// r.handleTicker(event.(p.TickerEvent))
			default:
				log.Println("Unknown Type")
			}
		case <-r.heartBeatTicker.C:
			if !r.isLeader {
				r.InitiateElection()
			}

		}
	}
}

func (r *Raft) InitiateElection() {
	log.Println(r.name, "Election started")

	r.hasVoted = true
	r.term += 1

	// TODO: Look into whether or not I need to pass anything in with this
	taskId := r.task.AddTask(p.ELECTION)

	// r.transponder.Write(fmt.Appendf([]byte(""), "%s%s%d%s%d%s%d", p.ELECTION, SEPERATOR, r.id, SEPERATOR, taskId, SEPERATOR, r.term))
	msg := r.transponder.CreateMsg(p.ELECTION, r.id, taskId, r.term)
	log.Println(r.name, "Created:", msg)
	r.transponder.Write(msg)
}

// Write to the connection with id
func (r *Raft) startHeartBeat(id int) {
	heartBeatTicker := time.NewTicker(3 * time.Second)

	for {
		if !r.isLeader {
			log.Println("I ws not the leader")
			return
		}

		select {
		case <-heartBeatTicker.C:
			// r.transponder.WriteTo(id, fmt.Appendf([]byte(""), "%s%s%d", p.APPEND_ENTRIES, SEPERATOR, r.id))
			r.transponder.WriteTo(id, r.transponder.CreateMsg(p.APPEND_ENTRIES, r.id))

		}
	}
}

func (r *Raft) SendNewOp(operation string, msg string) {
	log.Println(r.name, " Value of isleader is: ", r.isLeader)
	if !r.isLeader {
		r.transponder.Write(r.transponder.CreateMsg(p.NEW_OP, r.id, operation, msg))
		return
	}

	log.Println(r.name, "I tested and ran successfully")

	err := r.Db.AddOperation(operation, r.term, msg)
	if err != nil {
		log.Println("Unable to add operation: ", err)
	}

	taskId := r.task.AddTask("")

	// r.transponder.Write(fmt.Appendf([]byte(""), "%s%s%d%s%d%s%s", p.APPEND_ENTRIES, SEPERATOR, r.id, SEPERATOR, taskId, SEPERATOR, operation+SEPERATOR+msg))
	r.transponder.Write(r.transponder.CreateMsg(p.APPEND_ENTRIES, r.id, taskId, operation, msg))
}

func (r *Raft) newTimerEveryMin(wait int) *time.Ticker {
	time.Sleep(time.Duration(wait) * time.Second)

	return time.NewTicker(10 * time.Second)
}

func (r *Raft) AddConn(conn *websocket.Conn, id int) {
	r.transponder.AddConn(conn, id)
}

func (r *Raft) Info() string {
	return fmt.Sprint(r.isLeader, " - ", r.term)
}
