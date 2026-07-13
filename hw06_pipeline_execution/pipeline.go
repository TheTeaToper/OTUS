package hw06pipelineexecution

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
	for _, stage := range stages {
		if stage != nil {
			currentIn = stage(currentIn)
		}
	}

	finalOut := make(Bi)
	go func() {
		defer close(finalOut)
		for {
			select {
			case <-done:
				go func() {
					for range currentIn {
					}
				}()
				return
			case currentOut, ok := <-currentIn:
				if !ok {
					return
				}
				select {
				case <-done:
					go func() {
						for range currentIn {
						}
					}()
					return
				case finalOut <- currentOut:
				}
			}
		}
	}()
	return finalOut
}
