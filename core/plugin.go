package core

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/lru"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

// Plugin defines the interface for blockchain plugins
type Plugin interface {
	OnInit(chain *BlockChain)
	OnHead(header *types.Header)
	OnFinal(header *types.Header)
	OnClose()
	OnReorg(oldHeaders, newHeaders []*types.Header)
}

// IndexerPlugin implements blockchain indexing functionality
type IndexerPlugin struct {
	db    *IndexerDB
	chain *BlockChain
}

// NewIndexerPlugin creates a new indexer plugin instance
func NewIndexerPlugin(db *IndexerDB) *IndexerPlugin {
	if db == nil {
		log.Info("Creating indexer plugin without database connection")
		return &IndexerPlugin{}
	}
	log.Info("Creating new indexer plugin with database connection")
	return &IndexerPlugin{
		db: db,
	}
}

// OnInit implements Plugin
func (p *IndexerPlugin) OnInit(bc *BlockChain) {
	if p.db == nil {
		log.Info("Initializing indexer plugin without database - indexing disabled")
		return
	}
	log.Info("Initializing indexer plugin", "chainID", bc.Config().ChainID)
	p.chain = bc
}

// OnHead is called whenever a new head block is set
func (p *IndexerPlugin) OnHead(header *types.Header) {
	if p.db == nil {
		return
	}
	log.Info("Indexer processing new head block",
		"number", header.Number,
		"hash", header.Hash(),
		"parent", header.ParentHash,
		"timestamp", time.Unix(int64(header.Time), 0))

	tx, err := p.db.db.Beginx()
	if err != nil {
		log.Error("Failed to begin transaction",
			"block", header.Number,
			"hash", header.Hash(),
			"error", err)
		return
	}
	defer tx.Rollback()

	// Create base block record
	block := &Block{
		Number:                header.Number.Uint64(),
		Hash:                  header.Hash().Hex(),
		ParentHash:            header.ParentHash.Hex(),
		Timestamp:             time.Unix(int64(header.Time), 0),
		Nonce:                 fmt.Sprintf("%d", header.Nonce),
		BaseFeePerGas:         sql.NullString{String: header.BaseFee.String(), Valid: header.BaseFee != nil},
		BlobGasUsed:           sql.NullString{String: fmt.Sprintf("%d", *header.BlobGasUsed), Valid: header.BlobGasUsed != nil},
		Difficulty:            header.Difficulty.String(),
		ExcessBlobGas:         sql.NullString{String: fmt.Sprintf("%d", *header.ExcessBlobGas), Valid: header.ExcessBlobGas != nil},
		ExtraData:             fmt.Sprintf("0x%x", header.Extra),
		GasLimit:              fmt.Sprintf("%d", header.GasLimit),
		GasUsed:               fmt.Sprintf("%d", header.GasUsed),
		LogsBloom:             sql.NullString{String: fmt.Sprintf("0x%x", header.Bloom), Valid: true},
		Miner:                 header.Coinbase.Hex(),
		MixHash:               header.MixDigest.Hex(),
		ParentBeaconBlockRoot: sql.NullString{String: header.ParentBeaconRoot.String(), Valid: header.ParentBeaconRoot != nil},
		ReceiptsRoot:          header.ReceiptHash.Hex(),
		Sha3Uncles:            header.UncleHash.Hex(),
		StateRoot:             header.Root.Hex(),
		TransactionsRoot:      header.TxHash.Hex(),
		WithdrawalsRoot:       sql.NullString{String: header.WithdrawalsHash.String(), Valid: header.WithdrawalsHash != nil},
	}

	// Insert the block
	if err := p.db.InsertBlockWithTx(tx, block); err != nil {
		log.Error("Failed to index block",
			"number", block.Number,
			"hash", block.Hash,
			"error", err)
		return
	}

	// Get receipts from chain
	receipts := p.chain.GetReceiptsByHash(header.Hash())
	log.Debug("Processing receipts",
		"block", block.Number,
		"count", len(receipts),
		"hash", block.Hash)

	for i, receipt := range receipts {
		txHash := receipt.TxHash.Hex()

		// Index the receipt
		r := &Receipt{
			BlockNumber:      header.Number.Uint64(),
			BlockHash:        header.Hash().Hex(),
			TransactionHash:  txHash,
			TransactionIndex: uint(i),
			ContractAddress:  receipt.ContractAddress.Hex(),
			GasUsed:          receipt.GasUsed,
			Status:           receipt.Status,
		}

		if err := p.db.InsertReceiptWithTx(tx, r); err != nil {
			log.Error("Failed to index receipt",
				"block", block.Number,
				"tx", txHash,
				"error", err)
			continue
		}

		// Index the logs
		for _, logEntry := range receipt.Logs {
			l := &Log{
				BlockNumber:     header.Number.Uint64(),
				TransactionHash: txHash,
				LogIndex:        uint64(logEntry.Index),
				Address:         logEntry.Address.Hex(),
				Topics:          make([]string, len(logEntry.Topics)),
				Data:            hexutil.Encode(logEntry.Data),
			}
			for i, topic := range logEntry.Topics {
				l.Topics[i] = topic.Hex()
			}
			if err := p.db.InsertLogWithTx(tx, l); err != nil {
				log.Error("Failed to index log",
					"block", block.Number,
					"txHash", txHash,
					"logIndex", l.LogIndex,
					"error", err)
				continue
			}
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Error("Failed to commit transaction",
			"block", block.Number,
			"hash", block.Hash,
			"error", err)
		return
	}

	log.Info("Successfully indexed block",
		"number", block.Number,
		"hash", block.Hash,
		"txCount", len(receipts))
}

// OnFinal implements Plugin
func (p *IndexerPlugin) OnFinal(header *types.Header) {
	if p.db == nil {
		return
	}
	log.Info("Indexer processing finalized block",
		"number", header.Number,
		"hash", header.Hash())

	if err := p.db.MarkBlockFinalized(header.Number.Uint64()); err != nil {
		log.Error("Failed to mark block as finalized",
			"block", header.Number,
			"hash", header.Hash(),
			"error", err)
		return
	}

	log.Info("Successfully marked block as finalized",
		"number", header.Number,
		"hash", header.Hash())
}

// OnClose implements Plugin
func (p *IndexerPlugin) OnClose() {
	log.Info("Closing indexer plugin")
	if err := p.db.Close(); err != nil {
		log.Error("Failed to close database connection", "error", err)
	}
}

// OnReorg handles chain reorganizations
func (p *IndexerPlugin) OnReorg(oldHeaders, newHeaders []*types.Header) {
	log.Info("Indexer handling chain reorg",
		"oldLen", len(oldHeaders),
		"newLen", len(newHeaders),
		"firstOld", oldHeaders[0].Number,
		"firstNew", newHeaders[0].Number)

	tx, err := p.db.db.Beginx()
	if err != nil {
		log.Error("Failed to begin reorg transaction", "error", err)
		return
	}
	defer tx.Rollback()

	// Delete old blocks and their data
	for _, header := range oldHeaders {
		blockNumber := header.Number.Uint64()
		if err := p.db.DeleteBlockAndDescendantsWithTx(tx, blockNumber); err != nil {
			log.Error("Failed to delete reorged block", "number", blockNumber, "error", err)
			return
		}
	}

	// Insert new blocks
	for _, header := range newHeaders {
		p.OnHead(header)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Error("Failed to commit reorg transaction", "error", err)
		return
	}
}

// Add receipts cache implementation
var receiptsCache = lru.NewCache[common.Hash, types.Receipts](32)

// Update GetBlockReceipts to use the cache
func GetBlockReceipts(getReceipts func(hash common.Hash) types.Receipts, hash common.Hash, number uint64) types.Receipts {
	// Try to get from cache first
	if receipts, ok := receiptsCache.Get(hash); ok {
		return receipts
	}

	// Get from blockchain
	receipts := getReceipts(hash)
	if receipts == nil {
		return nil
	}

	// Cache for future use
	receiptsCache.Add(hash, receipts)
	return receipts
}
