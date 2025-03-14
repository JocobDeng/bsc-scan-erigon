package util

var (
	Transfer       = EthTransferEnum{"Transfer(address,address,uint256)", "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"}
	TransferSingle = EthTransferEnum{"TransferSingle(address,address,address,uint256,uint256)", "0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62"}
	TransferBatch  = EthTransferEnum{"TransferBatch(address,address,address,uint256[],uint256[])", "0x4a39dc06d4c0dbc64b70af90fd698a233a518aa5d07e595d983b8c0526c8f7fb"}
)

var EthTransferEnums = []EthTransferEnum{Transfer, TransferSingle, TransferBatch}

func GetByTopics0(topics0 string) *EthTransferEnum {
	for _, event := range EthTransferEnums {
		if event.Topics0 == topics0 {
			return &event
		}
	}
	return nil
}

// IsTransferEvent 检查给定的topics0是否对应一个转账事件
func IsTransferEvent(topics0 string) bool {
	return GetByTopics0(topics0) != nil
}
