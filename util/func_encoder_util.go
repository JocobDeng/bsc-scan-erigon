package util

import (
	"encoding/hex"
	"fmt"
	"github.com/ledgerwatch/erigon-lib/common/hexutility"
	"github.com/ledgerwatch/erigon/accounts/abi"
	"github.com/ledgerwatch/erigon/crypto"
	"golang.org/x/crypto/sha3"
	"math/big"
	"strings"
)

var (
	methodMapId = map[string]string{
		"name()":                    "06fdde03",
		"symbol()":                  "95d89b41",
		"balanceOf()":               "722713f7",
		"decimals()":                "313ce567",
		"totalSupply()":             "18160ddd",
		"supportsInterface(bytes4)": "01ffc9a7",
	}
)

// EncodeParameters encodes the parameters according to Ethereum ABI rules, including dynamic types.
func EncodeParameters(arguments abi.Arguments, values []interface{}) (string, error) {
	// First, we need to separate static and dynamic types because we need to calculate offsets.
	staticValues := []interface{}{}
	dynamicValues := []interface{}{}
	for i, arg := range arguments.NonIndexed() {
		if arg.Type.T == abi.SliceTy || arg.Type.T == abi.ArrayTy || arg.Type.T == abi.StringTy || arg.Type.T == abi.BytesTy {
			dynamicValues = append(dynamicValues, values[i])
		} else {
			staticValues = append(staticValues, values[i])
		}
	}
	// Calculate the starting offset for dynamic types. It starts after all static types.
	dynamicOffset := len(staticValues) * 32 // 32 bytes for each static value
	// We will collect the head part (offsets and static values) and the tail part (dynamic values) separately.
	var headPart, tailPart []byte

	// Encode static values and calculate offsets for dynamic values.
	for _, value := range staticValues {
		encodedValue, err := arguments.Pack(value)
		if err != nil {
			return "", err
		}
		headPart = append(headPart, encodedValue...)
	}
	for _, value := range dynamicValues {
		// For dynamic values, we need to encode the offset first.
		offsetValue := big.NewInt(int64(dynamicOffset))
		encodedOffset, err := arguments.Pack(offsetValue)
		if err != nil {
			return "", err
		}
		headPart = append(headPart, encodedOffset...)

		// Then, encode the dynamic value itself.
		encodedValue, err := arguments.Pack(value)
		if err != nil {
			return "", err
		}
		tailPart = append(tailPart, encodedValue...)
		// Increase the offset for the next dynamic value.
		dynamicOffset += len(encodedValue)
	}

	// Concatenate the head part and the tail part.
	fullEncoded := append(headPart, tailPart...)
	return hex.EncodeToString(fullEncoded), nil
}

// Function represents a contract function
type Function struct {
	Name             string
	InputParameters  []abi.Argument
	OutputParameters []abi.Argument
}

// EncodeFunction encodes a function call for Ethereum
func EncodeFunction(function Function, values []interface{}) (string, error) {
	//methodSignature := buildMethodSignature(function.Name, function.InputParameters)
	//methodId := BuildMethodId(methodSignature)
	encodedParams, err := EncodeParameters(function.InputParameters, values)
	if err != nil {
		return "", err
	}
	return encodedParams, nil
}

// buildMethodSignature builds the method signature from a function name and input parameters
func buildMethodSignature(name string, params []abi.Argument) string {
	paramTypes := make([]string, len(params))
	for i, param := range params {
		paramTypes[i] = param.Type.String()
	}
	return fmt.Sprintf("%s(%s)", name, strings.Join(paramTypes, ","))
}

func buildMethodId1(functionPrototype string) []byte {
	hash := crypto.Keccak256Hash([]byte(functionPrototype))
	return hash.Bytes()[:4]
}

// BuildMethodId builds the method id from a method signature
func BuildMethodId(methodSignature string) string {
	hash := crypto.Keccak256Hash([]byte(methodSignature))
	return hex.EncodeToString(hash.Bytes()[:4])
}

// buildMethodId2 builds the method id from a method signature
func buildMethodId2(methodSignature string) string {
	hash := sha3.NewLegacyKeccak256()
	hash.Write([]byte(methodSignature))
	return hexutility.Encode(hash.Sum(nil)[:4])
}

func main() {
	// Example usage
	//parameters := []Type{
	//	Uint{big.NewInt(123)},
	//	Uint{big.NewInt(456)},
	//	// Add other types as needed
	//}
	//
	//encodedParams := EncodeParameters(parameters)
	//fmt.Println(encodedParams)
	println(hex.EncodeToString(buildMethodId1("name()")))
	println(BuildMethodId("name()"))
	println(hex.EncodeToString(buildMethodId1("supportsInterface(bytes4)")))
	println(BuildMethodId("supportsInterface(bytes4)"))
	// Example usage
	//bytes4Type, _ := abi.NewType("bytes4", "", nil)
	//boolType, _ := abi.NewType("bool", "", nil)
	//function := Function{
	//	Name: "supportsInterface",
	//	InputParameters: abi.Arguments{
	//		{Type: bytes4Type},
	//	},
	//	OutputParameters: []abi.Argument{
	//		{Type: boolType},
	//	},
	//}
	//// 准备 supportsInterface 函数的参数值，即 ERC-721 接口的ID
	//interfaceID := [4]byte{0x80, 0xac, 0x58, 0xcd}
	//
	//// 使用 EncodeFunction 来编码 supportsInterface 函数调用的数据
	//encodedData, _ := util.EncodeParameters(function.InputParameters, []interface{}{interfaceID})
	//// 输出编码后的数据
	//fmt.Printf("Encoded Function Call: %s\n", encodedData)
	//// 假设 encodedData 是一个字节切片
	//fmt.Printf("Encoded Function Call Data Length (bytes): %d\n", len(encodedData))
}

type EthTransferEnum struct {
	Input   string
	Topics0 string
}

// EthApprovalEnum is a struct to hold the input and topics0 for the event.
type EthApprovalEnum struct {
	Input   string
	Topics0 string
}

// 定义一个枚举实例
var Approve = EthApprovalEnum{
	Input:   "approve(address,uint256)",
	Topics0: "0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925",
}

// EthApprovalEnums is a slice that holds all instances of EthApprovalEnum.
var EthApprovalEnums = []EthApprovalEnum{Approve}

// GetByTopics0 searches the EthApprovalEnums slice for an entry with the matching topics0.
func GetApprovalEventByTopics0(topics0 string) *EthApprovalEnum {
	for _, enum := range EthApprovalEnums {
		if enum.Topics0 == topics0 {
			return &enum
		}
	}
	return nil
}

// IsTransferEvent 检查给定的topics0是否对应一个转账事件
func IsApprovalEvent(topics0 string) bool {
	return GetApprovalEventByTopics0(topics0) != nil
}
