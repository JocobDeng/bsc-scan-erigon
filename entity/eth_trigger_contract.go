package entity

type ContractTransaction struct {
	TxID            string
	OwnerAddress    string
	ContractAddress string
	MethodID        string
	Timestamp       int64
	BlockNum        int64
	Code            string
	CreateType      int // 1: 非内部交易创建 2: 内部交易创建
}
