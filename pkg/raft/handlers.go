package raft

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	n "github.com/Kevin27954/liveness-sim-test/cmd/node"
	p "github.com/Kevin27954/liveness-sim-test/pkg"
)

func (r *Raft) handleEvent(message p.MessageEvent) error {

	switch message.Event {
	case n.ELECTION:
		r.handleElection(message.From, message.Contents)
	case n.VOTE_YES:
		r.handleVoteYes(message.From, message.Contents)
	case n.VOTE_NO:
	case n.HEARTBEAT:
		go func() { r.heartBeat <- 1 }()

	}

	return nil
}

func (r *Raft) handleTicker(tick p.TickerEvent) error {
	return nil
}

func (r *Raft) handleVoteYes(from int, data string) {
	log.Println(from, " voted yes")

	taskId, err := strconv.Atoi(data)
	if err != nil {
		log.Fatal("Unable to parse data")
	}

	if r.task.VoteOnTask(taskId) {
		r.isLeader = true
		go r.startHeartBeat()
	}
}

func (r *Raft) handleElection(from int, data string) {
	splitData := strings.Split(data, SEPERATOR)

	taskId, err := strconv.Atoi(splitData[0])
	if err != nil {
		log.Fatal("Unable to parse data")
	}
	newTerm, err := strconv.Atoi(splitData[1])
	if err != nil {
		log.Fatal("Unable to parse data")
	}

	if newTerm > r.term {
		r.isLeader = false
		r.hasVoted = false
		r.term = newTerm
	}

	if !r.hasVoted && !r.isLeader {
		r.transponder.WriteTo(from, fmt.Appendf([]byte(""), "%s%s%d%s%d", n.VOTE_YES, SEPERATOR, r.id, SEPERATOR, taskId))
		r.hasVoted = true
	}

}
