CREATE TABLE
    stocks (
        id uuid PRIMARY KEY,
        created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
        name VARCHAR(100) NOT NULL,
        symbol VARCHAR(50) NOT NULL,
        customSymbol VARCHAR(50) NOT NULL,
        scriptType VARCHAR(50) NOT NULL,
        industry VARCHAR(50),
        isin VARCHAR(50),
        fno boolean NOT NULL
    );