package types

type Direction int
type CallType int

// Direction is the direction of the call.
const (
	InvalidDir Direction = iota - 1
	Down
	Up
)

// CallType is the type of the call.
const (
	Cab CallType = iota
	Hall
)

// Call
type Call struct {
	Type       CallType
	Floor      int
	Dir        Direction
	ElevatorID string
}

// Order
type Order struct {
	Call
}

// Bid
type Bid struct {
	Call       Call
	Price      int
	ElevatorID string
}

// SoldTo
type SoldTo struct {
	Bid
}

// Ack
type Ack struct {
	Bid
}
