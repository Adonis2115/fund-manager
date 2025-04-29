-- name: GetStock :one
SELECT * FROM stocks
WHERE id = $1 LIMIT 1;

-- name: GetStocks :many
SELECT * FROM stocks
ORDER BY name;

-- name: CreateStock :one
INSERT INTO stocks (
    id, name, symbol, customSymbol, scriptType, industry, isin, fno
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: BulkCreateStocks :copyfrom
INSERT INTO stocks (
    id, name, symbol, customSymbol, scriptType, industry, isin, fno
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
);

-- name: BulkCreateDaily :copyfrom
INSERT INTO daily (
    id, stockId, open, high, low, close, adjClose, volume, timestamp
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
);