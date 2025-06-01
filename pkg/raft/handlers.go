package raft

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	p "github.com/Kevin27954/liveness-sim-test/pkg"
)

func (r *Raft) handleEvent(message p.MessageEvent) error {

	switch message.Event {
	case p.VOTE_YES:
		r.handleVoteYes(message.From, message.Contents)
	case p.VOTE_NO:
		r.handleVoteNo(message.From, message.Contents)

	case p.SYNC_REQ_INIT:
		r.handleSyncReqInit(message.From, message.Contents)
	case p.SYNC_REQ_ASK:
		r.handleSyncReqAsk(message.From, message.Contents)
	case p.SYNC_REQ_HAS, p.SYNC_REQ_NO_HAS:
		cpy := r.syncMap[message.From]
		switch message.Event {
		case p.SYNC_REQ_HAS:
			cpy.Low = cpy.Mid + 1
			cpy.Mid = (cpy.High + cpy.Low) / 2

		case p.SYNC_REQ_NO_HAS:
			cpy.High = cpy.Mid - 1
			cpy.Mid = (cpy.High + cpy.Low) / 2
		}
		r.syncMap[message.From] = cpy
		r.handleSyncRes(message.From, message.Contents)
	case p.SYNC_REQ_COMMIT:
		r.handleSyncCommit(message.From, message.Contents)

	case p.ELECTION:
		r.handleElection(message.From, message.Contents)
	case p.APPEND_ENTRIES:
		r.handleAppendEntries(message.From, message.Contents)

		// It appends the entry but will also immedaitely delete it if
		// it is not successful in the other nodes. Aka if no consensus was fully reached
		// it will just delete the log

		// Figure out how to make it empty and check those?

	case p.NEW_OP:
		r.handleNewOp(message.From, message.Contents)

	}

	return nil
}

func (r *Raft) handleNewOp(_ int, data string) {
	if !r.isLeader {
		return
	}

	dataSplit := strings.Split(data, SEPERATOR)
	operation := dataSplit[0]
	msg := strings.Join(dataSplit[1:], SEPERATOR)

	err := r.Db.AddOperation(operation, r.term, msg)
	if err != nil {
		log.Println("Unable to add operation: ", err)
	}

	taskId := r.task.AddTask(operation, msg)

	r.transponder.Write(r.transponder.CreateMsg(p.APPEND_ENTRIES, r.id, taskId, data))
}

func (r *Raft) handleSyncCommit(_ int, data string) {
	var opsArr []p.Operation

	opsStr := strings.Split(data, SEPERATOR)
	for _, ops := range opsStr {
		if ops == "" {
			continue
		}

		ops, err := p.ParseLog(ops)
		if err != nil {
			// DO I SKIP THIS LOG?
			log.Println("Unable to parse OPS, ", err)
			continue
		}

		opsArr = append(opsArr, ops)
	}

	if len(opsArr) == 0 {
		log.Println(r.name, "Commit Arr was Empty")
		return
	}
	_, err := r.Db.CommitLogs(opsArr)
	if err != nil {
		log.Println(r.name, "error commiting logs")
	}
}

func (r *Raft) handleSyncRes(from int, _ string) {
	cpy := r.syncMap[from]

	if cpy.Low <= cpy.High {

		nextLog, err := r.Db.GetLogByID(cpy.Mid)
		if err != nil {
			log.Println(r.name, "Unable to get Log")
		}
		// r.transponder.WriteTo(from, fmt.Appendf([]byte(""), "%s%s%d%s%s", p.SYNC_REQ_ASK, SEPERATOR, r.id, SEPERATOR, nextLog.String()))
		r.transponder.WriteTo(from, r.transponder.CreateMsg(p.SYNC_REQ_ASK, r.id, nextLog.String()))
	} else {
		opsArr, err := r.Db.GetLogsById(cpy.High)
		if err != nil {
			log.Println(r.name, "Unable to get Log")
		}

		// msg := fmt.Appendf([]byte(""), "%s%s%d%s", p.SYNC_REQ_COMMIT, SEPERATOR, r.id, SEPERATOR)
		msg := r.transponder.CreateMsg(p.SYNC_REQ_COMMIT, r.id, SEPERATOR)
		for _, ops := range opsArr {
			msg = fmt.Appendf(msg, "%s%s", ops, SEPERATOR)
		}

		r.transponder.WriteTo(from, msg)
	}

}

