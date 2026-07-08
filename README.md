# Electrum Dynamic Pricing Engine

Dynamic pricing engine for 2-wheel EV rentals. Computes optimal rental prices based on demand patterns, battery state (SoC), and zone fleet utilization. Built with **Go/Gin** (Clean Architecture + DDD) backend and **NextJS 14** (Emerald Design System) frontend.

---

## Features

- **Dynamic Pricing API** — `base_rate × demand × zone_surge × battery_discount × duration` with surge cap enforcement
- **Hot-Reload Configuration** — update pricing rules via API, applies within 20s without restart
- **Tamper-Evident Audit Log** — every calculation recorded with HMAC-SHA256 signature (append-only)
- **JWT Authentication** — role-based access (admin/viewer) for all endpoints
- **Admin Dashboard** — NextJS 14 UI with pricing calculator, audit log, settings, zone management
- **Table-Driven Tests** — 10 domain tests verifying PRD formula (6250 × 1.3 × 1.5 × 0.85 × 3 = 31078)
- **Docker** — single `docker-compose up` starts PostgreSQL + API

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| **Backend** | Go 1.23+ / Gin / pgx v5 |
| **Database** | PostgreSQL 16 |
| **Frontend** | Next.js 14 (App Router) / TypeScript / Tailwind CSS |
| **Auth** | JWT (HS256) + bcrypt |
| **Audit** | HMAC-SHA256 tamper-proof signing |
| **Design** | Emerald Design System — Inter + JetBrains Mono |

---

## Architecture

```
┌─────────────────────────────────────────────────┐
│                  Presentation                   │
│         Gin handlers, JWT middleware            │
└──────────────────┬──────────────────────────────┘
                   │  calls
┌──────────────────▼──────────────────────────────┐
│                 Application                     │
│         PricingUseCase (orchestration)          │
└────┬──────────────┬──────────────┬──────────────┘
     │              │              │
┌────▼────┐  ┌──────▼──────┐  ┌───▼──────────────┐
│ Domain  │  │Infrastructure│  │  Infrastructure  │
│ Pricing │  │  Postgres    │  │  JWT / HMAC      │
│ Service │  │  Repos       │  │  Config Watcher  │
└─────────┘  └─────────────┘  └──────────────────┘
```

**Dependency Rule**: Domain ← Application ← Infrastructure ← Presentation  
Domain has zero external dependencies. Infrastructure implements domain ports.

### Clean Architecture Layers

```
backend/internal/
├── domain/pricing/          ← Enterprise rules (pure, 4 interfaces)
│   ├── ports.go             # VehicleLookup, ZoneLookup, ConfigStore, AuditRecorder
│   ├── rules.go             # DemandMultipliers, ZoneSurgeConfig, BatteryDiscountTiers
│   ├── engine.go            # PricingService.Calculate() — pure function
│   └── engine_test.go       # 10 table-driven tests
├── application/
│   └── usecase.go           # PricingUseCase — orchestrates domain + repos
├── infrastructure/
│   ├── postgres/            # Repository implementations (pgxpool)
│   ├── config/              # JSON file store + 20s hot-reload watcher
│   └── auth/                # JWT service + bcrypt
└── presentation/http/
    ├── handler.go           # 8 REST endpoints
    ├── middleware.go         # AuthMiddleware + AdminMiddleware
    └── response.go          # JSON helpers
```

---

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.23+ (for local dev)
- Node.js 22+ (for frontend)

### Run with Docker

```bash
# Start PostgreSQL + API
docker-compose up -d

# API runs on http://localhost:8080
# Health check: curl http://localhost:8080/health
```

### Run Frontend

```bash
cd frontend
npm install
npm run dev

# Dashboard: http://localhost:3000
```

### Login

```
Username: admin
Password: admin123
```

---

## API Reference

Base URL: `http://localhost:8080/api/v1`

### Auth

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `POST` | `/auth/login` | — | Get JWT token |

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

### Pricing

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `GET` | `/pricing?vehicle_id=X&zone=X&duration_hours=N` | JWT | Calculate price |
| `GET` | `/pricing/breakdown?vehicle_id=X&zone=X&duration_hours=N` | JWT | Full step-by-step breakdown |

```bash
curl "http://localhost:8080/api/v1/pricing?vehicle_id=EV-10001&zone=south-jakarta&duration_hours=3" \
  -H "Authorization: Bearer <token>"
```

**Response:**
```json
{
  "vehicle_id": "EV-10001",
  "zone": "south-jakarta",
  "duration_hours": 3,
  "total_price": 31078,
  "currency": "IDR",
  "breakdown": {
    "base_rate_per_hour": 6250,
    "demand_multiplier": 1.3,
    "zone_surge_factor": 1.5,
    "battery_discount_factor": 0.85
  },
  "calculated_at": "2026-07-08T14:30:00Z"
}
```

