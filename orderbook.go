package main

import (
	"fmt"
	"sort"
	"time"
)

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

func (ob *OrderBook) Asks() Limits {
	sort.Sort(ByBestAsk{ob.asks})
	return ob.asks
}

func (ob *OrderBook) Bids() Limits {
	sort.Sort(ByBestBid{ob.bids})
	return ob.bids
}
