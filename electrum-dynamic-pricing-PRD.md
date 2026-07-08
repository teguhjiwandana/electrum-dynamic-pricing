# PRD: Dynamic Pricing Engine for EV Rentals (Electrum)

| Metadata | |
|---|---|
| **Product Name** | Dynamic Pricing Engine (DPE) |
| **Domain** | 2-Wheel EV Rental App + IoT Management |
| **Author** | Teguh Jiwandana |
| **Language** | Go |
| **Framework** | Gin |
| **Database** | PostgreSQL |
| **Auth** | JWT |
| **Deployment** | Local + Docker (Dockerfile + docker-compose) |
| **Status** | **Final** (after Q&A validation) |
| **Target Level** | L3–L4 (Senior / Staff Engineer) |
| **Timeline** | 72 jam |

---

## 1. Latar Belakang & Business Context

**Electrum** mengoperasikan armada kendaraan listrik (EV) roda 2 yang disewakan ke pelanggan via aplikasi mobile. Saat ini sistem **pricing bersifat statis**: Rp 50.000/hari tanpa mempertimbangkan:

- Tingginya permintaan pada jam sibuk vs jam sepi
- Kondisi baterai kendaraan (State of Charge / SoC)
- Utilisasi armada di setiap zona

**Akibatnya:**
- **High-demand periods undersold** — harga terlalu murah saat permintaan tinggi
- **Low-battery vehicles idle** — kendaraan dengan baterai rendah tidak laku padahal bisa didiskon
- **Revenue tidak optimal** — tidak ada fleksibilitas harga berdasarkan kondisi pasar

**Tujuan:** Membangun dynamic pricing engine yang mempertimbangkan demand patterns, battery state (SoC), dan fleet utilization di setiap zona untuk memaksimalkan revenue tanpa mengorbankan transparansi harga bagi pelanggan.

---

## 2. Goals & Objectives

| Goal | Objective | Metrik Keberhasilan |
|---|---|---|
| **Meningkatkan Revenue** | Optimasi harga berdasarkan demand & supply | Revenue per vehicle per hari meningkat ≥15% |
| **Utilisasi Armada** | Mendorong penyewaan kendaraan dengan baterai rendah via diskon | Utilisasi kendaraan dengan SoC <40% naik ≥20% |
| **Transparansi Harga** | Pelanggan bisa melihat breakdown harga | Semua kalkulasi price terekam di audit log |
| **Fairness** | Harga tidak eksploitatif, tetap wajar di jam sibuk | Surge cap maksimal 2× base price |
| **Performance** | Kalkulasi harga cepat | P95 <100ms |

---

## 3. Stakeholders

| Stakeholder | Kepentingan |
|---|---|
| **Renters (Pelanggan)** | Harga wajar, transparan, tidak berubah mendadak |
| **Electrum Operations** | Tool untuk mengatur strategi pricing tanpa deploy ulang |
| **Workshop Team** | Mengetahui kendaraan mana yang perlu didorong via diskon |
| **Data/Analytics Team** | Akses ke historical pricing data untuk analisis |
| **IoT Platform** | Menyediakan data SoC kendaraan secara real-time |

---

## 4. Functional Requirements

### 4.1 Must Have (P0) — Semua同等 Penting

| ID | Fitur | Deskripsi |
|---|---|---|
| **FR-01** | **Pricing API** | `GET /api/v1/pricing?vehicle_id={id}&zone={zone}&duration_hours={n}` → calculated price |
| **FR-02** | **Pricing Configuration** | Faktor pricing disimpan di config yang bisa di-update (base_price, demand_multiplier, battery_discount_tiers, zone_surge_config) |
| **FR-03** | **Admin Config API** | Admin bisa update pricing configuration via `PUT /api/v1/admin/config` tanpa redeployment; berlaku <30 detik |
| **FR-04** | **Pricing Breakdown API** | `GET /api/v1/pricing/breakdown` — detail kalkulasi lengkap dengan setiap faktor |
| **FR-05** | **Audit Log** | Setiap kalkulasi price dicatat (input, factors, output). Append-only + tamper-evident (HMAC signing) |
| **FR-06** | **Unit & Integration Tests** | Table-driven tests di Go untuk pricing logic + integration test untuk API endpoints |

### 4.2 Nice to Have (P1–P2)

| ID | Fitur | Prioritas |
|---|---|---|
| **FR-07** | IoT data feed integration (mock internal) untuk zone utilization real-time | P1 |
| **FR-08** | Price quote with TTL (valid N menit, diverifikasi saat booking) | P1 |
| **FR-09** | A/B testing support (dua strategi untuk segmen berbeda) | P2 |
| **FR-10** | Historical pricing analytics endpoint | P2 |

---

## 5. Pricing Logic & Formula

### Formula Dasar

```
final_price = base_price_per_hour
  × demand_multiplier(time_of_day, day_of_week)
  × zone_surge_factor(zone, current_utilization)
  × battery_discount_factor(vehicle_soc)
  × duration_hours
```

