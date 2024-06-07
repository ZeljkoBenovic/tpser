package txsigner

import (
	"bytes"
	"fmt"
	"io"
	"math/big"
	"net/http"

	"github.com/ZeljkoBenovic/tpser/pkg/conf"
	"github.com/ZeljkoBenovic/tpser/pkg/logger"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

type web3signer struct {
	publicKey string
	conf      conf.Conf
	log       logger.Logger
}

func newWeb3Signer(conf conf.Conf, log logger.Logger) *web3signer {
	return &web3signer{
		conf: conf,
		log:  log,
	}
}

func (w web3signer) SignTx(tx *types.Transaction, chainID *big.Int, pubKey string) (*types.Transaction, error) {
	var (
		signer       = types.NewEIP155Signer(chainID)
		dataToEncode = []interface{}{
			tx.Nonce(),
			tx.GasPrice(),
			tx.Gas(),
			tx.To(),
			tx.Value(),
			tx.Data(),
			chainID, uint(0), uint(0),
		}
	)

	rawTx, err := rlp.EncodeToBytes(dataToEncode)
	if err != nil {
		return nil, fmt.Errorf("could not rlp encode tx to bytes: %w", err)
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/api/v1/eth1/sign/%s", w.conf.Web3SignerURL, pubKey),
		"application/json",
		bytes.NewBuffer([]byte(fmt.Sprintf("{\"data\":\"%s\"}", hexutil.Encode(rawTx)))),
	)
	if err != nil {
		return nil, fmt.Errorf("web3signer error: %w", err)
	}

	defer resp.Body.Close()

	rawSig, _ := io.ReadAll(resp.Body)

	w.log.Debug("Signature received from web3signer", "sig", string(rawSig))

	if bytes.Contains(bytes.ToLower(rawSig), []byte("error")) || bytes.Contains(bytes.ToLower(rawSig), []byte("not found")) {
		return nil, fmt.Errorf("%s", string(rawSig))
	}

	sig, err := hexutil.Decode(string(rawSig))
	if err != nil {
		return nil, fmt.Errorf("could not decode raw web3signer signature: %w", err)
	}

	// fix legacy V
	if sig[64] == 28 || sig[64] == 27 {
		sig[64] -= 27
	}

	signedTx, err := tx.WithSignature(signer, sig)
	sender, err := signer.Sender(signedTx)
	if err != nil {
		return nil, err
	}

	w.log.Debug("Transaction sent", "sender", sender, "hash", signedTx.Hash().String())

	return signedTx, nil
}
