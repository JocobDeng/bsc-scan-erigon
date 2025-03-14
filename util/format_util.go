package util

import (
	"encoding/hex"
	"fmt"
	"github.com/ledgerwatch/log/v3"
	"math"
	"math/big"
	"strings"
)

// HexToBigInt 将给定的十六进制字符串转换为*big.Int类型的十进制数。
func HexToBigInt(hexStr string) *big.Int {
	hexStr = strings.TrimRight(hexStr, "0")
	hexStr = strings.TrimPrefix(hexStr, "0x")
	hexStr = strings.TrimPrefix(hexStr, "0X")
	bigInt := new(big.Int)
	bigInt.SetString(hexStr, 16) // 16代表输入字符串是十六进制
	return bigInt
}

func HexToBigIntWith64(b []byte) *big.Int {
	if len(b) > 32 {
		b = b[32:]
	}
	enc := make([]byte, len(b)*2)
	hex.Encode(enc, b)
	bigInt := new(big.Int)
	bigInt.SetString(string(enc), 16) // 16代表输入字符串是十六进制
	return bigInt
}
func HexToBigIntCutByteArr(b []byte) *big.Int {
	if len(b) > 32 {
		b = b[:32]
	}
	var eightBytes [32]byte
	copy(eightBytes[:], b)
	return new(big.Int).SetBytes(eightBytes[:])
}

func HexToBigIntNotCut(b []byte) *big.Int {
	return new(big.Int).SetBytes(b)
}

func Hex2Utf8(addrHex string, bytes []byte) (string, error) {
	// 检查初始偏移量数据是否足够 //0x51537Ea0510883A7a85c3aC25aA42eafA32F3655
	if len(bytes) < 32 {
		return "", fmt.Errorf("字节切片长度小于32字节，无法读取偏移量")
	}
	// 读取偏移量
	offsetBig := big.NewInt(0).SetBytes(bytes[:32])
	offset := uint64(0)
	if offsetBig.IsUint64() {
		offset = offsetBig.Uint64()
		if offset > math.MaxUint64-32 {
			log.Info(fmt.Sprintf("offset > math.MaxUint64 - 32 err addr : %s", addrHex))
			return "", fmt.Errorf("offset > math.MaxUint64 - 32")
		}
	}
	if offset+32 > uint64(len(bytes)) {
		return "", fmt.Errorf("字节切片长度小于偏移量加32字节，无法读取长度")
	}
	length := big.NewInt(0).SetBytes(bytes[offset : offset+32]).Uint64()
	start := offset + 32
	if start > math.MaxUint64-length {
		log.Info(fmt.Sprintf("start > math.MaxUint64-length err addr : %s", addrHex))
		return "", fmt.Errorf("start > math.MaxUint64-length")
	}
	//log.Info(fmt.Sprintf("len(bytes) : %d,start:%d,length:%d math.MaxUint64-length: %d", len(bytes), start, length, math.MaxUint64-length))
	if start+length > uint64(len(bytes)) {
		return "", fmt.Errorf("字节切片长度小于数据的结束位置，无法读取数据")
	}
	data := bytes[start : start+length]
	utf8String := string(data)
	return utf8String, nil
}

func FormatTopics(hexStr string) string {
	if strings.HasPrefix(hexStr, "0x") || strings.HasPrefix(hexStr, "0X") {
		hexStr = hexStr[2:]
	}
	firstNonZeroIndex := -1
	for i, ch := range hexStr {
		if ch != '0' {
			firstNonZeroIndex = i
			break
		}
	}
	var last string
	if firstNonZeroIndex == -1 {
		last = "0"
	} else {
		last = hexStr[firstNonZeroIndex:]
	}
	return "0x" + PadHexLen(last, 40)
}

func PadHexLen(hexStr string, length int) string {
	for len(hexStr) < length {
		hexStr = "0" + hexStr
	}
	return hexStr[:length]
}
