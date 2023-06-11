package main

type EventType byte

const (
	_                     = iota
	EventDelete EventType = iota
	EventPut
)

type Event struct {
	// unique ID
	Sequence  uint64
	EventType EventType
	// The key affected by this transaction
	Key string
	// The value of a PUT the transaction
	Value string
}
