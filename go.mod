module bsc-scan-erigon

go 1.23.0

toolchain go1.24.0

replace github.com/erigontech/erigon-lib => ./erigon-lib

require (
	github.com/ClickHouse/clickhouse-go/v2 v2.33.0
	github.com/erigontech/erigon v1.9.7-0.20250310084104-eef893abc581
	github.com/erigontech/erigon-lib v1.0.0
	github.com/holiman/uint256 v1.3.2
	github.com/ledgerwatch/log/v3 v3.9.0
	github.com/panjf2000/ants/v2 v2.11.2
	github.com/urfave/cli/v2 v2.27.6
	golang.org/x/crypto v0.36.0
)

require (
	github.com/ClickHouse/ch-go v0.65.1 // indirect
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/errors v0.7.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/paulmach/orb v0.11.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	go.opentelemetry.io/otel v1.35.0 // indirect
	go.opentelemetry.io/otel/trace v1.35.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
