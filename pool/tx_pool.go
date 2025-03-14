package pool

import (
	"context"
	"fmt"
	"github.com/erigontech/erigon-lib/kv"
	"github.com/ledgerwatch/log/v3"
)

type TxPool struct {
	pool chan kv.Tx
	db   kv.RwDB
}

func NewTxPool(db kv.RwDB, size int, ctx context.Context) (*TxPool, error) {
	pool := make(chan kv.Tx, size)
	for i := 0; i < size; i++ {
		tx, err := db.BeginRo(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize transaction: %w", err)
		}
		pool <- tx
	}
	log.Info(fmt.Sprintf("scan begin333... %s", len(pool)))

	return &TxPool{pool: pool, db: db}, nil
}

func (tp *TxPool) Get(ctx context.Context) (kv.Tx, error) {
	select {
	case tx := <-tp.pool:
		return tx, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (tp *TxPool) Release(tx kv.Tx) {
	// Simply put the transaction back into the pool for reuse
	tp.pool <- tx
}

func (tp *TxPool) Close() {
	close(tp.pool)
	for tx := range tp.pool {
		tx.Rollback() // Ensure all transactions are properly rolled back
	}
}
