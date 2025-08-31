# Pinjol - Loan Management System with Observability

Aplikasi manajemen pinjaman (Pinjol) dengan observability stack lengkap untuk monitoring dan alerting. Termasuk migrasi dari Promtail ke Grafana Alloy, business dashboard, dan smoke test simulation.

## ✨ Features

- ✅ **Loan Management**: Create, track, and manage loans with flat interest calculation
- ✅ **Payment Processing**: Weekly payment tracking with delinquency monitoring
- ✅ **REST API**: Clean REST endpoints for all operations
- ✅ **CLI Interface**: Command-line interface using Cobra framework
- ✅ **SQLite Database**: Embedded database for data persistence
- ✅ **Database Migrations**: Proper schema migration system
- ✅ **Structured Logging**: JSON logging with Loki integration
- ✅ **Metrics & Monitoring**: Prometheus metrics with Grafana dashboards
- ✅ **Grafana Alloy**: Latest log aggregation replacing deprecated Promtail
- ✅ **Business Dashboard**: Real-time business metrics and KPIs
- ✅ **Smoke Test Simulation**: Realistic user behavior simulation
- ✅ **Modular Architecture**: Clean separation of concerns with packages
- ✅ **Docker Support**: Full containerization with docker-compose
- ✅ **Nix Environment**: Reproducible development environment
- ✅ **Comprehensive Testing**: Unit, integration, and smoke tests

## 🚀 Quick Start

### Development Environment

```bash
# Clone repository
git clone <repository-url>
cd pinjol

# Setup development environment (Nix)
nix develop

# Initialize database
make db-init

# Start the application server
make run

# Or use the CLI command
./pinjol serve

# Access applications
# - Pinjol App: http://localhost:8080
# - Grafana: http://localhost:3000
# - Prometheus: http://localhost:9090
# - Loki: http://localhost:3100
```

### Using CLI Commands

The application now uses a CLI-based approach with Cobra:

```bash
# Start the server
./pinjol serve --port 8080 --env dev

# Initialize database
./pinjol db-init

# Run CLI scenarios
./pinjol scenario --scenario ontime --repeat 10
```

### Run Smoke Test Simulation

```bash
# Quick 5-minute simulation
make simulation-5m

# Standard 30-minute simulation
make simulation-30m

# Custom simulation
make simulation SIMULATION_DURATION=10m SIMULATION_USERS=8 SIMULATION_MAX_REQUESTS=30

# Or use the wrapper script
./run_simulation.sh -d 30m -u 5 -r 50
```

## 📊 Dashboards

### Business Dashboard
Monitor business metrics dan KPIs:
- Total active loans
- Revenue tracking
- Overdue loans
- Payment success rates

**URL**: http://localhost:3000/d/pinjol-business-dashboard

### Logs Dashboard
Monitor application logs dan errors:
- Log volume by level
- Error trends
- Recent error logs
- Application activity logs

**URL**: http://localhost:3000/d/pinjol-logs-dashboard

### Application Dashboard
Monitor system performance:
- Go runtime metrics
- Database statistics
- HTTP request metrics
- System health

**URL**: http://localhost:3000/d/pinjol-main-dashboard

## 🛠️ API Endpoints

### Loan Management
```bash
# Create loan
POST /loans
{
  "principal": 5000000,
  "annual_rate": 0.12,
  "start_date": "2025-01-01"
}

# Get loan details
GET /loans/{id}

# Make payment
POST /loans/{id}/pay
{
  "amount": 100000
}

# Get outstanding amount
GET /loans/{id}/outstanding

# Check delinquency status
GET /loans/{id}/delinquent
```

### System Endpoints
```bash
# Health check
GET /healthz

# Application metrics
GET /metrics

# Version info
GET /version
```

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CLI Commands  │    │   HTTP Server   │    │  Domain Service │
│   (cmd/)        │    │   (internal/)   │    │   (pkg/domain)  │
│                 │    │                 │    │                 │
│ - serve         │───▶│ - Handlers      │───▶│ - Business      │
│ - db-init       │    │ - Middleware    │    │   Logic         │
│ - scenario      │    │ - Routing       │    │ - Validation    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       ▲                     ▲
         ▼                       │                     │
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Repository    │    │   Database      │    │   Shared Libs   │
│  (internal/)    │    │  (migrations/)  │    │    (pkg/)       │
│                 │    │                 │    │                 │
│ - SQLite        │◄──►│ - Schema        │◄──►│ - Logging       │
│ - CRUD Ops      │    │ - Migrations    │    │ - Metrics       │
│ - Transactions  │    │ - Seeds         │    │ - Monitoring    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Architecture Overview

