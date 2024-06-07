package txsigner

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"

	"github.com/ZeljkoBenovic/tpser/pkg/conf"
	"github.com/ZeljkoBenovic/tpser/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

type txSignerEthClient interface {
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	ChainID(ctx context.Context) (*big.Int, error)
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
}

var (
	ErrPubKey = errors.New("could not get public key from private")
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

	privateKey                    *ecdsa.PrivateKey
	publicKey                     *ecdsa.PublicKey
	publicKeyStringFromWeb3Signer string
	from                          common.Address
	to                            common.Address
	nonce                         uint64
	gasPrice                      *big.Int
	gasLimit                      uint64
	chainId                       *big.Int

	web3signer *web3signer
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

func (t *TxSigner) UseWeb3Signer() {
	t.web3signer = newWeb3Signer(t.conf, t.log)
}

func (t *TxSigner) SetPrivateKey(opts ...SignerOpts) error {
	var (
		privKey *ecdsa.PrivateKey
		pubKey  *ecdsa.PublicKey
		address common.Address
		err     error
	)

	o := &Options{
		NumberOfAccounts: 1,
	}
	for _, f := range opts {
		f(o)
	}

	if t.conf.PrivateKey != "" {
		privKey, err = crypto.HexToECDSA(t.conf.PrivateKey)
		if err != nil {
			return fmt.Errorf("could not setup private key: %w", err)
		}
	}

	if t.conf.Mnemonic != "" {
		privKey, err = t.getPrivateKeyFromMnemonicDerivedNumber(o.NumberOfAccounts)
		if err != nil {
			return fmt.Errorf("could not get private key from mnemonic: %w", err)
		}
	}

	if t.conf.Web3SignerURL != "" {
		var pubKeyByte []byte

		//TODO: save public keys to local storage and parse them from there.
		// This must be done, as web3signer response is not reliable
		if t.conf.Web3SignerPublickey == "" {
			t.log.Debug("Using web3signer key number")
			resp, err := http.Get(fmt.Sprintf("%s/api/v1/eth1/publicKeys", t.conf.Web3SignerURL))
			if err != nil {
				return fmt.Errorf("could not fetch public keys from web3signer: %w", err)
			}

			defer resp.Body.Close()

			rawKeys, _ := io.ReadAll(resp.Body)

			if bytes.Contains(rawKeys, []byte("error")) || bytes.Contains(rawKeys, []byte("Error")) {
				return fmt.Errorf("%s", string(rawKeys))
			}

			keys := make([]string, 0)
			if err := json.Unmarshal(rawKeys, &keys); err != nil {
				return err
			}

			pubKeyByte, err = hexutil.Decode(keys[t.conf.Web3SignerPubKeyNum])
			if err != nil {
				return err
			}

			t.publicKeyStringFromWeb3Signer = keys[t.conf.Web3SignerPubKeyNum]
		} else {
			t.log.Debug("Using the provided public key", "key", t.conf.Web3SignerPublickey)
			pubKeyByte, err = hexutil.Decode(t.conf.Web3SignerPublickey)
			if err != nil {
				return err
			}

			t.publicKeyStringFromWeb3Signer = t.conf.Web3SignerPublickey
		}

		if len(pubKeyByte) == 65 && pubKeyByte[0] == 4 {
			pubKeyByte = pubKeyByte[1:]
		}

		hash := crypto.Keccak256(pubKeyByte)
		address = common.BytesToAddress(hash[len(hash)-20:])
	} else {
		pubKey = privKey.Public().(*ecdsa.PublicKey)
		address = crypto.PubkeyToAddress(*pubKey)
	}

	t.privateKey = privKey
	t.publicKey = pubKey
	t.from = address

	return nil
}

func (t *TxSigner) SetToAddress(toAddressString string) error {
	nonce, err := t.eth.PendingNonceAt(t.ctx, t.from)
	if err != nil {
		return fmt.Errorf("could not get pending nonce: %w", err)
	}

	gas, err := t.eth.SuggestGasPrice(t.ctx)
	if err != nil {
		t.log.Debug("Could not get suggested gas price", "default", DefaultGasPrice, "err", err.Error())
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
	if toAddressString == "" {
		t.to = t.from
	} else {
		t.to = common.HexToAddress(toAddressString)
	}

	bal, err := t.eth.BalanceAt(t.ctx, t.from, nil)
	if err != nil {
		return fmt.Errorf("could not fetch balance for senders account: %w", err)
	}

	t.log.Debug("Transaction details set",
		"nonce", nonce,
		"from", t.from.String(),
		"to", t.to.String(),
		"gas_price", gas.String(),
		"senders_balance", bal.String(),
	)

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

	if t.web3signer != nil {
		return t.web3signer.SignTx(newTx, t.chainId, t.publicKeyStringFromWeb3Signer)
	}

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
