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
    uncle_reward VARCHAR(255) NOT NULL,
    finalized BOOLEAN DEFAULT FALSE
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

CREATE TABLE IF NOT EXISTS receipts (
    id BIGSERIAL PRIMARY KEY,
    block_number BIGINT NOT NULL REFERENCES blocks(number),
    block_hash VARCHAR(66) NOT NULL,
    transaction_hash VARCHAR(66) NOT NULL REFERENCES transactions(hash),
    transaction_index BIGINT NOT NULL,
    contract_address VARCHAR(42),
    gas_used BIGINT NOT NULL,
    status SMALLINT NOT NULL,
    UNIQUE(transaction_hash)
);

CREATE INDEX IF NOT EXISTS idx_logs_address ON logs(address);
CREATE INDEX IF NOT EXISTS idx_logs_topics ON logs USING gin(topics);
CREATE INDEX IF NOT EXISTS idx_state_changes_address ON state_changes(address);
CREATE INDEX IF NOT EXISTS idx_access_lists_address ON access_lists(address);
CREATE INDEX IF NOT EXISTS idx_accounts_creator ON accounts(creator_address);
CREATE INDEX IF NOT EXISTS idx_receipts_block ON receipts(block_number);
CREATE INDEX IF NOT EXISTS idx_receipts_contract ON receipts(contract_address);