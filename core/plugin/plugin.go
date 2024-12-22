// Package plugin provides a plugin system for Geth, allowing external systems
// to hook into blockchain events like new heads, finalizations, and reorgs.
package plugin

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
)

// GethPlugin implements GethPlugin for database indexing
type GethPlugin struct {
	db *IndexerDB
}

// NewGethPlugin creates a new indexer plugin instance
func NewGethPlugin(db *IndexerDB) *GethPlugin {
	return &GethPlugin{
		db: db,
	}
}

// OnInit implements GethPlugin
func (p *GethPlugin) OnInit(bc interface{}) {
	// Initialize plugin with blockchain instance
}

// OnHead implements GethPlugin
func (p *GethPlugin) OnHead(header *types.Header) {
	block := &Block{
		Number:           header.Number.Uint64(),
		Hash:             header.Hash().Hex(),
		ParentHash:       header.ParentHash.Hex(),
		Timestamp:        time.Unix(int64(header.Time), 0),
		Nonce:            fmt.Sprintf("%d", header.Nonce),
		BaseFeePerGas:    sql.NullString{String: header.BaseFee.String(), Valid: header.BaseFee != nil},
		Difficulty:       header.Difficulty.String(),
		ExtraData:        fmt.Sprintf("0x%x", header.Extra),
		GasLimit:         fmt.Sprintf("%d", header.GasLimit),
		GasUsed:          fmt.Sprintf("%d", header.GasUsed),
		Miner:            header.Coinbase.Hex(),
		MixHash:          header.MixDigest.Hex(),
		ReceiptsRoot:     header.ReceiptHash.Hex(),
		Sha3Uncles:       header.UncleHash.Hex(),
		Size:             "0", // This needs to be calculated from the full block
		StateRoot:        header.Root.Hex(),
		TransactionsRoot: header.TxHash.Hex(),
		// Initialize empty arrays for these fields
		SealFields:   []string{},
		Transactions: []string{},
		Uncles:       []string{},
		BlockReward:  "0", // This needs to be calculated
		UncleReward:  "0", // This needs to be calculated
	}

	if err := p.db.InsertBlock(block); err != nil {
		// TODO log metrics and don't panic
		panic(fmt.Sprintf("Failed to index block %d: %v", block.Number, err))
	}
}

// OnFinal implements GethPlugin
func (p *GethPlugin) OnFinal(header *types.Header) {
	// Mark the block and all its data as "finalized"
	// This data will never need to be reverted
}

// OnClose implements GethPlugin
func (p *GethPlugin) OnClose() {
	// Cleanup resources
}

// On reorg roll back all the old blocks
func (p *GethPlugin) OnReorg(oldHeaders, newHeaders []*types.Header) {
	for _, header := range oldHeaders {
		blockNumber := header.Number.Uint64()
		if err := p.db.DeleteBlockAndDescendants(blockNumber); err != nil {
			log.Printf("Failed to delete reorged block %d: %v", blockNumber, err)
		}
	}

	for _, header := range newHeaders {
		p.OnHead(header)
	}
}
