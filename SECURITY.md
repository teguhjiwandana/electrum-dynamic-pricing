# Security — Electrum Dynamic Pricing Engine

> Optional document for take-home test. Describes JWT implementation, input validation, secret management, and security considerations.

---

## Authentication: JWT

### Implementation
- **Algorithm**: HMAC-SHA256 (HS256)
- **Token duration**: 24 hours
- **Claims**: `username`, `role`, `iat`, `exp`
- **Role-based access**: `admin` (full access) and `viewer` (read-only pricing)

```go
// Token generation
func GenerateToken(username, role string) (string, int64, error) {
    claims := jwt.MapClaims{
        "username": username,
        "role":     role,
        "iat":      now.Unix(),
        "exp":      now.Add(24 * time.Hour).Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(secret)
}
```

### Security Properties
| Property | Status | Detail |
|----------|--------|--------|
| Algorithm validation | ✓ | Rejects non-HMAC tokens |
| Expiration check | ✓ | Hard 24h expiry, no refresh |
| Signature verification | ✓ | Server-side secret key |
| Token storage | Client-side | localStorage (SPA) |
| Role enforcement | ✓ | Middleware checks `role` claim |

### Current Limitations
- **No token refresh**: After 24h, user must re-login. Would add refresh token with 7-day expiry for production.
- **No token revocation**: No blacklist. Acceptable for 24h tokens; would add Redis blacklist for production.
- **localStorage storage**: Vulnerable to XSS. Would use httpOnly cookie with CSRF token for production.

### Recommended Improvements
```go
// Refresh token flow
POST /auth/refresh  →  { "refresh_token": "..." }  →  { "token": "...", "expires_at": ... }

// Token blacklist (Redis)
func IsBlacklisted(tokenID string) bool {
    return redis.Exists(ctx, "blacklist:"+tokenID)
}
```

---

## Input Validation

### API-Level Validation

| Endpoint | Validation | Method |
|----------|-----------|--------|
| `POST /auth/login` | `username` required, 3–100 chars; `password` required, min 6 chars | Gin binding + manual check |
| `GET /pricing` | `vehicle_id` required, max 20 chars; `zone` required, max 100 chars; `duration_hours` 1–720 | Gin query binding + `min`/`max` tags |
| `PUT /admin/config` | `base_price_per_hour` > 0; `surge_cap_multiplier` ≥ 1.0; JSON body valid | Usecase-level validation |
| `GET /admin/*/audit` | `page` ≥ 1; `page_size` 1–100 | Default clamping |

```go
// Pricing request validation
type PricingRequest struct {
    VehicleID     string `form:"vehicle_id" binding:"required"`
    Zone          string `form:"zone" binding:"required"`
    DurationHours int    `form:"duration_hours" binding:"required,min=1,max=720"`
}

// Config update validation
if req.BasePricePerHour <= 0 {
    return nil, fmt.Errorf("base_price_per_hour must be > 0")
}
if req.SurgeCap < 1.0 {
    return nil, fmt.Errorf("surge_cap_multiplier must be >= 1.0")
}
```

### SQL Injection Prevention
- All queries use **parameterized placeholders** (`$1`, `$2`) — no string concatenation
- pgx driver prevents SQL injection at protocol level

```go
// Safe: parameterized
Pool.QueryRow(ctx, `SELECT id, zone, soc, model FROM vehicles WHERE id = $1`, vehicleID)

// Never: string concatenation
// Pool.QueryRow(ctx, "SELECT ... WHERE id = '" + vehicleID + "'")  ← NOT IN CODE
```

### Sanitization
- **HTML output**: NextJS/React auto-escapes by default
- **JSON output**: `encoding/json` produces valid UTF-8
- **File paths**: Config path from environment variable, validated at startup

---

## Secret Management

### Current Approach (V1)

| Secret | Source | Rotation |
|--------|--------|----------|
| `JWT_SECRET` | Environment variable | Manual restart |
| `AUDIT_HMAC_KEY` | Environment variable | Manual restart |
| `DATABASE_URL` | Environment variable | Manual restart |

