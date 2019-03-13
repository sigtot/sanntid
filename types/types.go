package types

type Direction int
type CallType int

const (
	InvalidDir Direction = iota - 1
	Down
	Up
)

const (
	Cab CallType = iota
	Hall
)

type Call struct {
	Type       CallType
	Floor      int
	Dir        Direction
	ElevatorID string
}

type Bid struct {
	Call       Call
	Price      int
	ElevatorID string
}

type SoldTo struct {
	Bid
}

type Ack struct {
	Bid
}
