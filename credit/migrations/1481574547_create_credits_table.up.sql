-- 1) Append-only event ledger
CREATE TABLE credit_events
(
    id        BIGSERIAL PRIMARY KEY,
    stream_id TEXT NOT NULL,
    org_id    TEXT NOT NULL,
    type      TEXT NOT NULL, -- 'debit'|'credit' etc
    amount    INT  NOT NULL,
    UNIQUE (stream_id)       -- ensure same stream message isn't inserted twice
);

-- 2) Current balances snapshot (upsert target)
CREATE TABLE balances
(
    org_id     TEXT PRIMARY KEY,
    amount     INT NULL    DEFAULT 0,
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (org_id)
);

-- 3) Consumer state (checkpoint)
CREATE TABLE persist_state
(
    id             SERIAL PRIMARY KEY,
    last_stream_id TEXT,
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (last_stream_id)
);
