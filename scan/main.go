package main

import (
	"context"
	"database/sql/driver"
	"fmt"
	"github.com/c2h5oh/datasize"
	"github.com/erigontech/erigon-lib/chain"
	libcommon "github.com/erigontech/erigon-lib/common"
	"github.com/erigontech/erigon-lib/common/cmp"
	"github.com/erigontech/erigon-lib/common/datadir"
	"github.com/erigontech/erigon-lib/common/dbg"
	"github.com/erigontech/erigon-lib/common/hexutil"
	"github.com/erigontech/erigon-lib/kv"
	"github.com/erigontech/erigon-lib/kv/kvcache"
	"github.com/erigontech/erigon-lib/kv/mdbx"
	"github.com/erigontech/erigon-lib/log/v3"
	"github.com/erigontech/erigon/cmd/rpcdaemon/cli"
	"github.com/erigontech/erigon/cmd/rpcdaemon/cli/httpcfg"
	"github.com/erigontech/erigon/core/rawdb"
	"github.com/erigontech/erigon/core/types"
	"github.com/erigontech/erigon/node/nodecfg"
	"github.com/erigontech/erigon/rpc"
	"github.com/erigontech/erigon/turbo/adapter/ethapi"
	"github.com/erigontech/erigon/turbo/jsonrpc"
	"github.com/erigontech/erigon/turbo/logging"
	"github.com/erigontech/erigon/turbo/snapshotsync/freezeblocks"
	"github.com/holiman/uint256"
	"github.com/panjf2000/ants/v2"
	urCli "github.com/urfave/cli/v2"
	"golang.org/x/sync/semaphore"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"bsc-scan-erigon/entity"
	"bsc-scan-erigon/pool"
	ethScanUtil "bsc-scan-erigon/util"
)

type SAny = []any

const (
	Step        = 100
	WorkerCount = 1
	//Empty                    = ""
	//Datasource               = "clickhouse://bitsky:PuVAroOHei#YRjqg@cc-gs5019c79btc46y5l.clickhouse.ads.aliyuncs.com:3306/default"
	//ethTransactionInnerTable = "eth_transaction_inner"
	//
	//TransferTable    = "eth_transaction"
	//TransferLogTable = "eth_transaction_log"
	//
	//ethTriggerContractsTable = "eth_trigger_contracts"
	//ethTransactionInnerSql   = `INSERT INTO ` +
	//	ethTransactionInnerTable +
	//	` (tx_id,status,owner_address,to_address,amount,timestamp,block_num, call_type,trace_type,gas,gas_used,sub_traces,trace_address,tx_idx)`
	//ethTriggerContractsSql = `INSERT INTO ` +
	//	ethTriggerContractsTable +
	//	` (tx_id,status,owner_address,contract_address,method_id,timestamp,block_num,code,create_type )`
	//TransferSql = `INSERT INTO ` +
	//	TransferTable +
	//	`(tx_id, status, owner_address, to_address, method_id, timestamp , block_num, amount,
	//    nonce,tx_idx,fee, gas_base_fees, gas_max_priority_fees, gas_max_fees, gas_price, gas)`
	//TransferLogSql = `INSERT INTO ` +
	//	TransferLogTable +
	//	` (tx_id,block_num,log_idx,tx_idx,timestamp,address,data,topics,removed)`
)

