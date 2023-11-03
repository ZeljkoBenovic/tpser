package txsigner

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"

	"github.com/ZeljkoBenovic/tpser/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	ErrPubKey = errors.New("could not get public key from private")
)

var (
	DefaultGasPrice = big.NewInt(30000000000)
	EOAGasLimit     = uint64(21000)
	EOAValue        = big.NewInt(100000)
)

type TxSigner struct {
	ctx context.Context
	log logger.Logger
	eth *ethclient.Client

	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	from       common.Address
	to         common.Address
	nonce      uint64
	gasPrice   *big.Int
	gasLimit   uint64
	chainId    *big.Int
}

func New(ctx context.Context, log logger.Logger, eth *ethclient.Client) *TxSigner {
	return &TxSigner{
		ctx: ctx,
		log: log,
		eth: eth,
	}
}

func (t *TxSigner) SetPrivateKey(privKey string) error {
	pk, err := crypto.HexToECDSA(privKey)
	if err != nil {
		return fmt.Errorf("could not setup private key: %w", err)
	}

	pubKey, ok := pk.Public().(*ecdsa.PublicKey)
	if !ok {
		return ErrPubKey
	}

	t.privateKey = pk
	t.publicKey = pubKey
	t.from = crypto.PubkeyToAddress(*pubKey)

	return nil
}

func (t *TxSigner) SetToAddress(toAddressString string) error {
	nonce, err := t.eth.PendingNonceAt(t.ctx, t.from)
	if err != nil {
		return fmt.Errorf("could not get pending nonce: %w", err)
	}

	gas, err := t.eth.SuggestGasPrice(t.ctx)
	if err != nil {
		t.log.Warn("Could not get suggested gas price, using default", "default", DefaultGasPrice)
		gas = DefaultGasPrice
	}

	chainId, err := t.eth.ChainID(t.ctx)
	if err != nil {
		return fmt.Errorf("could not get chain id: %w", err)
	}

	t.nonce = nonce
	t.gasPrice = gas
	t.gasLimit = EOAGasLimit
	t.chainId = chainId
	t.to = common.HexToAddress(toAddressString)

	return nil
}

func (t *TxSigner) GetNonce() uint64 {
	return t.nonce
}

func (t *TxSigner) GetNextSignedTx(nextNonce uint64) (*types.Transaction, error) {
	newTx := types.NewTx(&types.LegacyTx{
		Nonce:    nextNonce,
		GasPrice: t.gasPrice,
		Gas:      t.gasLimit,
		To:       &t.to,
		Value:    EOAValue,
		Data:     []byte{},
	})

	tx, err := types.SignTx(newTx, types.NewEIP155Signer(t.chainId), t.privateKey)
	if err != nil {
		return nil, fmt.Errorf("could not sign the transaction: %w", err)
	}

	return tx, nil
}
