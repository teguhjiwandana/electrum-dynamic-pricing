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
	ID    string
	Zone  string
	SoC   float64 // 0–100
	Model string
}

// Zone represents a geographic pricing zone with fleet utilization.
type Zone struct {
	Name        string
	Utilization float64 // 0–100
}

// PricingConfig holds the active pricing rules. This is the aggregate root
// for the configuration bounded context, eagerly parsed (no json.RawMessage).
type PricingConfig struct {
	BasePricePerHour float64
	Currency         string
	SurgeCap         float64
	DemandRules      DemandMultipliers
	ZoneSurge        ZoneSurgeConfig
	BatteryDiscounts BatteryDiscountTiers
	Version          int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// AuditEntry records a pricing calculation event. Tamper-evident via HMAC.
type AuditEntry struct {
	ID            string
	Timestamp     time.Time
	VehicleID     string
	Zone          string
	DurationHours int
	InputData     map[string]interface{}
	Factors       PricingFactors
	FinalPrice    float64
	ConfigVersion int
	Signature     string
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
	VehicleID     string
	Zone          string
	DurationHours int
}

// PricingOutput carries the calculated result with breakdown.
type PricingOutput struct {
	VehicleID     string
	Zone          string
	DurationHours int
	TotalPrice    float64
	Currency      string
	Factors       PricingFactors
	CalculatedAt  time.Time
}
