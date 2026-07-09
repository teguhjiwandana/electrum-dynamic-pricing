# Electrum Dynamic Pricing Engine — Take Home Test

Dynamic pricing engine for 2-wheel EV rentals. Computes optimal rental prices based on real-time demand patterns, battery state (SoC), and zone fleet utilization.

---

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.25+ (for local dev)
- Node.js 22+ (for frontend)

### Run Everything
```bash
# Clone
git clone https://github.com/teguhjiwandana/electrum-dynamic-pricing.git
cd electrum-dynamic-pricing

# Start PostgreSQL + API
docker compose up -d

# API runs on http://localhost:8080
curl http://localhost:8080/health

# Start Frontend (separate terminal)
cd frontend
npm install
npm run dev

# Dashboard: http://localhost:3000
# Login: admin / admin123
```

### Deployed
```
https://electrum-pricing.greatije.id
```

---

## Architecture

```
┌──────────────┐     ┌──────────────────────────────────────┐
│   Browser    │────▶│              Nginx :443              │
│   (NextJS)   │     │         (reverse proxy)              │
└──────────────┘     └──────┬───────────────┬──────────────┘
                            │               │
                    ┌───────▼──────┐ ┌──────▼───────┐
                    │  NextJS :3000│ │  Go API :8080│
                    │  (pm2)       │ │  (Docker)    │
                    └──────────────┘ └──────┬────────┘
                                            │
                                    ┌───────▼──────┐
                                    │ PostgreSQL   │
                                    │ (Docker)     │
                                    └──────────────┘
```

### Backend: Clean Architecture + DDD
```
backend/internal/
├── domain/pricing/          ← Pure enterprise rules (zero deps)
│   ├── ports.go             # 4 interfaces: VehicleLookup, ZoneLookup,
│   │                          ConfigStore, AuditRecorder
│   ├── rules.go             # DemandMultipliers, ZoneSurgeConfig,
│   │                          BatteryDiscountTiers (eager-parse)
│   ├── engine.go            # PricingService.Calculate() — pure function
│   └── engine_test.go       # 10 table-driven tests
├── application/
│   ├── usecase.go           # PricingUseCase — orchestration
│   └── usecase_test.go      # 8 tests (mocked repos)
├── infrastructure/
│   ├── postgres/            # Repository impls (pgxpool)
│   ├── config/              # JSON file store + hot-reload watcher
│   └── auth/                # JWT service + bcrypt
└── presentation/http/
    ├── handler.go           # 8 REST endpoints
    ├── middleware.go         # AuthMiddleware + AdminMiddleware
    └── response.go
```

**Dependency Rule:**
```
Presentation → Application → Domain ← Infrastructure
```

Domain has zero external dependencies. Infrastructure implements domain ports. Application orchestrates. Presentation only handles HTTP.

### Frontend: NextJS 14 + Emerald Design System
```
frontend/src/
├── app/
│   ├── (auth)/login/        # JWT login page
│   └── (dashboard)/
│       ├── pricing/         # Calculator with vehicle/zone dropdowns
│       ├── audit/           # Paginated audit log with filters
│       ├── settings/        # Config editor (sliders + JSON viewer)
│       └── zones/           # Zone utilization cards
├── components/layout/
│   ├── Sidebar.tsx          # 280px nav rail with dynamic user info
│   └── TopHeader.tsx        # Fixed top bar
└── lib/api.ts               # Typed API client (JWT, fetch wrapper)
```

### Design System: Emerald Analytics
- **Colors**: Deep Emerald `#064e3b` primary, slate-white surfaces
- **Typography**: Inter (UI) + JetBrains Mono (numbers/codes)
- **Spacing**: 4px baseline grid
- **Elevation**: Ambient shadows (Level 0–2)

### Database Schema
| Table | Purpose |
|-------|---------|
| `users` | JWT authentication (admin/viewer roles) |
| `pricing_config` | Active pricing rules (aggregate root) |
| `pricing_config_history` | Config change audit trail |
| `vehicles` | Fleet data (10 mock EVs) |
| `zone_utilization` | Zone fleet utilization (mock IoT) |
| `audit_log` | Tamper-evident pricing events (HMAC-SHA256) |

---

## API Reference

### Auth
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

### Pricing
```bash
TOKEN="<jwt-from-login>"
curl "http://localhost:8080/api/v1/pricing?vehicle_id=EV-10001&zone=south-jakarta&duration_hours=3" \
  -H "Authorization: Bearer $TOKEN"
```

Response:
```json
{
  "vehicle_id": "EV-10001",
  "zone": "south-jakarta",
  "duration_hours": 3,
  "total_price": 31078,
  "currency": "IDR",
  "breakdown": {
    "base_rate_per_hour": 6250,
    "demand_multiplier": 1.0,
    "zone_surge_factor": 1.5,
    "battery_discount_factor": 0.85
  }
}
```

### All Endpoints
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `POST` | `/auth/login` | — | Get JWT token |
| `GET` | `/pricing` | JWT | Calculate rental price |
| `GET` | `/pricing/breakdown` | JWT | Full step-by-step breakdown |
| `GET` | `/config` | JWT | Current pricing configuration |
| `GET` | `/vehicles` | — | List all vehicles |
| `GET` | `/zones` | — | List zone utilization |
| `PUT` | `/admin/config` | Admin | Update pricing rules (hot-reload) |
| `GET` | `/admin/config/history` | Admin | Config change history |
| `GET` | `/admin/pricing/audit` | Admin | Audit log (paginated) |