### Contoh Kalkulasi

```
Base rate:     Rp 6.250/jam
Demand mult:   1.3 (weekday 5–7 PM)
Zone surge:    1.5 (utilization 85%)
Battery disc:  0.85 (SoC 35%)
Duration:      3 jam

Final = 6.250 × 1.3 × 1.5 × 0.85 × 3 = Rp 31.078
```

### Faktor Pricing

| Faktor | Variabel | Contoh Nilai |
|---|---|---|
| Base Price | Harga per jam | Rp 6.250 |
| Demand Multiplier | time_of_day, day_of_week | 1.3 (weekday 5–7PM); 0.8 (12–4AM) |
| Zone Surge Factor | zone, utilization | 1.5 (util >80%); 1.0 (util <50%) |
| Battery Discount | vehicle SoC | 0.85 (SoC <40%); 1.0 (SoC ≥80%) |
| Duration | hours | Linear |

---

## 6. API Design

| Method | Endpoint | Auth | Deskripsi |
|---|---|---|---|
| `GET` | `/api/v1/pricing` | JWT | Hitung harga (query: vehicle_id, zone, duration_hours) |
| `GET` | `/api/v1/pricing/breakdown` | JWT | Breakdown kalkulasi lengkap |
| `GET` | `/api/v1/admin/config` | JWT + Admin | Lihat config pricing saat ini |
| `PUT` | `/api/v1/admin/config` | JWT + Admin | Update config pricing (hot reload) |
| `GET` | `/api/v1/admin/config/history` | JWT + Admin | Riwayat perubahan config |
| `GET` | `/api/v1/pricing/audit` | JWT + Admin | Audit log pricing (pagination) |
| `POST` | `/api/v1/auth/login` | Public | Login dapat JWT token |

### Response Pricing API

```json
{
  "vehicle_id": "EV-12345",
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

### Response Breakdown API

```json
{
  "vehicle_id": "EV-12345",
  "zone": "south-jakarta",
  "duration_hours": 3,
  "formula": "base_rate × demand_multiplier × zone_surge × battery_discount × duration",
  "calculation_steps": [
    {"step": "base_rate_per_hour", "value": 6250, "description": "Harga dasar per jam"},
    {"step": "demand_multiplier", "value": 1.3, "description": "Weekday 5-7 PM"},
    {"step": "zone_surge_factor", "value": 1.5, "description": "Utilisasi 85% > threshold 80%"},
    {"step": "battery_discount_factor", "value": 0.85, "description": "SoC 35% < threshold 40%"},
    {"step": "duration_hours", "value": 3, "description": "Durasi sewa"}
  ],
  "total_price": 31078,
  "currency": "IDR"
}
```

---

## 7. Non-Functional Requirements

| ID | Requirement | Spesifikasi |
|---|---|---|
| NFR-01 | Language | **Go** dengan **Gin** framework |
| NFR-02 | Database | **PostgreSQL** untuk config storage, audit log, dan data kendaraan |
| NFR-03 | Performance | P95 <100ms per kalkulasi harga |
| NFR-04 | Hot Reload Config | Perubahan config berlaku <30 detik, menggunakan file watcher + mutex |
| NFR-05 | Audit Integrity | Append-only, tiap entry di-HMAC sign dengan secret key |
| NFR-06 | Authentication | JWT untuk semua API (kecuali login). Admin endpoints require admin role |
| NFR-07 | Testing | Table-driven tests (unit) + integration tests (API) + coverage report |
| NFR-08 | Deployment | Docker + docker-compose untuk local run |

---

## 8. Data Model

### PricingConfig (disimpan di PostgreSQL)

```go
type PricingConfig struct {
    ID               int             `json:"id" db:"id"`
    BasePricePerHour float64         `json:"base_price_per_hour" db:"base_price_per_hour"`
    Currency         string          `json:"currency" db:"currency"`
    DemandRules      json.RawMessage `json:"demand_multipliers" db:"demand_multipliers"`
    ZoneSurge        json.RawMessage `json:"zone_surge_config" db:"zone_surge_config"`
    BatteryDiscounts json.RawMessage `json:"battery_discount_tiers" db:"battery_discount_tiers"`
    Version          int             `json:"version" db:"version"`
    CreatedAt        time.Time       `json:"created_at" db:"created_at"`
    UpdatedAt        time.Time       `json:"updated_at" db:"updated_at"`
}
```

### AuditLogEntry

```go
type AuditLogEntry struct {
    ID            string         `json:"id" db:"id"`
    Timestamp     time.Time      `json:"timestamp" db:"timestamp"`
    VehicleID     string         `json:"vehicle_id" db:"vehicle_id"`
    Zone          string         `json:"zone" db:"zone"`
    DurationHours int            `json:"duration_hours" db:"duration_hours"`
    InputData     json.RawMessage `json:"input_data" db:"input_data"`
    Factors       json.RawMessage `json:"factors_applied" db:"factors_applied"`
    FinalPrice    float64        `json:"final_price" db:"final_price"`
    ConfigVersion int            `json:"config_version" db:"config_version"`
    Signature     string         `json:"signature" db:"signature"`
}
```

### User (untuk JWT auth)

```go
type User struct {
    ID       int    `json:"id" db:"id"`
    Username string `json:"username" db:"username"`
    Password string `json:"-" db:"password_hash"`
    Role     string `json:"role" db:"role"` // "admin" atau "viewer"
}
```

---

## 9. Arsitektur (High-Level)

```
┌──────────────┐       ┌──────────────────────────────────────┐
│   Mobile App │       │         Go API Server (Gin)          │
│   / Client   │──JWT──▶                                      │
└──────────────┘       │  ┌────────────┐  ┌───────────────┐  │
                       │  │  Auth MW   │  │ Pricing Engine│  │
                       │  └────────────┘  └───────┬───────┘  │
                       │                          │          │
                       │  ┌──────────────────┐    │          │
                       │  │  Config Manager  │    │          │
                       │  │  (hot-reload)    │    │          │
                       │  └──────────────────┘    │          │
                       │                          │          │
                       │  ┌──────────────────┐    │          │
                       │  │  Audit Logger    │    │          │
                       │  │  (append-only)   │    │          │
                       │  └──────────────────┘    │          │
                       └──────────────────────────┼──────────┘
                                                  │
                    ┌─────────────────────────────┼──────────┐
                    │           PostgreSQL        │          │
                    │  ┌──────────┐ ┌──────────┐  │          │
                    │  │  Config  │ │Audit Log │  │          │
                    │  │  Table   │ │  Table   │  │          │
                    │  └──────────┘ └──────────┘  │          │
                    │  ┌──────────┐ ┌──────────┐  │          │
                    │  │  Users   │ │ Vehicles │  │          │
                    │  │  Table   │ │  Table   │  │          │
                    │  └──────────┘ └──────────┘  │          │
                    └─────────────────────────────┘          │
                                                  │
                    ┌─────────────────────────────┘
                    │
              ┌─────▼──────────┐
              │ IoT Data Mock  │
              │ (internal sim) │
              └────────────────┘
