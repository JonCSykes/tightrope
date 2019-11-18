package tightrope

import (
	"container/heap"
	"fmt"
	"sync"
	"time"
)

type Balancer struct {
	pool        Pool
	workerCount int
	// finishedWorkerCount int
	timeout    chan bool
	isTimedOut bool
	done       chan *Worker
	waitGroup  *sync.WaitGroup
}

func InitBalancer(workerCount int, maxWorkBuffer int, execute Execute, wg *sync.WaitGroup) *Balancer {
	done := make(chan *Worker, workerCount)
	timeout := make(chan bool)
	//balancer := &Balancer{make(Pool, 0, workerCount), workerCount, 0, timeout, false, done, wg}
	balancer := &Balancer{make(Pool, 0, workerCount), workerCount, timeout, false, done, wg}

	for i := 0; i < workerCount; i++ {

		worker := &Worker{Work: make(chan Request, maxWorkBuffer), Index: i, Closed: make(chan int)}
		heap.Push(&balancer.pool, worker)

		go func(w *Worker) {
			for {
				select {
				case request := <-w.Work:
					execute(request)
					done <- worker
				case <-w.Closed:
					fmt.Println("Closing after timeout:", w.Index)
					return
				}
			}
		}(worker)
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
	isClosed := make(chan bool)
	go func() {
		for {
			select {
			case request := <-req:
				b.Dispatch(request)
			case w := <-b.done:
				b.Completed(w)
			case <-b.timeout:
				fmt.Println("Timeout :::::::::::::::")
				isClosed <- true
				return
			}

			if printStats {
				b.Print()
			}

			// if b.workerCount == b.finishedWorkerCount {
			// 	fmt.Println("Leaving Balancer - " + strconv.Itoa(len(b.pool)) + ":" + strconv.Itoa(b.finishedWorkerCount))
			// 	return
			// }
		}
	}()
	<-isClosed
	for _, w := range b.pool {
		fmt.Println("Removing worker:", *w)
		w.Closed <- 1
		b.waitGroup.Done()
		fmt.Println("waitGroup is done:", w)
	}
}

func (b *Balancer) Dispatch(req Request) {
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
	//fmt.Println("worker is added back", w.Index)

	// if b.isTimedOut {
	// 	fmt.Println("Closing worker " + strconv.Itoa(index) + " of " + strconv.Itoa(len(b.pool)))
	// 	b.finishedWorkerCount++
	// 	fmt.Println(strconv.Itoa(len(b.pool)) + ":" + strconv.Itoa(b.finishedWorkerCount))

	// 	//w.Closed <- true
	// 	//close(w.Work)
	// 	b.waitGroup.Done()

	// } else {
	// 	fmt.Println("Putting Worker back on heap " + strconv.Itoa(index))
	// 	heap.Push(&b.pool, w)
	// }
}