var (
	conn        driver.Conn
	dbDir       = "/data/bsc/bscDataDir"
	dbPath      = dbDir + "/chaindata"
	api         *jsonrpc.TraceAPIImpl
	blockReader *freezeblocks.BlockReader
	logger      log.Logger
	ctx         context.Context
	dirs        datadir.Dirs
	poolSize    = int(Step) * WorkerCount
	//// WorkerCount * DBTableWorkerCount * DBWorkerCount * DBBatchSize
	//readEthWorkerCount = 5 * Step
	//DBTableWorkerCount = 2
	////DBWorkerCount      = 4
	//DBWorkerCount = 16
	//DBBatchSize   = 5120
	////DBBatchSize              = 10240

	//savePools          *ants.Pool
	//saveTablePools     *ants.Pool
	//readEthWorkerPools *ants.Pool
	txPool *pool.TxPool

	transferInnerMutex       sync.Mutex
	innerCreateContractMutex sync.Mutex
	createContractMutex      sync.Mutex
	transferLogMutex         sync.Mutex
	getEncodeCodeMutex       sync.Mutex

	stateCache *kvcache.Coherent
	db         kv.RwDB
	temporalDb kv.TemporalRoDB

	transactionAndReceiptCountMismatchCounter int32
	innerTransCounter                         int64
	triggerContractTransCounter               int64
	innerTriggerCodeMapSize                   int32

	minusPoolSize = 500
	addPoolSize   = 100
	poolMinSize   = Step * (WorkerCount - 8)
	poolMaxSize   = Step * WorkerCount

	roTxLimit = int64(poolMaxSize*10) + 1

	mainnetConfig = &chain.Config{
		ChainID:             big.NewInt(1),
		HomesteadBlock:      big.NewInt(1150000),
		DAOForkBlock:        big.NewInt(1920000),
		ByzantiumBlock:      big.NewInt(4370000),
		ConstantinopleBlock: big.NewInt(7280000),
		PetersburgBlock:     big.NewInt(7280000), // Assuming PetersburgBlock is the same as ConstantinopleBlock
		IstanbulBlock:       big.NewInt(9069000),
		MuirGlacierBlock:    big.NewInt(9200000),
		BerlinBlock:         big.NewInt(12244000),
		LondonBlock:         big.NewInt(12965000),
		ArrowGlacierBlock:   big.NewInt(13773000),
		GrayGlacierBlock:    big.NewInt(15050000),
	}
	innerTriggerCodeMap sync.Map

	batchSize = 20480
)

func init() {

	ctx = context.Background()
	//conn, _ = clickhouse.GetConnection(Datasource)

	app := &urCli.App{Flags: logging.Flags}
	cliCtx := urCli.NewContext(app, nil, nil)
	cliCtx.Set(logging.LogDirDisableFlag.Name, "true")
	cliCtx.Set(logging.LogDirPathFlag.Name, "/root/go-project/bsc-scan-erigon/logs")
	logger = logging.SetupLoggerCtx("scanInternalTxn", cliCtx, log.LvlInfo, log.LvlInfo, true)
	logger.SetHandler(LvlFilterHandler(log.LvlInfo, log.StdoutHandler))

	logger.Info(fmt.Sprintf("LogDirDisableFlag111 : %s", cliCtx.String(logging.LogDirDisableFlag.Name)))
	logger.Info(fmt.Sprintf("LogDirPathFlag : %s", cliCtx.String(logging.LogDirPathFlag.Name)))
	logger.Info(fmt.Sprintf("LogDirPathFlag : %s", cliCtx.String("datadir")))

	dirs = datadir.New(dbDir)

	api = getApi()

	//go func() {
	//	log.Info("ListenAndServe......  ")
	//	log.Info(fmt.Sprintf("ListenAndServe......  %+v", http.ListenAndServe("0.0.0.0:55577", nil)))
	//}()
}

func LvlFilterHandler(maxLvl log.Lvl, h log.Handler) log.Handler {
	return log.FilterHandler(func(r *log.Record) (pass bool) {
		return r.Lvl <= maxLvl
	}, h)
}

func main() {

	pools, _ := ants.NewPool(100, ants.WithPreAlloc(false))
	outPools, _ := ants.NewPool(1, ants.WithPreAlloc(true))

	var begin = uint64(400000001)
	var end = uint64(400000002)

	Mode(begin, end, outPools, pools)

}

