# Pinjol - Loan Management System with Observability

Aplikasi manajemen pinjaman (Pinjol) dengan observability stack lengkap untuk monitoring dan alerting. Termasuk migrasi dari Promtail ke Grafana Alloy, business dashboard, dan smoke test simulation.

## ✨ Features

- ✅ **Loan Management**: Create, track, and manage loans with flat interest calculation
- ✅ **Payment Processing**: Weekly payment tracking with delinquency monitoring
- ✅ **REST API**: Clean REST endpoints for all operations
- ✅ **SQLite Database**: Embedded database for data persistence
- ✅ **Structured Logging**: JSON logging with Loki integration
- ✅ **Metrics & Monitoring**: Prometheus metrics with Grafana dashboards
- ✅ **Grafana Alloy**: Latest log aggregation replacing deprecated Promtail
- ✅ **Business Dashboard**: Real-time business metrics and KPIs
- ✅ **Smoke Test Simulation**: Realistic user behavior simulation
- ✅ **Docker Support**: Full containerization with docker-compose
- ✅ **Nix Environment**: Reproducible development environment

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

# Start full development stack
make dev-full

# Access applications
# - Pinjol App: http://localhost:8080
# - Grafana: http://localhost:3000
# - Prometheus: http://localhost:9090
# - Loki: http://localhost:3100
```

### Run Smoke Test Simulation

```bash
# Quick 5-minute simulation
make simulation-5m

# Standard 30-minute simulation
make simulation-30m

# Custom simulation
make simulation SIMULATION_DURATION=10m SIMULATION_USERS=8 SIMULATION_MAX_REQUESTS=30
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
│   Pinjol App    │    │   Prometheus     │    │     Grafana     │
│                 │    │                 │    │                 │
│ - REST API      │◄──►│ - Metrics       │◄──►│ - Dashboards    │
│ - Business Logic│    │ - Alerting      │    │ - Visualization │
│ - SQLite DB     │    │ - Targets       │    │ - Alerts        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       ▲                     ▲
         ▼                       │                     │
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     Loki        │    │   Alloy         │    │   Node          │
│                 │    │                 │    │   Exporter      │
│ - Log Storage   │◄──►│ - Log Shipping  │◄──►│ - System        │
│ - Query Engine  │    │ - Relabeling    │    │   Metrics       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

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

## 📚 Documentation

- [API Documentation](./API.md)
- [Business Dashboard Guide](./BUSINESS_DASHBOARD_README.md)
- [Smoke Test Guide](./SMOKE_TEST_README.md)
- [Monitoring Setup](./monitoring/README.md)

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
   ```bash
   nix develop
   ```

3. Initialize the database:
   ```bash
   make db-init
   ```

4. Run the application:
   ```bash
   make run
   ```

5. Run tests:
   ```bash
   make test
   ```

## Database

The application uses SQLite for data persistence. The database file is created automatically when you run `make db-init`. You can configure the database path using the `DATABASE_PATH` environment variable:

```bash
export DATABASE_PATH=./my-database.db
make db-init
```

## Docker Setup

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

## Performance Optimizations

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

---

## Penjelasan

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
│   HTTP Handlers │───▶│  Domain Logic   │───▶│   Repository    │
│   (Echo Routes) │    │   (loans.go)    │    │   (SQLite)      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
   RESTful API          Business Rules          Data Persistence
```

**Prinsip Desain:**
- **Domain-Driven Design (DDD):** Logika bisnis inti dipisahkan di `loans.go` dengan struct `Loan` dan method-method seperti `MakePayment()`, `IsDelinquent()`
- **Repository Pattern:** Abstraksi data access melalui interface `LoanRepository` dengan implementasi `SQLiteLoanRepository`
- **Clean Architecture:** Handler hanya sebagai thin layer untuk HTTP, domain logic tidak bergantung pada framework
- **Test-Driven Development:** Coverage >80% untuk komponen core, dengan unit test dan integration test
- **Error Handling:** Custom error types (`ErrInvalidRequest`, `ErrAlreadyPaid`, dll.) untuk handling yang spesifik

**Teknologi Stack:**
- **Go 1.24+:** Untuk performance dan concurrency
- **Echo Framework:** Lightweight HTTP router
- **SQLite:** Embedded database untuk simplicity
- **Nix:** Reproducible development environment
- **Docker:** Containerization untuk deployment

### 3. Skenario Smoke Test

**Langkah Smoke Test untuk Validasi Sistem:**

1. **Setup Environment:**
   ```bash
   nix develop
   make db-init
   make run
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
   make cli-ontime    # Test pembayaran tepat waktu
   make cli-skip2     # Test delinquency dan catch-up
   make cli-fullpay   # Test pembayaran penuh
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
- **Extract common utilities ke shared packages:**
  - Buat package `pkg/common` untuk utilities seperti validation, formatting, dan error handling
  - Extract database connection helpers ke `pkg/database`
  - Buat shared middleware package untuk authentication dan logging
  - Implementasi: `pkg/common/validation.go`, `pkg/database/connection.go`

- **Implement circuit breaker pattern:**
  - Tambahkan circuit breaker untuk database connections menggunakan library seperti `gobreaker`
  - Implement fallback mechanisms untuk external API calls
  - Add health check endpoints untuk monitoring circuit breaker status
  - Konfigurasi threshold untuk failure detection dan recovery

- **Add integration tests dengan database real:**
  - Setup test database menggunakan Docker containers (PostgreSQL/MySQL)
  - Implementasi test fixtures untuk data seeding
  - Add database migration testing untuk schema changes
  - Continuous integration dengan database testing di CI/CD pipeline

- **Performance benchmarking dan profiling:**
  - Implementasi pprof endpoints untuk CPU dan memory profiling
  - Add benchmarking tests untuk critical paths (loan creation, payment processing)
  - Database query optimization dengan EXPLAIN ANALYZE
  - Load testing menggunakan tools seperti Apache Bench atau hey
  - Monitoring response times dan throughput metrics

---

*Proyek ini mendemonstrasikan implementasi clean architecture dengan fokus pada domain logic yang solid, testing yang komprehensif, dan deployment yang reproducible menggunakan Nix dan Docker.*

## License

This project is licensed under the MIT License.
