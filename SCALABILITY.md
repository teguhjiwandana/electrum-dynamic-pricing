# Scalability Plan — Electrum Dynamic Pricing Engine

> Optional document for take-home test. Describes how the system would scale from V1 production to enterprise-grade.

---

## Current State (V1)

| Component | Current | Limits |
|-----------|---------|--------|
| API server | 1 instance (Docker) | ~1000 req/s (Gin) |
| Database | 1 PostgreSQL instance | ~500 QPS (single node) |
| Frontend | 1 NextJS instance (pm2) | ~500 concurrent users |
| Config hot-reload | File watcher (20s poll) | 1 config per 20s |
| Audit log | Append-only table | ~10M rows before partitioning |

---

## Phase 1: Production Readiness (Week 1–2)

### 1. API Horizontal Scaling
```
                    ┌──────────────┐
                    │   Nginx LB   │
                    │  (round-robin)│
                    └──┬───┬───┬──┘
                       │   │   │
               ┌───────▼───▼───▼───────┐
               │  API-1  API-2  API-3 │
               │  (Docker, auto-scale) │
               └───────────┬───────────┘
                           │
                   ┌───────▼───────┐
                   │  PostgreSQL   │
                   │  (primary)    │
                   └───────────────┘
```

**Changes:**
- Remove `gin.Default()` → `gin.New()` (production mode)
- Add `GIN_MODE=release` environment variable
- Docker Compose → Kubernetes or Docker Swarm with 3+ replicas
- Nginx upstream block with health checks
- Stateless API — no session affinity needed

### 2. Database Optimization
- **Connection pooling**: Already using pgxpool (20 max, 2 min). Tune per replica.
- **Read replicas**: Add 1–2 PostgreSQL read replicas for audit queries + zone/vehicle reads
- **Connection string split**: `DATABASE_URL` (write) + `DATABASE_READ_URL` (read)
- **Indexes**: Already present on `audit_log(vehicle_id)`, `audit_log(timestamp DESC)`, `vehicles(zone)`. Add:
  ```sql
  CREATE INDEX idx_audit_config_version ON audit_log(config_version);
  CREATE INDEX idx_config_active ON pricing_config(version DESC, updated_at);
  ```

### 3. Config Hot-Reload at Scale
- **Current**: File watcher polls `pricing_config.json` every 20s (single instance)
- **Scale**: Replace file watcher with PostgreSQL `LISTEN/NOTIFY`
  ```go
  // On config update: NOTIFY config_changed, 'v' || new_version;
  // All API instances subscribe and reload on notification
  ```
- **Or**: Use Redis pub/sub with config cached in Redis (TTL 30s)

### 4. Monitoring & Observability
```go
// Add to main.go
import "github.com/prometheus/client_golang/prometheus"

var (
    pricingRequests = prometheus.NewCounterVec(...)
    pricingLatency  = prometheus.NewHistogramVec(...)
    auditWriteFail  = prometheus.NewCounter(...)
)
```
- **Metrics**: Prometheus `/metrics` endpoint
- **Logging**: Structured logging (zerolog/logrus) with request IDs
- **Alerting**: Grafana dashboards for P95 latency, error rate, DB pool exhaustion

---

## Phase 2: High Traffic (Month 1–3)

### 1. Caching Layer
```
API → Redis → PostgreSQL
```
- **Cache pricing config**: TTL 30s (configurable). Invalidated on config update.
- **Cache zone utilization**: TTL 10s (IoT data changes frequently).
- **Cache vehicle SoC**: TTL 60s.

**Expected impact**: 80% reduction in DB reads for pricing requests.

### 2. Audit Log Scaling
- **Partition by month**: `audit_log_2026_07`, `audit_log_2026_08`, etc.
- **Archive**: Move partitions >6 months to S3-compatible storage (Parquet format)
- **Async write**: Use channel-based buffer or outbox pattern
  ```go
  // Fire-and-forget → buffered channel → batch insert
  auditChan := make(chan AuditEntry, 1000)
  go batchWriter(auditChan)
  ```

### 3. Rate Limiting
```go
// Per-user rate limiting
limiter := rate.NewLimiter(10, 20) // 10 req/s, burst 20
```
- **Pricing API**: 100 req/min per user
- **Admin API**: 30 req/min per user
- **Login**: 5 attempts/min per IP (brute force protection)

### 4. Multi-Region
```
Region A (Jakarta)          Region B (Singapore)
     │                            │
     ├── API replicas             ├── API replicas
     ├── PostgreSQL (write)       ├── PostgreSQL (read replica)
     └── Redis                     └── Redis
```
- **Geo-routing**: Route users to nearest region via DNS (Route 53 latency-based)
- **Write to primary**: All writes go to Jakarta primary
- **Read from local**: Reads served from local replica

---

## Phase 3: Enterprise (Month 3–6)

### 1. Event Sourcing for Audit
- Replace direct DB writes with Kafka/event stream
- Each pricing calculation → `pricing.calculated` event
- Audit log consumer writes to DB + S3 archive
- Enables replay, reprocessing, and analytics

### 2. ML-Based Predictive Pricing
- Replace rule-based demand multipliers with ML model
- Model inputs: historical demand, weather, events, time features
- A/B testing framework: route X% of traffic to ML pricing engine
- Feedback loop: actual rentals vs predicted demand → model retraining

### 3. Multi-Tenancy
- Add `tenant_id` to all tables
- Row-level security (RLS) in PostgreSQL
- Per-tenant config overrides
- Separate JWT signing keys per tenant

---

## Scalability Checklist

| Area | Current (V1) | Phase 1 | Phase 2 | Phase 3 |
|------|-------------|---------|---------|---------|
| API instances | 1 Docker | 3+ replicas | Auto-scaled | Multi-region |
| DB reads | Direct | Read replicas | Redis cache | Local replicas |
| Config reload | 20s poll (1 inst) | NOTIFY pub/sub | Redis pub/sub | Per-tenant |
| Audit writes | Direct INSERT | Partitioned | Async batch | Kafka event sourcing |
| Monitoring | Docker logs | Prometheus | Grafana | Distributed tracing |
| Pricing engine | Rule-based | Rule-based | Caching | ML hybrid |
| Auth | JWT 24h | JWT + refresh | OAuth2 | SSO/SAML |
| Nginx | ✓ (single) | LB upstream | — | — |
| DB indexes | ✓ (basic) | ✓ (optimized) | — | — |
| Connection pool | ✓ (pgxpool) | ✓ (tuned) | — | — |

> **Bold = already implemented in this codebase.** All Phase 1–3 items are recommendations for production deployment, not yet built.

---

## What I'd Do Differently With More Time

1. **Start with PostgreSQL LISTEN/NOTIFY**: File watcher was quick to implement but doesn't scale beyond 1 instance. Would replace with DB-native pub/sub from day 1.
2. **Audit outbox pattern**: Current fire-and-forget loses audit entries on crash. Outbox table + background worker guarantees delivery.
3. **API versioning headers**: `Accept: application/vnd.electrum.v1+json` instead of URL-based versioning for cleaner evolution.
4. **gRPC for internal services**: If IoT data feed becomes real-time, gRPC streaming would be more efficient than polling REST.
