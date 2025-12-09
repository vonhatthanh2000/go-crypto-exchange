package main

import (
	"encoding/json"
	"go-crypto-exchange/orderbook"
	"net/http"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	exchange := NewExchange()

	e.POST("/order", exchange.handlePlaceOrder)

	e.Start(":3000")

}

type Market string

const (
	MarketETH Market = "ETH"
)

type OrderType string

const (
	OrderTypeLimit  OrderType = "LIMIT"
	OrderTypeMarket OrderType = "MARKET"
)

type Exchange struct {
	orderbooks map[Market]*orderbook.OrderBook
}

func NewExchange() *Exchange {

	orderbooks := make(map[Market]*orderbook.OrderBook)
	orderbooks[MarketETH] = orderbook.NewOrderBook()

	return &Exchange{
		orderbooks: orderbooks,
	}
}

type PlaceOrderRequest struct {
	Type   OrderType
	Market Market
	Bid    bool
	Size   float64
	Price  float64
}

func (ex *Exchange) handlePlaceOrder(c echo.Context) error {

	var placeOrderData PlaceOrderRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&placeOrderData); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	market := Market(placeOrderData.Market)
	ob := ex.orderbooks[market]

	requestOrderType := OrderType(placeOrderData.Type)

	order := orderbook.NewOrder(placeOrderData.Bid, placeOrderData.Size)

	if requestOrderType == OrderTypeLimit {
		ob.PlaceLimitOrder(placeOrderData.Price, order)
	} else if requestOrderType == OrderTypeMarket {
		ob.PlaceMarketOrder(order)
	} else {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid order type"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Order placed"})
}
