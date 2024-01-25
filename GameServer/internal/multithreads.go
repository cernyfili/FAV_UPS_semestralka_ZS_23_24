package internal

import (
	"fmt"
	"sync"
)

var numWorkersCounter = 0
var numTasksCounter = 0

var tasks = make(chan Task)
var results = make(chan int)

var wg sync.WaitGroup

var workerChannels = make(map[int]chan Task)

type Task struct {
	ID int
}

func worker(id int) {
	defer wg.Done()
	for task := range tasks {
		fmt.Printf("Worker %d processing task %d\n", id, task.ID)
		// Simulating some task processing
		results <- task.ID * 2 // Sending result back
	}
}

func addWorker(workerID int) error {
	numWorkersCounter++
	_, ok := workerChannels[workerID]
	if ok {
		return fmt.Errorf("worker ID already exists")
	}
	workerChannels[workerID] = make(chan Task)
	wg.Add(1)
	go worker(workerID)

	return nil
}

func sendTasks(task Task, workerID int) error {
	numTasksCounter++
	channel, ok := workerChannels[workerID]
	if !ok {
		return fmt.Errorf("worker ID not found")
	}
	channel <- task

	return nil
}

func closeWorkerChannels() {
	go func() {
		wg.Wait()
		for _, ch := range workerChannels {
			close(ch)
		}
		close(results)
	}()
}

func getResultForWorker(workerID int) (int, error) {
	if numTasksCounter == 0 {
		return -1, fmt.Errorf("no tasks found")
	}
	for result := range results {
		if result%numWorkersCounter == workerID {
			return result, nil
		}
	}
	return -1, fmt.Errorf("no result found for worker") // Worker ID not found
}
