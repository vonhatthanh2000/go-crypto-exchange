package orderbook

import (
	"reflect"
	"testing"
)

const userId = int64(1)

func assert(t *testing.T, a, b any) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("Expected %v, got %v", a, b)
	}
}

func TestLimit(t *testing.T) {
	l := NewLimit(10_000)
	orderA := NewOrder(true, 10, userId)
	orderB := NewOrder(true, 20, userId)
	orderC := NewOrder(false, 30, userId)

	l.addOrder(orderA)
	l.addOrder(orderB)
	l.addOrder(orderC)

	l.deleteOrder(orderB)

}

func TestPlaceLimitOrder(t *testing.T) {
	ob := NewOrderBook()

	buyOrderA := NewOrder(true, 100, userId)
	ob.PlaceLimitOrder(10_000, buyOrderA)

	buyOrderB := NewOrder(true, 200, userId)
	ob.PlaceLimitOrder(10_000, buyOrderB)

	assert(t, len(ob.Bids()), 1)
	assert(t, len(ob.Orders), 2)
	assert(t, ob.Orders[buyOrderA.ID], buyOrderA)
}

func TestPlaceMarketOrder(t *testing.T) {

	ob := NewOrderBook()

	buyOrderA := NewOrder(true, 100, userId)
	ob.PlaceLimitOrder(10_000, buyOrderA)
	sellOrderA := NewOrder(false, 50.0, userId)
	matches := ob.PlaceMarketOrder(sellOrderA)
	assert(t, len(matches), 1)
	assert(t, matches[0].Bid.Size, 50.0)
	assert(t, ob.BidLimits[10_000].TotalVolume, 50.0)

}

func TestPlaceMarketOrderMultiFill(t *testing.T) {
	ob := NewOrderBook()

	buyOrderA := NewOrder(true, 10.0, userId)
	buyOrderB := NewOrder(true, 15.0, userId)
	buyOrderC := NewOrder(true, 20.0, userId)
	buyOrderD := NewOrder(true, 1.0, userId)

	// place limit orders
	ob.PlaceLimitOrder(10_000, buyOrderA)
	ob.PlaceLimitOrder(10_000, buyOrderD)
	ob.PlaceLimitOrder(11_000, buyOrderB)
	ob.PlaceLimitOrder(12_000, buyOrderC)

	assert(t, ob.BidTotalVolume(), 46.0)

	// place market order
	sellOrderA := NewOrder(false, 34.0, userId)
	matchesA := ob.PlaceMarketOrder(sellOrderA)
	assert(t, len(matchesA), 2)

	sellOrderB := NewOrder(false, 2.0, userId)
	matchesB := ob.PlaceMarketOrder(sellOrderB)
	assert(t, len(matchesB), 2)

}

func TestPlaceMarketOrderFillLimit(t *testing.T) {

	ob := NewOrderBook()

	buyOrderA := NewOrder(true, 15.0, userId)
	buyOrderB := NewOrder(true, 10.0, userId)
	buyOrderC := NewOrder(true, 20.0, userId)

	ob.PlaceLimitOrder(10_000, buyOrderA)
	ob.PlaceLimitOrder(10_000, buyOrderB)
	ob.PlaceLimitOrder(11_000, buyOrderC)

	assert(t, ob.BidTotalVolume(), 45.0)

	sellOrderA := NewOrder(false, 34.0, userId)
	matchesA := ob.PlaceMarketOrder(sellOrderA)

	assert(t, len(matchesA), 2)

	assert(t, len(ob.Bids()), 1)

}

func TestCancelOrder(t *testing.T) {
	ob := NewOrderBook()

	buyOrderA := NewOrder(true, 10.0, userId)
	buyOrderB := NewOrder(true, 15.0, userId)
	ob.PlaceLimitOrder(10_000, buyOrderA)
	ob.PlaceLimitOrder(11_000, buyOrderB)
	assert(t, len(ob.Bids()), 2)

	ob.CancelOrder(buyOrderA)
	assert(t, len(ob.Bids()), 1)

	_, ok := ob.Orders[buyOrderA.ID]
	assert(t, ok, false)
}
