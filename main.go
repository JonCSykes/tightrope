package main

func main() {
	work := make(chan Request)
	for i := 0; i < nRequester; i++ {
		go CreateRequest(work)
	}
	InitBalancer().balance(work)
}