### Pricing Formula
```
final_price = base_price_per_hour (Rp 6.250)
  × demand_multiplier(time_of_day, day_of_week)
  × zone_surge_factor(zone, fleet_utilization)
  × battery_discount_factor(vehicle_soc)
  × duration_hours

Surge cap: max 2× base price (configurable)
```

---

## Running Tests

```bash
cd backend
go test ./... -v -count=1
```

```
ok  domain/pricing       10 tests — PRD formula, surge cap, edge cases, factors
ok  application           8 tests — usecase with mocked repos
ok  infrastructure/auth   7 tests — JWT gen/validate, bcrypt, authenticate
────────────────────────────────────────────────
                   TOTAL: 25 tests pass
```

**Domain test example** (PRD verification):
```go
// Input: base 6250, demand 1.3, zone surge 1.5, battery 0.85, 3 hours
// Expected: 6250 × 1.3 × 1.5 × 0.85 × 3 = 31078
func TestCalculate_PRDExample(t *testing.T) {
    svc := pricing.NewService()
    output, _ := svc.Calculate(ctx, input, cfg, vehicle, zone, tuesday)
    assert.Equal(t, float64(31078), output.TotalPrice)
}
```

---

## Key Decisions

### 1. Clean Architecture + DDD (not flat packages)
**Why**: PRD specified "L3–L4 (Senior/Staff Engineer) target". Clean architecture demonstrates dependency inversion, testability, and separation of concerns.

**Tradeoff**: ~20 files vs ~15 in flat layout. But every layer testable in isolation — domain tests need zero infrastructure, usecase tests use mocked repos.

**What I'd change**: Merge `application/` and `presentation/` into a single `app/` layer if the team is <3 people. The domain/infrastructure split is the high-value boundary.

### 2. Eager Config Parsing (not json.RawMessage per request)
**Why**: PRD v1 stored `demand_multipliers` as `json.RawMessage`, parsed on every pricing request. Moved parsing to config load time — `DemandMultipliers` becomes a typed struct. 20-second hot-reload interval means amortized cost is negligible.

**Tradeoff**: Type-safe but harder to add new factor types dynamically. Acceptable for V1 rule-based engine.

### 3. PostgreSQL + Docker (not SQLite/file-based)
**Why**: PRD required PostgreSQL. Docker compose for zero-config local setup. Seed migration auto-runs on first start.

**Tradeoff**: Docker required for local dev. Without Docker, need local PostgreSQL install.

### 4. Audit as Decorator Pattern (not separate bounded context)
**Why**: Audit has zero business logic — it's a fire-and-forget side effect. Made it a `AuditRecorder` interface in domain, implemented by infrastructure. Oracle review confirmed: "cross-cutting concern, not a domain."

**Tradeoff**: Auditor doesn't participate in transactions (best-effort recording). For financial-grade auditing, would add outbox pattern or event sourcing.

### 5. HMAC-SHA256 Tamper-Evident Signatures
**Why**: PRD required append-only + tamper-evident audit. Each entry signed with `HMAC(vehicle_id|zone|duration|price|config_version|timestamp)`. Signature stored alongside data — verification possible without re-querying.

**Tradeoff**: HMAC key must be kept secret. Signature verification is manual (no automated checker in V1).

### 6. Zones/Vehicles as Public Endpoints
**Why**: Zone list and vehicle list are read-only reference data with no sensitive info. Making them public avoids auth failures on dashboard load (user hasn't logged in yet).

---

## AI Tool Usage

### Tools Used
- **OpenCode** (powered by DeepSeek V4 Pro) — primary AI coding assistant with specialist sub-agents
- **Specialist sub-agents**: oracle (architecture review), designer (UI/UX), fixer (implementation), explorer (codebase search)
- **Model**: DeepSeek V4 Pro (via OpenCode orchestrator)

### What Worked Well
- **Oracle architecture review**: Caught over-engineering early — simplified from 4 bounded contexts to 1, from 30+ files to 20. Saved hours of unnecessary abstraction.
- **Designer agent**: Generated full Emerald design system (CSS, fonts, spacing) and 5 dashboard pages from screenshot references in one shot. Consistent visual language across all pages.
- **Parallel dispatches**: 5 backend packages written simultaneously by separate fixer agents. Domain tests, JWT tests, and usecase tests also dispatched in parallel.
- **Domain-first approach**: Wrote pure domain engine + tests first (10 tests, no infrastructure), then built outward. Domain tests caught formula bugs before any HTTP code was written.

### What I Had to Fix/Correct
- **Go version mismatch**: Local Go 1.26.2 vs Docker golang:1.23 → updated Dockerfile to 1.25, downgraded go.mod directive.
- **JSON serialization**: Domain structs initially had no `json` tags → API returned PascalCase fields. Added snake_case tags to all domain types.
- **Audit log silent failure**: Usecase created audit entries without UUID and HMAC signature → DB insert failed silently. Added UUID generation and HMAC signing in usecase layer.
- **Frontend API types mismatch**: Designer built pages against initial API response shapes. Had to update all 5 pages when backend response format changed (nested `breakdown` object, paginated wrapper).
- **Nginx Host header issue**: `proxy_set_header Host \$host` (escaped in shell heredoc) → nginx received literal `$host`. Fixed with proper heredoc quoting.
- **DNS migration**: Domain greatije.id had broken SOA record at SumoPod. NS migrated to Cloudflare, added A record, installed Let's Encrypt SSL.
