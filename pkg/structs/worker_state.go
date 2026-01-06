package structs

import "sync"

type WorkerState[I any, O any] struct {
	InputChannel  chan I
	OutputChannel chan O
	WaitGroup     sync.WaitGroup
}

func NewWorkerState[I any, O any](bufferSize int) *WorkerState[I, O] {
	return &WorkerState[I, O]{
		InputChannel:  make(chan I, bufferSize),
		OutputChannel: make(chan O, bufferSize),
		WaitGroup:     sync.WaitGroup{},
	}
}

func (w *WorkerState[I, O]) Stop() {
	close(w.InputChannel)
}
