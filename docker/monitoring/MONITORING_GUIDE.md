# Pinjol Monitoring & Dashboard Setup

## **Industrial-Grade Monitoring Solutions**

### **Option 1: Grafana + Prometheus (RECOMMENDED)**
**Implemented** - Production-ready monitoring stack

**Features:**
- **Prometheus**: Time-series metrics database
- **Grafana**: Advanced visualization dashboards
- **AlertManager**: Intelligent alert management
- **Node Exporter**: System metrics collection
- **cAdvisor**: Container metrics
- **Pre-built Dashboards**: Application, System, Business metrics

---

### **Option 2: DataDog (SaaS Alternative)**
**For teams preferring managed solutions:**

```yaml
# datadog.yaml
api_key: "your-datadog-api-key"
app_key: "your-datadog-app-key"
tags:
  - env:production
  - service:pinjol
```

**Benefits:**
- Zero infrastructure management
- Built-in APM and tracing
- Advanced anomaly detection
- 15+ integrations out-of-the-box

---

### **Option 3: New Relic (APM Focused)**
**For application performance monitoring:**

```yaml
# newrelic.yml
common: &default_settings
  license_key: 'your-license-key'
  app_name: 'Pinjol Production'
  distributed_tracing:
    enabled: true
  attributes:
    enabled: true
```

**Benefits:**
- Code-level performance insights
- Automatic error tracking
- Transaction tracing
- Infrastructure monitoring

---

## **Quick Start with Grafana + Prometheus**

### **1. Start Monitoring Stack**
```bash
# Make script executable
chmod +x monitoring.sh

# Start all services
./monitoring.sh start
```

### **2. Access Dashboards**
- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **AlertManager**: http://localhost:9093

### **3. Pre-configured Dashboards**
1. **Application Dashboard** - Go runtime metrics, HTTP endpoints
2. **System Dashboard** - CPU, Memory, Disk, Network
3. **Business Dashboard** - Loan metrics, revenue, performance

---

## **Dashboard Features**

### **Application Dashboard**
- ✅ Health status monitoring
- ✅ HTTP request rate & latency
- ✅ Error rate tracking
- ✅ Go runtime metrics (goroutines, memory, GC)
- ✅ Database connection pooling

### **System Dashboard**
- ✅ CPU usage & load average
- ✅ Memory utilization
- ✅ Disk I/O and usage
- ✅ Network traffic monitoring

### **Business Dashboard**
- ✅ Loan creation & approval rates
- ✅ Payment success tracking
- ✅ Revenue metrics
- ✅ Database query performance
- ✅ Status distribution charts

---

## **Configuration**

### **Prometheus Configuration**
```yaml
# docker/monitoring/prometheus.yml
scrape_configs:
  - job_name: 'pinjol'
    static_configs:
      - targets: ['host.docker.internal:8080']
    metrics_path: '/metrics'
```

### **Alert Rules**
```yaml
# Add to prometheus.yml
rule_files:
  - "alert_rules.yml"

# alert_rules.yml
groups:
  - name: pinjol
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~'5..'}[5m]) > 0.1
        for: 5m
        labels:
          severity: critical
```

---

## **Custom Metrics Integration**

### **Add Business Metrics to Your Code**
```go
// In your handlers
import "github.com/prometheus/client_golang/prometheus"

// Define metrics
var (
    loansCreated = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "pinjol_loans_created_total",
            Help: "Total number of loans created",
        },
        []string{"status"},
    )
)

// Register metrics
func init() {
    prometheus.MustRegister(loansCreated)
}

// Use in handlers
loansCreated.WithLabelValues("approved").Inc()
```

---

## 🎛️ **Management Commands**

```bash
# Start monitoring
./monitoring.sh start

# Check status
./monitoring.sh status

# View logs
./monitoring.sh logs grafana
./monitoring.sh logs prometheus

# Stop monitoring
./monitoring.sh stop

# Clean up (removes all data)
./monitoring.sh clean
```

---

## **Security Considerations**

### **Production Deployment**
1. **Change default passwords**
2. **Use HTTPS with certificates**
3. **Configure authentication**
4. **Network segmentation**
5. **Regular security updates**

### **Grafana Security**
```yaml
# grafana.ini
[security]
admin_password = "strong-password"
[auth]
disable_login_form = false
[users]
allow_sign_up = false
```

---

## **Additional Resources**

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Monitoring Best Practices](https://prometheus.io/docs/practices/)
- [AlertManager Configuration](https://prometheus.io/docs/alerting/latest/alertmanager/)

---

## **Recommended for Production**

1. **Start with Grafana + Prometheus** (implemented)
2. **Add business-specific metrics** to your application
3. **Configure alerts** for critical issues
4. **Set up log aggregation** (ELK stack)
5. **Implement distributed tracing** (Jaeger)
6. **Add APM** (New Relic or DataDog) for complex deployments

This monitoring setup provides enterprise-grade observability for your Pinjol application! 🚀
