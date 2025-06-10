# dynapool

`dynapool` — легковесный динамический worker pool на Go с возможностью добавлять и удалять воркеров во время выполнения.

## Особенности

- Динамическое управление количеством воркеров
- Потокобезопасный интерфейс
- Буферизированный канал задач
- Корректное завершение работы пула
- Примеры и тесты с покрытием основных сценариев

## Установка

```bash
go get github.com/Sereneum/dynapool
```

## Быстрый старт

```go
package main

import (
	"fmt"
	"dynapool"
	"time"
)

func main() {
	// Создаем пул с буфером 10 задач
	pool := dynapool.NewWorkerPool(10)

	// Добавляем 3 воркера
	for i := 0; i < 3; i++ {
		pool.AddWorker()
	}

	// Отправляем задачи в пул
	for i := 0; i < 10; i++ {
		err := pool.Submit(fmt.Sprintf("job-%d", i))
		if err != nil {
			fmt.Println("Submit error:", err)
		}
	}

	// Даем воркерам время на обработку
	time.Sleep(time.Second)

	// Корректно закрываем пул
	pool.Shutdown()
}
```

## API

```go
func NewWorkerPool(bufferSize int) *WorkerPool
```
Создает новый пул с указанным размером буфера для очереди задач.
```go
func (p *WorkerPool) AddWorker()
```
Добавляет нового воркера в пул.
```go
func (p *WorkerPool) RemoveWorker()
```
Удаляет одного воркера из пула, если есть активные воркеры.
```go
func (p *WorkerPool) Submit(job string) error
```
Отправляет задачу в пул. Возвращает ошибку, если пул завершил работу.
```go
func (p *WorkerPool) Shutdown()
```
Завершение работы пула: закрывает все воркеры, останавливает приём задач.

## Тесты
```go
go test -race
```
В проекте предусмотрены тесты на:

- Базовую работу пула с несколькими воркерами и задачами.

- Отказ отправки задач после вызова Shutdown().

- Динамическое добавление и удаление воркеров в параллельных горутинах.

- Проверку отсутствия гонок при интенсивной нагрузке.

## Пример с профилированием (pprof)
Для запуска примера профилирования запустите:
```go
go run ./cmd/pprof-example
```
Профилирование будет доступно по адресу:
```bash
http://localhost:6060/debug/pprof/
```
Либо используя команды:
```bash
go tool pprof http://localhost:6060/debug/pprof/heap
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

## Структура проекта
```text
dynapool/
├── cmd/
│   └── pprof-example/
│       └── main.go       # Пример с pprof-сервером
├── dynapool.go           # Основной код пула
├── dynapool_test.go      # Тесты
├── example/
│   └── main.go           # Простой пример использования
├── go.mod
└── README.md
```