**Modular Package Structure:**
- **`cmd/`**: CLI commands using Cobra framework for command-line interface
- **`internal/`**: Private application code (handlers, database operations)
- **`pkg/domain/`**: Core business logic and domain models
- **`pkg/common/`**: Shared utilities and common functionality
- **`pkg/logging/`**: Structured logging with Loki integration
- **`pkg/metrics/`**: Prometheus metrics collection
- **`pkg/monitoring/`**: Application monitoring endpoints
- **`pkg/profiling/`**: Performance profiling setup
- **`migrations/`**: Database schema migrations
- **`scripts/`**: Automation scripts for development and deployment

**Clean Architecture Principles:**
- **Dependency Injection**: Repository pattern with interface-based design
- **Domain-Driven Design**: Business logic separated from infrastructure
- **Layered Architecture**: Clear separation between CLI, HTTP, Domain, and Data layers
- **Testability**: Each layer can be tested independently
- **Modularity**: Shared packages can be reused across different parts of the application

## 📈 Monitoring Stack

### Metrics Collected
- **Business Metrics**: Loans created, payments processed, revenue
- **Performance Metrics**: Response times, database queries, cache hits
- **System Metrics**: CPU, memory, disk, network
- **Application Metrics**: Goroutines, GC stats, heap usage

### Log Aggregation
- **Application Logs**: Structured JSON logs with context
- **Error Logs**: Detailed error information with stack traces
- **Audit Logs**: User actions and system events
- **Performance Logs**: Slow queries and bottlenecks

## 🧪 Testing

### Unit Tests
```bash
make test-unit
```

### API Tests
```bash
make test-api
```

### Smoke Test Simulation
```bash
# Various simulation presets
make simulation-5m   # 5 minutes
make simulation-30m  # 30 minutes (recommended)
make simulation-1h   # 1 hour

# Custom simulation
make simulation SIMULATION_DURATION=15m SIMULATION_USERS=10
```

### Load Testing
```bash
# Using the simulation script directly
./run_simulation.sh -d 30m -u 20 -r 100
```

### CLI Testing Tools

The project includes CLI tools for testing various loan scenarios:

#### On-time Payment Scenario
```bash
make cli-ontime
# Or: ./pinjol scenario --scenario ontime --repeat 10 --verbose
```

#### Delinquency and Catch-up Scenario
```bash
make cli-skip2
# Or: ./pinjol scenario --scenario skip2 --verbose
```

#### Full Payment Scenario
```bash
make cli-fullpay
# Or: ./pinjol scenario --scenario fullpay --verbose
```

#### Custom CLI Usage
```bash
./pinjol scenario --scenario ontime --principal 5000000 --rate 0.10 --repeat 10 --verbose
```

#### Database Operations
```bash
./pinjol db-init --db-path ./data/pinjol.db
```

## 🐳 Docker Deployment

### Development
```bash
# Build and run with docker-compose
make compose

# Run in detached mode
make compose-detached

# View logs
make compose-logs
```

### Production
```bash
# Build optimized image
make prod-build

# Push to registry
make docker-push
```

### Optimized Production Build (Recommended)

The project includes an optimized Dockerfile using Distroless for production-ready deployments:

```bash
# Build optimized image with BuildKit for faster builds
DOCKER_BUILDKIT=1 make docker-build

# Run with Docker Compose (includes security hardening)
make compose

# Or run directly
docker run -p 8080:8080 pinjol:latest
```

#### Production Features:
- **Distroless base image** for minimal attack surface
- **Static binary linking** for CGO compatibility
- **Non-root user** for security
- **Health checks** for container orchestration
- **Security hardening** with read-only filesystem
- **Only 32MB image size** for fast deployments

### Development Build

For development with debugging tools:

