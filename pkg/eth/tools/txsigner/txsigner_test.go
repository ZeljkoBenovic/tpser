package txsigner

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

type ethClientMock struct{}

func (e ethClientMock) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	return 10, nil
}

func (e ethClientMock) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return big.NewInt(100), nil
}

func (e ethClientMock) ChainID(ctx context.Context) (*big.Int, error) {
	return big.NewInt(1000), nil
}

func TestTxSigner_SetPrivateKey(t *testing.T) {
	t.Parallel()
	txs := TxSigner{}

	testPrivKey, _ := crypto.GenerateKey()
	pkBytes := crypto.FromECDSA(testPrivKey)
	pkString := hexutil.Encode(pkBytes)[2:]
	testPubKey, _ := testPrivKey.Public().(*ecdsa.PublicKey)
	testFromAddr := crypto.PubkeyToAddress(*testPubKey)

	var privKeyTests = []struct {
		name         string
		privKeyInput string
		wantPrivKey  *ecdsa.PrivateKey
		wantPubKey   *ecdsa.PublicKey
		wantFrom     common.Address
		shouldError  bool
	}{
		{
			name:         "Private key not set",
			privKeyInput: "",
			wantPrivKey:  nil,
			wantPubKey:   nil,
			wantFrom:     common.Address{},
			shouldError:  true,
		},
		{
			name:         "Invalid private key format",
			privKeyInput: "2137814708090789796567587656574",
			wantPrivKey:  nil,
			wantPubKey:   nil,
			wantFrom:     common.Address{},
			shouldError:  true,
		},
		{
			name:         "Valid private key",
			privKeyInput: pkString,
			wantPrivKey:  testPrivKey,
			wantPubKey:   testPubKey,
			wantFrom:     testFromAddr,
			shouldError:  false,
		},
	}
	for _, tt := range privKeyTests {
		t.Run(tt.name, func(t *testing.T) {
			err := txs.SetPrivateKey(tt.privKeyInput)
			if tt.shouldError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			assert.Equal(t, tt.wantPrivKey, txs.privateKey)
			assert.Equal(t, tt.wantPubKey, txs.publicKey)
			assert.Equal(t, tt.wantFrom, txs.from)
		})
	}
}

func TestTxSigner_SetToAddress(t *testing.T) {
	tx := TxSigner{
		eth: ethClientMock{},
		ctx: context.Background(),
	}

	testPrivKey, _ := crypto.GenerateKey()
	testPubKey := testPrivKey.Public().(*ecdsa.PublicKey)
	testToAddress := crypto.PubkeyToAddress(*testPubKey)

	var setToAddressTests = []struct {
		name         string
		toAddrString string
		wantNonce    uint64
		wantGasPrice *big.Int
		wantGasLimit uint64
		wantChainId  *big.Int
		wantToAddr   common.Address
		shouldErr    bool
	}{
		{
			name:         "Valid address",
			toAddrString: testToAddress.Hex(),
			wantNonce:    10,
			wantGasPrice: big.NewInt(100),
			wantGasLimit: EOAGasLimit,
			wantChainId:  big.NewInt(1000),
			wantToAddr:   testToAddress,
			shouldErr:    false,
		},
	}

	for _, tt := range setToAddressTests {
		t.Run(tt.name, func(t *testing.T) {
			err := tx.SetToAddress(tt.toAddrString)
			if tt.shouldErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			assert.Equal(t, tt.wantNonce, tx.nonce)
			assert.Equal(t, tt.wantGasPrice, tx.gasPrice)
			assert.Equal(t, tt.wantGasLimit, tx.gasLimit)
			assert.Equal(t, tt.wantChainId, tx.chainId)
			assert.Equal(t, tt.wantToAddr, tx.to)
		})
	}
}

func TestTxSigner_GetNonce(t *testing.T) {
	tx := TxSigner{}

	var nonceTestCases = []struct {
		name  string
		input uint64
		want  uint64
	}{
		{
			name: "valid nonce",
			want: 123,
		},
	}
	tx.nonce = 123

	for _, tt := range nonceTestCases {
		t.Run(tt.name, func(t *testing.T) {
			tx.GetNonce()
			assert.Equal(t, tt.want, tx.nonce)
		})
	}
}

func TestTxSigner_GetNextSignedTx(t *testing.T) {
	testPrivKey, _ := crypto.GenerateKey()
	testPubKey := testPrivKey.Public().(*ecdsa.PublicKey)
	testToAddress := crypto.PubkeyToAddress(*testPubKey)

	tx := TxSigner{}
	tx.gasPrice = big.NewInt(21000)
	tx.gasLimit = EOAGasLimit
	tx.to = testToAddress
	tx.chainId = big.NewInt(1000)
	tx.privateKey = testPrivKey

	var testCases = []struct {
		name              string
		nonce             uint64
		nextNonce         uint64
		txHashShouldMatch bool
	}{
		{
			name:              "sequential nonce",
			nonce:             1,
			nextNonce:         2,
			txHashShouldMatch: false,
		},
		{
			name:              "same nonce",
			nonce:             1,
			nextNonce:         1,
			txHashShouldMatch: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tx1, err := tx.GetNextSignedTx(tt.nonce)
			assert.Nil(t, err)

			tx2, err := tx.GetNextSignedTx(tt.nextNonce)
			assert.Nil(t, err)

			if tt.txHashShouldMatch {
				assert.Equal(t, tx1.Hash().String(), tx2.Hash().String())
			} else {
				assert.NotEqual(t, tx1.Hash().String(), tx2.Hash().String())
			}
		})
	}
}
