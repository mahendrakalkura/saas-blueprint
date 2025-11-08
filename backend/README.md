# Backend

Go backend service using Chi router and sqlc for database queries.

## Structure

```
backend/
├── main.go              # Entry point with Chi router setup
├── db/
│   ├── migrations/      # SQL schema files
│   ├── queries/         # SQL query files for sqlc
│   └── sqlc/           # Generated Go code from sqlc
├── sqlc.yaml           # sqlc configuration
└── go.mod
```

## Using sqlc

1. Add your schema in `db/migrations/*.sql`
2. Add your queries in `db/queries/*.sql`
3. Run `sqlc generate` to generate Go code
4. Import and use the generated code from `db/sqlc` package

## Example Query File

See `db/queries/.gitkeep` for example query syntax.

## Running Locally

```bash
# Install dependencies
go mod download

# Generate sqlc code
sqlc generate

# Run the server
go run main.go
```

## Environment Variables

See root `.env` file for configuration.
