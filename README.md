# Pinjol

Pinjol is a billing engine for flat interest loans built using the Echo framework. This project demonstrates a flat project structure and includes a Nix development environment for easy reproducibility.

## Features

- **Billing Engine**: 50-week flat interest loan management
- **RESTful API**: Complete loan lifecycle management
- **Payment Processing**: FIFO payment enforcement with exact amount validation
- **Delinquency Detection**: Automatic calculation based on payment history
- **CLI Testing Tools**: Command-line scenarios for testing loan behavior
- **Middleware**: Request logging and error handling
- **Dockerized**: Containerized deployment support
- **Nix-based**: Reproducible development environment

## Loan Product Specifications

- **Term**: 50 weeks
- **Default Principal**: Rp 5,000,000
- **Default Interest**: 10% p.a. (flat)
- **Weekly Payment**: Constant amount (Rp 110,000 for default)
- **Payment Order**: FIFO (First-In, First-Out) - no skipping weeks
- **Delinquency**: Triggered when latest 2 scheduled weeks are unpaid

## Prerequisites

- [Nix](https://nixos.org/): A tool for reproducible builds and development environments.

## Installing Nix

To install Nix, follow the instructions below or visit the [Zero to Nix](https://zero-to-nix.com/start/install/) website for more details.

Run the following command to install Nix:

```bash
curl --proto '=https' --tlsv1.2 -sSf -L https://install.determinate.systems/nix | sh -s -- install
```

After installation, open a new terminal session to ensure the `nix` command is available.

## Getting Started

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd pinjol
   ```

2. Enter the Nix development shell:
   ```bash
   nix develop
   ```

3. Run the application:
   ```bash
   make run
   ```

4. Run tests:
   ```bash
   make test
   ```

## API Endpoints

### Create Loan
```bash
POST /loans
Content-Type: application/json

{
  "principal": 5000000,
  "annual_rate": 0.10,
  "start_date": "2025-08-15"
}
```

### Make Payment
```bash
POST /loans/{id}/pay
Content-Type: application/json

{
  "amount": 110000
}
```

### Check Outstanding Balance
```bash
GET /loans/{id}/outstanding
```

### Check Delinquency Status
```bash
GET /loans/{id}/delinquent[?now=YYYY-MM-DD]
```

### Get Loan Details
```bash
GET /loans/{id}
```

## CLI Testing Tools

The project includes CLI tools for testing various loan scenarios:

### On-time Payment Scenario
```bash
make cli-ontime
```

### Delinquency and Catch-up Scenario
```bash
make cli-skip2
```

### Full Payment Scenario
```bash
make cli-fullpay
```

### Custom CLI Usage
```bash
go run . cli --scenario ontime --principal 5000000 --rate 0.10 --repeat 10 --verbose
```

## Project Structure

```plaintext
.
├── main.go              // Application bootstrap and routing
├── handlers.go          // HTTP handlers (thin layer)
├── loans.go             // Domain logic: loans, payments, delinquency
├── middleware.go        // Request logging middleware
├── config.go            // Environment variable helpers
├── errors.go            // Error types and definitions
├── version.go           // Version information structure
├── cli.go               // CLI testing tools
├── runner_spec.md       // Detailed specification document
├── loans_test.go        // Unit tests for domain logic
├── handlers_test.go     // Basic handler tests
├── api_test.go          // API integration tests
├── tests/               // Test directories
├── Makefile             // Build and test commands
├── Dockerfile           // Container configuration
├── docker-compose.yml   // Container orchestration
├── flake.nix            // Nix development environment
└── README.md            // This file
```

## Development Commands

```bash
# Run application
make run

# Run tests
make test
make test-verbose
make test-coverage

# CLI tools
make cli-ontime
make cli-skip2
make cli-fullpay

# Build
make build

# Docker
make docker
make compose

# Lint (requires golangci-lint)
make lint
```

## Architecture

This billing engine follows these principles:

- **Domain-Driven Design**: Core business logic in `loans.go`
- **FIFO Payment Processing**: Strict order enforcement
- **Deterministic Calculations**: No randomness, all math is exact
- **UTC Time Handling**: All internal times use UTC
- **Integer Amounts**: All monetary values are int64 rupiah (no floats)
- **Structured Logging**: Request logging with correlation IDs
- **Error Handling**: Specific error types for different failure modes

## License

This project is licensed under the MIT License.
