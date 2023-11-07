package conf

import (
	"errors"
	"flag"
	"fmt"
	"log"
)

type Mode string

func (m Mode) String() string {
	return string(m)
}

const (
	BlocksFetcher Mode = "blocks-fetcher"
	LongSender    Mode = "long-sender"
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
	TxSendTimeoutMin int64
	IncludeTPSReport bool
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
	includeTpsReport bool
}

func New() (Conf, error) {
	raw := &rawConf{}
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
	flag.Int64Var(&c.txSendTimeoutMin, "duration", 60, "the number of minutes after witch to stop the send")
	flag.BoolVar(&c.includeTpsReport, "report", false, "set to true to include tps report after the long-sender node")
	flag.StringVar(&c.mnemonic, "mnemonic", "", "mnemonic string to derive accounts from")
	flag.IntVar(&c.totalAccounts, "mnemonic-addr", 1, "total number of account to send transactions from")
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

	return Conf{
		JsonRPC: c.jsonRpc,
		Blocks: Blocks{
			Start: c.blockStart,
			End:   c.blockEnd,
			Range: c.blockRange,
		},
		Mode:             Mode(c.mode),
		PrivateKey:       c.privKey,
		Mnemonic:         c.mnemonic,
		ToAddress:        c.toAddr,
		TxPerSec:         c.txPerSec,
		TxSendTimeoutMin: c.txSendTimeoutMin,
		LogLevel:         c.logLevel,
		IncludeTPSReport: c.includeTpsReport,
		TotalAccounts:    c.totalAccounts,
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

	return nil
}
