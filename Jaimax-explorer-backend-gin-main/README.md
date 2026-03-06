# Blockchain API — Go + Gin

A production-ready REST API for the blockchain indexer database, built with Go and Gin.

## Setup

### 1. Install dependencies
```bash
go mod tidy
```

### 2. Configure environment
Edit `.env` with your database credentials:
```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=blockchain_db
DB_SSLMODE=disable
PORT=8080
GIN_MODE=debug   # or "release"
```

### 3. Run
```bash
go run main.go
```

---

## API Reference

All endpoints are prefixed with `/api/v1`.

### Health & Stats

| Method | Endpoint      | Description              |
|--------|---------------|--------------------------|
| GET    | `/health`     | Health check             |
| GET    | `/stats`      | Chain-wide stats summary |
| GET    | `/search?q=`  | Search by hash/address/height |

**Search example:** `GET /api/v1/search?q=cosmos1abc...`  
Returns matches across blocks, transactions, validators, contracts, and addresses.

---

### Blocks

| Method | Endpoint                          | Description                        |
|--------|-----------------------------------|------------------------------------|
| GET    | `/blocks`                         | List blocks (paginated)            |
| GET    | `/blocks/latest`                  | Latest block                       |
| GET    | `/blocks/:height`                 | Block by height                    |
| GET    | `/blocks/:height/transactions`    | All transactions in a block        |

---

### Transactions

| Method | Endpoint                          | Description                       |
|--------|-----------------------------------|-----------------------------------|
| GET    | `/transactions`                   | List all transactions (paginated) |
| GET    | `/transactions/recent`            | Recent transactions view          |
| GET    | `/transactions/:hash`             | Transaction by hash               |
| GET    | `/transactions/:hash/messages`    | Messages in a transaction         |
| GET    | `/transactions/:hash/events`      | Events in a transaction           |

---

### Addresses

| Method | Endpoint                              | Description                       |
|--------|---------------------------------------|-----------------------------------|
| GET    | `/addresses/:address/balances`        | Balances for an address           |
| GET    | `/addresses/:address/transactions`    | Transactions for an address       |
| GET    | `/addresses/:address/bank-transfers`  | Bank transfers for an address     |
| GET    | `/addresses/:address/cw20-transfers`  | CW20 transfers for an address     |
| GET    | `/addresses/:address/votes`           | Governance votes by address       |
| GET    | `/balances?denom=`                    | All balances (filter by denom)    |

---

### Bank Transfers

| Method | Endpoint              | Description               |
|--------|-----------------------|---------------------------|
| GET    | `/bank-transfers`     | List all bank transfers   |
| GET    | `/bank-transfers/:id` | Get transfer by ID        |

---

### Validators

| Method | Endpoint                         | Description                       |
|--------|----------------------------------|-----------------------------------|
| GET    | `/validators`                    | List validators (`?status=`)      |
| GET    | `/validators/:address`           | Validator by operator address     |
| GET    | `/validators/:address/history`   | Validator history                 |
| GET    | `/validators/:address/stats`     | Validator stats                   |
| GET    | `/validator-stats`               | All validator stats               |

**Filter example:** `GET /api/v1/validators?status=BOND_STATUS_BONDED`

---

### Governance

| Method | Endpoint                    | Description                      |
|--------|-----------------------------|----------------------------------|
| GET    | `/proposals`                | List proposals (`?status=`)      |
| GET    | `/proposals/:id`            | Proposal by ID                   |
| GET    | `/proposals/:id/votes`      | Votes on a proposal              |

---

### Wasm

| Method | Endpoint                                  | Description                        |
|--------|-------------------------------------------|------------------------------------|
| GET    | `/wasm/codes`                             | List all wasm codes                |
| GET    | `/wasm/codes/:code_id`                    | Wasm code by ID                    |
| GET    | `/wasm/codes/:code_id/contracts`          | Contracts for a code ID            |
| GET    | `/wasm/contracts`                         | List all contracts                 |
| GET    | `/wasm/contracts/:address`                | Contract by address                |
| GET    | `/wasm/contracts/:address/executions`     | Executions for a contract          |
| GET    | `/wasm/contracts/:address/events`         | Wasm events for a contract         |
| GET    | `/wasm/contracts/:address/migrations`     | Migrations for a contract          |
| GET    | `/wasm/executions`                        | All wasm executions (paginated)    |
| GET    | `/wasm/instantiations`                    | All wasm instantiations (paginated)|
| GET    | `/wasm/activity`                          | Contract activity leaderboard      |
| GET    | `/wasm/activity/recent`                   | Recent wasm activity               |

---

### CW20

| Method | Endpoint                      | Description                       |
|--------|-------------------------------|-----------------------------------|
| GET    | `/cw20/transfers`             | All CW20 transfers (`?contract=`) |
| GET    | `/cw20/address-activity`      | CW20 address activity             |

---

### Authority Accounts

| Method | Endpoint                          | Description                         |
|--------|-----------------------------------|-------------------------------------|
| GET    | `/authority-accounts`             | List accounts (`?active=true/false`)|
| GET    | `/authority-accounts/:address`    | Get account by address              |

---

### Sync State

| Method | Endpoint                  | Description              |
|--------|---------------------------|--------------------------|
| GET    | `/state/indexer`          | Indexer sync state       |
| GET    | `/state/validator-sync`   | Validator sync state     |
| GET    | `/state/proposal-sync`    | Proposal sync state      |

---

## Pagination

All list endpoints support:
```
?page=1&per_page=20
```

Response format:
```json
{
  "data": [...],
  "page": 1,
  "per_page": 20,
  "total": 1500,
  "total_pages": 75
}
```

## Project Structure

```
blockchain-api/
├── main.go               # Entry point
├── .env                  # Environment config
├── go.mod
├── config/
│   └── config.go         # DB connection
├── models/
│   └── models.go         # All structs
├── handlers/
│   ├── blocks.go
│   ├── transactions.go
│   ├── addresses.go
│   ├── bank_transfers.go
│   ├── validators.go
│   ├── proposals.go
│   ├── wasm.go
│   ├── misc.go           # CW20, authority accounts, sync state
│   └── stats.go          # Stats & search
├── routes/
│   └── routes.go         # All route registration
└── middleware/
    └── middleware.go     # CORS, logger, error handler
```
