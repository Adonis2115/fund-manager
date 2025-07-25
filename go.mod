module fund-manager

go 1.23.1

require (
	github.com/jackc/pgx/v5 v5.7.2
	github.com/joho/godotenv v1.5.1
)

require github.com/stretchr/testify v1.9.0 // indirect

require (
	github.com/google/uuid v1.6.0
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/text v0.21.0 // indirect
)

replace github.com/piquette/finance-go => github.com/psanford/finance-go v0.0.0-20250222221941-906a725c60a0
