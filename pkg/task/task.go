package Task

import (
	"fmt"
	"strings"

	p "github.com/Kevin27954/liveness-sim-test/pkg"
)

const SEPERATOR = "≡" // A Hambuger menu looking thing

type Task struct {
	msg    p.MessageEvent
	taskId int
	votes  int
}

type TaskManager struct {
	taskQueue    map[int]int //taskId - votes
	taskInfo     map[int]string
	taskComplete int
	taskStarted  int
	totalNodes   int
	queuedTask   []string
}

func Init() TaskManager {
	return TaskManager{
		taskQueue:    make(map[int]int),
		taskInfo:     make(map[int]string),
		taskComplete: 0,
		taskStarted:  1,
		queuedTask:   make([]string, 0),
	}
}

func (t *TaskManager) NumMsg() int {
	return len(t.queuedTask)
}

func (t *TaskManager) GetQueuedMsg() []string {
	var msgs []string

	for _, msg := range t.queuedTask {
		msgs = append(msgs, msg)
	}

	// Making sure there is nothing left
	t.queuedTask = []string{}

	return msgs
}

func (t *TaskManager) AddQueueTask(event string, msg string) {
	t.queuedTask = append(t.queuedTask, event+SEPERATOR+msg)
}

func (t *TaskManager) AddTask(event string, msg string) int {
	if event == p.ELECTION {
		t.taskQueue[1] = 0
		return 1
	}

	t.taskStarted += 1
	t.taskQueue[t.taskStarted] = 0
	t.taskInfo[t.taskStarted] = event + SEPERATOR + msg
	return t.taskStarted
}

// Returns true if minimum consensus reached. Otherwise false
func (t *TaskManager) VoteOnTask(taskId int, vote bool) bool {
	if taskId <= t.taskComplete {
		return true
	}

	t.taskQueue[taskId] += 1

	if t.taskQueue[taskId] > t.totalNodes/2 {
		t.taskComplete += 1
		delete(t.taskQueue, taskId)
		delete(t.taskInfo, taskId)
		return true
	}

	return false
}

func (t *TaskManager) SetMaxNodes(totalNodes int) {
	t.totalNodes = totalNodes
}

func (t *TaskManager) String() string {
	var sb strings.Builder

	first := true

	sb.WriteString(fmt.Sprintf("  Task Started: %d\n", t.taskStarted-1))
	sb.WriteString(fmt.Sprintf("  Task Completed: %d\n", t.taskComplete))
	sb.WriteString("  Task Queue Items: [")
	for k, v := range t.taskQueue {
		if !first {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("taskId: %d, votes: %d", k, v))
		first = false
	}
	sb.WriteString("]\n")

	return sb.String()
}