func Mode(begin uint64, end uint64, outPools *ants.Pool, pools *ants.Pool) {
	//var consumerWg sync.WaitGroup
	// 启动消费者 goroutine
	//startConsumer := func(ch chan [][]any, sql string, table string) {
	//	defer consumerWg.Done()
	//	buffer := make([][]any, 0, batchSize)
	//	for data := range ch {
	//		buffer = append(buffer, data...)
	//		if len(buffer) >= batchSize {
	//			now := time.Now()
	//			//BatchSave(buffer[:batchSize], sql, doSave(), begin, end)
	//			log.Debug(fmt.Sprintf("BatchSave inin   【%d - %d】size : %d, table : %s 执行耗时: %d ms ",
	//				begin, end, len(buffer), table, time.Now().Sub(now).Milliseconds()))
	//			//buffer = buffer[batchSize:]
	//			//buffer = append(make([][]any, 0, batchSize), buffer...)
	//			// 使用 copy 函数将剩余数据移动到切片的开头
	//			bufferRemain := copy(buffer, buffer[batchSize:])
	//			buffer = buffer[:bufferRemain]
	//		}
	//	}
	//	// 处理剩余数据
	//	if len(buffer) > 0 {
	//		now := time.Now()
	//		//BatchSave(buffer, sql, doSave(), begin, end)
	//		log.Debug(fmt.Sprintf("BatchSave out  【%d - %d】size : %d, 执行耗时: %d ms ", begin, end, len(buffer), time.Now().Sub(now).Milliseconds()))
	//	}
	//	buffer = nil
	//}
	//consumerWg.Add(4)
	//go startConsumer(transferInnerCh, ethTransactionInnerSql, ethTransactionInnerTable)
	//go startConsumer(createContractCh, ethTriggerContractsSql, ethTriggerContractsTable)
	//go startConsumer(transferCh, TransferSql, TransferTable)
	//go startConsumer(transferLogCh, TransferLogSql, TransferLogTable)

	// 分批次扫描集合数据一起入库
	for start := begin; start < end; {
		next := start + Step
		if next > end {
			next = end
		}
		fStart := start
		fNext := next
		_ = outPools.Submit(func() {
			//defer wg.Done()
			doBatchScan(fStart, fNext)
		})
		start = next
	}
	// 等待所有扫描任务完成后关闭通道
	time.Sleep(time.Second * 1)
	//go func() {
	//	wg.Wait()
	//	log.Info("wg wait finish....")
	//	close(transferInnerCh)
	//	close(transferCh)
	//	close(transferLogCh)
	//	close(createContractCh)
	//}()
	//consumerWg.Wait()
	log.Info("consumerWg wait finish....")
}

func doBatchScan(begin, end uint64) {
	log.Info(fmt.Sprintf("#######ants  doBatchScan2222 inner、log、triggerContract、 txn 【 %d - %d 】#######\n", begin, end))
	var now = time.Now()
	//var producerWg sync.WaitGroup
	// 生产者任务
	for i := begin; i < end; i++ {

		b := i
		var (
			transferInnerDbs            [][]any
			createContractInnerTransDbs [][]any
			transferDbs                 [][]any
			transferLogDbs              [][]any
			createContractTransDbs      [][]any
		)
		doScan(&transferInnerDbs, &createContractInnerTransDbs, &transferDbs, &transferLogDbs, &createContractTransDbs, b)
	}
	//producerWg.Wait()
	log.Info(fmt.Sprintf("doBatchScan2222 inner、log、triggerContract、 txn 【 %d - %d 】cost :  %d ms", begin, end, time.Now().Sub(now).Milliseconds()))
}

