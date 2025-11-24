package main

import (
	"fmt"
	"testing"
)

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

func TestOrderBook(t *testing.T) {
	ob := NewOrderBook()

	buyOrderA := NewOrder(true, 100)
	ob.PlaceOrder(10_000, buyOrderA)

	buyOrderB := NewOrder(true, 200)
	ob.PlaceOrder(10_000, buyOrderB)

	for i := 0; i < len(ob.Bids); i++ {
		fmt.Println(ob.Bids[i])
	}
}
