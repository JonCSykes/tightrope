package tightrope

type Worker struct {
	Index    int
	Work     chan Request
	Pending  int
	Complete int
	Closed   chan bool
}

type Execute func(*Worker, chan *Worker)
