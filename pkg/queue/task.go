package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusCompleted TaskStatus = "completed"
	StatusFailed    TaskStatus = "failed"
)

type Task struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	Status    TaskStatus             `json:"status"`
	Result    interface{}            `json:"result,omitempty"`
	Error     string                 `json:"error,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	StartedAt *time.Time             `json:"started_at,omitempty"`
	EndedAt   *time.Time             `json:"ended_at,omitempty"`
}

type TaskHandler func(ctx context.Context, payload map[string]interface{}) (interface{}, error)

type TaskQueue struct {
	tasks    map[string]*Task
	handlers map[string]TaskHandler
	mu       sync.RWMutex
	ch       chan *Task
	workers  int
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewTaskQueue(workers int, bufferSize int) *TaskQueue {
	ctx, cancel := context.WithCancel(context.Background())
	return &TaskQueue{
		tasks:    make(map[string]*Task),
		handlers: make(map[string]TaskHandler),
		ch:       make(chan *Task, bufferSize),
		workers:  workers,
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (q *TaskQueue) RegisterHandler(taskType string, handler TaskHandler) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.handlers[taskType] = handler
}

func (q *TaskQueue) Submit(taskType string, payload map[string]interface{}) string {
	task := &Task{
		ID:        generateID(),
		Type:      taskType,
		Payload:   payload,
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}

	q.mu.Lock()
	q.tasks[task.ID] = task
	q.mu.Unlock()

	q.ch <- task
	return task.ID
}

func (q *TaskQueue) Start() {
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}
}

func (q *TaskQueue) worker(id int) {
	defer q.wg.Done()
	for {
		select {
		case <-q.ctx.Done():
			return
		case task, ok := <-q.ch:
			if !ok {
				return
			}
			q.processTask(task)
		}
	}
}

func (q *TaskQueue) processTask(task *Task) {
	q.mu.Lock()
	task.Status = StatusRunning
	now := time.Now()
	task.StartedAt = &now
	q.mu.Unlock()

	q.mu.RLock()
	handler, exists := q.handlers[task.Type]
	q.mu.RUnlock()

	if !exists {
		q.mu.Lock()
		task.Status = StatusFailed
		task.Error = "handler not found"
		now := time.Now()
		task.EndedAt = &now
		q.mu.Unlock()
		return
	}

	result, err := handler(q.ctx, task.Payload)

	q.mu.Lock()
	defer q.mu.Unlock()

	now = time.Now()
	task.EndedAt = &now

	if err != nil {
		task.Status = StatusFailed
		task.Error = err.Error()
	} else {
		task.Status = StatusCompleted
		task.Result = result
	}
}

func (q *TaskQueue) GetTask(id string) (*Task, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	task, exists := q.tasks[id]
	return task, exists
}

func (q *TaskQueue) ListTasks(status TaskStatus) []*Task {
	q.mu.RLock()
	defer q.mu.RUnlock()

	var tasks []*Task
	for _, task := range q.tasks {
		if status == "" || task.Status == status {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

func (q *TaskQueue) Stop() {
	q.cancel()
	close(q.ch)
	q.wg.Wait()
}

func (q *TaskQueue) Stats() map[string]int {
	q.mu.RLock()
	defer q.mu.RUnlock()

	stats := map[string]int{
		"total":     len(q.tasks),
		"pending":   0,
		"running":   0,
		"completed": 0,
		"failed":    0,
	}

	for _, task := range q.tasks {
		switch task.Status {
		case StatusPending:
			stats["pending"]++
		case StatusRunning:
			stats["running"]++
		case StatusCompleted:
			stats["completed"]++
		case StatusFailed:
			stats["failed"]++
		}
	}

	return stats
}

func (t *Task) ToJSON() string {
	data, _ := json.Marshal(t)
	return string(data)
}

func generateID() string {
	return fmt.Sprintf("%d%06d", time.Now().UnixNano(), rand.Intn(1000000))
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
