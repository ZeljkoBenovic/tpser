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

	PrivateKey string
	ToAddress  string

	TxPerSec         int64
	TxSendTimeoutMin int64
	IncludeTPSReport bool
}

type Blocks struct {
	Start int64
	End   int64
}

var (
	ErrJsonRPCNotDefined       = errors.New("json-rpc endpoint not defined")
	ErrEndBlockNotDefined      = errors.New("end block not defined")
	ErrPrivKeyToAddrNotDefined = errors.New("to address and private key must be set")
)

type rawConf struct {
	mode     string
	jsonRpc  string
	logLevel string

	blockStart int64
	blockEnd   int64

	privKey string
	toAddr  string

	txPerSec         int64
	txSendTimeoutMin int64
	includeTpsReport bool
}

func New() (Conf, error) {
	raw := &rawConf{}
	conf, err := raw.getConfig()
	if err != nil {
		log.Fatalln("Could not initialize config", "err", err.Error())
	}

	return conf, nil
}

func (c *rawConf) getConfig() (Conf, error) {
	flag.StringVar(&c.logLevel, "log-level", "info", "log output level")
	flag.StringVar(&c.jsonRpc, "json-rpc", "", "JSON-RPC or WS endpoint")
	flag.Int64Var(&c.blockStart, "block-start", 1, "the start block range")
	flag.Int64Var(&c.blockEnd, "block-end", 0, "the end block range")
	flag.StringVar(&c.privKey, "priv-key", "", "the private key for the sender account")
	flag.StringVar(&c.toAddr, "to", "", "address to which the funds will be sent")
	flag.Int64Var(&c.txPerSec, "tx-per-sec", 100, "the number of transactions per second to send")
	flag.Int64Var(&c.txSendTimeoutMin, "tx-send-timeout", 60, "the number of minutes after witch to stop the send")
	flag.BoolVar(&c.includeTpsReport, "include-tps-report", false, "set to true to include tps report after the long-sender node")
	flag.StringVar(
		&c.mode,
		"mode",
		BlocksFetcher.String(),
		fmt.Sprintf("mode of operation (%s, %s)", BlocksFetcher.String(), LongSender.String()),
	)
	flag.Parse()

	if c.jsonRpc == "" {
		return Conf{}, ErrJsonRPCNotDefined
	}
	if c.mode == BlocksFetcher.String() && c.blockEnd == 0 {
		return Conf{}, ErrEndBlockNotDefined
	}

	if c.mode == LongSender.String() && c.toAddr == "" || c.privKey == "" {
		return Conf{}, ErrPrivKeyToAddrNotDefined
	}

	return Conf{
		JsonRPC: c.jsonRpc,
		Blocks: Blocks{
			Start: c.blockStart,
			End:   c.blockEnd,
		},
		Mode:             Mode(c.mode),
		PrivateKey:       c.privKey,
		ToAddress:        c.toAddr,
		TxPerSec:         c.txPerSec,
		TxSendTimeoutMin: c.txSendTimeoutMin,
		LogLevel:         c.logLevel,
		IncludeTPSReport: c.includeTpsReport,
	}, nil
}
