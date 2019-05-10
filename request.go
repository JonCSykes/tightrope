package tightrope

import (
	"math/rand"
	"time"
)

type Request struct {
	data     int
	response chan float64
}

func CreateRequest(request chan Request) {
	response := make(chan float64)
	// spawn requests indefinitely
	for {
		// wait before next request
		time.Sleep(time.Duration(rand.Int63n(int64(time.Millisecond))))
		request <- Request{int(rand.Int31n(90)), response}
		// read value from RESP channel
		<-response
	}
}