```bash
# Build development image
make docker

# Run with development features
docker run -p 8080:8080 -v pinjol_data:/data pinjol:dev
```

## 📚 Documentation

- [API Documentation](./API.md)
- [Business Dashboard Guide](./BUSINESS_DASHBOARD_README.md)
- [Smoke Test Guide](./SMOKE_TEST_README.md)
- [Monitoring Setup](./monitoring/README.md)
- [Monitoring Guide](./monitoring/MONITORING_GUIDE.md)
- [Production Deployment](./monitoring/PRODUCTION_DEPLOYMENT.md)

## 🔧 Configuration

### Environment Variables
```bash
# Application
PORT=8080
APP_ENV=production
DATABASE_PATH=/data/pinjol.db
LOG_LEVEL=info

# Simulation (for smoke tests)
SIMULATION_DURATION=30m
SIMULATION_USERS=5
SIMULATION_MAX_REQUESTS=50
PINJOL_URL=http://localhost:8081
```

### Database Schema
```sql
-- Loans table
CREATE TABLE loans (
    id TEXT PRIMARY KEY,
    principal INTEGER NOT NULL,
    annual_rate REAL NOT NULL,
    start_date TEXT NOT NULL,
    weekly_due INTEGER NOT NULL,
    schedule TEXT NOT NULL, -- JSON
    paid_count INTEGER DEFAULT 0,
    outstanding INTEGER NOT NULL,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP
);
```

## 📁 Project Structure

```plaintext
.
├── main.go                    // Application entry point (calls cmd.Execute())
├── cmd/                       // CLI commands using Cobra
│   ├── root.go               // Root CLI command
│   ├── serve.go              // HTTP server command
│   ├── dbinit.go             // Database initialization command
│   ├── scenario.go           // CLI testing scenarios
│   └── smoke/
│       └── smoke_simulation.go // Smoke test simulation
├── internal/                  // Internal application code
│   ├── handlers.go           // HTTP handlers
│   ├── database.go           // Database operations
│   ├── handlers_test.go      // Handler tests
│   ├── handlers_edge_test.go // Edge case tests
│   └── database_test.go      // Database tests
├── pkg/                      // Shared packages
│   ├── common/               // Common utilities
│   │   ├── config.go         // Configuration helpers
│   │   ├── errors.go         // Common error types
│   │   └── validation.go     // Validation utilities
│   ├── domain/               // Business domain logic
│   │   ├── loan.go           // Loan domain model
│   │   ├── loan_service.go   // Loan business logic
│   │   ├── repository.go     // Repository interface
│   │   ├── errors.go         // Domain errors
│   │   └── loan_test.go      // Domain tests
│   ├── logging/              // Structured logging
│   │   ├── echo.go           // Echo middleware
│   │   └── logger.go         // Logger implementation
│   ├── metrics/              // Prometheus metrics
│   │   ├── metrics.go        // Metrics collection
│   │   └── integration_example.go // Metrics examples
│   ├── monitoring/           // Application monitoring
│   │   └── monitoring.go     // Monitoring endpoints
│   └── profiling/            // Performance profiling
│       └── profiling.go      // Profiling setup
├── migrations/               // Database migrations
│   └── 001_create_tables.sql // Initial schema
├── scripts/                  // Utility scripts
│   ├── monitoring.sh         // Monitoring stack management
│   ├── run_simulation.sh     // Simulation runner
│   ├── test-dashboard.sh     // Dashboard testing
│   └── test-logs.sh          // Log testing
├── tests/                    // Integration tests
│   ├── docker-compose.test.yml // Test environment
│   └── integration/
│       ├── fixtures.go       // Test data fixtures
│       └── repository_test.go // Repository integration tests
├── data/                     // Data directory
│   └── pinjol.db             // SQLite database file
├── docker/                   // Docker configurations
│   └── monitoring/           // Monitoring stack
├── smoke_simulation          // Smoke test binary
├── pinjol                    // Main application binary
├── coverage.out              // Test coverage output
├── Dockerfile                // Container configuration
├── docker-compose.yml        // Development orchestration
├── flake.nix                 // Nix development environment
├── Makefile                  // Build and development tasks
├── go.mod & go.sum           // Go module files
└── README.md                 // This file
```

