package tightrope

import (
	"container/heap"
	"fmt"
	"strconv"
	"sync"
	"time"
)

type Balancer struct {
	pool                Pool
	workerCount         int
	finishedWorkerCount int
	timeout             chan bool
	isTimedOut          bool
	done                chan *Worker
	waitGroup           *sync.WaitGroup
}

//func InitBalancer(workerCount int, maxWorkBuffer int, execute Execute) *Balancer {
func InitBalancer(workerCount int, maxWorkBuffer int, execute Execute, wg *sync.WaitGroup) *Balancer {
	done := make(chan *Worker, workerCount)
	timeout := make(chan bool)
	balancer := &Balancer{make(Pool, 0, workerCount), workerCount, 0, timeout, false, done, wg}

	for i := 0; i < workerCount; i++ {

		worker := &Worker{Work: make(chan Request, maxWorkBuffer), Index: i, Closed: make(chan bool)}
		heap.Push(&balancer.pool, worker)
		go execute(worker, balancer.done)

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

	for {
		select {
		case request := <-req:
			b.Dispatch(request)
		case w := <-b.done:
			b.Completed(w)
		case <-b.timeout:
			b.isTimedOut = true
		}

		if printStats {
			b.Print()
		}

		if b.workerCount == b.finishedWorkerCount {
			fmt.Println("Leaving Balancer - " + strconv.Itoa(len(b.pool)) + ":" + strconv.Itoa(b.finishedWorkerCount))
			return
		}
	}
}

func (b *Balancer) Dispatch(req Request) {
	w := heap.Pop(&b.pool).(*Worker)
	w.Work <- req
	w.Pending++
	heap.Push(&b.pool, w)
}

func (b *Balancer) Completed(w *Worker) {
	index := w.Index
	fmt.Println("Completing Work for Worker " + strconv.Itoa(index))
	w.Pending--
	w.Complete++
	heap.Remove(&b.pool, index)

	fmt.Println("Timeout: " + strconv.FormatBool(b.isTimedOut))

	if b.isTimedOut {
		fmt.Println("Closing worker " + strconv.Itoa(index) + " of " + strconv.Itoa(len(b.pool)))
		b.finishedWorkerCount++
		fmt.Println(strconv.Itoa(len(b.pool)) + ":" + strconv.Itoa(b.finishedWorkerCount))

		//w.Closed <- true
		//close(w.Work)
		b.waitGroup.Done()

	} else {
		fmt.Println("Putting Worker back on heap " + strconv.Itoa(index))
		heap.Push(&b.pool, w)
	}
}

func (b *Balancer) Purge(i int) {

	heap.Remove(&b.pool, i)
	b.pool[i].Closed <- true
	close(b.pool[i].Work)
	b.waitGroup.Done()
}
