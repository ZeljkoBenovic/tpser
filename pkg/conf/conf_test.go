package conf

import (
	"testing"
)

func TestFlagValidation(t *testing.T) {
	cnf := rawConf{}
	var jsonRpcFlagTest = []struct {
		name  string
		input string
		want  error
	}{
		{
			name:  "Empty JSON-RPC",
			input: "",
			want:  ErrJsonRPCNotDefined,
		},
		{
			name:  "JSON-RPC provided",
			input: "https://json-rpc.example.com",
			want:  nil,
		},
	}

	var blockFetcherModeFlagTest = []struct {
		name       string
		blockEnd   int64
		blockRange int64
		want       error
	}{
		{
			name:       "BlockStart and BlockRange not defined",
			blockEnd:   0,
			blockRange: 0,
			want:       ErrEndBlockNotDefined,
		},
		{
			name:       "BlockStart defined",
			blockEnd:   100,
			blockRange: 0,
			want:       nil,
		},
		{
			name:       "BlockRange defined",
			blockEnd:   0,
			blockRange: 100,
			want:       nil,
		},
	}

	var longSenderFlagsTest = []struct {
		name     string
		toAddr   string
		privKey  string
		mnemonic string
		want     error
	}{
		{
			name:     "Addr and Key and Mnemonic not provided",
			toAddr:   "",
			privKey:  "",
			mnemonic: "",
			want:     ErrToAddrNotProvided,
		},
		{
			name:     "Addr provided Key and Mnemonic not provided",
			toAddr:   "0x124155436436",
			privKey:  "",
			mnemonic: "",
			want:     ErrPrivKeyOrMnemonicNotProvided,
		},
		{
			name:     "Addr not provided Key provided",
			toAddr:   "",
			privKey:  "jnkdfv-2j42-838yhi9-0u9-0",
			mnemonic: "",
			want:     ErrToAddrNotProvided,
		},
		{
			name:    "Both to addr and key provided",
			toAddr:  "0x21415545435",
			privKey: "fjndksafpj9f[m2-jgfi42-9",
			want:    nil,
		},
		{
			name:     "Addr and mnemonic provided",
			toAddr:   "0x12354135564",
			privKey:  "",
			mnemonic: "test test test",
			want:     nil,
		},
	}

	for _, tt := range jsonRpcFlagTest {
		t.Run(tt.name, func(t *testing.T) {
			cnf.jsonRpc = tt.input
			vErr := cnf.validateRawFlags()
			if vErr != tt.want {
				t.Errorf("json-rpc mandatory test failed")
			}
		})
	}

	for _, tt := range blockFetcherModeFlagTest {
		t.Run(tt.name, func(t *testing.T) {
			cnf.mode = BlocksFetcher.String()
			cnf.blockRange = tt.blockRange
			cnf.blockEnd = tt.blockEnd

			bErr := cnf.validateRawFlags()
			if bErr != tt.want {
				t.Errorf("mandatory blocks-fetcher flags not passed")
			}
		})
	}

	for _, tt := range longSenderFlagsTest {
		t.Run(tt.name, func(t *testing.T) {
			cnf.mode = LongSender.String()
			cnf.toAddr = tt.toAddr
			cnf.privKey = tt.privKey
			cnf.mnemonic = tt.mnemonic

			lErr := cnf.validateRawFlags()
			if lErr != tt.want {
				t.Errorf("long-sender flags test not passed")
				t.Logf("ERR: %s", lErr.Error())
			}
		})
	}

}

func TestDefaultFlags(t *testing.T) {
	var defaultConf = rawConf{
		mode:             BlocksFetcher.String(),
		jsonRpc:          "",
		logLevel:         "info",
		blockStart:       1,
		blockEnd:         0,
		blockRange:       0,
		privKey:          "",
		toAddr:           "",
		txPerSec:         100,
		txSendTimeoutMin: 60,
		includeTpsReport: false,
	}

	conf, err := defaultConf.getConfig(true)
	if err != nil {
		t.Errorf("could not get config: %s", err.Error())
	}

	if conf.Mode != Mode(defaultConf.mode) {
		t.Errorf("got: %s have: %s", conf.Mode.String(), defaultConf.mode)
	}

	if conf.LogLevel != defaultConf.logLevel {
		t.Errorf("got: %s have: %s", conf.LogLevel, defaultConf.logLevel)
	}

	if conf.Blocks.Start != defaultConf.blockStart {
		t.Errorf("got: %d have: %d", conf.Blocks.Start, defaultConf.blockStart)
	}

	if conf.Blocks.End != defaultConf.blockEnd {
		t.Errorf("got: %d have: %d", conf.Blocks.End, defaultConf.blockEnd)
	}

	if conf.Blocks.Range != defaultConf.blockRange {
		t.Errorf("got: %d have: %d", conf.Blocks.Range, defaultConf.blockRange)
	}

	if conf.TxPerSec != defaultConf.txPerSec {
		t.Errorf("got: %d have: %d", conf.TxPerSec, defaultConf.txPerSec)
	}

	if conf.TxSendTimeoutMin != defaultConf.txSendTimeoutMin {
		t.Errorf("got: %d have: %d", conf.TxSendTimeoutMin, defaultConf.txSendTimeoutMin)
	}

	if conf.IncludeTPSReport != defaultConf.includeTpsReport {
		t.Errorf("got: %t have: %t", conf.IncludeTPSReport, defaultConf.includeTpsReport)
	}
}
