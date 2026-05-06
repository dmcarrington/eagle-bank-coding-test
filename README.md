# Eagle Bank API

A REST API for a fictional bank, built in Go. Implements user management, bank accounts, and transactions conforming to the provided OpenAPI specification.

**Stack**: Go 1.22 · Gin · SQLite (modernc, pure Go) · JWT (HS256) · bcrypt

---

## Run locally

```bash
export JWT_SECRET=$(openssl rand -base64 32)
go run ./cmd/server
```

The server starts on port `8080` and creates `eagle.db` in the working directory on first run.

---

## Run with Docker

Copy the example env file and set a real secret:

```bash
cp .env.example .env
# edit .env and set JWT_SECRET using openssl rand -base64 32
docker compose up
```

The database is persisted to a named Docker volume (`eagle-data`).

---

## Configuration

| Variable     | Default        | Description                               |
|--------------|----------------|-------------------------------------------|
| `JWT_SECRET` | *(required)*   | HMAC-SHA256 signing secret for JWTs       |
| `PORT`       | `8080`         | Port the server listens on                |
| `DB_PATH`    | `./eagle.db`   | Path to the SQLite database file          |
| `JWT_TTL`    | `24h`          | JWT expiry duration (e.g. `1h`, `168h`)   |

---

## API walkthrough

All endpoints except `POST /v1/users` and `POST /v1/auth/login` require an `Authorization: Bearer <token>` header.

### 1. Create a user

```bash
curl -s -X POST http://localhost:8080/v1/users \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Alice Smith",
    "email": "alice@example.com",
    "password": "supersecret1",
    "phoneNumber": "+447700900000",
    "address": {
      "line1": "1 High Street",
      "town": "London",
      "county": "Greater London",
      "postcode": "SW1A 1AA"
    }
  }'
```

### 2. Login

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","password":"supersecret1"}' \
  | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
```

### 3. Fetch your user

```bash
USER_ID=<id from create response>
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/v1/users/$USER_ID
```

### 4. Create a bank account

```bash
ACC=$(curl -s -X POST http://localhost:8080/v1/accounts \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Main Account","accountType":"personal"}' \
  | grep -o '"accountNumber":"[^"]*"' | cut -d'"' -f4)
```

### 5. Deposit money

```bash
curl -s -X POST http://localhost:8080/v1/accounts/$ACC/transactions \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"amount":250.00,"currency":"GBP","type":"deposit","reference":"initial deposit"}'
```

### 6. Check balance

```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/v1/accounts/$ACC
```

### 7. Withdraw money

```bash
curl -s -X POST http://localhost:8080/v1/accounts/$ACC/transactions \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"amount":50.00,"currency":"GBP","type":"withdrawal"}'
```

### 8. List transactions

```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/v1/accounts/$ACC/transactions
```

---

## Testing

```bash
go test ./...
```

Tests cover every BDD scenario from the brief at the handler level, plus targeted unit tests for auth (bcrypt, JWT), ID generation, money arithmetic, and concurrent balance updates.

---

## Project structure

```
cmd/server/         entry point
internal/
  api/              Gin handlers + middleware
  auth/             bcrypt + JWT
  service/          business rules (ownership, balance, conflicts)
  store/            SQLite repositories
  domain/           core types
  ids/              ID generators
  config/           env-driven config
```

Three-layer architecture: **handler → service → store**. Ownership and conflict checks live in the service layer so they are independently testable.

---

## Caveats / out of scope

The following were intentionally excluded to keep the scope appropriate for a take-home exercise:

- **Pagination, filtering, sorting** on list endpoints — the OpenAPI spec doesn't define them.
- **Refresh tokens, password reset, email verification, account lockout** — auth is JWT-only with configurable TTL.
- **Rate limiting** — assumed handled by an upstream gateway in a production deployment.
- **Request logging beyond slog defaults** — structured logging is present but not request-level middleware.
- **Multi-currency or non-personal account types** — the spec defines single-value enums for both; no logic was written to support extending them.
- **Schema migration tooling** (e.g. goose, golang-migrate) — the single embedded migration runs idempotently on startup. A migration tool would be added before production use.
- **Production-grade secret management** — the JWT signing secret is read from an environment variable. In production, this would come from a secrets manager (AWS Secrets Manager, Vault, etc.).
- **TLS termination** — the server speaks plain HTTP; TLS is assumed to be handled upstream (load balancer, reverse proxy).
- **gRPC or any non-HTTP interface**.
- **BDD framework** (godog/cucumber) — scenarios from the brief are exercised via descriptively-named handler tests rather than a separate BDD layer.
- **Generated server stubs from OpenAPI** (e.g. oapi-codegen) — handlers are handwritten and reviewed against the spec to keep the code transparent for review.

---

## OpenAPI deviations

The submitted `openapi.yaml` differs from the original in three places:

1. **`password` field added to `CreateUserRequest`** — required to support authentication; not present in the original spec.
2. **`POST /v1/auth/login` added** — the brief asked us to design and document an auth endpoint; this is it.
3. **Transaction ID regex fixed** — `^tan-[A-Za-z0-9]$` → `^tan-[A-Za-z0-9]+$`. The original matches only a single character after the prefix, but the spec's own example (`tan-123abc`) is six characters. The `+` quantifier is the evident intent.
