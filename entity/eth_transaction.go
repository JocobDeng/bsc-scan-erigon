package entity

import (
	libcommon "github.com/erigontech/erigon-lib/common"
	"github.com/erigontech/erigon-lib/common/hexutil"
	"github.com/erigontech/erigon/core/types"
	"math/big"
)

type EthTransactionVo struct {
	TransactionHash string `json:"transactionHash"` // 交易哈希
	BlockHash       string `json:"blockHash"`       // 区块哈希
	Status          bool   `json:"status"`          // 本次交易的结果 (成功或失败)

	Block              int64             `json:"block"`                    // 交易被存储的区块编号
	Timestamp          int64             `json:"timestamp"`                // 交易被打包的时间
	From               string            `json:"from"`                     // 这笔交易的发起方
	To                 string            `json:"to"`                       // 这笔交易的接收方
	ContractAddress    libcommon.Address `json:"contractAddress"`          // 创建合约的交易中，此字段有值，为合约地址
	TransactionValue   *big.Int          `json:"transactionValue"`         // 交易数量
	TransactionFee     *big.Int          `json:"transactionFee"`           // 交易手续费
	GasBaseFees        *big.Int          `json:"gasBaseFees"`              // 基础费用
	GasMaxPriorityFees *big.Int          `json:"gasMaxPriorityFees"`       // 最大附加小费
	GasPrice           *big.Int          `json:"gasPrice"`                 // 合约Gas价格
	EffectiveGasPrice  *big.Int          `json:"effectiveGasPrice"`        // 为此交易指定的每单位Gas成本
	Gas                *big.Int          `json:"gas"`                      // Gas上限
	GasUsed            *big.Int          `json:"gasUsed,receipt_gas_used"` // Gas消耗 receipt_gas_used
	MinFeePerGas       *big.Int          `json:"minFeePerGas"`             // Gas费-最小附加小费
	MaxFeePerGas       *big.Int          `json:"maxFeePerGas"`             // Gas费-最大手续费  覆盖基础费用和优先费用的总费用
	DestroyFee         *big.Int          `json:"destroyFee"`               // 销毁手续费
	ChangeHandingFee   *big.Int          `json:"changeHandingFee"`         // 手续费找零
	TransactionType    uint8             `json:"transactionType"`          // 交易类型
	Nonce              uint64            `json:"nonce"`                    // 代表此交易是发起者地址发起的第几笔交易
	TransactionIndex   uint32            `json:"transactionIndex"`         // 区块内交易编号
	//InputData         string                 `json:"inputData"`         // 在调用智能合约时，发起方调用的方法与传入的参数
	Input             *hexutil.Bytes
	Function          string                 `json:"function"`          // 功能函数
	MethodId          string                 `json:"methodId"`          // 调用方法
	MethodName        string                 `json:"methodName"`        // 调用方法名称
	HasReceiptLogs    bool                   `json:"hasReceiptLogs"`    // 是否有票据日志
	TokenTransferList *[]*EthTokenTransferVo `json:"tokenTransferList"` // 代币转账

	//TokenTransferList *[]*EthTokenTransferVo `json:"tokenTransferList"` // 代币转账
	Logs *types.Logs `json:"Logs"` //日志

	ReceiptRoot              uint8    `json:"receiptRoot"`              // 交易后的状态根（拜占庭之前） TODO
	ReceiptStatus            uint8    `json:"receiptStatus"`            // 交易状态：1（成功）或 0（失败）（拜占庭之后）
	ReceiptCumulativeGasUsed uint64   `json:"receiptCumulativeGasUsed"` //此交易执行时区块中使用的总 Gas
	MaxFeePerBlobGas         *big.Int `json:"maxFeePerBlobGas"`         //用户愿意支付的每单位 Blob Gas 的最大费用
	BlobVersionedHashes      []string `json:"blobVersionedHashes"`      //kzg_to_versioned_hash 的哈希输出列表
	ReceiptBlobGasPrice      *big.Int `json:"receiptBlobGasPrice"`      //Blob Gas 价格
	ReceiptBlobGasUsed       uint64   `json:"receiptBlobGasUsed"`       //使用的 Blob Gas
}

// EthTokenTransferVo represents an Ethereum token transfer.
type EthTokenTransferVo struct {
	//IsNFT             bool       `json:"isNFT"`             // 是否NFT
	ERC               string     `json:"erc"`               // 代币所属协议 ERC20/ERC721/ERC1155
	TokenID           *big.Int   `json:"tokenId"`           // 代币ID (ERC721/ERC1155)
	Address           string     `json:"address"`           // 代币合约地址
	Symbol            string     `json:"symbol"`            // 代币象征
	Name              string     `json:"name"`              // 代币名称
	Decimals          *big.Int   `json:"decimals"`          // 代币使用的小数位数(计算当前代币)
	TotalSupply       *big.Int   `json:"totalSupply"`       // 代币总供应量
	From              string     `json:"from"`              // 发送方
	To                string     `json:"to"`                // 接收方
	Amount            *big.Int   `json:"amount"`            // 数量
	Topics0           string     `json:"topics0"`           // 事件标识
	LogIdx            uint32     `json:"logIdx"`            // 事件标识
	TransferType      int64      `json:"transferType"`      // 转账方法区分 Transfer-0, TransferBatch-1, TransferSingle-2
	TransferBatchSize int64      `json:"transferBatchSize"` // TransferBatch解析出来的Size
	IDS               []*big.Int `json:"IDS"`               // uint256[] ids
	Amounts           []*big.Int `json:"amounts"`           // uint256[] amounts
	NeedInsert        bool       `json:"needInsert"`        // uint256[] amounts
}

type EthTransactionInner struct {
	TxID         string   `json:"txId"`
	Status       bool     `json:"status"` // Assuming 1 for true and 0 for false
	OwnerAddress string   `json:"ownerAddress"`
	ToAddress    string   `json:"toAddress"`
	Amount       *big.Int `json:"amount"`
	Timestamp    int64    `json:"timestamp"`
	BlockNum     int64    `json:"blockNum"`
	Input        string   `json:"input"`
	Output       string   `json:"output"`
	CallType     string   `json:"callType"`
	TraceType    string   `json:"traceType"`
	Gas          *big.Int `json:"gas"`
	GasUsed      *big.Int `json:"gasUsed"`
	SubTraces    int      `json:"subTraces"`
	TraceAddress string   `json:"traceAddress"`
	TxIdx        int      `json:"txIdx"`
}
