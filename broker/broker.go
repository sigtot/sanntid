package broker

import (
	"github.com/sigtot/sanntid/hotchan"
	"github.com/sigtot/sanntid/order"
	"time"
)

const TTL = 400

type Price struct {
	Price  int
	ElevID string
}

func StartSelling(newOrders chan order.Order) {
	forSale := hotchan.HotChan{}
	forSale.Start()
	defer forSale.Stop()

	for {
		select {
		case val := <-newOrders:
			hcItem := hotchan.Item{Val: val, TTL: TTL * time.Millisecond}
			forSale.In <- hcItem
		case itemForSale := <-forSale.Out:
			prices := getPricesOnNetwork(itemForSale)
			if len(prices) == 0 {
				forSale.In <- itemForSale
			} else {
				lowestPrice := lowestPrice(prices)
				ord, ok := itemForSale.Val.(order.Order)
				if ok {
					announceSale(ord, lowestPrice.ElevID)
				} else {
					panic("bad type conversion")
				}
			}
		}
	}
}

func getPricesOnNetwork(orderItem hotchan.Item) []Price {
	// Try to sell or 5 ms
	return []Price{Price{Price: 1, ElevID: "hoawhidh"}}
}

func announceSale(ord order.Order, elevID string) {

}

func lowestPrice(prices []Price) Price {
	lowestPrice := prices[0]
	for _, price := range prices {
		if price.Price < lowestPrice.Price {
			lowestPrice = price
		}
	}
	return lowestPrice
}