### Admin (JWT + admin role)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/admin/config` | Current pricing configuration |
| `PUT` | `/admin/config` | Update pricing rules (hot-reload) |
| `GET` | `/admin/config/history?page=1&page_size=10` | Config change history |
| `GET` | `/admin/pricing/audit?page=1&page_size=20` | Audit log (paginated) |

### Zones

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `GET` | `/zones` | JWT | List zone utilization |

---

## Pricing Formula

```
final_price = base_price_per_hour (Rp 6.250)
  × demand_multiplier(time_of_day, day_of_week)
  × zone_surge_factor(zone, fleet_utilization)
  × battery_discount_factor(vehicle_soc)
  × duration_hours
```

**Surge cap**: maximum 2× base price (configurable).

### Example (PRD)

| Factor | Value | Condition |
|--------|-------|-----------|
| Base rate | 6.250 | — |
| Demand | 1.3× | Weekday 5–7 PM |
| Zone surge | 1.5× | Utilization 85% |
| Battery discount | 0.85× | SoC 35% |
| Duration | 3 hours | — |
| **Final** | **Rp 31.078** | 6.250 × 1.3 × 1.5 × 0.85 × 3 |

---

## Environment Variables

Copy `backend/.env.example` → `backend/.env`:

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://electrum:...` | PostgreSQL connection string |
| `JWT_SECRET` | `electrum-jwt-secret-...` | JWT signing key |
| `AUDIT_HMAC_KEY` | `electrum-audit-hmac-...` | HMAC key for audit integrity |
| `CONFIG_PATH` | `./config/pricing_config.json` | Pricing config file path |
| `SERVER_PORT` | `8080` | HTTP port |

---

## Testing

```bash
cd backend
go test ./... -v -count=1
```

```
ok  domain/pricing    10 tests pass (PRD formula, surge cap, edge cases)
ok  application       (no DB-dependent tests)
ok  infrastructure    (requires live PostgreSQL)
ok  presentation      (requires live DB)
```

---

## Frontend Pages

| Route | Description |
|-------|-------------|
| `/login` | Login with username/password |
| `/pricing` | Main pricing calculator dashboard |
| `/audit` | Audit log with filters + pagination |
| `/settings` | Config management (view/edit pricing rules) |
| `/zones` | Zone utilization cards with status bars |

---

## Design System

**Emerald Analytics** — built for high-stakes financial environments.

- **Colors**: Deep Emerald `#064e3b` primary, slate-white surfaces
- **Typography**: Inter (UI) + JetBrains Mono (numbers, codes)
- **Spacing**: 4px baseline grid
- **Elevation**: Ambient shadows (Level 0–2), hover lift
- **Shapes**: 4px inputs/buttons, 8px cards

See `stitch_dynamic_pricing_engine_dashboard/emerald_analytics/DESIGN.md` for full spec.

---

## Database

6 tables, auto-created via `docker-compose` migration:

| Table | Purpose |
|-------|---------|
| `users` | JWT authentication |
| `pricing_config` | Active pricing rules (aggregate root) |
| `pricing_config_history` | Config change audit trail |
| `vehicles` | Fleet data (10 mock EVs) |
| `zone_utilization` | Zone fleet utilization (mock IoT) |
| `audit_log` | Tamper-evident pricing events (HMAC) |

Seed data: 1 admin user, 10 vehicles across 3 zones, 5 zones with mock utilization.

---

## Project Structure

```
electrum-dynamic-pricing/
├── docker-compose.yml
├── backend/                         # Go/Gin API
│   ├── cmd/server/main.go           # DI wiring, router setup
│   ├── internal/
│   │   ├── domain/pricing/          # Pure business rules
│   │   ├── application/             # Usecases
│   │   ├── infrastructure/          # DB, auth, config
│   │   └── presentation/http/       # HTTP delivery
│   ├── config/pricing_config.json
│   ├── Dockerfile
│   └── .env.example
├── frontend/                        # NextJS 14 Admin UI
│   └── src/
│       ├── app/
│       │   ├── (auth)/login/
│       │   └── (dashboard)/
│       │       ├── pricing/         # Calculator
│       │       ├── audit/           # Audit log table
│       │       ├── settings/        # Config editor
│       │       └── zones/           # Zone management
│       ├── components/layout/       # Sidebar + TopHeader
│       └── lib/api.ts               # API client
└── electrum-dynamic-pricing-PRD.md  # Product requirements
```

---

## Out of Scope (V1)

- Payment gateway integration
- ML-based predictive pricing
- A/B testing
- Multi-currency support
- Mobile app

---

## License

MIT