## ⚡ Development Commands

```bash
# Application
make run                    # Start the application
make build                  # Build the application
make build-static           # Build static binary

# Database
make db-init               # Initialize database
make db-migrate            # Run database migrations

# Testing
make test                  # Run all tests
make test-verbose          # Run tests with verbose output
make test-coverage         # Run tests with coverage
make test-race             # Run tests with race detection

# CLI Scenarios
make cli-ontime            # Test on-time payment scenario
make cli-skip2             # Test delinquency scenario
make cli-fullpay           # Test full payment scenario

# Simulation
make simulation-5m         # 5-minute simulation
make simulation-30m        # 30-minute simulation (recommended)
make simulation-1h         # 1-hour simulation
make simulation-custom     # Custom simulation parameters

# Docker
make docker-build          # Build Docker image
make docker-run            # Run Docker container
make compose               # Start with docker-compose
make compose-detached      # Start in background
make compose-down          # Stop containers
make compose-logs          # View container logs

# Monitoring
make monitoring-start      # Start monitoring stack
make monitoring-stop       # Stop monitoring stack
make monitoring-status     # Check monitoring status
make monitoring-logs       # View monitoring logs

# Full Development Environment
make dev-full              # Start everything (app + monitoring)
make dev-stop              # Stop everything

# Health Checks
make health-check          # Check all services health

# Code Quality
make lint                  # Run linter
make fmt                   # Format code
make vet                   # Run go vet

# Production
make prod-build            # Production build with all checks
```

## 🚀 Performance Optimizations

### Build Time Optimizations:
- **Docker layer caching** with dependency-first copying
- **BuildKit support** for parallel builds
- **Multi-stage builds** to reduce final image size
- **Selective file copying** with .dockerignore

### Runtime Optimizations:
- **Static binary** with stripped symbols
- **Minimal base image** (Distroless)
- **Connection pooling** for database
- **Optimized Go build flags** for performance

### Security Features:
- **Non-root execution**
- **Read-only filesystem**
- **Minimal attack surface**
- **No shell in production image**

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

