package main

import (
	"math"
	"math/rand"
	"time"

	"github.com/joncsykes/tightrope"
)

const maxWorkBuffer = 2000
const workerCount = 10

func main() {
	work := make(chan tightrope.Request)

	go CreateWork(work)

	tightrope.InitBalancer(workerCount, maxWorkBuffer, Execute).Balance(work, false, time.Duration(30)*time.Second)
}

func CreateWork(request chan tightrope.Request) {
	response := make(chan interface{})

	for {
		request <- tightrope.Request{int(rand.Int31n(90)), response}
		<-response
	}
}

func Execute(worker *tightrope.Worker, done chan *tightrope.Worker) {

	for {

		request := <-worker.Work
		request.Response <- math.Sin(float64(request.Data.(int)))

		time.Sleep(time.Duration(rand.Int63n(int64(time.Second * 1))))
		done <- worker
	}
}
