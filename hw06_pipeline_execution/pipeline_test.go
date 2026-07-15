package hw06pipelineexecution

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	sleepPerStage = time.Millisecond * 100
	fault         = sleepPerStage / 2
)

func TestPipeline(t *testing.T) {
	// Stage generator
	g := func(_ string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			return out
		}
	}

	stages := []Stage{
		g("Dummy", func(v interface{}) interface{} { return v }),
		g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		g("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }),
	}

	t.Run("simple case", func(t *testing.T) {
		in := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, nil, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Equal(t, []string{"102", "104", "106", "108", "110"}, result)
		require.Less(t,
			int64(elapsed),
			// ~0.8s for processing 5 values in 4 stages (100ms every) concurrently
			int64(sleepPerStage)*int64(len(stages)+len(data)-1)+int64(fault))
	})

	t.Run("done case", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		// Abort after 200ms
		abortDur := sleepPerStage * 2
		go func() {
			<-time.After(abortDur)
			close(done)
		}()

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Len(t, result, 0)
		require.Less(t, int64(elapsed), int64(abortDur)+int64(fault))
	})
}

func TestAllStageStop(t *testing.T) {
	wg := sync.WaitGroup{}
	// Stage generator
	g := func(_ string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			return out
		}
	}

	stages := []Stage{
		g("Dummy", func(v interface{}) interface{} { return v }),
		g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		g("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }),
	}

	t.Run("done case", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		// Abort after 200ms
		abortDur := sleepPerStage * 2
		go func() {
			<-time.After(abortDur)
			close(done)
		}()

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		wg.Wait()

		require.Len(t, result, 0)
	})
}

func passthroughStage(in In) Out {
	out := make(Bi)
	go func() {
		defer close(out)
		for val := range in {
			out <- val
		}
	}()
	return out
}

func TestBoundaryConditions(t *testing.T) {
	t.Run("Empty stages slice", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)

		// Должен просто вернуть исходный канал
		out := ExecutePipeline(in, done)

		if out != in {
			t.Error("функция должна вернуть исходный канал 'in'")
		}
	})

	t.Run("Contains nil stages", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)

		out := ExecutePipeline(in, done, passthroughStage, nil, passthroughStage)

		go func() {
			in <- "test"
			close(in)
		}()

		select {
		case val, ok := <-out:
			if !ok || val != "test" {
				t.Errorf("пайплайн поврежден nil-стадией. Получено: %v", val)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Таймаут: пайплайн завис из-за nil-стадии")
		}
	})

	// Сценарий 3: Входной канал закрывается сразу (нет данных)
	t.Run("Immediate close of input channel", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)

		out := ExecutePipeline(in, done, passthroughStage, passthroughStage)

		close(in)

		select {
		case _, ok := <-out:
			if ok {
				t.Error("Финальный канал должен закрыться пустым")
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Таймаут: пайплайн завис при мгновенном закрытии входа")
		}
	})
}

func TestExecutePipeline_TrueCpuStop(t *testing.T) {
	var wg sync.WaitGroup

	infiniteCpuStage := func(in In) Out {
		out := make(Bi)
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer close(out)

			for val := range in {
				res := val.(int)
				for i := 0; i < 1000; i++ {
					res += i
				}
				out <- res
			}
		}()
		return out
	}

	in := make(Bi)
	done := make(Bi)

	pipelineOut := ExecutePipeline(in, done, infiniteCpuStage)

	go func() {
		for i := 0; ; i++ {
			select {
			case in <- i:
			case <-done:
				close(in)
				return
			}
		}
	}()

	time.Sleep(50 * time.Millisecond)

	close(done)

	for range pipelineOut {
	}

	doneChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneChan)
	}()

	select {
	case <-doneChan:
		t.Log("Успех: Стадия мгновенно прекратила вычисления после сигнала done")
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Провал: Стадия продолжает нагружать CPU после done (утечка процессора)!")
	}
}
