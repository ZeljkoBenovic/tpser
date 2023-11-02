package types

type BlockInfo struct {
	TransactionNum int
	GasLimit       uint64
	GasUsed        uint64
	Hash           string
	Number         uint64
	Time           uint64
}
