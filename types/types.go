package types

type Direction int
type CallType int

// Direction is the direction of the call.
const (
	InvalidDir Direction = iota - 1
	Down
	Up
)

// CallType is the type of the call, i.e a hall call or a cab call.
const (
	Cab CallType = iota
	Hall
)

// Call has a floor and call type, and a direction if it is a hall call or
// an id if it is a cab call (as cab calls can only be delivered by the elevator that received it).
// A call is the sellers responsibility until it is sold and thus transformed to an order.
type Call struct {
	Type       CallType
	Floor      int
	Dir        Direction
	ElevatorID string
}

// Order is a call that has been bought by an elevator and as such is now the buyers responsibility.
type Order struct {
	Call
}

// Bid consists of a call, as well as the bidders id and the bidder's price on the call.
type Bid struct {
	Call       Call
	Price      int
	ElevatorID string
}

// SoldTo is the signal sent by the seller to communicate which bidder wins the bidding round.
type SoldTo struct {
	Bid
}

// Ack is the signal sent by the buyer to acknowledge that the sale is completed and the buyer now has
// the responsibility of delivering the order.
type Ack struct {
	Bid
}
