package tightrope

type Request struct {
	Data     interface{}
	Response chan interface{}
}