func doScan(allTransferInnerDbs *[][]any, allCreateContractTransferDbs *[][]any,
	allTransferDbs *[][]any, allTransferLogDbs *[][]any, allCreateContractTransDbs *[][]any, blockNumber uint64) {
	var now = time.Now()
	tx, err := txPool.Get(ctx)
	if err != nil {
		log.Info(fmt.Sprintf("txPool.Get err : %+v", err))
		return
	}
	defer txPool.Release(tx)

	//获取区块hash
	blkHash, err := rawdb.ReadCanonicalHash(tx, blockNumber)
	if err != nil {
		log.Info(fmt.Sprintf("ReadCanonicalHash err : %+v", err))
	}
	//获取区块体
	//blkBody1, _, _ := rawdb.ReadBody(tx, blkHash, blockNumber)
	blkBody, err := blockReader.BodyWithTransactions(ctx, tx, blkHash, blockNumber)
	if err != nil {
		log.Info(fmt.Sprintf("BodyWithTransactions err  : %+v", err))
		return
	}
	if blkBody == nil {
		log.Info(fmt.Sprintf("ReadBody nil,blockNumber: %d ", blockNumber))
		return
	}
	//获取区块时间
	blkHeader, err := blockReader.Header(ctx, tx, blkHash, blockNumber)
	if err != nil {
		return
	}
	if blkHeader == nil {
		log.Info("blkHeader is nil")
		return
	}

	//获取收据
	receipts := rawdb.ReadRawReceipts(tx, blockNumber)
	if receipts == nil || len(receipts) == 0 {
		log.Info(fmt.Sprintf("receipts : %s,blockNum:%d", receipts, blockNumber))
		return
	}
	if err := receipts.DeriveFields(blkHash, blockNumber, blkBody.Transactions, blkBody.SendersFromTxs()); err != nil {
		log.Error("Failed to derive block receipts fields", "hash", blkHash, "number", blockNumber, "err", err, "stack", dbg.Stack())
		log.Info(fmt.Sprintf("receipts size : %d blkBody size : %d", len(receipts), len(blkBody.Transactions)))
		if len(receipts) != len(blkBody.Transactions) {
			atomic.AddInt32(&transactionAndReceiptCountMismatchCounter, 1)
		}
		return
	}
	//tx hash - receipt map
	receiptsMap := make(map[libcommon.Hash]*types.Receipt)
	for _, receipt := range receipts {
		receiptsMap[receipt.TxHash] = receipt
	}
	//tx hash - transaction map
	transactionMap := make(map[libcommon.Hash]*entity.EthTransactionVo)
	for txIndex, tx := range blkBody.Transactions {
		transactionMap[tx.Hash()] = parseTransaction(&tx, blkHash, blockNumber, blkHeader, txIndex, receiptsMap[tx.Hash()])
	}
	//获取整个块的内部交易
	//traces := make([]jsonrpc.ParityTrace, 0)
	traces, err := api.Block(ctx, rpc.BlockNumber(blockNumber), new(bool), nil)
	if err != nil {
		logger.Info(fmt.Sprintf("tarace  err: %+v", err))
	}
	log.Debug(fmt.Sprintf("doScan readData: 执行耗时: %d ms ", time.Now().Sub(now).Milliseconds()))

	now = time.Now()
	//追加这批区块的内部交易信息
	contractAddrCodeMap := appendInnerTrans(traces, &transactionMap, &receiptsMap, blkHeader, allTransferInnerDbs, allCreateContractTransferDbs)
	//追加这批区块的外部交易信息
	appendTrans(&transactionMap, contractAddrCodeMap, allTransferDbs, allTransferLogDbs, allCreateContractTransDbs)
	log.Debug(fmt.Sprintf("doScan append readData: 执行耗时: %d ms ", time.Now().Sub(now).Milliseconds()))
}

