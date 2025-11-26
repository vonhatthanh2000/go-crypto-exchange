package main

import (
	"fmt"
	"reflect"
	"testing"
)

func assert(t *testing.T, a, b any) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("Expected %v, got %v", a, b)
	}
}

func TestLimit(t *testing.T) {
	l := NewLimit(10_000)
	orderA := NewOrder(true, 10)
	orderB := NewOrder(true, 20)
	orderC := NewOrder(false, 30)

	l.addOrder(orderA)
	l.addOrder(orderB)
	l.addOrder(orderC)

	l.deleteOrder(orderB)

	fmt.Println(l)
}

func TestPlaceLimitOrder(t *testing.T) {
	ob := NewOrderBook()

	buyOrderA := NewOrder(true, 100)
	ob.PlaceLimitOrder(10_000, buyOrderA)

	buyOrderB := NewOrder(true, 200)
	ob.PlaceLimitOrder(10_000, buyOrderB)

	assert(t, len(ob.Bids()), 1)
}

func TestPlaceMarketOrder(t *testing.T) {

	ob := NewOrderBook()

	buyOrderA := NewOrder(true, 100)
	ob.PlaceLimitOrder(10_000, buyOrderA)
	sellOrderA := NewOrder(false, 50.0)
	matches := ob.PlaceMarketOrder(sellOrderA)
	assert(t, len(matches), 1)
	assert(t, matches[0].Bid.Size, 50.0)
	assert(t, ob.BidLimits[10_000].TotalVolume, 50.0)

}
