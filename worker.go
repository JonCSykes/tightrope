package tightrope

type Worker struct {
	Index           int
	WorkerID        int
	Work            chan Request
	Pending         int
	Complete        int
}

type Execute func(Request)
