package simulation

import "sync"

type workerData[I any, O any] struct {
	inputChannel  chan I
	outputChannel chan O
	wg            sync.WaitGroup
}

func (w *workerData[I, O]) reset(outputBufferSize int) {
	if w.inputChannel != nil {
		close(w.inputChannel)
	}

	w.inputChannel = make(chan I)
	w.outputChannel = make(chan O, outputBufferSize)
	w.wg = sync.WaitGroup{}
}
