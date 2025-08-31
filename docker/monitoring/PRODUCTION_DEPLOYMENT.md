# Production Deployment Guide for Pinjol Monitoring

## **Quick Start**

### **1. Development Setup**
```bash
# Start monitoring stack
./monitoring.sh start

# Access dashboards
# Grafana: http://localhost:3000 (admin/admin)
# Prometheus: http://localhost:9090
```

### **2. Production Deployment**
```bash
# Copy environment file
cp docker/monitoring/.env.example docker/monitoring/.env

# Edit with your values
nano docker/monitoring/.env

# Deploy with production config
docker-compose -f docker/monitoring/docker-compose.prod.yml up -d
```

---

## **Dashboard Overview**

### **Application Dashboard**
- **Health Status**: Real-time application health
- **HTTP Metrics**: Request rate, latency, error rates
- **Go Runtime**: Goroutines, memory usage, GC stats
- **Database**: Connection pool status

### **System Dashboard**
- **CPU Usage**: Core utilization and load averages
- **Memory**: RAM usage and swap statistics
- **Disk I/O**: Read/write operations and usage
- **Network**: Traffic and error statistics

### **Business Dashboard**
- **Loan Metrics**: Creation, approval, rejection rates
- **Revenue**: Total and monthly revenue tracking
- **Payment Success**: Payment processing statistics
- **Performance**: Database query performance

---

## **Configuration**

### **Environment Variables**
```bash
# Required for production
GRAFANA_ADMIN_PASSWORD=your-secure-password
SMTP_USER=alerts@pinjol.com
SMTP_PASSWORD=your-smtp-password
DOMAIN=monitoring.pinjol.com
```

### **Prometheus Configuration**
```yaml
# docker/monitoring/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'pinjol'
    static_configs:
      - targets: ['your-app-host:8080']
    metrics_path: '/metrics'
```

### **AlertManager Configuration**
```yaml
# docker/monitoring/alertmanager.yml
global:
  smtp_smarthost: 'smtp.gmail.com:587'
  smtp_from: 'alerts@pinjol.com'

route:
  receiver: 'email'
  routes:
  - match:
      severity: critical
    receiver: 'email'
```

---

## **Adding Custom Metrics**

### **1. Import Metrics Package**
```go
import "pinjol/pkg/metrics"
```

### **2. Initialize Metrics**
```go
func init() {
    metrics.InitMetrics()
}
```

### **3. Use in Your Code**
```go
// Record loan creation
metrics.RecordLoanCreated("approved", "individual")

// Record payment
metrics.RecordPaymentReceived(1000.0, "bank_transfer", "success")

// Record revenue
metrics.RecordRevenue(1000.0)

// Record errors
metrics.RecordBusinessError("validation_failed", "loan_creation")
```

### **4. Add Metrics Endpoint**
```go
// In your main.go
e.GET("/metrics", echo.WrapHandler(metrics.GetPrometheusHandler()))
```

---

## 🚨 **Alert Configuration**

### **Application Alerts**
- **Service Down**: Alert when application is unreachable
- **High Error Rate**: Alert when 5xx errors > 5%
- **High Latency**: Alert when 95th percentile > 2s
- **Memory Usage**: Alert when heap usage > 90%

### **System Alerts**
- **High CPU**: Alert when CPU usage > 90%
- **Low Memory**: Alert when available memory < 10%
- **Disk Space**: Alert when disk usage > 90%
- **High Load**: Alert when load average > 4

### **Business Alerts**
- **Database Issues**: Alert on connection pool saturation
- **Circuit Breaker**: Alert when circuit breaker opens
- **Business Errors**: Alert on high business logic errors

---

## **Security Best Practices**

### **1. Change Default Passwords**
```bash
# In docker-compose.prod.yml
environment:
  - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_ADMIN_PASSWORD}
```

### **2. Enable HTTPS**
```yaml
# Add to Grafana configuration
[server]
protocol = https
cert_file = /path/to/cert.pem
cert_key = /path/to/key.pem
```

### **3. Network Security**
```yaml
# Use internal networks
networks:
  monitoring:
    internal: true
```

### **4. Access Control**
```yaml
# Grafana configuration
[auth]
disable_login_form = false
[users]
allow_sign_up = false
```

---

## **Monitoring Best Practices**

### **1. Key Metrics to Monitor**
- **Application Health**: Uptime, response time, error rates
- **Resource Usage**: CPU, memory, disk, network
- **Business Metrics**: Revenue, user activity, conversion rates
- **Performance**: Database queries, cache hit rates, API latency

### **2. Alert Thresholds**
- **Warning**: 80% resource usage, 1% error rate
- **Critical**: 90% resource usage, 5% error rate
- **Immediate**: Service down, data loss

### **3. Dashboard Organization**
- **Real-time**: Current status and recent activity
- **Historical**: Trends over time (1h, 24h, 7d, 30d)
- **Comparative**: Compare with previous periods
- **Predictive**: Forecast based on trends

---

## **Backup & Recovery**

### **1. Data Backup**
```bash
# Backup Grafana data
docker run --rm -v pinjol_grafana_data:/data -v $(pwd)/backup:/backup alpine tar czf /backup/grafana-$(date +%Y%m%d).tar.gz -C /data .

# Backup Prometheus data
docker run --rm -v pinjol_prometheus_data:/data -v $(pwd)/backup:/backup alpine tar czf /backup/prometheus-$(date +%Y%m%d).tar.gz -C /data .
```

### **2. Configuration Backup**
```bash
# Backup configurations
cp docker/monitoring/prometheus.yml backup/
cp docker/monitoring/alertmanager.yml backup/
cp docker/monitoring/grafana/provisioning/ backup/ -r
```

---

## **Scaling Considerations**

### **1. High Availability**
```yaml
# Multiple Prometheus instances
prometheus:
  replicas: 2
  loadBalancer:
    type: LoadBalancer
```

### **2. Federation**
```yaml
# Global view across multiple clusters
scrape_configs:
  - job_name: 'federate'
    honor_labels: true
    metrics_path: '/federate'
    params:
      'match[]':
        - '{job="prometheus"}'
        - '{__name__=~".+"}'
```

### **3. Long-term Storage**
```yaml
# Use remote storage for long retention
remote_write:
  - url: "http://remote-storage:9201/write"
remote_read:
  - url: "http://remote-storage:9201/read"
```

---

## **Troubleshooting**

### **Common Issues**

1. **Grafana not accessible**
   ```bash
   # Check container status
   docker ps | grep grafana

   # Check logs
   docker logs pinjol_grafana
   ```

2. **Metrics not appearing**
   ```bash
   # Check Prometheus targets
   curl http://localhost:9090/api/v1/targets

   # Check metrics endpoint
   curl http://localhost:8080/metrics
   ```

3. **Alerts not working**
   ```bash
   # Check AlertManager status
   curl http://localhost:9093/api/v1/status

   # Check alert rules
   curl http://localhost:9090/api/v1/rules
   ```

---

## **Additional Resources**

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Monitoring Best Practices](https://prometheus.io/docs/practices/)
- [AlertManager Guide](https://prometheus.io/docs/alerting/latest/alertmanager/)

---

## **Next Steps**

1. **Customize Dashboards**: Add business-specific panels
2. **Configure Alerts**: Set up notification channels
3. **Add More Metrics**: Implement custom business metrics
4. **Set up Backup**: Regular backup of monitoring data
5. **Security Hardening**: Implement production security measures
6. **High Availability**: Set up redundant monitoring instances

This monitoring setup provides enterprise-grade observability for your Pinjol application!
