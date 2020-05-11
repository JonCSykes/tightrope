package tightrope

import (
	"container/heap"
	"fmt"
	"time"
)

type Balancer struct {
	pool            Pool
	timeout         chan bool
	done            chan *Worker
	shutdown        chan bool
	finishedWorkers chan int
	workerCount     int
}

//func InitBalancer(workerCount int, maxWorkBuffer int, execute Execute) *Balancer {
func InitBalancer(workerCount int, maxWorkBuffer int, execute Execute) *Balancer {
	done := make(chan *Worker, workerCount)
	finishedWorkers := make(chan int, workerCount)
	timeout := make(chan bool)
	shutdown := make(chan bool)
	balancer := &Balancer{make(Pool, 0, workerCount), timeout, done, shutdown, finishedWorkers, workerCount}

	for i := 0; i < workerCount; i++ {

		worker := &Worker{Work: make(chan Request, maxWorkBuffer), Index: i, WorkerID: i}
		heap.Push(&balancer.pool, worker)
		go func(worker *Worker, done chan *Worker) {
			for {
				select {
				case request := <-worker.Work:
					if request.Data == nil {
						balancer.finishedWorkers <- 0
						close(worker.Work)
						return
					}
					execute(request)
					done <- worker
				}
			}

		}(worker, balancer.done)
	}
	return balancer
}

// TimeOut :
func TimeOut(timeoutDuration time.Duration, timeout chan<- bool) {
	time.Sleep(timeoutDuration)
	timeout <- true
}

func (b *Balancer) Print() {
	sum := 0
	sumsq := 0
	totalCompleted := 0

	for _, w := range b.pool {
		fmt.Printf("wid: %d, pnd: %d, cmplt: %d | ", w.Index, w.Pending, w.Complete)

		sum += w.Pending
		sumsq += w.Pending * w.Pending
		totalCompleted += w.Complete
	}

	avg := float64(sum) / float64(len(b.pool))
	variance := float64(sumsq)/float64(len(b.pool)) - avg*avg
	fmt.Printf(" avg: %.2f, var: %.2f, ttl cmplt: %d, ts: %s\n", avg, variance, totalCompleted, time.Now().Format("15:04:05.999999"))

}

func (b *Balancer) Balance(req chan Request, printStats bool, timeoutDuration time.Duration) {
	go TimeOut(timeoutDuration, b.timeout)
	finishedWorkersCount := 0
	go func() {
		for {
			select {
			case request := <-req:
				b.Dispatch(request)
			case w := <-b.done:
				b.Completed(w)
			case <-b.timeout:
				b.StopAllWorkersGraceFully()
			case <-b.finishedWorkers:
				finishedWorkersCount++
			}
			if finishedWorkersCount == b.workerCount {
				b.shutdown <- true
				close(b.shutdown)
				return
			}
			if printStats {
				b.Print()
			}
		}
	}()
	<-b.shutdown
}

func (b *Balancer) StopAllWorkersGraceFully() {
	for i := 0; i < b.pool.Len(); i++ {
		worker := b.pool[i]
		emptyRequest := Request{Data: nil}
		worker.Work <- emptyRequest
	}
}

func (b *Balancer) Dispatch(req Request) {
	defer handlePanic()
	w := heap.Pop(&b.pool).(*Worker)
	w.Work <- req
	w.Pending++
	heap.Push(&b.pool, w)
}

func (b *Balancer) Completed(w *Worker) {
	w.Pending--
	w.Complete++
	heap.Remove(&b.pool, w.Index)
	heap.Push(&b.pool, w)
}

func (b *Balancer) Purge() {

	for b.pool.Len() != 0 {
		w := heap.Pop(&b.pool).(*Worker)
		w.Work <- Request{}
		close(w.Work)
	}
}

func handlePanic() {
	if r := recover(); r != nil {
		fmt.Println(r)
	}
}
