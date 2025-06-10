package main

import (
	"dynapool"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"runtime/debug"
	"time"
)

func printMem(label string) uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("%s: Alloc = %v KB\n", label, m.Alloc/1024)
	return m.Alloc
}

func main() {
	/*
		Быстрый доступ к командам
		> go tool pprof http://localhost:6060/debug/pprof/heap
		> go tool pprof http://localhost:6060/debug/pprof/goroutine
	*/

	// Запускаем pprof сервер
	go func() {
		log.Println("pprof server running on http://localhost:6060/debug/pprof/")
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// Создаём пул с буфером 10
	pool := dynapool.NewWorkerPool(10)

	// Добавляем 5 воркеров
	for i := 0; i < 5; i++ {
		pool.AddWorker()
	}

	// Отправляем 20 задач
	for i := 0; i < 20; i++ {
		err := pool.Submit(fmt.Sprintf("job-%d", i))
		if err != nil {
			fmt.Println("Submit error:", err)
		}
	}

	// Ждём немного, чтобы воркеры обработали задачи
	time.Sleep(2 * time.Second)

	// Убираем 3 воркера
	for i := 0; i < 3; i++ {
		pool.RemoveWorker()
	}

	// Ждём ещё
	time.Sleep(1 * time.Second)

	// Завершаем пул
	pool.Shutdown()

	// После Shutdown выводим количество горутин
	fmt.Printf("Number of goroutines after Shutdown: %d\n", runtime.NumGoroutine())

	// Заставляем сборщик мусора сработать
	runtime.GC()

	// для жёсткого освобождения памяти
	debug.FreeOSMemory()

	fmt.Printf("Number of goroutines after GC: %d\n", runtime.NumGoroutine())

	// Задержка для ручного подключения к pprof, если надо
	fmt.Println("Press Ctrl+C to exit...")
	select {}
}
