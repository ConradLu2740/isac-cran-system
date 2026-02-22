package queue

import (
	"context"
	"testing"
	"time"
)

func TestNewTaskQueue(t *testing.T) {
	q := NewTaskQueue(5, 10)
	if q == nil {
		t.Fatal("queue should not be nil")
	}
}

func TestTaskQueueSubmit(t *testing.T) {
	q := NewTaskQueue(2, 10)
	q.RegisterHandler("test", func(ctx context.Context, payload map[string]interface{}) (interface{}, error) {
		return payload["value"], nil
	})
	q.Start()
	defer q.Stop()

	taskID := q.Submit("test", map[string]interface{}{"value": 42})
	if taskID == "" {
		t.Error("task ID should not be empty")
	}

	time.Sleep(100 * time.Millisecond)

	task, exists := q.GetTask(taskID)
	if !exists {
		t.Fatal("task should exist")
	}

	if task.Status != StatusCompleted {
		t.Errorf("expected status completed, got %s", task.Status)
	}

	if task.Result.(int) != 42 {
		t.Errorf("expected result 42, got %v", task.Result)
	}
}

func TestTaskQueueHandlerNotFound(t *testing.T) {
	q := NewTaskQueue(1, 5)
	q.Start()
	defer q.Stop()

	taskID := q.Submit("unknown", nil)

	time.Sleep(100 * time.Millisecond)

	task, _ := q.GetTask(taskID)
	if task.Status != StatusFailed {
		t.Errorf("expected status failed, got %s", task.Status)
	}
}

func TestTaskQueueStats(t *testing.T) {
	q := NewTaskQueue(2, 10)
	q.RegisterHandler("test", func(ctx context.Context, payload map[string]interface{}) (interface{}, error) {
		time.Sleep(10 * time.Millisecond)
		return nil, nil
	})
	q.Start()
	defer q.Stop()

	for i := 0; i < 5; i++ {
		q.Submit("test", map[string]interface{}{"id": i})
	}

	time.Sleep(300 * time.Millisecond)

	stats := q.Stats()
	if stats["completed"] != 5 {
		t.Errorf("expected 5 completed tasks, got %d", stats["completed"])
	}
}

func TestTaskQueueMultipleTasks(t *testing.T) {
	q := NewTaskQueue(3, 20)
	q.RegisterHandler("compute", func(ctx context.Context, payload map[string]interface{}) (interface{}, error) {
		val := payload["input"].(int)
		return val * 2, nil
	})
	q.Start()
	defer q.Stop()

	type taskInfo struct {
		id    string
		input int
	}

	tasks := make([]taskInfo, 10)
	for i := 0; i < 10; i++ {
		tasks[i] = taskInfo{
			id:    q.Submit("compute", map[string]interface{}{"input": i}),
			input: i,
		}
	}

	time.Sleep(300 * time.Millisecond)

	for _, ti := range tasks {
		task, exists := q.GetTask(ti.id)
		if !exists {
			t.Errorf("task %s should exist", ti.id)
			continue
		}

		if task.Status != StatusCompleted {
			t.Errorf("task with input %d should be completed, got %s", ti.input, task.Status)
			continue
		}

		expected := ti.input * 2
		if task.Result.(int) != expected {
			t.Errorf("task with input %d: expected %d, got %v", ti.input, expected, task.Result)
		}
	}
}

func TestTaskToJSON(t *testing.T) {
	task := &Task{
		ID:        "test-123",
		Type:      "test",
		Status:    StatusCompleted,
		CreatedAt: time.Now(),
	}

	jsonStr := task.ToJSON()
	if jsonStr == "" {
		t.Error("JSON should not be empty")
	}
}
