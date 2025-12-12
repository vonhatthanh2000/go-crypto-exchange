package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	eth_client "go-crypto-exchange/eth-client"
	"go-crypto-exchange/orderbook"
	"go-crypto-exchange/pkg/response"
	"log"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	exchange := NewExchange(ExPrivateKey)

	e.HTTPErrorHandler = errorHandler

	e.POST("/order", exchange.handlePlaceOrder)
	e.DELETE("/order", exchange.handleCancelOrder)
	e.GET("/book/:market", exchange.handleGetOrderBook)

	e.Start(":3000")

}

func errorHandler(err error, c echo.Context) {
	fmt.Println("err", err)
}

type (
	Market    string
	OrderType string

	Exchange struct {
		users      map[int64]*User
		orders     map[int64]*orderbook.Order
		PrivateKey *ecdsa.PrivateKey
		orderbooks map[Market]*orderbook.OrderBook
		ethClient  *ethclient.Client
	}

	PlaceOrderRequest struct {
		Market Market    `json:"market"`
		UserID int64     `json:"userID"`
		Type   OrderType `json:"type"`
		Bid    bool      `json:"bid"`
		Size   float64   `json:"size"`
		Price  float64   `json:"price"`
	}

	MatchedOrder struct {
		Price float64
		Size  float64
		ID    int64
	}

	Order struct {
		ID        int64
		Price     float64
		Size      float64
		Bid       bool
		Timestamp int64
	}

	OrderBookData struct {
		TotalBidVolume float64 `json:"totalBidVolume"`
		TotalAskVolume float64 `json:"totalAskVolume"`
		Asks           []*Order
		Bids           []*Order
	}

	CancelOrderRequest struct {
		Market Market `json:"market"`
		ID     int64  `json:"id"`
	}
)

type User struct {
	UserID     int64
	PrivateKey *ecdsa.PrivateKey
}

func NewUser(userID int64, privKey string) *User {
	privateKey, err := crypto.HexToECDSA(privKey)
	if err != nil {
		log.Fatal(err)
	}
	return &User{UserID: userID, PrivateKey: privateKey}
}

const (
	MarketETH Market = "ETH"

	OrderTypeLimit  OrderType = "LIMIT"
	OrderTypeMarket OrderType = "MARKET"
)

func NewExchange(privKey string) *Exchange {

	orderbooks := make(map[Market]*orderbook.OrderBook)
	orderbooks[MarketETH] = orderbook.NewOrderBook()

	privateKey, err := crypto.HexToECDSA(privKey)
	if err != nil {
		log.Fatal(err)
	}

	ethclient := eth_client.NewEthClient(GANACHE_URL)

	userMap := make(map[int64]*User)

	userMap[1] = NewUser(1, UserInfo.PrivateKey)

	return &Exchange{
		users:      userMap,
		orders:     make(map[int64]*orderbook.Order),
		PrivateKey: privateKey,
		orderbooks: orderbooks,
		ethClient:  ethclient,
	}
}

func (ex *Exchange) handlePlaceLimitOrder(c echo.Context, ob *orderbook.OrderBook, price float64, order *orderbook.Order) error {

	limit := ob.PlaceLimitOrder(price, order)

	orderData := Order{
		ID:        order.ID,
		Price:     limit.Price,
		Size:      order.Size,
		Bid:       order.Bid,
		Timestamp: order.Timestamp,
	}

	// transfer ETH from user to exchange
	exPublicKey := ex.PrivateKey.Public()
	exPublicKeyECDSA, ok := exPublicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	exAddress := crypto.PubkeyToAddress(*exPublicKeyECDSA)

	eth_client.TransferETH(ex.ethClient, ex.users[1].PrivateKey, exAddress, big.NewInt(int64(order.Size)))

	return response.SuccessResponse(c, response.Success, orderData)
}

func (ex *Exchange) handlePlaceMarketOrder(c echo.Context, ob *orderbook.OrderBook, order *orderbook.Order) error {
	matches := ob.PlaceMarketOrder(order)
	// Convert matches to response format
	matchedOrders := make([]MatchedOrder, len(matches))

	// handle matches
	for i, match := range matches {

		id := match.Bid.ID
		if order.Bid {
			id = match.Ask.ID
		}

		matchedOrders[i] = MatchedOrder{
			Size:  match.SizeFilled,
			Price: match.Price,
			ID:    id,
		}
	}

	return response.SuccessResponse(c, response.Success, matchedOrders)
}

func (ex *Exchange) handlePlaceOrder(c echo.Context) error {

	var placeOrderData PlaceOrderRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&placeOrderData); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	market := Market(placeOrderData.Market)
	ob := ex.orderbooks[market]

	requestOrderType := OrderType(placeOrderData.Type)

	order := orderbook.NewOrder(placeOrderData.Bid, placeOrderData.Size, placeOrderData.UserID)

	if requestOrderType == OrderTypeLimit {
		return ex.handlePlaceLimitOrder(c, ob, placeOrderData.Price, order)
	} else if requestOrderType == OrderTypeMarket {
		fmt.Println("order", order)

		return ex.handlePlaceMarketOrder(c, ob, order)
	} else {
		return response.FailResponse(c, response.Fail, map[string]string{"error": "invalid order type"})
	}
}

// /

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
