// TODO add retries
package plugin

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// Block represents the blocks table schema
type Block struct {
	Number                uint64         `db:"number"`
	Hash                  string         `db:"hash"`
	ParentHash            string         `db:"parent_hash"`
	Timestamp             time.Time      `db:"timestamp"`
	Nonce                 string         `db:"nonce"`
	BaseFeePerGas         sql.NullString `db:"base_fee_per_gas"`
	BlobGasUsed           sql.NullString `db:"blob_gas_used"`
	Difficulty            string         `db:"difficulty"`
	ExcessBlobGas         sql.NullString `db:"excess_blob_gas"`
	ExtraData             string         `db:"extra_data"`
	GasLimit              string         `db:"gas_limit"`
	GasUsed               string         `db:"gas_used"`
	LogsBloom             sql.NullString `db:"logs_bloom"`
	Miner                 string         `db:"miner"`
	MixHash               string         `db:"mix_hash"`
	ParentBeaconBlockRoot sql.NullString `db:"parent_beacon_block_root"`
	ReceiptsRoot          string         `db:"receipts_root"`
	Sha3Uncles            string         `db:"sha3_uncles"`
	Size                  string         `db:"size"`
	StateRoot             string         `db:"state_root"`
	TotalDifficulty       sql.NullString `db:"total_difficulty"`
	TransactionsRoot      string         `db:"transactions_root"`
	WithdrawalsRoot       sql.NullString `db:"withdrawals_root"`
	SealFields            []string       `db:"seal_fields"`
	Transactions          []string       `db:"transactions"`
	Uncles                []string       `db:"uncles"`
	BlockReward           string         `db:"block_reward"`
	UncleReward           string         `db:"uncle_reward"`
}

// Transaction represents the transactions table schema
type Transaction struct {
	Hash           string         `db:"hash"`
	BlockNumber    uint64         `db:"block_number"`
	From           string         `db:"from"`
	To             sql.NullString `db:"to"`
	Value          string         `db:"value"`
	Nonce          uint64         `db:"nonce"`
	GasPrice       string         `db:"gas_price"`
	GasLimit       uint64         `db:"gas_limit"`
	GasUsed        uint64         `db:"gas_used"`
	Input          string         `db:"input"`
	Status         uint64         `db:"status"`
	Type           uint64         `db:"type"`
	MaxFeePerGas   sql.NullString `db:"max_fee_per_gas"`
	MaxPriorityFee sql.NullString `db:"max_priority_fee"`
	BlobGasUsed    sql.NullString `db:"blob_gas_used"`
	BlobGasPrice   sql.NullString `db:"blob_gas_price"`
	Error          sql.NullString `db:"error"`
}

// Log represents the logs table schema
type Log struct {
	ID              uint64   `db:"id"`
	TransactionHash string   `db:"transaction_hash"`
	BlockNumber     uint64   `db:"block_number"`
	Address         string   `db:"address"`
	Topics          []string `db:"topics"`
	Data            string   `db:"data"`
	LogIndex        uint64   `db:"log_index"`
	Removed         bool     `db:"removed"`
}

// StateChange represents state changes within a transaction
type StateChange struct {
	ID              uint64         `db:"id"`
	BlockNumber     uint64         `db:"block_number"`
	TransactionHash string         `db:"transaction_hash"`
	Address         string         `db:"address"`
	StorageKey      sql.NullString `db:"storage_key"`
	PrevValue       string         `db:"prev_value"`
	NewValue        string         `db:"new_value"`
	ChangeType      string         `db:"change_type"` // balance, nonce, code, storage
}

// AccessList represents transaction access lists
type AccessList struct {
	ID              uint64 `db:"id"`
	TransactionHash string `db:"transaction_hash"`
	Address         string `db:"address"`
	StorageKey      string `db:"storage_key"`
}