func parseTransaction(tx *types.Transaction, blkHash libcommon.Hash, blockNumber uint64, blkHeader *types.Header,
	txIndex int, receipt *types.Receipt) *entity.EthTransactionVo {
	if tx == nil {
		log.Info("tx is nil")
		return nil
	}
	wrapperTx := ethapi.NewRPCTransaction(*tx, blkHash, blockNumber, uint64(txIndex), blkHeader.BaseFee)
	transaction := &entity.EthTransactionVo{
		TransactionHash: (*tx).Hash().String(),
		Block:           int64(blockNumber),
		Timestamp:       int64(blkHeader.Time),
		From:            hexutil.Encode(wrapperTx.From.Bytes()),
		Nonce:           uint64(wrapperTx.Nonce),
		TransactionFee:  new(big.Int),
	}
	if (*tx).GetTo() != nil {
		transaction.To = hexutil.Encode((*tx).GetTo().Bytes()) //是空？ 0x55baf8abf10f41fc0fa75f87bdb90013cd2c185e949f07aa15937474090eefee
	}
	if receipt != nil {
		transaction.Status = receipt.Status == 1
		transaction.GasUsed = new(big.Int).SetUint64(receipt.GasUsed)
		transaction.TransactionType = receipt.Type
		transaction.TransactionIndex = uint32(receipt.TransactionIndex)
		if receipt.ContractAddress != (libcommon.Address{}) {
			transaction.ContractAddress = receipt.ContractAddress
		}
	}
	if wrapperTx.Value != nil {
		transaction.TransactionValue = wrapperTx.Value.ToInt()
	} else {
		transaction.TransactionValue = big.NewInt(0)
	}
	if len(wrapperTx.Input) >= 4 { //f39b5b9b00000000000000000000000000000000000000000000001d2a1c3028201f4f97000000000000000000000000000000000000000000000000000000005eb022cf
		transaction.MethodId = hexutil.Encode(wrapperTx.Input[:4])
	}
	transaction.Input = &wrapperTx.Input
	//设置相关费用属性
	setFee(transaction, wrapperTx, receipt, tx, blkHeader, blockNumber)

	if receipt.Logs != nil && len(receipt.Logs) > 0 {
		transaction.Logs = &receipt.Logs
	}
	log.Debug(fmt.Sprintf("扫描 transaction........【%+v】", transaction))
	log.Debug(fmt.Sprintf("扫描 receipt........【%+v】", receipt))
	return transaction
}

func setFee(transaction *entity.EthTransactionVo, wrapperTx *ethapi.RPCTransaction, receipt *types.Receipt, txn *types.Transaction, blkHeader *types.Header, blockNumber uint64) {
	transaction.GasBaseFees = blkHeader.BaseFee
	if wrapperTx.MaxPriorityFeePerGas != nil {
		transaction.GasMaxPriorityFees = wrapperTx.MaxPriorityFeePerGas.ToInt()
	}
	if wrapperTx.MaxFeePerGas != nil {
		transaction.MaxFeePerGas = wrapperTx.MaxFeePerGas.ToInt()
	}
	if !mainnetConfig.IsLondon(blockNumber) {
		transaction.EffectiveGasPrice = (*txn).GetTipCap().ToBig()
	} else {
		baseFeeInt, _ := uint256.FromBig(blkHeader.BaseFee)
		gasPrice := new(big.Int).Add(blkHeader.BaseFee, (*txn).GetEffectiveGasTip(baseFeeInt).ToBig())
		transaction.EffectiveGasPrice = gasPrice
	}
	if receipt != nil {
		transaction.GasUsed = new(big.Int).SetUint64(receipt.GasUsed)
	}
	if transaction.GasUsed != nil {
		if transaction.EffectiveGasPrice != nil {
			transaction.TransactionFee = new(big.Int).Mul(transaction.EffectiveGasPrice, transaction.GasUsed)
		} else if wrapperTx.GasPrice != nil {
			transaction.TransactionFee = new(big.Int).Mul(wrapperTx.GasPrice.ToInt(), transaction.GasUsed)
		}
		// EIP-1559 提出的交易类型
		if transaction.TransactionType == types.DynamicFeeTxType {
			if transaction.EffectiveGasPrice != nil && wrapperTx.GasPrice != nil {
				transaction.DestroyFee = new(big.Int).Mul(new(big.Int).Sub(transaction.EffectiveGasPrice, wrapperTx.GasPrice.ToInt()), transaction.GasUsed)
			}
			if wrapperTx.MaxFeePerGas != nil && transaction.EffectiveGasPrice != nil {
				transaction.ChangeHandingFee = new(big.Int).Mul(new(big.Int).Sub(wrapperTx.MaxFeePerGas.ToInt(), transaction.EffectiveGasPrice), transaction.GasUsed)
			}
		}
	}

	if wrapperTx.GasPrice != nil {
		transaction.MinFeePerGas = wrapperTx.GasPrice.ToInt()
		transaction.GasPrice = wrapperTx.GasPrice.ToInt()
	}
	transaction.Gas = new(big.Int).SetUint64(uint64(wrapperTx.Gas))

}

