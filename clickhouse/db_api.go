package clickhouse

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// CreateTable 创表
func CreateTable(conn driver.Conn, createSql string) error {
	if err := conn.Exec(context.Background(), createSql); err != nil {
		return err
	}
	return nil
}

// GetRowsCount 查询总数
func GetRowsCount(conn driver.Conn, table string) (uint64, error) {
	var count uint64
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s`, table)
	if err := conn.QueryRow(context.Background(), query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// DeduplicateTable OPTIMIZE TABLE PARTITION FINAL DEDUPLICATE BY
// 去重，若没有后续的“BY”子句，则按照行完全相同去重（所有字段值相同）
// “BY”：配合“DEDUPLICATE”关键词使用，指定依据哪些列去重
func DeduplicateTable(conn driver.Conn, table string) error {
	if err := conn.Exec(context.Background(), fmt.Sprintf(`OPTIMIZE TABLE %s DEDUPLICATE`, table)); err != nil {
		return err
	}
	return nil
}
