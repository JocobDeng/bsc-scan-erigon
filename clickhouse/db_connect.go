package clickhouse

import (
	"errors"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"strings"
	"time"
)

/*var dialCount = 0*/

func optionsConfig(hostPort, database, username, password string) *clickhouse.Options {
	return &clickhouse.Options{
		Addr: []string{hostPort},
		Auth: clickhouse.Auth{
			Database: database,
			Username: username,
			Password: password,
		},
		/*		DialContext: func(ctx context.Context, addr string) (net.Conn, error) {
					dialCount++
					var d net.Dialer
					return d.DialContext(ctx, "tcp", addr)
				},
				Debug: true,
				Debugf: func(format string, v ...any) {
					fmt.Printf(format+"\n", v...)
				},*/
		Settings: clickhouse.Settings{
			"database_replicated_enforce_synchronous_settings": "1",
			"insert_quorum":                 1,
			"insert_quorum_parallel":        0,
			"select_sequential_consistency": 1,
			"max_execution_time":            60,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		DialTimeout:          time.Duration(30) * time.Second,
		MaxIdleConns:         50,
		MaxOpenConns:         2000,
		ConnMaxLifetime:      time.Duration(10) * time.Minute,
		ConnOpenStrategy:     clickhouse.ConnOpenInOrder,
		BlockBufferSize:      10,
		MaxCompressionBuffer: 10485760,
		ClientInfo: clickhouse.ClientInfo{
			Products: []struct {
				Name    string
				Version string
			}{
				{Name: "geth-demo", Version: "1.0.0"},
			},
		},
		TLS: nil,
	}
}

// GetConnection 连接连接
func GetConnection(dataSource string) (driver.Conn, error) {
	upArgs, hpdArgs, err := parseDataSource(dataSource)
	if err != nil {
		return nil, err
	}
	return clickhouse.Open(optionsConfig(hpdArgs[0], hpdArgs[1], upArgs[0], upArgs[1]))
}

func parseDataSource(dataSource string) ([]string, []string, error) {
	ds := strings.Split(dataSource, "://")
	if ds == nil || len(ds) < 2 {
		return nil, nil, errors.New("dataSource parse error")
	}

	infos := strings.Split(ds[1], "@")
	if infos == nil || len(infos) < 2 {
		return nil, nil, errors.New("dataSource parse error")
	}

	ups := strings.Split(infos[0], ":")
	if ups == nil || len(ups) < 2 {
		return nil, nil, errors.New("dataSource parse error")
	}

	hpds := strings.Split(infos[1], "/")
	if hpds == nil || len(hpds) < 2 || !strings.Contains(hpds[0], ":") {
		return nil, nil, errors.New("dataSource parse error")
	}
	return ups, hpds, nil
}