func appendInnerTrans(traces jsonrpc.ParityTraces, transactionMap *map[libcommon.Hash]*entity.EthTransactionVo, receiptsMap *map[libcommon.Hash]*types.Receipt, blkHeader *types.Header, allTransferInnerDbs *[][]any, allCreateContractTransferDbs *[][]any) *map[string]string {
	var (
		innerTransferDbs    []SAny
		createContractDbs   []SAny
		contractAddrCodeMap = make(map[string]string)
	)
	if len(traces) == 0 {
		logger.Info("taraces size = 0")
		return &contractAddrCodeMap
	}
	for _, trace := range traces {
		if trace.TransactionHash == nil {
			continue
		}
		txHash := *trace.TransactionHash
		txnVo := (*transactionMap)[txHash]
		if len(*txnVo.Input) == 0 {
			logger.Debug(fmt.Sprintf("not have inner txn, txHash : %s", txHash))
			continue
		}
		innerTradeBeans, createTradeBeans, contractTriggerTrans := convertInnerTradeBeans(trace, (*receiptsMap)[txHash], txnVo, blkHeader.Time)
		if innerTradeBeans != nil {
			innerTransferDbs = append(innerTransferDbs, innerTradeBeans)
		}
		if createTradeBeans != nil {
			createContractDbs = append(createContractDbs, createTradeBeans)
		}
		if contractTriggerTrans != nil && contractTriggerTrans.ContractAddress != "" {
			contractAddrCodeMap[contractTriggerTrans.ContractAddress] = contractTriggerTrans.Code
		}
	}
	//transferInnerMutex.Lock()
	*allTransferInnerDbs = append(*allTransferInnerDbs, innerTransferDbs...)
	//transferInnerMutex.Unlock()
	//innerCreateContractMutex.Lock()
	*allCreateContractTransferDbs = append(*allCreateContractTransferDbs, createContractDbs...)
	//innerCreateContractMutex.Unlock()
	return &contractAddrCodeMap
}

func convertInnerTradeBeans(trace jsonrpc.ParityTrace, receipt *types.Receipt, txnVo *entity.EthTransactionVo,
	timestamp uint64) (SAny, SAny, *entity.ContractTransaction) {
	if trace.Action == nil {
		log.Info("trace.Action is nil")
		return nil, nil, nil
	}
	var (
		from            libcommon.Address
		to              libcommon.Address
		contractAddrStr string
		amount          *big.Int
		gasUsed         *big.Int
		gas             *big.Int
		callType        string
		encodeCode      string
		methodId        string
		rewardType      string
	)
	//本笔交易是外部创建合约的交易，不必记录，同时返回编码后的 合约字节码
	var onlyCreateContractInner = txnVo.To != ""

	if trace.Result != nil {
		switch trace.Result.(type) {
		case *jsonrpc.TraceResult:
			gasUsed = trace.Result.(*jsonrpc.TraceResult).GasUsed.ToInt()
		case *jsonrpc.CreateTraceResult:
			gasUsed = trace.Result.(*jsonrpc.CreateTraceResult).GasUsed.ToInt()
			encodeCode = getInnerEncodeContractCode(trace.Result.(*jsonrpc.CreateTraceResult).Code.String())
			contractAddrStr = hexutil.Encode(trace.Result.(*jsonrpc.CreateTraceResult).Address.Bytes())
		}
	}
	switch action := trace.Action.(type) {
	case *jsonrpc.CallTraceAction:
		from = action.From
		to = action.To
		amount = action.Value.ToInt()
		callType = action.CallType
		gas = action.Gas.ToInt()
	case *jsonrpc.CreateTraceAction:
		from = action.From
		gas = action.Gas.ToInt()
		if len(action.Init) >= 4 {
			methodId = hexutil.Encode(action.Init[:4])
		}
		amount = action.Value.ToInt()
	case *jsonrpc.SuicideTraceAction:
		from = action.Address
		to = action.RefundAddress
		amount = action.Balance.ToInt()
	case *jsonrpc.RewardTraceAction:
		from = action.Author
		amount = action.Value.ToInt()
		rewardType = action.RewardType
	}
	fromStr := ""
	toStr := ""
	if from != (libcommon.Address{}) {
		fromStr = hexutil.Encode(from.Bytes())
	}
	if to != (libcommon.Address{}) {
		toStr = hexutil.Encode(to.Bytes())
	}
	traceAddrArr := make([]int, len(trace.TraceAddress))
	if len(trace.TraceAddress) > 0 {
		for i, traceAddr := range trace.TraceAddress {
			traceAddrArr[i] = traceAddr
		}
	}
	innerTrade := SAny{
		trace.TransactionHash.String(),
		receipt.Status == 1,
		fromStr, toStr,
		amount,
		timestamp,
		int64(*trace.BlockNumber),
		callType,
		trace.Type,
		rewardType,
		gas,
		gasUsed,
		trace.Subtraces,
		traceAddrArr,
		uint32(*trace.TransactionPosition),
	}
	if _, ok := trace.Action.(*jsonrpc.CreateTraceAction); ok {
		if onlyCreateContractInner {
			createContractTrade := SAny{
				trace.TransactionHash.String(),
				receipt.Status == 1,
				fromStr,
				contractAddrStr,
				methodId,
				timestamp,
				int64(*trace.BlockNumber),
				encodeCode,
				2,
			}
			return innerTrade, createContractTrade, nil
		}
		// 外部交易的创建合约的信息
		return innerTrade, nil, &entity.ContractTransaction{
			ContractAddress: contractAddrStr,
			Code:            encodeCode,
		}
	}
	return innerTrade, nil, nil
}

