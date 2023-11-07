package txsender

import (
	"context"
	"fmt"

	"github.com/ZeljkoBenovic/tpser/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type TxSender struct {
	ctx context.Context
	eth *ethclient.Client
	log logger.Logger
}

func New(ctx context.Context, log logger.Logger, eth *ethclient.Client) *TxSender {
	return &TxSender{
		ctx: ctx,
		eth: eth,
		log: log,
	}
}

func (t *TxSender) SendSignedTransaction(signedTx *types.Transaction) (common.Hash, error) {
	if err := t.eth.SendTransaction(t.ctx, signedTx); err != nil {
		t.log.Debug("Could not send transaction", "err", err, "tx_hash", signedTx.Hash())
		return common.Hash{}, fmt.Errorf("could not send transacton: %w", err)
	}

	return signedTx.Hash(), nil
}
