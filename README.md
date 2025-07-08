# Internal Transfers Service

A Go-based microservice for managing accounts and internal money transfers, using PostgreSQL as the backend database and Iris as the web framework.

---

## 1. Prerequisites

- **Go**: v1.24+ (see `go.mod`)
- **Docker** and **Docker Compose**: for local development and running the database/service containers
- **Make** (optional): for convenience scripts

---

## 2. Installation, Setup, Running, and Testing

### Clone the repository

```bash
git clone git@github.com:PaulShpilsher/internal-transfers.git
cd internal-transfers
```

### Environment Configuration

Create a `.env.docker` file in the project root with the following content (edit as needed):

```env
APP_ENV=docker
SERVER_PORT=3000
POSTGRES_HOST=db
POSTGRES_PORT=5432
POSTGRES_USER=user
POSTGRES_PASSWORD=password
POSTGRES_DB=accounts_data

```

### Build and Run with Docker Compose

```bash
docker-compose up --build
```

- The API will be available at [http://localhost:3000](http://localhost:3000)
- The Postgres database will be available at port 5432

### Tear Down and Clean Volumes

```bash
docker-compose down -v
```

### Run Tests (locally, not in container)

```bash
go test ./...
```

---

## 3. Assumptions

- All monetary values are handled as strings to avoid floating-point errors, using the `shopspring/decimal` library.
- All monetary values support a maximum decimal precision of 8 digits, as enforced by the service and database.
- Only one table (`accounts`) is present; transactions are not persisted, only balances are updated.
- The API is stateless and does not implement authentication.
- The service expects the database to be initialized with the correct schema (see below).

---

## 4. API Descriptions & Example `curl` Usage

### Create Account

- **POST** `/accounts`
- **Request Body:**
  ```json
  {
    "account_id": 1,
    "initial_balance": "100.00"
  }
  ```
- **Responses:**
  - `201 Created`: Account successfully created.
  - `400 Bad Request`: 
    - Invalid request body (malformed JSON)
    - Validation error (missing/invalid fields)
    - Invalid initial balance (not a number)
    - Account ID not positive, balance negative, or precision too high
  - `409 Conflict`: Account ID already exists.
  - `500 Internal Server Error`: Any other error (e.g., database error).

**Example:**
```bash
curl -X POST http://localhost:3000/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id":1,"initial_balance":"100.00"}'
```

---

### Get Account

- **GET** `/accounts/{id}`
- **Responses:**
  - `200 OK`: Account found. 
    - **Response Body:**
      ```json
      {
          "account_id": 2,
          "balance": "100.12"
      }
      ```
  - `400 Bad Request`: Invalid account ID (not a number).
  - `404 Not Found`: Account not found.
  - `500 Internal Server Error`: Any other error (e.g., database error, response write error).

**Example:**
```bash
curl http://localhost:3000/accounts/1
```

---

### Submit Transaction

- **POST** `/transactions`
- **Request Body:**
  ```json
  {
    "source_account_id": 1,
    "destination_account_id": 2,
    "amount": "10.00"
  }
  ```
- **Responses:**
  - `200 OK`: Transaction successful.
  - `400 Bad Request`: 
    - Invalid request body (malformed JSON)
    - Validation error (missing/invalid fields)
    - Invalid amount (not a number)
    - Source/destination account ID not positive, same account, amount not positive, or precision too high
    - Insufficient funds
  - `404 Not Found`: Source or destination account not found.
  - `500 Internal Server Error`: Any other error (e.g., database error).

**Example:**
```bash
curl -X POST http://localhost:3000/transactions \
  -H "Content-Type: application/json" \
  -d '{"source_account_id":1,"destination_account_id":2,"amount":"10.00"}'
```

---

## 5. Project Architecture & Methodology

- **Layered Architecture**: The project is organized into API handlers, services (business logic), repositories (data access), and models (domain).
- **Dependency Injection**: Services and repositories are injected into handlers for testability.
- **Validation**: Uses `go-playground/validator` for request validation.
- **Testing**: Includes unit tests and mocks for services and repositories.
- **Error Handling**: Centralized error handling middleware for API responses.
- **Configuration**: Loaded from environment variables, with `.env.docker` for local/dev.

**Directory Structure:**
```
internal/
  api/        # HTTP handlers, DTOs, routing
  config/     # Configuration loading
  db/         # Database access and repository interfaces
  model/      # Domain models and errors
  services/   # Business logic
  mocks/      # Generated mocks for testing
data/
  postgres/   # Database schema
cmd/
  main.go     # Application entrypoint
```

---

## 6. Configuration

All configuration is via environment variables (see `.env.docker`):

- `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`: Database connection
- `SERVER_PORT`: Port for the API server (default: 3000)
- `APP_ENV`: Application environment (default: development)

---

## 7. Database

- **Schema**: See `data/postgres/schema.sql`
  ```sql
  CREATE TABLE IF NOT EXISTS accounts (
      account_id BIGINT PRIMARY KEY,
      balance NUMERIC(20, 8) NOT NULL CHECK (balance >= 0),
      created_at TIMESTAMP NOT NULL DEFAULT NOW(),
      updated_at TIMESTAMP NOT NULL DEFAULT NOW()
  );
  -- Trigger to automatically update updated_at on row update
  CREATE OR REPLACE FUNCTION update_updated_at_column()
  RETURNS TRIGGER AS $$
  BEGIN
      NEW.updated_at = NOW();
      RETURN NEW;
  END;
  $$ LANGUAGE plpgsql;

  DROP TRIGGER IF EXISTS set_updated_at ON accounts;
  CREATE TRIGGER set_updated_at
  BEFORE UPDATE ON accounts
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at_column();
  ```
- **Initialization**: The schema is automatically loaded into the database on first run via Docker Compose volume mount.
- **Note**: The `updated_at` column is automatically updated via a database trigger whenever a row is updated.

---

## 8. Used Packages

- **github.com/kataras/iris/v12**: Web framework for HTTP server and routing.
- **github.com/lib/pq**: PostgreSQL driver for Go's `database/sql` package.
- **github.com/shopspring/decimal**: Arbitrary-precision decimal arithmetic for handling money safely.
- **github.com/go-playground/validator/v10**: Struct and field validation for incoming API requests.
- **github.com/joho/godotenv**: Loads environment variables from `.env` files for configuration.
- **github.com/golang/mock**: Mocking framework for unit tests.
- **github.com/stretchr/testify**: Assertions and test helpers for Go tests.
- **github.com/DATA-DOG/go-sqlmock**: SQL driver mock for testing database interactions.

---

## 9. Docker Build Process

- The Docker build uses a **multi-stage build**:
  1. **Build Stage**: Compiles the Go binary.
  2. **Test Stage**: Runs all Go tests. The build will fail if any test fails.
  3. **Final Stage**: Copies only the compiled binary and required data into a fresh Alpine image, resulting in a minimal, production-ready image.
- This approach ensures that only tested, minimal artifacts are shipped in the final container, reducing attack surface and image size.

---

## 10. Areas of Improvement

- Add authentication and authorization for API endpoints.
- Implement transaction history and persistence.
- Add pagination and filtering for account listings.
- Improve error messages and API documentation (e.g., Swagger/OpenAPI).
- Add health checks and metrics endpoints.
- Support for running migrations (e.g., with `golang-migrate`).
- Add CI/CD pipeline for automated testing and deployment.
- Enhance test coverage, including integration tests.
- Refactor database transaction handling and consider moving the funds transfer logic from the service layer to the repository layer for better transactional consistency. However, note that this would move business logic out of the service layer, which is generally not desirable, but possible if stricter transactional guarantees are needed. 
