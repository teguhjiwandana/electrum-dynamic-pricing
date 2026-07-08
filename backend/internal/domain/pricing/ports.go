// Package pricing defines the domain model for the Dynamic Pricing Engine.
// It contains enterprise business rules, pure from infrastructure concerns.
package pricing

import (
	"context"
	"time"
)

// ---------------------------------------------------------------------------
// Ports (interfaces the domain needs, implemented by infrastructure)
// ---------------------------------------------------------------------------

// VehicleLookup retrieves a vehicle by its ID.
type VehicleLookup interface {
	GetVehicle(ctx context.Context, vehicleID string) (*Vehicle, error)
}

// ZoneLookup retrieves zone utilization data.
type ZoneLookup interface {
	GetZone(ctx context.Context, zone string) (*Zone, error)
	ListZones(ctx context.Context) ([]Zone, error)
}

// ConfigStore persists and retrieves pricing configuration.
type ConfigStore interface {
	GetActive(ctx context.Context) (*PricingConfig, error)
	Save(ctx context.Context, config *PricingConfig, changedBy string) error
	GetHistory(ctx context.Context, page, pageSize int) ([]PricingConfig, int, error)
}

// AuditRecorder records tamper-evident pricing calculation events.
type AuditRecorder interface {
	Record(ctx context.Context, entry *AuditEntry) error
	List(ctx context.Context, page, pageSize int, vehicleID, zone string) ([]AuditEntry, int, error)
}

// ---------------------------------------------------------------------------
// Domain entities (identity + behavior)
// ---------------------------------------------------------------------------

// Vehicle represents a rented EV unit.
type Vehicle struct {
	ID    string  `json:"id"`
	Zone  string  `json:"zone"`
	SoC   float64 `json:"soc"`
	Model string  `json:"model"`
}

// Zone represents a geographic pricing zone with fleet utilization.
type Zone struct {
	Name        string  `json:"name"`
	Utilization float64 `json:"utilization"`
}

// PricingConfig holds the active pricing rules. This is the aggregate root
// for the configuration bounded context, eagerly parsed (no json.RawMessage).
type PricingConfig struct {
	BasePricePerHour float64             `json:"base_price_per_hour"`
	Currency         string              `json:"currency"`
	SurgeCap         float64             `json:"surge_cap_multiplier"`
	DemandRules      DemandMultipliers   `json:"demand_multipliers"`
	ZoneSurge        ZoneSurgeConfig     `json:"zone_surge_config"`
	BatteryDiscounts BatteryDiscountTiers `json:"battery_discount_tiers"`
	Version          int                 `json:"version"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
}

// AuditEntry records a pricing calculation event. Tamper-evident via HMAC.
type AuditEntry struct {
	ID            string                 `json:"id"`
	Timestamp     time.Time              `json:"timestamp"`
	VehicleID     string                 `json:"vehicle_id"`
	Zone          string                 `json:"zone"`
	DurationHours int                    `json:"duration_hours"`
	InputData     map[string]interface{} `json:"input_data"`
	Factors       PricingFactors         `json:"factors_applied"`
	FinalPrice    float64                `json:"final_price"`
	ConfigVersion int                    `json:"config_version"`
	Signature     string                 `json:"signature"`
}

// PricingFactors captures the multipliers used in a calculation.
type PricingFactors struct {
	BaseRatePerHour       float64 `json:"base_rate_per_hour"`
	DemandMultiplier      float64 `json:"demand_multiplier"`
	ZoneSurgeFactor       float64 `json:"zone_surge_factor"`
	BatteryDiscountFactor float64 `json:"battery_discount_factor"`
}

// PricingInput carries the raw request data for a calculation.
type PricingInput struct {
	VehicleID     string `json:"vehicle_id"`
	Zone          string `json:"zone"`
	DurationHours int    `json:"duration_hours"`
}

// PricingOutput carries the calculated result with breakdown.
type PricingOutput struct {
	VehicleID     string          `json:"vehicle_id"`
	Zone          string          `json:"zone"`
	DurationHours int             `json:"duration_hours"`
	TotalPrice    float64         `json:"total_price"`
	Currency      string          `json:"currency"`
	Factors       PricingFactors  `json:"breakdown"`
	CalculatedAt  time.Time       `json:"calculated_at"`
}
