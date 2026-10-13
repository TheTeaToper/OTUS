package hw06pipelineexecution

import "sync"

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	if len(stages) == 0 {
		return in
	}

	currentIn := in
	stagesOuts := make([]Out, 0, len(stages))
	var wg sync.WaitGroup

	for _, stage := range stages {
		if stage == nil {
			continue
		}

		proxyIn := make(Bi)
		stageOut := stage(proxyIn)
		stagesOuts = append(stagesOuts, stageOut)

		wg.Add(1)
		go proxyStageInput(&wg, currentIn, proxyIn, done)

		currentIn = stageOut
	}

	finalOut := make(Bi)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(finalOut)
		for {
			select {
			case <-done:
				return
			case currentOut, ok := <-currentIn:
				if !ok {
					return
				}
				select {
				case <-done:
					return
				case finalOut <- currentOut:
				}
			}
		}
	}()

	go func() {
		<-done
		for _, ch := range stagesOuts {
			go func(c Out) {
				for range c {
				}
			}(ch)
		}
		go func(c Out) {
			for range c {
			}
		}(currentIn)
	}()

	go func() {
		wg.Wait()
	}()

	return finalOut
}

func proxyStageInput(wg *sync.WaitGroup, src In, dst Bi, done In) {
	defer wg.Done()
	defer close(dst)
	for {
		select {
		case <-done:
			go func(ch In) {
				for range ch {
				}
			}(src)
			return
		case val, ok := <-src:
			if !ok {
				return
			}
			select {
			case <-done:
				go func(ch In) {
					for range ch {
					}
				}(src)
				return
			case dst <- val:
			}
		}
	}
}