```bash
# .env (not committed to git)
DATABASE_URL=postgres://electrum:electrum_secret@localhost:5432/electrum_pricing
JWT_SECRET=electrum-jwt-secret-change-in-production
AUDIT_HMAC_KEY=electrum-audit-hmac-key-change-in-production
```

### Production Recommendations

**1. Vault / Secrets Manager**
```yaml
# Kubernetes Secret
apiVersion: v1
kind: Secret
metadata:
  name: electrum-secrets
data:
  jwt-secret: <base64>
  audit-hmac-key: <base64>
  database-url: <base64>
```

**2. Secret Rotation**
- JWT secret: rotate with overlapping validity window (old + new both valid during rotation)
- DB password: rotate via PostgreSQL `ALTER ROLE`, update env var, rolling restart API instances
- Audit HMAC key: cannot rotate without breaking signature verification — keep stable, rotate only with audit log re-signing script

**3. Environment Isolation**
```
Dev  → .env.local (git-ignored)
CI   → GitHub Secrets
Prod → Vault / K8s Secrets / AWS Secrets Manager
```

---

## Network Security

| Layer | Protection |
|-------|-----------|
| **HTTPS** | Let's Encrypt SSL (TLS 1.2+) with auto-renewal |
| **API** | Internal Docker network (not exposed directly) |
| **DB port** | Mapped to `5433` (non-standard), not `5432` |
| **Nginx** | Reverse proxy — only port 80/443 exposed |
| **Firewall** | UFW enabled, only 22/80/443 open |

```bash
# UFW rules on production VPS
sudo ufw allow 22    # SSH
sudo ufw allow 80    # HTTP → HTTPS redirect
sudo ufw allow 443   # HTTPS
sudo ufw enable
```

---

## Audit Integrity

### HMAC-SHA256 Signing
```go
// Payload: vehicle_id|zone|duration_hours|final_price|config_version|timestamp
payload := fmt.Sprintf("%s|%s|%d|%.2f|%d|%s",
    entry.VehicleID, entry.Zone, entry.DurationHours,
    entry.FinalPrice, entry.ConfigVersion,
    entry.Timestamp.Format(time.RFC3339),
)
mac := hmac.New(sha256.New, []byte(hmacKey))
mac.Write([]byte(payload))
signature := hex.EncodeToString(mac.Sum(nil))
```

### Verification
```go
// To verify an audit entry hasn't been tampered with:
// 1. Re-compute signature from stored fields
// 2. Compare with stored signature
// 3. Mismatch = tampering detected
```

### Attack Surface
| Threat | Mitigation |
|--------|-----------|
| Signature forgery | HMAC key never exposed to client |
| Entry modification | Signature mismatch on verification |
| Entry deletion | Append-only table (no DELETE permissions) |
| Replay attack | Timestamp included in signature payload |
| Key compromise | Rotate key, re-sign audit log with migration script |

---

## Dependency Security

```bash
# Go vulnerability check
cd backend && go vet ./...
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

# npm audit
cd frontend && npm audit
```

### Key Dependencies
| Package | Version | Purpose |
|---------|---------|---------|
| `gin-gonic/gin` | v1.12 | HTTP framework |
| `golang-jwt/jwt` | v5 | JWT signing/validation |
| `jackc/pgx` | v5 | PostgreSQL driver |
| `golang.org/x/crypto` | latest | bcrypt |
| `next` | 14 | Frontend framework |
| `tailwindcss` | v4 | CSS framework |

---

## What I'd Do Differently With More Time

1. **httpOnly cookies + CSRF**: Replace localStorage JWT with httpOnly cookie + CSRF token to prevent XSS token theft.
2. **Rate limiting**: Add `golang.org/x/time/rate` middleware for brute-force protection on login endpoint.
3. **Request ID tracing**: Inject `X-Request-ID` header for audit trail correlation across services.
4. **CORS policy**: Currently permissive (`*`). Would restrict to known frontend origins.
5. **Content Security Policy**: Add CSP headers to prevent XSS in frontend.
6. **Database encryption**: Encrypt sensitive columns (user emails, audit signatures) at rest with pgcrypto.
