package main

import (
	"encoding/json"
	"go-crypto-exchange/orderbook"
	"go-crypto-exchange/pkg/response"
	"net/http"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	exchange := NewExchange()

	e.POST("/order", exchange.handlePlaceOrder)
	e.DELETE("/order", exchange.handleCancelOrder)
	e.GET("/book/:market", exchange.handleGetOrderBook)

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
	Market Market    `json:"market"`
	Type   OrderType `json:"type"`
	Bid    bool      `json:"bid"`
	Size   float64   `json:"size"`
	Price  float64   `json:"price"`
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
		limit := ob.PlaceLimitOrder(placeOrderData.Price, order)

		orderData := Order{
			ID:        order.ID,
			Price:     limit.Price,
			Size:      order.Size,
			Bid:       order.Bid,
			Timestamp: order.Timestamp,
		}
		return response.SuccessResponse(c, response.Success, orderData)
	} else if requestOrderType == OrderTypeMarket {
		matches := ob.PlaceMarketOrder(order)

		// Convert matches to response format
		matchResponses := make([]MatchResponse, len(matches))
		for i, match := range matches {
			matchResponses[i] = MatchResponse{
				SizeFilled: match.SizeFilled,
				Price:      match.Price,
			}
		}

		return response.SuccessResponse(c, response.Success, map[string]interface{}{
			"matches": matchResponses,
		})
	} else {
		return response.FailResponse(c, response.Fail, map[string]string{"error": "invalid order type"})
	}
}

// /
type Order struct {
	ID        int64
	Price     float64
	Size      float64
	Bid       bool
	Timestamp int64
}

type MatchResponse struct {
	SizeFilled float64 `json:"sizeFilled"`
	Price      float64 `json:"price"`
}

type OrderBookData struct {
	TotalBidVolume float64 `json:"totalBidVolume"`
	TotalAskVolume float64 `json:"totalAskVolume"`
	Asks           []*Order
	Bids           []*Order
}

func (ex *Exchange) handleGetOrderBook(c echo.Context) error {
	market := Market(c.Param("market"))

	ob, ok := ex.orderbooks[market]
	if !ok {
		return response.FailResponse(c, response.Fail, map[string]string{"error": "market not found"})
	}

	orderBookData := OrderBookData{
		TotalBidVolume: ob.BidTotalVolume(),
		TotalAskVolume: ob.AskTotalVolume(),
		Asks:           []*Order{},
		Bids:           []*Order{},
	}

	for _, bidLimit := range ob.Bids() {
		for _, order := range bidLimit.Orders {
			orderBookData.Bids = append(orderBookData.Bids, &Order{
				ID:        order.ID,
				Price:     bidLimit.Price,
				Size:      order.Size,
				Bid:       true,
				Timestamp: order.Timestamp,
			})
		}
	}

	for _, askLimit := range ob.Asks() {
		for _, order := range askLimit.Orders {
			orderBookData.Asks = append(orderBookData.Asks, &Order{
				ID:        order.ID,
				Price:     askLimit.Price,
				Size:      order.Size,
				Bid:       false,
				Timestamp: order.Timestamp,
			})
		}
	}

	return response.SuccessResponse(c, response.Success, orderBookData)
}

type CancelOrderRequest struct {
	Market Market `json:"market"`
	ID     int64  `json:"id"`
}

func (ex *Exchange) handleCancelOrder(c echo.Context) error {
	var cancelOrderData CancelOrderRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&cancelOrderData); err != nil {
		return response.FailResponse(c, response.Fail, map[string]string{"error": "invalid cancel order data"})
	}

	market := Market(cancelOrderData.Market)
	ob := ex.orderbooks[market]

	order, ok := ob.Orders[cancelOrderData.ID]
	if !ok {
		return response.FailResponse(c, response.Fail, map[string]string{"error": "order not found"})
	}

	ob.CancelOrder(order)

	return response.SuccessResponse(c, response.Success, map[string]string{"message": "Order cancelled"})
}
