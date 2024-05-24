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
	rawTx, err := tx.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal unsigned tx: %w", err)
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

	if bytes.Contains(rawSig, []byte("error")) || bytes.Contains(rawSig, []byte("Error")) || bytes.Contains(rawSig, []byte("Resource not found")) {
		return nil, fmt.Errorf("%s", string(rawSig))
	}

	sig, err := hexutil.Decode(string(rawSig))
	if err != nil {
		return nil, err
	}

	// V must be 0 or 1
	if sig[64] > 1 {
		sig[64] = 1
	}

	return tx.WithSignature(types.NewEIP155Signer(chainID), sig)
}