```

---

## 10. Struktur Repository

```
electrum-dynamic-pricing/
├── README.md
├── SCALABILITY.md       # (opsional)
├── SECURITY.md          # (opsional)
├── Dockerfile
├── docker-compose.yml
├── .env.example
├── go.mod
├── go.sum
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   │   ├── config.go        # Struktur & loader config
│   │   └── watcher.go       # Hot-reload watcher
│   ├── pricing/
│   │   ├── engine.go        # Core pricing logic
│   │   ├── engine_test.go   # Table-driven tests
│   │   └── factors.go       # Demand, zone, battery factors
│   ├── api/
│   │   ├── handler.go       # HTTP handlers
│   │   ├── middleware.go    # JWT auth middleware
│   │   ├── response.go      # Response helpers
│   │   └── handler_test.go  # Integration tests
│   ├── audit/
│   │   ├── logger.go        # Audit log writer
│   │   └── signer.go        # HMAC signing
│   ├── auth/
│   │   ├── jwt.go           # JWT generation & verification
│   │   └── auth_test.go
│   └── db/
│       ├── postgres.go      # DB connection
│       └── migrations/      # SQL migrations
├── docs/
│   ├── api.md
│   └── openapi.yaml
└── scripts/
    └── seed.go              # Seed data untuk IoT mock
```

---

## 11. Out of Scope (V1)

| Item | Alasan |
|---|---|
| Integrasi Payment Gateway | Fokus backend pricing engine |
| ML-based Predictive Pricing | Cukup rule-based untuk V1 |
| A/B Testing | Ditunda ke V2 |
| Multi-currency | Hanya IDR |
| Mobile App UI | Backend-only |

---

## 12. Deliverables

| Item | Keterangan |
|---|---|
| Source code | GitHub repository (Go + Gin) |
| Database migrations | SQL file untuk setup tabel PostgreSQL |
| Docker | Dockerfile + docker-compose.yml |
| README.md | Setup, architecture, decisions, AI tool usage |
| SCALABILITY.md | (Opsional) Scaling plan |
| SECURITY.md | (Opsional) JWT, input validation, secret management |
| API docs | OpenAPI spec + contoh curl |
| Test suite | Table-driven unit tests + integration tests |

---

## 13. Evaluation Criteria Mapping

| Criterion | Weight | Strategi Pemenuhan |
|---|---|---|
| Pricing Model Design | 25% | Factors composable via config JSON; mudah tambah faktor baru |
| Configuration Management | 20% | Hot reload via file watcher; validasi config; versioning + history |
| Audit & Transparency | 20% | Setiap price di-audit; breakdown API jelas; HMAC signing |
| API Design | 20% | RESTful dengan Gin; JWT auth; proper HTTP status codes; pagination |
| Testing | 15% | Table-driven tests (Go style) untuk unit + integration tests |
