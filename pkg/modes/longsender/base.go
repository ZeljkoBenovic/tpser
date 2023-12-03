package longsender

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/ZeljkoBenovic/tpser/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/sync/errgroup"
)

var (
	defaultEOAGasLimit = uint64(21000)
	defaultEOAValue    = big.NewInt(100000)
)

type base struct {
	log    logger.Logger
	client client

	config Config
}

func newBase(log logger.Logger, client client, config Config) *base {
	return &base{
		log:    log,
		client: client,
		config: config,
	}
}

// sendTransactions sends transactions at a given rate
// from a list of senders
func (b *base) sendTransactions(
	ctx context.Context,
	senders []*ecdsa.PrivateKey,
) error {
	var (
		ticker = time.NewTicker(time.Second)
		numTxs = b.config.TPS / uint64(len(senders))

		toAddress = common.HexToAddress(b.config.ToSendAddress)
	)

	defer ticker.Stop()

	gasPrice, err := b.client.SuggestGasPrice(ctx)
	if err != nil {
		return fmt.Errorf("unable to get gas price, %w", err)
	}

	chainID, err := b.client.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("could not get chain id: %w", err)
	}

	signer := types.NewEIP155Signer(chainID)

	for {
		select {
		case <-ctx.Done():
			b.log.Info("Stop signal captured")

			return nil
		case <-ticker.C:
			group, groupCtx := errgroup.WithContext(ctx)

			for _, sender := range senders {
				sender := sender

				group.Go(func() error {
					var (
						senderFrom = crypto.PubkeyToAddress(sender.PublicKey)
					)

					// Fetch the latest account nonce
					nonce, err := b.client.PendingNonceAt(groupCtx, senderFrom)
					if err != nil {
						return fmt.Errorf("could not get pending nonce: %w", err)
					}

					// Each account needs to send a batch of transactions
					for i := uint64(0); i < numTxs; i++ {
						// Construct the transaction
						tx := &types.LegacyTx{
							Nonce:    nonce,
							GasPrice: gasPrice,
							Gas:      defaultEOAGasLimit,
							To:       &toAddress,
							Value:    defaultEOAValue,
							Data:     []byte{},
						}

						// Sign the transaction
						signedTx, err := types.SignTx(
							types.NewTx(tx),
							signer,
							sender,
						)
						if err != nil {
							return fmt.Errorf(
								"unable to sign the transaction: %w",
								err,
							)
						}

						// Send the transaction
						if err := b.client.SendTransaction(ctx, signedTx); err != nil {
							b.log.Error(
								"Transaction send error",
								"err", err,
								"from", senderFrom.String(),
								"nonce", nonce,
								"hash", signedTx.Hash().String(),
							)

							return fmt.Errorf(
								"unable to send transaction, %w",
								err,
							)
						}

						b.log.Debug("Transaction sent",
							"hash", signedTx.Hash().String(),
							"from", senderFrom.String(),
							"nonce", nonce,
						)

						// Increment the nonce.
						// Nonce values are kept locally
						// per sender, to avoid unnecessary fetches
						// on each transaction send
						nonce++
					}

					return nil
				})
			}

			b.log.Debug("Transaction batch sent")
		}
	}
}