// Account represents both EOAs and Contracts
type Account struct {
	Address        string         `db:"address"`
	Balance        string         `db:"balance"`
	Nonce          uint64         `db:"nonce"`
	Code           sql.NullString `db:"code"`            // null for EOA, populated for contracts
	CreatorAddress sql.NullString `db:"creator_address"` // null for EOA
	CreatorTxHash  sql.NullString `db:"creator_tx_hash"` // null for EOA
	CreatedAt      sql.NullTime   `db:"created_at"`      // block timestamp when created
}

// CreateTablesSQL contains the SQL statements to create the tables
const CreateTablesSQL = `
CREATE TABLE IF NOT EXISTS blocks (
    number BIGINT PRIMARY KEY,
    hash VARCHAR(66) NOT NULL UNIQUE,
    parent_hash VARCHAR(66) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    nonce VARCHAR(255) NOT NULL,
    base_fee_per_gas VARCHAR(255),
    blob_gas_used VARCHAR(255),
    difficulty VARCHAR(255) NOT NULL,
    excess_blob_gas VARCHAR(255),
    extra_data VARCHAR(255) NOT NULL,
    gas_limit VARCHAR(255) NOT NULL,
    gas_used VARCHAR(255) NOT NULL,
    logs_bloom TEXT,
    miner VARCHAR(42) NOT NULL,
    mix_hash VARCHAR(66) NOT NULL,
    parent_beacon_block_root VARCHAR(66),
    receipts_root VARCHAR(66) NOT NULL,
    sha3_uncles VARCHAR(66) NOT NULL,
    size VARCHAR(255) NOT NULL,
    state_root VARCHAR(66) NOT NULL,
    total_difficulty VARCHAR(255),
    transactions_root VARCHAR(66) NOT NULL,
    withdrawals_root VARCHAR(66),
    seal_fields TEXT[],
    transactions TEXT[],
    uncles TEXT[],
    block_reward VARCHAR(255) NOT NULL,
    uncle_reward VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS transactions (
    hash VARCHAR(66) PRIMARY KEY,
    block_number BIGINT NOT NULL REFERENCES blocks(number),
    "from" VARCHAR(42) NOT NULL,
    "to" VARCHAR(42),
    value VARCHAR(255) NOT NULL,
    nonce BIGINT NOT NULL,
    gas_price VARCHAR(255) NOT NULL,
    gas_limit BIGINT NOT NULL,
    gas_used BIGINT NOT NULL,
    input TEXT NOT NULL,
    status SMALLINT NOT NULL,
    type SMALLINT NOT NULL,
    max_fee_per_gas VARCHAR(255),
    max_priority_fee VARCHAR(255),
    blob_gas_used VARCHAR(255),
    blob_gas_price VARCHAR(255),
    error TEXT
);

CREATE TABLE IF NOT EXISTS logs (
    id BIGSERIAL PRIMARY KEY,
    transaction_hash VARCHAR(66) NOT NULL REFERENCES transactions(hash),
    block_number BIGINT NOT NULL REFERENCES blocks(number),
    address VARCHAR(42) NOT NULL,
    topics TEXT[] NOT NULL,
    data TEXT NOT NULL,
    log_index BIGINT NOT NULL,
    removed BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS state_changes (
    id BIGSERIAL PRIMARY KEY,
    block_number BIGINT NOT NULL REFERENCES blocks(number),
    transaction_hash VARCHAR(66) NOT NULL REFERENCES transactions(hash),
    address VARCHAR(42) NOT NULL,
    storage_key VARCHAR(66),
    prev_value TEXT NOT NULL,
    new_value TEXT NOT NULL,
    change_type VARCHAR(20) NOT NULL
);

CREATE TABLE IF NOT EXISTS access_lists (
    id BIGSERIAL PRIMARY KEY,
    transaction_hash VARCHAR(66) NOT NULL REFERENCES transactions(hash),
    address VARCHAR(42) NOT NULL,
    storage_key VARCHAR(66) NOT NULL
);

CREATE TABLE IF NOT EXISTS accounts (
    address VARCHAR(42) PRIMARY KEY,
    balance VARCHAR(255) NOT NULL,
    nonce BIGINT NOT NULL,
    code TEXT,
    creator_address VARCHAR(42),
    creator_tx_hash VARCHAR(66) REFERENCES transactions(hash),
    created_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_logs_address ON logs(address);
CREATE INDEX IF NOT EXISTS idx_logs_topics ON logs USING gin(topics);
CREATE INDEX IF NOT EXISTS idx_state_changes_address ON state_changes(address);
CREATE INDEX IF NOT EXISTS idx_access_lists_address ON access_lists(address);
CREATE INDEX IF NOT EXISTS idx_accounts_creator ON accounts(creator_address);
`

