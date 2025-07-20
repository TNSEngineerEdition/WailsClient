package simulation

import "sync"

type workerData[I any, O any] struct {
	inputChannel  chan I
	outputChannel chan O
	wg            sync.WaitGroup
}

func (w *workerData[I, O]) reset(bufferSize int) {
	if w.inputChannel != nil {
		close(w.inputChannel)
	}

	w.inputChannel = make(chan I, bufferSize)
	w.outputChannel = make(chan O, bufferSize)
	w.wg = sync.WaitGroup{}
}
