package util

import (
	"fmt"
	"github.com/erigontech/erigon-lib/log/v3"
	"math"
	"strconv"
	"strings"
)

func Uint64ToInt64(value *uint64) int64 {
	if value == nil {
		log.Info("value is nil")
		return 0
	}
	if *value > math.MaxInt64 {
		log.Info(fmt.Sprintf("value is nil : %d", *value))
		return 0
	}
	return int64(*value)
}

func IntListToStrForCk(list []int) string {
	if len(list) == 0 {
		return "[]"
	}

	var strList []string
	for _, num := range list {
		strList = append(strList, "'"+strconv.Itoa(num)+"'")
	}

	return "[" + strings.Join(strList, ", ") + "]"
}
