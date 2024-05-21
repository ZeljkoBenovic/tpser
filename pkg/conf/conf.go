package conf

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"
)

type Mode string

func (m Mode) String() string {
	return string(m)
}

const (
	BlocksFetcher Mode = "blocks-fetcher"
	LongSender    Mode = "long-sender"
	TxInfo        Mode = "tx-info"
)

type Conf struct {
	Mode Mode

	JsonRPC  string
	Blocks   Blocks
	LogLevel string

	PrivateKey    string
	Mnemonic      string
	TotalAccounts int
	ToAddress     string

	TxPerSec         int64
	TxSendInterval   int64
	TxSendTimeoutMin int64
	IncludeTPSReport bool

	WaitForConfirm        bool
	WaitForConfirmTimeout int64

	TxHashes    []string
	TxCostInEth bool

	StartingNonce *int64

	MetricsPort string
}

type Blocks struct {
	Start int64
	End   int64
	Range int64
}

var (
	ErrJsonRPCNotDefined            = errors.New("json-rpc endpoint not defined")
	ErrEndBlockNotDefined           = errors.New("end block or block range not defined")
	ErrToAddrNotProvided            = errors.New("to address not provided")
	ErrPrivKeyOrMnemonicNotProvided = errors.New("private key or mnemonic not provided")
	ErrTxHashNotProvided            = errors.New("transaction hash must be provided")
)

type rawConf struct {
	mode     string
	jsonRpc  string
	logLevel string

	blockStart int64
	blockEnd   int64
	blockRange int64

	privKey       string
	mnemonic      string
	totalAccounts int
	toAddr        string

	txPerSec         int64
	txSendTimeoutMin int64
	txSendInterval   int64
	includeTpsReport bool

	waitForConfirm        bool
	waitForConfirmTimeout int64

	txHash      string
	txHashes    []string
	txCostInEth bool

	metricsPort string
}

func New() (Conf, error) {
	raw := &rawConf{
		txHashes: make([]string, 0),
	}
	conf, err := raw.getConfig(false)
	if err != nil {
		log.Fatalln("Could not initialize config: ", err.Error())
	}

	return conf, nil
}

func (c *rawConf) getConfig(test bool) (Conf, error) {
	flag.StringVar(&c.logLevel, "log-level", "info", "log output level")
	flag.StringVar(&c.jsonRpc, "json-rpc", "", "JSON-RPC or WS endpoint")
	flag.Int64Var(&c.blockStart, "block-start", 1, "the start block range")
	flag.Int64Var(&c.blockEnd, "block-end", 0, "the end block range")
	flag.Int64Var(&c.blockRange, "block-range", 0, "the range of blocks to fetch from latest")
	flag.StringVar(&c.privKey, "pk", "", "the private key for the sender account")
	flag.StringVar(&c.toAddr, "to", "", "address to which the funds will be sent")
	flag.Int64Var(&c.txPerSec, "tps", 100, "the number of transactions per second to send")
	flag.Int64Var(&c.txSendInterval, "tx-sec", 1, "the number of seconds to wait between sending transactions")
	flag.Int64Var(&c.txSendTimeoutMin, "duration", 60, "the number of minutes after witch to stop the send")
	flag.BoolVar(&c.includeTpsReport, "report", false, "set to true to include tps report after the long-sender node")
	flag.BoolVar(&c.waitForConfirm, "confirm", false, "wait for transactions to be confirmed")
	flag.Int64Var(&c.waitForConfirmTimeout, "confirm-timeout", 10, "wait for tx confirmation timeout in minutes")
	flag.StringVar(&c.mnemonic, "mnemonic", "", "mnemonic string to derive accounts from")
	flag.IntVar(&c.totalAccounts, "mnemonic-addr", 1, "total number of account to send transactions from")
	flag.StringVar(&c.txHash, "tx-hashes", "", "comma delimited transaction hashes to get details for")
	flag.BoolVar(&c.txCostInEth, "tx-cost-eth", false, "present transaction costs in wei instead of eth")
	flag.StringVar(&c.metricsPort, "metrics-port", "3000", "port where the prometheus metrics will be exposed")
	flag.StringVar(
		&c.mode,
		"mode",
		BlocksFetcher.String(),
		fmt.Sprintf("mode of operation (%s, %s)", BlocksFetcher.String(), LongSender.String()),
	)
	flag.Parse()

	if !test {
		if err := c.validateRawFlags(); err != nil {
			return Conf{}, err
		}
	}

	c.processFlags()

	return Conf{
		JsonRPC: c.jsonRpc,
		Blocks: Blocks{
			Start: c.blockStart,
			End:   c.blockEnd,
			Range: c.blockRange,
		},
		Mode:                  Mode(c.mode),
		PrivateKey:            c.privKey,
		Mnemonic:              c.mnemonic,
		ToAddress:             c.toAddr,
		TxPerSec:              c.txPerSec,
		TxSendInterval:        c.txSendInterval,
		TxSendTimeoutMin:      c.txSendTimeoutMin,
		LogLevel:              c.logLevel,
		IncludeTPSReport:      c.includeTpsReport,
		TotalAccounts:         c.totalAccounts,
		WaitForConfirm:        c.waitForConfirm,
		WaitForConfirmTimeout: c.waitForConfirmTimeout,
		TxHashes:              c.txHashes,
		TxCostInEth:           c.txCostInEth,
		MetricsPort:           c.metricsPort,
	}, nil
}

func (c *rawConf) validateRawFlags() error {
	if c.jsonRpc == "" {
		return ErrJsonRPCNotDefined
	}
	if c.mode == BlocksFetcher.String() && c.blockEnd == 0 && c.blockRange == 0 {
		return ErrEndBlockNotDefined
	}

	if c.mode == LongSender.String() {
		if c.toAddr == "" {
			return ErrToAddrNotProvided
		}

		if c.privKey == "" && c.mnemonic == "" {
			return ErrPrivKeyOrMnemonicNotProvided
		}
	}

	if c.mode == TxInfo.String() && c.txHash == "" {
		return ErrTxHashNotProvided
	}

	return nil
}

func (c *rawConf) processFlags() {
	rawHashes := strings.Split(strings.TrimSpace(c.txHash), ",")
	txHashes := make([]string, 0)
	c.txHashes = append(txHashes, rawHashes...)
}