For support and questions:
- 📧 Email: support@pinjol.com
- 📖 Documentation: [docs.pinjol.com](https://docs.pinjol.com)
- 🐛 Issues: [GitHub Issues](https://github.com/pinjol/issues)

---

**Made with ❤️ for better loan management**

## Penjelasan (Indonesian Documentation)

### 1. Problem Statement

**Masalah yang Diatasi:**
Dalam industri fintech khususnya peer-to-peer lending (pinjol), diperlukan sistem billing yang dapat:
- Mengelola pinjaman dengan bunga flat selama 50 minggu
- Memproses pembayaran secara FIFO (First-In, First-Out) tanpa boleh skip minggu
- Mendeteksi delinquency secara otomatis berdasarkan riwayat pembayaran
- Menyediakan API RESTful untuk integrasi dengan frontend/mobile
- Memastikan persistensi data yang reliable menggunakan database
- Menangani skenario testing yang beragam untuk validasi bisnis logic

**Konteks Bisnis:**
- Produk pinjaman default: Rp 5.000.000 dengan bunga 10% p.a. (flat)
- Pembayaran mingguan konstan: Rp 110.000
- Delinquency trigger: 2 minggu terakhir belum dibayar
- Waktu mulai pinjaman dapat bervariasi

### 2. Approach Design Code

**Arsitektur yang Diambil:**
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CLI Commands  │    │   HTTP Server   │    │  Domain Service │
│     (cmd/)      │───▶│   (internal/)   │───▶│  (pkg/domain)   │
│                 │    │                 │    │                 │
│ - serve         │    │ - Handlers      │    │ - Business      │
│ - db-init       │    │ - Middleware    │    │   Logic         │
│ - scenario      │    │ - Routing       │    │ - Validation    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       ▲                     ▲
         ▼                       │                     │
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Repository    │    │   Database      │    │   Shared Libs   │
│  (internal/)    │    │  (migrations/)  │    │    (pkg/)       │
│                 │    │                 │    │                 │
│ - SQLite        │◄──►│ - Schema        │◄──►│ - Logging       │
│ - CRUD Ops      │    │ - Migrations    │    │ - Metrics       │
│ - Transactions  │    │ - Seeds         │    │ - Monitoring    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

**Prinsip Desain:**
- **Modular Architecture**: Kode terorganisir dalam package terpisah (cmd/, internal/, pkg/)
- **CLI-First Design**: Aplikasi menggunakan Cobra untuk command-line interface
- **Domain-Driven Design (DDD):** Logika bisnis inti dipisahkan di `pkg/domain/`
- **Repository Pattern:** Abstraksi data access melalui interface
- **Clean Architecture:** Handler hanya sebagai thin layer untuk HTTP
- **Dependency Injection:** Semua dependencies diinject melalui constructor
- **Structured Logging:** Logging terstruktur dengan Loki integration
- **Metrics Collection:** Prometheus metrics untuk monitoring
- **Test-Driven Development:** Coverage >80% dengan unit dan integration tests

**Teknologi Stack:**
- **Go 1.24+:** Untuk performance dan concurrency
- **Cobra:** CLI framework untuk command-line interface
- **Echo Framework:** Lightweight HTTP router
- **SQLite:** Embedded database untuk simplicity
- **Nix:** Reproducible development environment
- **Docker:** Containerization untuk deployment
- **Prometheus:** Metrics collection
- **Grafana:** Visualization dashboards
- **Loki:** Log aggregation
- **Grafana Alloy:** Log shipping (replacing Promtail)

### 3. Skenario Smoke Test

**Langkah Smoke Test untuk Validasi Sistem:**

1. **Setup Environment:**
   ```bash
   nix develop
   make db-init
   ./pinjol serve &
   ```

2. **Test Case 1: Pembuatan Pinjaman Normal**
   ```bash
   curl -X POST http://localhost:8080/loans \
     -H "Content-Type: application/json" \
     -d '{"principal": 5000000, "annual_rate": 0.10, "start_date": "2025-08-15"}'
   # Expected: HTTP 201, loan ID returned
   ```

3. **Test Case 2: Pembayaran Pertama**
   ```bash
   curl -X POST http://localhost:8080/loans/{loan_id}/pay \
     -H "Content-Type: application/json" \
     -d '{"amount": 110000}'
   # Expected: HTTP 200, payment processed
   ```

4. **Test Case 3: Cek Outstanding Balance**
   ```bash
   curl http://localhost:8080/loans/{loan_id}/outstanding
   # Expected: Outstanding = 5,500,000 - 110,000 = 5,390,000
   ```

5. **Test Case 4: Cek Status Delinquency**
   ```bash
   curl http://localhost:8080/loans/{loan_id}/delinquent
   # Expected: false (masih on-time)
   ```

6. **Test Case 5: CLI Scenario Testing**
   ```bash
   # Using CLI commands
   ./pinjol scenario --scenario ontime --repeat 10 --verbose
   ./pinjol scenario --scenario skip2 --verbose
   ./pinjol scenario --scenario fullpay --verbose

   # Or using make commands
   make cli-ontime
   make cli-skip2
   make cli-fullpay
   ```

7. **Test Case 6: Smoke Test Simulation**
   ```bash
   # Run smoke test simulation
   make simulation-30m

   # Or use the CLI directly
   CGO_ENABLED=1 go run ./cmd/smoke/smoke_simulation.go
   ```

### 4. Edge Cases dan Analisis

**Edge Cases Utama:**

#### 4.1 Pembayaran dengan Jumlah Salah
```
Scenario: User mencoba bayar Rp 100,000 padahal harus Rp 110,000
Expected: Error "wrong amount"
```
**ASCII Diagram:**
```
Loan Schedule (Minggu 1-5):
┌─────┬─────┬─────┬─────┬─────┐
│  1  │  2  │  3  │  4  │  5  │  ← Belum dibayar
│  ✓  │     │     │     │     │  ← Sudah dibayar
└─────┴─────┴─────┴─────┴─────┘
      ↑
   Next Payment (Minggu 1)
   
Input: amount = 100,000 (wrong)
Output: Error - Amount must be exactly 110,000
```

#### 4.2 Pinjaman Sudah Lunas
```
Scenario: Semua 50 minggu sudah dibayar, user coba bayar lagi
Expected: Error "already paid"
```
**ASCII Diagram:**
```
Loan Schedule (Minggu 46-50):
┌─────┬─────┬─────┬─────┬─────┐
│ 46  │ 47  │ 48  │ 49  │ 50  │
│  ✓  │  ✓  │  ✓  │  ✓  │  ✓  │  ← Semua sudah dibayar
└─────┴─────┴─────┴─────┴─────┘
   
Input: amount = 110,000
Output: Error - Loan already fully paid
```

#### 4.3 Delinquency Detection
```
Scenario: Minggu ke-3, minggu 1 & 2 belum dibayar
Expected: Delinquent = true, streak = 2
```
**ASCII Diagram:**
```
Timeline: Start Date = 2025-08-01
Current: 2025-08-15 (Week 3)

Week Index:     1     2     3
┌─────────────┬─────┬─────┬─────┐
│ 2025-08-01 │ 08  │ 15  │ 22  │
├─────────────┼─────┼─────┼─────┤
│   Unpaid   │  ✗  │  ✗  │     │  ← Latest 2 weeks unpaid
│   Status   │     │     │     │
└─────────────┴─────┴─────┴─────┘

Delinquency Check:
- Observed Week: 3
- Check Weeks: 1 & 2 (latest 2 scheduled)
- Result: DELINQUENT (streak = 2)
```

#### 4.4 Tanggal di Luar Range
```
Scenario: WeekIndexAt() untuk tanggal > 50 minggu
Expected: Return 50 (capped)
```

#### 4.5 Database Constraint Violations
```
Scenario: Duplicate loan ID creation
Expected: UNIQUE constraint failed error
```

### 5. Pengembangan dan Improvement Selanjutnya

**Short-term Improvements:**
- **Performance Optimization:**
  - Connection pooling untuk database
  - Caching untuk frequent queries
  - Database indexing optimization

- **Monitoring & Observability:**
  - Structured logging dengan levels
  - Metrics collection (Prometheus)
  - Health check endpoints enhancement

- **Security Enhancements:**
  - Input validation middleware
  - Rate limiting
  - Authentication/Authorization

**Medium-term Features:**
- **Multi-tenancy Support:** Multiple loan products
- **Batch Operations:** Bulk payment processing
- **Notification System:** Email/SMS alerts untuk delinquency
- **API Versioning:** v2 API dengan backward compatibility

**Long-term Vision:**
- **Microservices Architecture:** Split domain services
- **Event-Driven Design:** Event sourcing untuk audit trail
- **Machine Learning:** Risk assessment untuk loan approval
- **Multi-currency Support:** International expansion

**Technical Debt & Refactoring:**
- **Package Organization:** Struktur package sudah diorganisir dengan baik (cmd/, internal/, pkg/)
- **CLI Framework:** Implementasi Cobra untuk CLI yang robust
- **Modular Design:** Kode terpisah dalam package yang dapat di-test secara independen
- **Dependency Injection:** Semua dependencies menggunakan constructor injection
- **Configuration Management:** Centralized configuration dengan Viper
- **Error Handling:** Structured error handling dengan custom error types
- **Logging & Monitoring:** Comprehensive logging dan metrics collection
- **Database Migrations:** Proper migration system untuk schema changes
- **Testing Strategy:** Unit tests, integration tests, dan smoke tests
- **CI/CD Ready:** Makefile dengan semua commands untuk automation

**Current Architecture Benefits:**
- **Scalability:** Modular design memungkinkan easy scaling
- **Maintainability:** Clear separation of concerns
- **Testability:** Each package dapat di-test independently
- **Reusability:** Shared packages dapat digunakan di berbagai bagian aplikasi
- **Observability:** Comprehensive monitoring dan logging
- **Developer Experience:** CLI tools dan scripts untuk development workflow

---

*Proyek ini mendemonstrasikan implementasi clean architecture dengan fokus pada domain logic yang solid, testing yang komprehensif, dan deployment yang reproducible menggunakan Nix dan Docker. Aplikasi telah di-restructure menjadi modular architecture dengan CLI framework, comprehensive monitoring, dan development workflow yang efisien.*
