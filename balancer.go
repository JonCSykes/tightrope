package tightrope

import (
	"container/heap"
	"fmt"
)

type Balancer struct {
	pool Pool
	done chan *Worker
}

const nRequester = 100
const nWorker = 10

func InitBalancer() *Balancer {
	done := make(chan *Worker, nWorker)

	balancer := &Balancer{make(Pool, 0, nWorker), done}

	for i := 0; i < nWorker; i++ {
		worker := &Worker{work: make(chan Request, nRequester)}
		heap.Push(&balancer.pool, worker)
		go worker.Execute(balancer.done)
	}

	return balancer
}

func (b *Balancer) print() {
	sum := 0
	sumsq := 0

	for _, w := range b.pool {
		fmt.Printf("%d ", w.pending)
		sum += w.pending
		sumsq += w.pending * w.pending
	}

	avg := float64(sum) / float64(len(b.pool))
	variance := float64(sumsq)/float64(len(b.pool)) - avg*avg
	fmt.Printf(" %.2f %.2f\n", avg, variance)
}

func (b *Balancer) balance(req chan Request) {
	for {
		select {
		case request := <-req:
			b.dispatch(request)
		case w := <-b.done:
			b.completed(w)
		}
		b.print()
	}
}

func (b *Balancer) dispatch(req Request) {
	w := heap.Pop(&b.pool).(*Worker)
	w.work <- req
	w.pending++
	heap.Push(&b.pool, w)
}

func (b *Balancer) completed(w *Worker) {
	w.pending--
	heap.Remove(&b.pool, w.index)
	heap.Push(&b.pool, w)
}
