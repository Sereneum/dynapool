package main

import (
	dynopool "dynapool"
	"fmt"
	"sync"
	"time"
)

func main() {
	pool := dynopool.NewWorkerPool(32)

	// Запускаем 5 воркеров
	for i := 0; i < 5; i++ {
		pool.AddWorker()
	}

	var wg sync.WaitGroup

	// Отправляем 20 задач из разных горутин
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				task := fmt.Sprintf("task %d", j)
				if err := pool.Submit(task); err != nil {
					fmt.Println("Submit error:", err)
					return
				}
				time.Sleep(10 * time.Millisecond)
			}
		}()
	}

	wg.Wait()
	pool.Shutdown()

	fmt.Println("All tasks completed, pool shutdown.")
}
