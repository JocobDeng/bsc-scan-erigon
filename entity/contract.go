package entity

import "math/big"

// ContractRo is the Go equivalent of the Java ContractRo class.
type ContractRo struct {
	Address     string   `json:"address"`
	Symbol      string   `json:"symbol"`
	Name        string   `json:"name"`
	Decimals    *big.Int `json:"decimals"`
	TotalSupply *big.Int `json:"totalSupply"`
	IsErc20     bool     `json:"isErc721"`
	IsErc721    bool     `json:"isErc721"`
	IsErc1155   bool     `json:"isErc1155"`
}

// NewContractRo creates a new ContractRo with default values.
func NewContractRo(address string) *ContractRo {
	return &ContractRo{
		Address:     address,
		Symbol:      "",
		Name:        "",
		Decimals:    big.NewInt(0),
		TotalSupply: big.NewInt(0),
		IsErc20:     false,
		IsErc721:    false,
		IsErc1155:   false,
	}
}
