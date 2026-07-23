package hw05parallelexecution

import (
	"errors"
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestRun(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("if were errors in first M tasks, than finished not more N+M tasks", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			err := fmt.Errorf("error from task %d", i)
			tasks = append(tasks, func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				atomic.AddInt32(&runTasksCount, 1)
				return err
			})
		}

		workersCount := 10
		maxErrorsCount := 23
		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.LessOrEqual(t, runTasksCount, int32(workersCount+maxErrorsCount), "extra tasks were started")
	})

	t.Run("tasks without errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var sumTime time.Duration

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			sumTime += taskSleep

			tasks = append(tasks, func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		workersCount := 5
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		require.NoError(t, err)

		require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were completed")
		require.LessOrEqual(t, int64(elapsedTime), int64(sumTime/2), "tasks were run sequentially?")
	})
}

func TestRun_IgnoreErrorsWhenMLessOrEqualZero(t *testing.T) {
	var count int32
	tasks := make([]Task, 5)
	for i := 0; i < 5; i++ {
		tasks[i] = func() error {
			atomic.AddInt32(&count, 1)
			return errors.New("some error")
		}
	}

	err := Run(tasks, 2, 0)
	assert.NoError(t, err, "При m <= 0 ошибки должны игнорироваться, а ошибка лимита — не возвращаться")
	assert.Equal(t, int32(5), count, "Должны выполниться все задачи")

	atomic.StoreInt32(&count, 0)
	err = Run(tasks, 2, -1)
	assert.NoError(t, err)
	assert.Equal(t, int32(5), count)
}

func TestRun_MaxTasksExecutedLimit(t *testing.T) {
	n := 3
	m := 2
	var totalExecuted int32
	tasks := make([]Task, 20)
	for i := 0; i < 20; i++ {
		tasks[i] = func() error {
			atomic.AddInt32(&totalExecuted, 1)
			return errors.New("task error")
		}
	}

	err := Run(tasks, n, m)
	assert.ErrorIs(t, err, ErrErrorsLimitExceeded)

	executed := atomic.LoadInt32(&totalExecuted)
	assert.LessOrEqual(t, executed, int32(n+m), "Количество выполненных задач превысило лимит n + m")
}

func TestRun_ConcurrencyWithoutSleep(t *testing.T) {
	n := 4
	tasksCount := 8

	startCh := make(chan struct{})
	runningTasks := int32(0)

	tasks := make([]Task, tasksCount)
	for i := 0; i < tasksCount; i++ {
		tasks[i] = func() error {
			atomic.AddInt32(&runningTasks, 1)
			<-startCh
			return nil
		}
	}

	done := make(chan struct{})
	go func() {
		_ = Run(tasks, n, 1)
		close(done)
	}()

	require.Eventually(t, func() bool {
		return atomic.LoadInt32(&runningTasks) == int32(n)
	}, 2*time.Second, 10*time.Millisecond, "Воркеры должны запуститься параллельно")

	assert.Equal(t, int32(n), atomic.LoadInt32(&runningTasks))

	close(startCh)
	<-done
}
