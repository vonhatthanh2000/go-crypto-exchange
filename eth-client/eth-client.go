package eth_client

import (
	"context"
	"crypto/ecdsa"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func NewEthClient(url string) *ethclient.Client {
	client, err := ethclient.Dial(url)
	if err != nil {
		panic(err)
	}

	return client
}

func getAccountBalance(client *ethclient.Client, address string) (*big.Int, error) {
	balance, err := client.BalanceAt(context.Background(), common.HexToAddress(address), nil)
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func TransferETH(client *ethclient.Client, fromPrivKey *ecdsa.PrivateKey, toAddress common.Address, amount *big.Int) {
	publicKey := fromPrivKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	value := amount.Mul(amount, big.NewInt(1000000000000000000))

	gasLimit := uint64(21000) // in units

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)

	chainID := big.NewInt(1337)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), fromPrivKey)
	if err != nil {
		log.Fatal(err)
	}
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

}
