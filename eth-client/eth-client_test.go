package eth_client

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const ganacheUrl = "http://localhost:7545"

func TestGetAccountBalance(t *testing.T) {

	testAccount := "0x2977ef60Ec801bd31951B725CC127bB5A89FEbFb"

	ethClient := NewEthClient(ganacheUrl)

	accountBalance, err := getAccountBalance(ethClient, testAccount)
	if err != nil {
		t.Fatalf("failed to get account balance: %v", err)
	}

	fmt.Println("account balance", accountBalance)
}

func TestTransferETH(t *testing.T) {

	fromPrivKeyStr := "f8582382e7d1509139099517bb420ddd81ed15322f343d6e5d95b0d90020177f"

	fromPrivKey, err := crypto.HexToECDSA(fromPrivKeyStr)
	if err != nil {
		t.Fatalf("failed to convert private key to ECDSA: %v", err)
	}

	toAddress := common.HexToAddress("0xF8d986780871ca37E8E7bA9C973609b48dfBE679")

	ethClient := NewEthClient(ganacheUrl)

	TransferETH(ethClient, fromPrivKey, toAddress, big.NewInt(1))
}
