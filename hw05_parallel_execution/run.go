package hw05parallelexecution

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if len(tasks) == 0 {
		return nil
	}
	if n <= 0 {
		n = 1
	}
	context, cancelExecution := context.WithCancel(context.Background())
	defer cancelExecution()

	tasksChannel := make(chan Task, len(tasks))
	for _, task := range tasks {
		tasksChannel <- task
	}
	close(tasksChannel)

	var totalErrorsCount int32
	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-context.Done():
					return
				case taskForExecution, ok := <-tasksChannel:
					if !ok {
						return
					}

					err := taskForExecution()
					if err != nil {
						if m > 0 {
							currentErrorsCount := atomic.AddInt32(&totalErrorsCount, 1)
							if currentErrorsCount >= int32(m) {
								cancelExecution()
								return
							}
						}
					}
				}
			}
		}()
	}

	wg.Wait()

	if m > 0 && atomic.LoadInt32(&totalErrorsCount) >= int32(m) {
		return ErrErrorsLimitExceeded
	}
	return nil
}