func getInnerEncodeContractCode(code string) string {
	if len(code) == 0 {
		return ""
	}
	if enCodeCode, ok := innerTriggerCodeMap.Load(code); ok {
		return enCodeCode.(string)
	}
	return doGetEncodeCode(code)
}

func doGetEncodeCode(code string) string {
	getEncodeCodeMutex.Lock()
	defer getEncodeCodeMutex.Unlock()
	if enCodeCode, ok := innerTriggerCodeMap.Load(code); ok {
		return enCodeCode.(string)
	}
	encodeCode := ethScanUtil.EncodeContractStrCode(code)
	innerTriggerCodeMap.Store(code, encodeCode)
	atomic.AddInt32(&innerTriggerCodeMapSize, 1)
	return encodeCode
}

func appendTrans(transactionMap *map[libcommon.Hash]*entity.EthTransactionVo, codeMap *map[string]string,
	allTransferDbs *[][]any, allTransferLogDbs *[][]any, allCreateContractTransDbs *[][]any) {
	var (
		transferDbs            []SAny
		transferLogDbs         []SAny
		createContractTransDbs []SAny
	)
	// 遍历 map
	for _, transaction := range *transactionMap {
		transferDbs = append(transferDbs, convertTransBeans(transaction))
		ethLogs := convertEthLogBeans(transaction)
		if ethLogs != nil && len(ethLogs) > 0 {
			transferLogDbs = append(transferLogDbs, ethLogs...)
		}
		createContractTran := convertTradeBeans(transaction, codeMap)
		if createContractTran != nil {
			createContractTransDbs = append(createContractTransDbs, createContractTran)
		}
	}
	//transferInnerMutex.Lock()
	*allTransferDbs = append(*allTransferDbs, transferDbs...)
	//transferInnerMutex.Unlock()
	//transferLogMutex.Lock()
	*allTransferLogDbs = append(*allTransferLogDbs, transferLogDbs...)
	//transferLogMutex.Unlock()
	//createContractMutex.Lock()
	*allCreateContractTransDbs = append(*allCreateContractTransDbs, createContractTransDbs...)
	//createContractMutex.Unlock()
}

