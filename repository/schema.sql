CREATE TABLE
    stocks (
        id BIGSERIAL PRIMARY KEY,
        created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
        name VARCHAR(50) NOT NULL,
        symbol VARCHAR(50) NOT NULL,
        customSymbol VARCHAR(50) NOT NULL,
        scriptType VARCHAR(50) NOT NULL,
        industry VARCHAR(50) NOT NULL,
        isin VARCHAR(50) NOT NULL,
        fno boolean NOT NULL
    );