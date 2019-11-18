package tightrope

type Worker struct {
	Index    int
	Work     chan Request
	Pending  int
	Complete int
	Closed   chan int
}

type Execute func(request Request)