func convertTransBeans(txnVo *entity.EthTransactionVo) SAny {
	return SAny{
		txnVo.TransactionHash,
		txnVo.Status,
		txnVo.From,
		txnVo.To,
		txnVo.MethodId,
		txnVo.Timestamp,
		txnVo.Block,
		txnVo.TransactionValue,
		txnVo.Nonce,
		txnVo.TransactionIndex,
		txnVo.TransactionFee,
		txnVo.GasBaseFees,
		txnVo.GasMaxPriorityFees,
		txnVo.MaxFeePerGas,
		txnVo.GasPrice,
		txnVo.Gas,
	}

}

func convertTradeBeans(txnVo *entity.EthTransactionVo, codeMap *map[string]string) SAny {
	if txnVo.To != "" {
		return nil
	}
	contractAddr := hexutil.Encode(txnVo.ContractAddress.Bytes())
	code := (*codeMap)[contractAddr]
	if code == "" {
		//code = encodeContractCode(tx, txnVo)
	} else {
		//log.Info(fmt.Sprintf("code not need general contractAddr: %s", contractAddr))
	}
	return SAny{
		txnVo.TransactionHash,
		txnVo.Status,
		txnVo.From,
		contractAddr,
		txnVo.MethodId,
		txnVo.Timestamp,
		txnVo.Block,
		code,
		1,
	}
}

func convertEthLogBeans(txnvo *entity.EthTransactionVo) []SAny {
	beansCap := 0
	if txnvo.Logs != nil {
		beansCap = len(*txnvo.Logs)
	}
	if beansCap == 0 {
		return nil
	}
	ethLogs := make([]SAny, 0, beansCap)
	for _, txnvoLog := range *txnvo.Logs {
		logTopicArr := make([]string, len(txnvoLog.Topics))
		if len(txnvoLog.Topics) > 0 {
			for i, topic := range txnvoLog.Topics {
				logTopicArr[i] = topic.Hex()
			}
		}

		ethLogs = append(ethLogs, SAny{
			txnvo.TransactionHash,
			txnvo.Block,
			txnvoLog.Index,
			txnvo.TransactionIndex,
			txnvo.Timestamp,
			hexutil.Encode(txnvoLog.Address.Bytes()),
			hexutil.Encode(txnvoLog.Data),
			logTopicArr,
			txnvoLog.Removed,
		})
	}
	return ethLogs
}

func getApi() *jsonrpc.TraceAPIImpl {
	db = openDb()
	_, rootCancel := libcommon.RootContext()
	cfg := &httpcfg.HttpCfg{}
	temporalDb, _, _, _, stateCache, blockReader, engine, ff, bridgeReader, _, err := cli.RemoteServices(ctx, cfg, logger, rootCancel)
	logger.Info(fmt.Sprintf("getApi RemoteServices err: %s", err))
	baseApi := jsonrpc.NewBaseApi(ff, stateCache, blockReader, cfg.WithDatadir, cfg.EvmCallTimeout, engine, dirs, bridgeReader)
	api := jsonrpc.NewTraceAPI(baseApi, temporalDb, &httpcfg.HttpCfg{})
	logger.Info("NewTraceAPI finish")
	return api
}

func openDb() kv.RwDB {
	config := &nodecfg.Config{
		Name:              "scan internal txn",
		DatabaseVerbosity: kv.DBVerbosityLvl(2),
		Http:              httpcfg.HttpCfg{DBReadConcurrency: cmp.InRange(10, 9_000, runtime.GOMAXPROCS(-1)*64)},
		Dirs:              dirs,
	}

	//roTxLimit := int64(40000)
	//if config.Http.DBReadConcurrency > 0 {
	//	roTxLimit = int64(config.Http.DBReadConcurrency)
	//}
	roTxsLimiter := semaphore.NewWeighted(roTxLimit) // 1 less than max to allow unlocking to happen

	chainKv, err := mdbx.New(kv.ChainDB, logger).
		Path(dbPath).
		GrowthStep(16 * datasize.MB).
		DBVerbosity(config.DatabaseVerbosity).
		RoTxsLimiter(roTxsLimiter).
		Readonly(true).
		Open(ctx)

	logger.Info(fmt.Sprintf("OpenDatabase finish, err: %s", err))
	if err != nil {
		return nil
	}
	return chainKv
}
