package tightrope

import "math"

type Worker struct {
	// heap index
	index int
	// WOK channel
	work chan Request
	// number of pending request this worker is working on
	pending int
}

func (worker *Worker) Execute(done chan *Worker) {
	// worker works indefinitely
	for {
		// extract request from WOK channel
		request := <-worker.work
		// write to RESP channel
		request.response <- math.Sin(float64(request.data))
		// write to DONE channel
		done <- worker
	}
}
