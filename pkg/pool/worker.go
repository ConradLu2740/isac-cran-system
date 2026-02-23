package pool

import (
	"context"
	"sync"
)

type Task func() (interface{}, error)

type Result struct {
	Data  interface{}
	Error error
}

type WorkerPool struct {
	workers   int
	taskQueue chan Task
	results   chan Result
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewWorkerPool(workers int, queueSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		workers:   workers,
		taskQueue: make(chan Task, queueSize),
		results:   make(chan Result, queueSize),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (p *WorkerPool) Start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()
	for {
		select {
		case <-p.ctx.Done():
			return
		case task, ok := <-p.taskQueue:
			if !ok {
				return
			}
			data, err := task()
			p.results <- Result{Data: data, Error: err}
		}
	}
}

func (p *WorkerPool) Submit(task Task) bool {
	select {
	case p.taskQueue <- task:
		return true
	default:
		return false
	}
}

func (p *WorkerPool) SubmitAndWait(task Task) (interface{}, error) {
	resultChan := make(chan Result, 1)
	p.taskQueue <- func() (interface{}, error) {
		data, err := task()
		resultChan <- Result{Data: data, Error: err}
		return data, err
	}
	result := <-resultChan
	return result.Data, result.Error
}

func (p *WorkerPool) Results() <-chan Result {
	return p.results
}

func (p *WorkerPool) Stop() {
	p.cancel()
	close(p.taskQueue)
	p.wg.Wait()
	close(p.results)
}

func (p *WorkerPool) QueueSize() int {
	return len(p.taskQueue)
}

func (p *WorkerPool) WorkerCount() int {
	return p.workers
}
