package application

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/electrum/dynamic-pricing-engine/internal/domain/pricing"
)

// PricingUseCase orchestrates pricing calculations using domain services
// and infrastructure adapters injected via constructor.
type PricingUseCase struct {
	vehicleRepo pricing.VehicleLookup
	zoneRepo    pricing.ZoneLookup
	configStore pricing.ConfigStore
	auditRepo   pricing.AuditRecorder
	pricingSvc  *pricing.Service
}

// NewPricingUseCase creates the usecase with its dependencies.
func NewPricingUseCase(
	vehicleRepo pricing.VehicleLookup,
	zoneRepo pricing.ZoneLookup,
	configStore pricing.ConfigStore,
	auditRepo pricing.AuditRecorder,
) *PricingUseCase {
	return &PricingUseCase{
		vehicleRepo: vehicleRepo,
		zoneRepo:    zoneRepo,
		configStore: configStore,
		auditRepo:   auditRepo,
		pricingSvc:  pricing.NewService(),
	}
}

// CalculatePriceResponse mirrors the API response for pricing.
type CalculatePriceResponse struct {
	VehicleID     string                  `json:"vehicle_id"`
	Zone          string                  `json:"zone"`
	DurationHours int                     `json:"duration_hours"`
	TotalPrice    float64                 `json:"total_price"`
	Currency      string                  `json:"currency"`
	Breakdown     BreakdownResponse       `json:"breakdown"`
	CalculatedAt  string                  `json:"calculated_at"`
}

// BreakdownResponse contains the factors applied.
type BreakdownResponse struct {
	BaseRatePerHour       float64 `json:"base_rate_per_hour"`
	DemandMultiplier      float64 `json:"demand_multiplier"`
	ZoneSurgeFactor       float64 `json:"zone_surge_factor"`
	BatteryDiscountFactor float64 `json:"battery_discount_factor"`
}

// StepBreakdownResponse is the detailed step-by-step response.
type StepBreakdownResponse struct {
	VehicleID        string               `json:"vehicle_id"`
	Zone             string               `json:"zone"`
	DurationHours    int                  `json:"duration_hours"`
	Formula          string               `json:"formula"`
	CalculationSteps []CalculationStep    `json:"calculation_steps"`
	TotalPrice       float64              `json:"total_price"`
	Currency         string               `json:"currency"`
}

// CalculationStep is one step in the breakdown.
type CalculationStep struct {
	Step        string  `json:"step"`
	Value       float64 `json:"value"`
	Description string  `json:"description,omitempty"`
}

// PaginatedResponse wraps list responses.
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// CalculatePrice is the main usecase: fetch data → calculate → record audit.
func (uc *PricingUseCase) CalculatePrice(ctx context.Context, input pricing.PricingInput) (*CalculatePriceResponse, error) {
	// Fetch vehicle
	vehicle, err := uc.vehicleRepo.GetVehicle(ctx, input.VehicleID)
	if err != nil {
		return nil, fmt.Errorf("get vehicle: %w", err)
	}
	if vehicle == nil {
		return nil, fmt.Errorf("vehicle %s not found", input.VehicleID)
	}

	// Fetch zone
	zone, err := uc.zoneRepo.GetZone(ctx, input.Zone)
	if err != nil {
		return nil, fmt.Errorf("get zone: %w", err)
	}
	if zone == nil {
		return nil, fmt.Errorf("zone %s not found", input.Zone)
	}

	// Get current config
	cfg, err := uc.configStore.GetActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("get config: %w", err)
	}

	// Calculate
	now := time.Now()
	output, err := uc.pricingSvc.Calculate(ctx, input, *cfg, *vehicle, *zone, now)
	if err != nil {
		return nil, fmt.Errorf("calculate: %w", err)
	}

	// Audit
	auditID, _ := newUUID()
	auditEntry := &pricing.AuditEntry{
		ID:            auditID,
		Timestamp:     now,
		VehicleID:     input.VehicleID,
		Zone:          input.Zone,
		DurationHours: input.DurationHours,
		InputData: map[string]interface{}{
			"vehicle_id":     input.VehicleID,
			"zone":           input.Zone,
			"duration_hours": input.DurationHours,
			"soc":            vehicle.SoC,
			"utilization":    zone.Utilization,
		},
		Factors:       output.Factors,
		FinalPrice:    output.TotalPrice,
		ConfigVersion: cfg.Version,
	}
	auditEntry.Signature = signAuditEntry(auditEntry)
	_ = uc.auditRepo.Record(ctx, auditEntry) // fire-and-forget, best-effort

	return &CalculatePriceResponse{
		VehicleID:     output.VehicleID,
		Zone:          output.Zone,
		DurationHours: output.DurationHours,
		TotalPrice:    output.TotalPrice,
		Currency:      output.Currency,
		Breakdown: BreakdownResponse{
			BaseRatePerHour:       output.Factors.BaseRatePerHour,
			DemandMultiplier:      output.Factors.DemandMultiplier,
			ZoneSurgeFactor:       output.Factors.ZoneSurgeFactor,
			BatteryDiscountFactor: output.Factors.BatteryDiscountFactor,
		},
		CalculatedAt: output.CalculatedAt.Format(time.RFC3339),
	}, nil
}

