package orderbook

import (
	"fmt"
	"sort"
	"time"
)

type Match struct {
	Ask        *Order
	Bid        *Order
	SizeFilled float64
	Price      float64
}

type Order struct {
	Size      float64
	Bid       bool
	Limit     *Limit
	Timestamp int64
}

type Orders []*Order

func (o Orders) Len() int           { return len(o) }
func (o Orders) Less(i, j int) bool { return o[i].Timestamp < o[j].Timestamp }
func (o Orders) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }

func NewOrder(bid bool, size float64) *Order {
	return &Order{
		Size:      size,
		Bid:       bid,
		Timestamp: time.Now().UnixNano(),
	}
}

func (o *Order) String() string {
	return fmt.Sprintf("Order(Size: %f)", o.Size)
}

func (o *Order) isEmpty() bool {
	return o.Size == 0.0
}

type Limit struct {
	Price       float64
	Orders      Orders
	TotalVolume float64
}

func NewLimit(price float64) *Limit {
	return &Limit{
		Price:  price,
		Orders: Orders{},
	}
}

func (l *Limit) String() string {
	return fmt.Sprintf("Limit(Price: %f, TotalVolume: %f)", l.Price, l.TotalVolume)
}

func (l *Limit) addOrder(o *Order) {
	o.Limit = l
	l.Orders = append(l.Orders, o)
	l.TotalVolume += o.Size

	sort.Sort(l.Orders)
}

func (l *Limit) deleteOrder(o *Order) {
	for i, order := range l.Orders {
		if order == o {
			l.Orders = append(l.Orders[:i], l.Orders[i+1:]...)
			l.TotalVolume -= o.Size
			break
		}
	}
	sort.Sort(l.Orders)

	o.Limit = nil
}

func (l *Limit) isEmpty() bool {
	return len(l.Orders) == 0 || l.TotalVolume == 0.0
}

func (l *Limit) fill(o *Order) []Match {
	matches := []Match{}

	orderToDelete := []*Order{}

	for _, order := range l.Orders {

		match := l.fillOrder(o, order)
		matches = append(matches, match)

		if order.isEmpty() {
			orderToDelete = append(orderToDelete, order)
		}
		if o.isEmpty() {
			break
		}
	}

	// delete empty orders
	for _, order := range orderToDelete {
		l.deleteOrder(order)
	}

	return matches

}

func (l *Limit) fillOrder(a, b *Order) Match {
	var (
		ask        *Order
		bid        *Order
		sizeFilled float64
	)

	if a.Bid {
		bid = a
		ask = b
	} else {
		ask = a
		bid = b
	}

	if a.Size > b.Size {
		sizeFilled = b.Size
		a.Size -= sizeFilled
		b.Size = 0.0
	} else {
		sizeFilled = a.Size
		b.Size -= sizeFilled
		a.Size = 0.0
	}

	l.TotalVolume -= sizeFilled

	return Match{
		Ask:        ask,
		Bid:        bid,
		SizeFilled: sizeFilled,
		Price:      l.Price,
	}

}

type Limits []*Limit

type ByBestAsk struct{ Limits }

func (a ByBestAsk) Len() int           { return len(a.Limits) }
func (a ByBestAsk) Less(i, j int) bool { return a.Limits[i].Price < a.Limits[j].Price }
func (a ByBestAsk) Swap(i, j int)      { a.Limits[i], a.Limits[j] = a.Limits[j], a.Limits[i] }

type ByBestBid struct{ Limits }

func (b ByBestBid) Len() int           { return len(b.Limits) }
func (b ByBestBid) Less(i, j int) bool { return b.Limits[i].Price > b.Limits[j].Price }
func (b ByBestBid) Swap(i, j int)      { b.Limits[i], b.Limits[j] = b.Limits[j], b.Limits[i] }

type OrderBook struct {
	bids []*Limit
	asks []*Limit

	BidLimits map[float64]*Limit
	AskLimits map[float64]*Limit
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		bids:      []*Limit{},
		asks:      []*Limit{},
		BidLimits: make(map[float64]*Limit),
		AskLimits: make(map[float64]*Limit),
	}
}

func (ob *OrderBook) PlaceMarketOrder(o *Order) []Match {
	matches := []Match{}
	limitOrders := Limits{}

	var limitToDelete []*Limit

	// fill bid orders with ask limit orders and vice versa

	if o.Bid {
		if ob.AskTotalVolume() < o.Size {
			panic(fmt.Errorf("not enough volume [size: %f] to fill bid order [ask total volume: %f]", o.Size, ob.AskTotalVolume()))
		}
		limitOrders = ob.Asks()
	} else {
		if ob.BidTotalVolume() < o.Size {
			panic(fmt.Errorf("not enough volume [size: %f] to fill ask order [bid total volume: %f]", o.Size, ob.BidTotalVolume()))
		}
		limitOrders = ob.Bids()
	}

	for _, limit := range limitOrders {
		filledMatches := limit.fill(o)
		matches = append(matches, filledMatches...)

		if limit.isEmpty() {
			limitToDelete = append(limitToDelete, limit)
		}

		if o.isEmpty() {
			break
		}
	}

	// if order is bid, clear ask limit, otherwise clear bid limit

	for _, limit := range limitToDelete {
		ob.clearLimit(!o.Bid, limit)
	}

	return matches

}

func (ob *OrderBook) PlaceLimitOrder(price float64, o *Order) {

	var limit *Limit
	if o.Bid {
		limit = ob.BidLimits[price]
		if limit == nil {
			limit = NewLimit(price)
			ob.bids = append(ob.bids, limit)
			ob.BidLimits[price] = limit
		}
	} else {
		limit = ob.AskLimits[price]
		if limit == nil {
			limit = NewLimit(price)
			ob.asks = append(ob.asks, limit)
			ob.AskLimits[price] = limit
		}
	}

	limit.addOrder(o)
}

func (ob *OrderBook) BidTotalVolume() float64 {
	totalVolume := 0.0
	for _, limit := range ob.bids {
		totalVolume += limit.TotalVolume
	}
	return totalVolume
}

func (ob *OrderBook) AskTotalVolume() float64 {
	totalVolume := 0.0
	for _, limit := range ob.asks {
		totalVolume += limit.TotalVolume
	}
	return totalVolume
}

func (ob *OrderBook) Asks() Limits {
	sort.Sort(ByBestAsk{ob.asks})
	return ob.asks
}

func (ob *OrderBook) Bids() Limits {
	sort.Sort(ByBestBid{ob.bids})
	return ob.bids
}

func (ob *OrderBook) clearLimit(bid bool, l *Limit) {
	// take the limit out of the slice
	if bid {
		delete(ob.BidLimits, l.Price)
		for i, limit := range ob.bids {
			if limit == l {
				ob.bids = append(ob.bids[:i], ob.bids[i+1:]...)
				break
			}
		}
	} else {
		delete(ob.AskLimits, l.Price)
		for i, limit := range ob.asks {
			if limit == l {
				ob.asks = append(ob.asks[:i], ob.asks[i+1:]...)
				break
			}
		}
	}

}

func (ob *OrderBook) CancelOrder(o *Order) {
	limit := o.Limit
	limit.deleteOrder(o)

	if limit.isEmpty() {
		ob.clearLimit(o.Bid, limit)
	}
}
