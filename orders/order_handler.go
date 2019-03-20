package orders

import "github.com/sigtot/sanntid/types"

const numFloors = 4 // TODO: Move this maybe
const topFloor = numFloors - 1
const bottomFloor = 0

type OrderHandler struct {
}

func (oh *OrderHandler) GetPrice(order types.Call) int {
	// TODO: Implement GetPrice
	return 0
}
