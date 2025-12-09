package simulation

import "sync"

type workerData[I any, O any] struct {
	inputChannel  chan I
	outputChannel chan O
	wg            sync.WaitGroup
}

func newWorkerData[I any, O any](bufferSize int) *workerData[I, O] {
	data := workerData[I, O]{}

	data.inputChannel = make(chan I, bufferSize)
	data.outputChannel = make(chan O, bufferSize)
	data.wg = sync.WaitGroup{}

	return &data
}

func (w *workerData[I, O]) stop() {
	close(w.inputChannel)
}
