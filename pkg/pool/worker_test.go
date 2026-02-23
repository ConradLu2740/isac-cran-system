package pool

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestNewWorkerPool(t *testing.T) {
	pool := NewWorkerPool(5, 10)
	if pool == nil {
		t.Fatal("pool should not be nil")
	}
	if pool.WorkerCount() != 5 {
		t.Errorf("expected 5 workers, got %d", pool.WorkerCount())
	}
}

func TestWorkerPoolSubmit(t *testing.T) {
	pool := NewWorkerPool(3, 10)
	pool.Start()
	defer pool.Stop()

	var counter int64

	for i := 0; i < 10; i++ {
		pool.Submit(func() (interface{}, error) {
			atomic.AddInt64(&counter, 1)
			return nil, nil
		})
	}

	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt64(&counter) != 10 {
		t.Errorf("expected 10 tasks completed, got %d", counter)
	}
}

func TestWorkerPoolSubmitAndWait(t *testing.T) {
	pool := NewWorkerPool(2, 5)
	pool.Start()
	defer pool.Stop()

	result, err := pool.SubmitAndWait(func() (interface{}, error) {
		return 42, nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result.(int) != 42 {
		t.Errorf("expected 42, got %v", result)
	}
}

func TestWorkerPoolConcurrent(t *testing.T) {
	pool := NewWorkerPool(10, 100)
	pool.Start()
	defer pool.Stop()

	var counter int64
	tasks := 100

	for i := 0; i < tasks; i++ {
		pool.Submit(func() (interface{}, error) {
			time.Sleep(1 * time.Millisecond)
			atomic.AddInt64(&counter, 1)
			return nil, nil
		})
	}

	time.Sleep(200 * time.Millisecond)

	if atomic.LoadInt64(&counter) != int64(tasks) {
		t.Errorf("expected %d tasks completed, got %d", tasks, counter)
	}
}

func TestWorkerPoolQueueFull(t *testing.T) {
	pool := NewWorkerPool(1, 2)
	pool.Start()
	defer pool.Stop()

	pool.Submit(func() (interface{}, error) {
		time.Sleep(100 * time.Millisecond)
		return nil, nil
	})

	pool.Submit(func() (interface{}, error) {
		time.Sleep(100 * time.Millisecond)
		return nil, nil
	})

	success := pool.Submit(func() (interface{}, error) {
		return nil, nil
	})

	if success {
		t.Error("expected queue to be full")
	}
}
