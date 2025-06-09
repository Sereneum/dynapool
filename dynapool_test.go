package dynopool

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Тест базовой подачи задач и корректного выполнения воркерами
func TestWorkerPoolBasic(t *testing.T) {
	pool := NewWorkerPool(10)
	defer pool.Shutdown()

	// 3 воркера
	for i := 0; i < 3; i++ {
		pool.AddWorker()
	}

	// 20 задач
	for i := 0; i < 20; i++ {
		job := fmt.Sprintf("task %d", i)
		err := pool.Submit(job)
		if err != nil {
			t.Fatalf("Submit failed: %v", err)
		}
	}

	time.Sleep(300 * time.Millisecond)
}

// Проверка, что Submit после Shutdown возвращает ошибку
func TestWorkerPoolShutdown(t *testing.T) {
	pool := NewWorkerPool(1)
	pool.AddWorker()

	err := pool.Submit("job before shutdown")
	if err != nil {
		t.Fatalf("Submit before shutdown should succeed, got error: %v", err)
	}

	pool.Shutdown()

	err = pool.Submit("job after shutdown")
	if err == nil {
		t.Fatalf("Submit after shutdown should fail")
	}
}

// Тест динамического добавления и удаления воркеров
func TestAddRemoveWorkers(t *testing.T) {
	pool := NewWorkerPool(10)
	defer pool.Shutdown()

	// Добавляем и удаляем воркеров параллельно
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			pool.AddWorker()
			time.Sleep(10 * time.Millisecond)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			pool.RemoveWorker()
			time.Sleep(15 * time.Millisecond)
		}
	}()

	wg.Wait()

	// Отправить задачу и проверить, что пул работает
	pool.AddWorker()
	err := pool.Submit("final job")
	if err != nil {
		t.Fatalf("Submit failed after add/remove workers: %v", err)
	}

	time.Sleep(300 * time.Millisecond)
}

// Тест гонок при интенсивной работе
func TestRaceCondition(t *testing.T) {
	pool := NewWorkerPool(100)
	for i := 0; i < 10; i++ {
		pool.AddWorker()
	}

	var wg sync.WaitGroup
	var counter int64

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				taskNum := atomic.AddInt64(&counter, 1)
				job := fmt.Sprintf("goroutine %d task %d", id, taskNum)
				err := pool.Submit(job)
				if err != nil {
					return
				}
			}
		}(i)
	}

	wg.Wait()
	pool.Shutdown()
}
