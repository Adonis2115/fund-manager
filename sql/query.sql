-- name: GetStock :one
SELECT * FROM stocks
WHERE id = $1 LIMIT 1;

-- name: GetStocks :many
SELECT * FROM stocks
ORDER BY name;

-- name: CreateStock :one
INSERT INTO stocks (
    id, name, symbol, scriptType, industry, isin, fno
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: BulkCreateStocks :copyfrom
INSERT INTO stocks (
    id, name, symbol, scriptType, industry, isin, fno
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
);

-- name: BulkCreateDaily :copyfrom
INSERT INTO daily (
    id, stockId, open, high, low, close, volume, timestamp
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
);

-- name: GetTopStocksByReturn :many
WITH one_year_ago_prices AS (
    SELECT DISTINCT ON (d.stockid)
        d.stockid,
        d.close
    FROM
        daily d
    JOIN stocks s ON d.stockid = s.id
    WHERE
        d.timestamp <= ($1::timestamp - make_interval(months => $2::int))
        AND d.close IS NOT NULL
        AND d.close != 0
        AND ($3 = 'all' OR s.scriptType = $3)
    ORDER BY d.stockid, d.timestamp DESC
),
latest_prices AS (
    SELECT DISTINCT ON (d.stockid)
        d.stockid,
        d.close
    FROM
        daily d
    JOIN stocks s ON d.stockid = s.id
    WHERE
        d.timestamp <= $1::timestamp
        AND d.close IS NOT NULL
        AND d.close != 0
        AND ($3 = 'all' OR s.scriptType = $3)
    ORDER BY d.stockid, d.timestamp DESC
),
stock_returns AS (
    SELECT
        l.stockid,
        ROUND((l.close - o.close) / o.close * 100)::int AS return_percentage
    FROM latest_prices l
    JOIN one_year_ago_prices o ON l.stockid = o.stockid
)
SELECT
    s.id,
    s.name,
    s.symbol,
    sr.return_percentage
FROM stock_returns sr
JOIN stocks s ON sr.stockid = s.id
ORDER BY sr.return_percentage DESC
LIMIT $4;

-- name: GetLatestClosePrice :one
SELECT close
FROM daily d
JOIN stocks s ON d.stockid = s.id
WHERE s.symbol = $1 AND d.timestamp <= $2
AND d.close IS NOT NULL
ORDER BY d.timestamp DESC
LIMIT 1;

-- name: GetHistoricalStockPrices :many
SELECT d.timestamp, d.close
FROM daily d
JOIN stocks s ON d.stockid = s.id
WHERE s.symbol = $1
  AND d.timestamp >= $2
  AND d.timestamp <= $3
  AND d.close IS NOT NULL
ORDER BY d.timestamp;