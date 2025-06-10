package dynapool

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

// WorkerPool управляет пулом воркеров
type WorkerPool struct {
	jobs         chan string
	addWorker    chan struct{}
	removeWorker chan struct{}

	mu         sync.Mutex
	wg         sync.WaitGroup
	isShutdown atomic.Bool

	ctx    context.Context
	cancel context.CancelFunc

	nextID  int
	workers []*worker
}

// Worker представляет отдельного обработчика
type worker struct {
	id   int
	stop chan struct{}

	stopOnce sync.Once
}

func (w *worker) stopWorker() {
	w.stopOnce.Do(func() {
		close(w.stop)
	})
}

// NewWorkerPool создает новый пул воркеров
func NewWorkerPool(bufferSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPool{
		jobs:         make(chan string, bufferSize),
		addWorker:    make(chan struct{}, 1),
		removeWorker: make(chan struct{}, 1),
		ctx:          ctx,
		cancel:       cancel,
	}

	go pool.runManager()
	return pool
}

// Submit добавляет задачу в пул
func (p *WorkerPool) Submit(job string) error {
	if p.isShutdown.Load() {
		return errors.New("worker pool is shut down")
	}

	select {
	case p.jobs <- job:
		return nil
	case <-p.ctx.Done():
		return errors.New("worker pool is shut down")
	}
}

// Shutdown останавливает весь пул
func (p *WorkerPool) Shutdown() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.isShutdown.CompareAndSwap(false, true) {
		return
	}

	p.cancel()
	for _, w := range p.workers {
		w.stopWorker()
	}
	p.workers = nil
	close(p.jobs)
	p.wg.Wait()
}

// AddWorker добавляет нового воркера
func (p *WorkerPool) AddWorker() {
	if p.isShutdown.Load() {
		return
	}

	select {
	case p.addWorker <- struct{}{}:
	case <-p.ctx.Done():
	}
}

func (p *WorkerPool) RemoveWorker() {
	if p.isShutdown.Load() {
		return
	}

	select {
	case p.removeWorker <- struct{}{}:
	case <-p.ctx.Done():
	}
}

// runManager управляет жизненным циклом воркеров
func (p *WorkerPool) runManager() {
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-p.addWorker:
			p.startWorker()
		case <-p.removeWorker:
			p.stopWorker()
		}
	}
}

func (p *WorkerPool) startWorker() {
	p.mu.Lock()
	defer p.mu.Unlock()

	id := p.nextID
	p.nextID++

	w := &worker{id: id, stop: make(chan struct{})}
	p.workers = append(p.workers, w)

	p.wg.Add(1)
	go p.workerLoop(w)
}

func (p *WorkerPool) stopWorker() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.workers) == 0 {
		return
	}

	lastIdx := len(p.workers) - 1
	w := p.workers[lastIdx]
	p.workers = p.workers[:lastIdx]
	w.stopWorker()
}

func (p *WorkerPool) workerLoop(w *worker) {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-w.stop:
			return
		case job, ok := <-p.jobs:
			if !ok {
				return
			}
			fmt.Printf("Worker %d:\t%s\n", w.id, job)
		}
	}
}