// GetBreakdown returns a detailed step-by-step breakdown of the pricing.
func (uc *PricingUseCase) GetBreakdown(ctx context.Context, input pricing.PricingInput) (*StepBreakdownResponse, error) {
	result, err := uc.CalculatePrice(ctx, input)
	if err != nil {
		return nil, err
	}

	steps := []CalculationStep{
		{Step: "base_rate_per_hour", Value: result.Breakdown.BaseRatePerHour, Description: "Harga dasar per jam"},
		{Step: "demand_multiplier", Value: result.Breakdown.DemandMultiplier, Description: "Faktor permintaan berdasarkan waktu"},
		{Step: "zone_surge_factor", Value: result.Breakdown.ZoneSurgeFactor, Description: "Faktor lonjakan zona berdasarkan utilisasi"},
		{Step: "battery_discount_factor", Value: result.Breakdown.BatteryDiscountFactor, Description: "Faktor diskon berdasarkan SoC baterai"},
		{Step: "duration_hours", Value: float64(result.DurationHours), Description: "Durasi sewa (jam)"},
	}

	return &StepBreakdownResponse{
		VehicleID:        result.VehicleID,
		Zone:             result.Zone,
		DurationHours:    result.DurationHours,
		Formula:          "base_rate × demand_multiplier × zone_surge × battery_discount × duration",
		CalculationSteps: steps,
		TotalPrice:       result.TotalPrice,
		Currency:         result.Currency,
	}, nil
}

// GetConfig returns the current pricing configuration.
func (uc *PricingUseCase) GetConfig(ctx context.Context) (*pricing.PricingConfig, error) {
	cfg, err := uc.configStore.GetActive(ctx)
	if err != nil {
		return nil, err
	}

	// Strip version/timestamps for API response if needed
	return cfg, nil
}

// UpdateConfig validates and persists new pricing configuration.
func (uc *PricingUseCase) UpdateConfig(ctx context.Context, basePrice float64, currency string, surgeCap float64, demandRules json.RawMessage, zoneSurge json.RawMessage, batteryDiscounts json.RawMessage, changedBy string) (*pricing.PricingConfig, error) {
	if basePrice <= 0 {
		return nil, fmt.Errorf("base_price_per_hour must be > 0")
	}
	if surgeCap < 1.0 {
		return nil, fmt.Errorf("surge_cap_multiplier must be >= 1.0")
	}

	var dm pricing.DemandMultipliers
	if err := json.Unmarshal(demandRules, &dm); err != nil {
		return nil, fmt.Errorf("invalid demand_multipliers: %w", err)
	}

	var zs pricing.ZoneSurgeConfig
	if err := json.Unmarshal(zoneSurge, &zs); err != nil {
		return nil, fmt.Errorf("invalid zone_surge_config: %w", err)
	}

	var bd pricing.BatteryDiscountTiers
	if err := json.Unmarshal(batteryDiscounts, &bd); err != nil {
		return nil, fmt.Errorf("invalid battery_discount_tiers: %w", err)
	}

	cfg, err := uc.configStore.GetActive(ctx)
	if err != nil {
		return nil, err
	}

	newCfg := &pricing.PricingConfig{
		BasePricePerHour: basePrice,
		Currency:         currency,
		SurgeCap:         surgeCap,
		DemandRules:      dm,
		ZoneSurge:        zs,
		BatteryDiscounts: bd,
		Version:          cfg.Version + 1,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := uc.configStore.Save(ctx, newCfg, changedBy); err != nil {
		return nil, fmt.Errorf("save config: %w", err)
	}

	return newCfg, nil
}

// GetConfigHistory returns paginated config change history.
func (uc *PricingUseCase) GetConfigHistory(ctx context.Context, page, pageSize int) (*PaginatedResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	history, total, err := uc.configStore.GetHistory(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}

	if history == nil {
		history = []pricing.PricingConfig{}
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	if totalPages < 1 {
		totalPages = 1
	}

	return &PaginatedResponse{
		Data:       history,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetAuditLogs returns paginated audit log entries.
func (uc *PricingUseCase) GetAuditLogs(ctx context.Context, page, pageSize int, vehicleID, zone string) (*PaginatedResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	entries, total, err := uc.auditRepo.List(ctx, page, pageSize, vehicleID, zone)
	if err != nil {
		return nil, err
	}

	if entries == nil {
		entries = []pricing.AuditEntry{}
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	if totalPages < 1 {
		totalPages = 1
	}

	return &PaginatedResponse{
		Data:       entries,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// ListZones returns all zone utilization data.
func (uc *PricingUseCase) ListZones(ctx context.Context) ([]pricing.Zone, error) {
	return uc.zoneRepo.ListZones(ctx)
}

// ListVehicles returns all vehicles.
func (uc *PricingUseCase) ListVehicles(ctx context.Context) ([]pricing.Vehicle, error) {
	return uc.vehicleRepo.ListVehicles(ctx)
}

// ---------------------------------------------------------------------------
// Audit helpers
// ---------------------------------------------------------------------------

func newUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

func signAuditEntry(entry *pricing.AuditEntry) string {
	key := os.Getenv("AUDIT_HMAC_KEY")
	if key == "" {
		key = "electrum-audit-hmac-key"
	}
	payload := fmt.Sprintf("%s|%s|%d|%.2f|%d|%s",
		entry.VehicleID, entry.Zone, entry.DurationHours,
		entry.FinalPrice, entry.ConfigVersion,
		entry.Timestamp.Format(time.RFC3339),
	)
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