func (r *Raft) handleSyncReqAsk(from int, data string) {
	ops, err := p.ParseLog(data)
	if err != nil {
		log.Println(r.name, "Unable to parse data")
		return
	}

	has, err := r.Db.HasLog(ops)
	if err != nil {
		log.Println(r.name, "Unable to check for log")
	}

	if has {
		r.transponder.WriteTo(from, r.transponder.CreateMsg(p.SYNC_REQ_HAS, r.id))
	} else {
		r.transponder.WriteTo(from, r.transponder.CreateMsg(p.SYNC_REQ_NO_HAS, r.id))
	}
}

func (r *Raft) handleSyncReqInit(from int, _ string) {
	low := 1
	high, err := r.Db.GetNumLogs()
	if err != nil {
		// If this breaks, the whole thing doesn't work
		log.Fatal(r.name, "DB Query err: ", err)
	}

	r.syncMap[from] = BinarySearch{Low: low, Mid: (low + high) / 2, High: high}

	if r.syncMap[from].Mid == 0 {
		log.Println(r.name, "There is nothing to sync")
		return
	}

	nextLog, err := r.Db.GetLogByID(r.syncMap[from].Mid)
	if err != nil {
		log.Println(r.name, "Unable to get Log, ", err)
		return
	}

	r.transponder.WriteTo(from, r.transponder.CreateMsg(p.NEW_OP, p.SYNC_REQ_ASK, r.id, nextLog.String()))
}

func (r *Raft) handleAppendEntries(from int, data string) {
	// If it is empty then it is heartbeat
	if len(data) == 0 {
		r.heartBeatTicker.Reset(10 * time.Second)

		if r.task.NumMsg() > 0 {
			// Send the data that was receieved over to leader.
			msgs := r.task.GetQueuedMsg()

			// create the msg and send it over.
			// Honestly I don't feel like changing thingstoo much now. I'll just take this.
			for _, msg := range msgs {
				r.transponder.WriteTo(from, r.transponder.CreateMsg(p.NEW_OP, msg))
			}

		}

		return
	}

	// Check if I can store the log or operation
	// For now I would just say YES always
	dataSplit := strings.Split(data, SEPERATOR)

	err := r.Db.AddOperation(dataSplit[1], r.term, strings.Join(dataSplit[2:], SEPERATOR))
	if err != nil {
		log.Println("Unable to add operation: ", err)
	}

	taskId := dataSplit[0]
	r.transponder.WriteTo(from, r.transponder.CreateMsg(p.VOTE_YES, r.id, taskId))

	// Else I want to append data upon receiving messages and data

	// I want to send any messages that were nto able to be send before
	// because there wasn't a leader elected yet.
}

func (r *Raft) handleVoteYes(from int, data string) {
	log.Println(from, " voted yes")

	taskId, err := strconv.Atoi(data)
	if err != nil {
		log.Fatal("Unable to parse data")
	}

	if r.task.VoteOnTask(taskId, true) {
		if taskId == 1 {
			r.isLeader = true
			log.Println(r.name, "I ran: isleader:", r.isLeader)
			go r.startHeartBeat(from)
		}
	}
}

func (r *Raft) handleVoteNo(from int, data string) {
	log.Println(from, " voted yes")

	taskId, err := strconv.Atoi(data)
	if err != nil {
		log.Fatal("Unable to parse data")
	}

	if r.task.VoteOnTask(taskId, false) {
		if taskId == 1 {
			r.isLeader = true
			log.Println(r.name, "I ran: isleader:", r.isLeader)
			go r.startHeartBeat(from) // needs a new way to start heartbeat
		}
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

	log.Println(r.name, " handleElection:", r.isLeader)

	if !r.hasVoted && !r.isLeader {
		r.transponder.WriteTo(from, r.transponder.CreateMsg(p.VOTE_YES, r.id, taskId))
		r.hasVoted = true
	}
}
