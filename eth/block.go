package eth

type blockInfo struct {
	transactionNum int
	gasLimit       uint64
	gasUsed        uint64
	hash           string
	number         uint64
	time           uint64
}