// IndexerDB handles the database connection and operations
type IndexerDB struct {
	db *sqlx.DB
}

// Config holds database connection configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewDB creates a new database connection
func NewDB(config Config) (*IndexerDB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	// Connect to database
	db, err := sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("error connecting to the database: %v", err)
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error pinging the database: %v", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Create new IndexerDB instance
	indexerDB := &IndexerDB{
		db: db,
	}

	// Initialize database schema
	if err := indexerDB.initSchema(); err != nil {
		return nil, fmt.Errorf("error initializing database schema: %v", err)
	}

	return indexerDB, nil
}

// initSchema creates the database tables if they don't exist
func (idb *IndexerDB) initSchema() error {
	_, err := idb.db.Exec(CreateTablesSQL)
	if err != nil {
		return fmt.Errorf("error creating tables: %v", err)
	}
	return nil
}

// Close closes the database connection
func (idb *IndexerDB) Close() error {
	return idb.db.Close()
}

// InsertBlock inserts a block into the database
func (idb *IndexerDB) InsertBlock(block *Block) error {
	query := `
		INSERT INTO blocks (
			number, hash, parent_hash, timestamp, nonce, base_fee_per_gas,
			blob_gas_used, difficulty, excess_blob_gas, extra_data, gas_limit,
			gas_used, logs_bloom, miner, mix_hash, parent_beacon_block_root,
			receipts_root, sha3_uncles, size, state_root, total_difficulty,
			transactions_root, withdrawals_root, seal_fields, transactions,
			uncles, block_reward, uncle_reward
		) VALUES (
			:number, :hash, :parent_hash, :timestamp, :nonce, :base_fee_per_gas,
			:blob_gas_used, :difficulty, :excess_blob_gas, :extra_data, :gas_limit,
			:gas_used, :logs_bloom, :miner, :mix_hash, :parent_beacon_block_root,
			:receipts_root, :sha3_uncles, :size, :state_root, :total_difficulty,
			:transactions_root, :withdrawals_root, :seal_fields, :transactions,
			:uncles, :block_reward, :uncle_reward
		)`

	_, err := idb.db.NamedExec(query, block)
	if err != nil {
		return fmt.Errorf("error inserting block: %v", err)
	}
	return nil
}

// DeleteBlockAndDescendants deletes a block and all its associated data
func (idb *IndexerDB) DeleteBlockAndDescendants(blockNumber uint64) error {
	tx, err := idb.db.Beginx()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}

	// Delete in reverse order of dependencies to respect foreign key constraints
	deleteQueries := []string{
		`DELETE FROM access_lists WHERE transaction_hash IN (
			SELECT hash FROM transactions WHERE block_number >= $1
		)`,
		`DELETE FROM state_changes WHERE block_number >= $1`,
		`DELETE FROM logs WHERE block_number >= $1`,
		`DELETE FROM transactions WHERE block_number >= $1`,
		`DELETE FROM blocks WHERE number >= $1`,
	}

	for _, query := range deleteQueries {
		_, err := tx.Exec(query, blockNumber)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error executing delete query: %v", err)
		}
	}

	return tx.Commit()
}
