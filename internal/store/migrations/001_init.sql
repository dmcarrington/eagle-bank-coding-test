CREATE TABLE IF NOT EXISTS users (
    id                 TEXT PRIMARY KEY,
    name               TEXT NOT NULL,
    email              TEXT NOT NULL UNIQUE,
    password_hash      TEXT NOT NULL,
    phone_number       TEXT NOT NULL,
    address_line1      TEXT NOT NULL,
    address_line2      TEXT,
    address_line3      TEXT,
    address_town       TEXT NOT NULL,
    address_county     TEXT NOT NULL,
    address_postcode   TEXT NOT NULL,
    created_timestamp  TEXT NOT NULL,
    updated_timestamp  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS accounts (
    account_number     TEXT PRIMARY KEY,
    user_id            TEXT NOT NULL REFERENCES users(id),
    name               TEXT NOT NULL,
    account_type       TEXT NOT NULL,
    balance_pence      INTEGER NOT NULL DEFAULT 0,
    currency           TEXT NOT NULL DEFAULT 'GBP',
    sort_code          TEXT NOT NULL DEFAULT '10-10-10',
    created_timestamp  TEXT NOT NULL,
    updated_timestamp  TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_accounts_user ON accounts(user_id);

CREATE TABLE IF NOT EXISTS transactions (
    id                 TEXT PRIMARY KEY,
    account_number     TEXT NOT NULL REFERENCES accounts(account_number),
    user_id            TEXT NOT NULL REFERENCES users(id),
    amount_pence       INTEGER NOT NULL,
    currency           TEXT NOT NULL,
    type               TEXT NOT NULL CHECK (type IN ('deposit','withdrawal')),
    reference          TEXT,
    created_timestamp  TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_tx_account ON transactions(account_number);
