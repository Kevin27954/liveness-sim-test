package Task

import (
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
}

func Init() TaskManager {
	return TaskManager{
		taskQueue:    make(map[int]int),
		taskComplete: 0,
		taskStarted:  0,
	}
}

func (t *TaskManager) AddTask(msg p.MessageEvent) int {
	t.taskStarted += 1
	// t.taskQueue = append(t.taskQueue, Task{msg: msg, taskId: t.taskStarted, votes: 0})
	t.taskQueue[t.taskStarted] = 0
	return t.taskStarted
}

// Returns true if minimum consensus reached. Otherwise false
func (t *TaskManager) VoteOnTask(taskId int) bool {
	t.taskQueue[taskId] += 1

	if t.taskQueue[taskId] >= 0 {
		t.taskComplete += 1
		delete(t.taskQueue, taskId)
		return true
	}

	return false
}
