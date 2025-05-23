package Task

import (
	"fmt"
	"strings"

	p "github.com/Kevin27954/liveness-sim-test/pkg"
)

type Task struct {
	msg    p.MessageEvent
	taskId int
	votes  int
}

type TaskManager struct {
	taskQueue    map[int]int //taskId - votes
	taskComplete int
	taskStarted  int
	totalNodes   int
}

func Init() TaskManager {
	return TaskManager{
		taskQueue:    make(map[int]int),
		taskComplete: 0,
		taskStarted:  1,
	}
}

func (t *TaskManager) AddTask(event string) int {
	if event == p.ELECTION {
		t.taskQueue[1] = 0
		return 1
	}

	t.taskStarted += 1
	t.taskQueue[t.taskStarted] = 0
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
