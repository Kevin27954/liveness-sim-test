package raft

import (
	"fmt"
	"log"
	"time"

	"github.com/Kevin27954/liveness-sim-test/cmd/node"
	"github.com/Kevin27954/liveness-sim-test/db"
	p "github.com/Kevin27954/liveness-sim-test/pkg"
	task "github.com/Kevin27954/liveness-sim-test/pkg/task"
	"github.com/Kevin27954/liveness-sim-test/pkg/transponder"
	"github.com/gorilla/websocket"
)

const (
	HEARTBEAT = 3 * time.Second
	WAIT_TIME = 7

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

	syncMap   map[*websocket.Conn]BinarySearch // The thing that keeps track of binary serach idx
	heartBeat chan int

	hasVoted bool
	term     int
	isLeader bool

	eventCh chan any // We only want specific types (structs); will be filtered.

	task        task.TaskManager
	db          db.DB
	transponder transponder.Transponder
}

func Init(name string, id int, db db.DB, portList string, size int) Raft {
	r := Raft{
		name: name,
		id:   id,

		heartBeat: make(chan int),
		eventCh:   make(chan any),

		hasVoted: false,
		term:     -1,
		isLeader: false,

		task:        task.Init(),
		db:          db,
		transponder: transponder.Init(id, size),
	}

	r.transponder.OnRecv(func(msg p.MessageEvent) {
		go func() { r.eventCh <- msg }()
	})

	go r.transponder.StartConns(portList)
	go r.Run()

	return r
}

func (r *Raft) Run() {
	defer r.db.Close()

	r.syncMap = make(map[*websocket.Conn]BinarySearch)

	var totalWait int
	for range r.id {
		totalWait = totalWait + WAIT_TIME
	}
	newTicker := r.newTimerEveryMin(totalWait)
	defer newTicker.Stop()

	for {
		// Time of every 3 seconds
		select {
		case event := <-r.eventCh:
			switch event.(type) {
			case p.MessageEvent:
				r.handleEvent(event.(p.MessageEvent))
			case p.TickerEvent:
				log.Println("Ticker")
			default:
				log.Println("Unknown Type")
			}
		case <-newTicker.C:
			if !r.isLeader {
				r.InitiateElection()
			}

		case <-r.heartBeat:
			newTicker.Reset(10 * time.Second)
			go func() { r.eventCh <- p.TickerEvent{} }()
		}
	}
}

func (r *Raft) InitiateElection() {
	log.Println(r.name, "Election started")

	r.hasVoted = true
	r.term += 1

	// TODO: Look into whether or not I need to pass anything in with this
	taskId := r.task.AddTask(p.MessageEvent{})

	r.transponder.Write(fmt.Appendf([]byte(""), "%s%s%d%s%d%s%d", node.ELECTION, SEPERATOR, r.id, SEPERATOR, taskId, SEPERATOR, r.term))
}

func (r *Raft) startHeartBeat() {
	heartBeatTicker := time.NewTicker(3 * time.Second)

	for {
		if !r.isLeader {
			log.Println("I ws not the leader")
			return
		}

		select {
		case <-heartBeatTicker.C:
			r.transponder.Write(fmt.Appendf([]byte(""), "%s%s%d", node.HEARTBEAT, SEPERATOR, r.id))
		}
	}
}

func (r *Raft) newTimerEveryMin(wait int) *time.Ticker {
	// ONE_MIN := 60
	// currTime := time.Now()
	// timeLeft := time.Duration(((ONE_MIN-currTime.Second())+wait)%ONE_MIN) * time.Second

	time.Sleep(time.Duration(wait) * time.Second)

	return time.NewTicker(10 * time.Second)
}

func (r *Raft) AddConn(conn *websocket.Conn, id int) {
	r.transponder.AddConn(conn, id)
}
