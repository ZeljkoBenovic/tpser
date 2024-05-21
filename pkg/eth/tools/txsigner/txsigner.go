package txsigner

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"

	"github.com/ZeljkoBenovic/tpser/pkg/conf"
	"github.com/ZeljkoBenovic/tpser/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

type txSignerEthClient interface {
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	ChainID(ctx context.Context) (*big.Int, error)
}

var (
	ErrPubKey                  = errors.New("could not get public key from private")
	ErrPKOrMnemonicNotProvided = errors.New("private key or mnemonic not provided")
)

var (
	DefaultGasPrice = big.NewInt(30000000000)
	EOAGasLimit     = uint64(21000)
	EOAValue        = big.NewInt(100000)
)

type TxSigner struct {
	ctx  context.Context
	log  logger.Logger
	eth  txSignerEthClient
	conf conf.Conf

	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	from       common.Address
	to         common.Address
	nonce      uint64
	gasPrice   *big.Int
	gasLimit   uint64
	chainId    *big.Int
}

type Options struct {
	NumberOfAccounts int
}

type SignerOpts func(*Options)

func WithNumberOfAccounts(accNo int) SignerOpts {
	return func(options *Options) {
		options.NumberOfAccounts = accNo
	}
}

func New(ctx context.Context, log logger.Logger, eth *ethclient.Client, conf conf.Conf) *TxSigner {
	return &TxSigner{
		ctx:  ctx,
		log:  log,
		eth:  eth,
		conf: conf,
	}
}

func (t *TxSigner) SetPrivateKey(opts ...SignerOpts) error {
	var (
		pk  *ecdsa.PrivateKey
		err error
	)

	o := &Options{
		NumberOfAccounts: 1,
	}
	for _, f := range opts {
		f(o)
	}

	if t.conf.PrivateKey != "" {
		pk, err = crypto.HexToECDSA(t.conf.PrivateKey)
		if err != nil {
			return fmt.Errorf("could not setup private key: %w", err)
		}
	}

	if t.conf.Mnemonic != "" {
		pk, err = t.getPrivateKeyFromMnemonicDerivedNumber(o.NumberOfAccounts)
	}

	if t.conf.PrivateKey == "" && t.conf.Mnemonic == "" {
		return ErrPKOrMnemonicNotProvided
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

	t.log.Debug("Pending nonce fetched", "nonce", nonce, "from", t.from.String())

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
func (t *TxSigner) GetFreshNonce() (uint64, error) {
	return t.eth.PendingNonceAt(t.ctx, t.from)
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

func (t *TxSigner) GetFromAddress() string {
	return t.from.String()
}

func (t *TxSigner) getPrivateKeyFromMnemonicDerivedNumber(accNo int) (*ecdsa.PrivateKey, error) {
	wallet, err := hdwallet.NewFromMnemonic(t.conf.Mnemonic)
	if err != nil {
		return nil, fmt.Errorf("could not process mnemonic: %w", err)
	}

	path := hdwallet.MustParseDerivationPath(fmt.Sprintf("m/44'/60'/0'/0/%d", accNo))
	account, err := wallet.Derive(path, false)
	if err != nil {
		return nil, fmt.Errorf("could not derive account from mnemonic: %w", err)
	}

	return wallet.PrivateKey(account)
}
