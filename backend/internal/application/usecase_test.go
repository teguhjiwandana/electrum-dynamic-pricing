package application_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/electrum/dynamic-pricing-engine/internal/application"
	"github.com/electrum/dynamic-pricing-engine/internal/domain/pricing"
)

// ---------------------------------------------------------------------------
// Mock implementations
// ---------------------------------------------------------------------------

type mockVehicleRepo struct {
	getVehicleFn  func(ctx context.Context, id string) (*pricing.Vehicle, error)
	listVehiclesFn func(ctx context.Context) ([]pricing.Vehicle, error)
}

func (m *mockVehicleRepo) GetVehicle(ctx context.Context, id string) (*pricing.Vehicle, error) {
	return m.getVehicleFn(ctx, id)
}

func (m *mockVehicleRepo) ListVehicles(ctx context.Context) ([]pricing.Vehicle, error) {
	if m.listVehiclesFn != nil {
		return m.listVehiclesFn(ctx)
	}
	return nil, nil
}

type mockZoneRepo struct {
	getZoneFn   func(ctx context.Context, zone string) (*pricing.Zone, error)
	listZonesFn func(ctx context.Context) ([]pricing.Zone, error)
}

func (m *mockZoneRepo) GetZone(ctx context.Context, zone string) (*pricing.Zone, error) {
	return m.getZoneFn(ctx, zone)
}

func (m *mockZoneRepo) ListZones(ctx context.Context) ([]pricing.Zone, error) {
	return m.listZonesFn(ctx)
}

type mockConfigStore struct {
	getActiveFn  func(ctx context.Context) (*pricing.PricingConfig, error)
	saveFn       func(ctx context.Context, config *pricing.PricingConfig, changedBy string) error
	getHistoryFn func(ctx context.Context, page, pageSize int) ([]pricing.PricingConfig, int, error)
}

func (m *mockConfigStore) GetActive(ctx context.Context) (*pricing.PricingConfig, error) {
	return m.getActiveFn(ctx)
}

func (m *mockConfigStore) Save(ctx context.Context, config *pricing.PricingConfig, changedBy string) error {
	return m.saveFn(ctx, config, changedBy)
}

func (m *mockConfigStore) GetHistory(ctx context.Context, page, pageSize int) ([]pricing.PricingConfig, int, error) {
	return m.getHistoryFn(ctx, page, pageSize)
}

type mockAuditRepo struct {
	recordFn func(ctx context.Context, entry *pricing.AuditEntry) error
	listFn   func(ctx context.Context, page, pageSize int, vehicleID, zone string) ([]pricing.AuditEntry, int, error)
}

func (m *mockAuditRepo) Record(ctx context.Context, entry *pricing.AuditEntry) error {
	return m.recordFn(ctx, entry)
}

func (m *mockAuditRepo) List(ctx context.Context, page, pageSize int, vehicleID, zone string) ([]pricing.AuditEntry, int, error) {
	return m.listFn(ctx, page, pageSize, vehicleID, zone)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func defaultConfig() pricing.PricingConfig {
	return pricing.PricingConfig{
		BasePricePerHour: 6250,
		Currency:         "IDR",
		SurgeCap:         2.0,
		DemandRules: pricing.DemandMultipliers{
			Default: 1.0,
			Rules: []pricing.DemandRule{
				{Days: []int{1, 2, 3, 4, 5}, Hours: []int{17, 18, 19}, Multiplier: 1.3},
				{Days: []int{1, 2, 3, 4, 5}, Hours: []int{0, 1, 2, 3, 4}, Multiplier: 0.8},
				{Days: []int{6, 0}, Hours: []int{9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21}, Multiplier: 1.2},
			},
		},
		ZoneSurge: pricing.ZoneSurgeConfig{
			Thresholds: []pricing.ZoneSurgeThreshold{
				{MaxUtilization: 50, Factor: 1.0},
				{MaxUtilization: 80, Factor: 1.2},
				{MaxUtilization: 100, Factor: 1.5},
			},
		},
		BatteryDiscounts: pricing.BatteryDiscountTiers{
			Thresholds: []pricing.BatteryDiscountThreshold{
				{MaxSOC: 40, DiscountFactor: 0.85},
				{MaxSOC: 60, DiscountFactor: 0.92},
				{MaxSOC: 80, DiscountFactor: 0.97},
				{MaxSOC: 100, DiscountFactor: 1.0},
			},
		},
		Version: 1,
	}
}

func newUseCase(vehicleRepo pricing.VehicleLookup, zoneRepo pricing.ZoneLookup, configStore pricing.ConfigStore, auditRepo pricing.AuditRecorder) *application.PricingUseCase {
	return application.NewPricingUseCase(vehicleRepo, zoneRepo, configStore, auditRepo)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestCalculatePrice_Success(t *testing.T) {
	cfg := defaultConfig()

	vehicleRepo := &mockVehicleRepo{
		getVehicleFn: func(ctx context.Context, id string) (*pricing.Vehicle, error) {
			return &pricing.Vehicle{ID: "EV-10001", Zone: "south-jakarta", SoC: 35, Model: "E1"}, nil
		},
	}
	zoneRepo := &mockZoneRepo{
		getZoneFn: func(ctx context.Context, zone string) (*pricing.Zone, error) {
			return &pricing.Zone{Name: "south-jakarta", Code: "SJ", Utilization: 85}, nil
		},
	}
	configStore := &mockConfigStore{
		getActiveFn: func(ctx context.Context) (*pricing.PricingConfig, error) {
			return &cfg, nil
		},
	}
	auditRepo := &mockAuditRepo{
		recordFn: func(ctx context.Context, entry *pricing.AuditEntry) error {
			return nil
		},
	}

	uc := newUseCase(vehicleRepo, zoneRepo, configStore, auditRepo)

	resp, err := uc.CalculatePrice(context.Background(), pricing.PricingInput{
		VehicleID:     "EV-10001",
		Zone:          "south-jakarta",
		DurationHours: 3,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.VehicleID != "EV-10001" {
		t.Errorf("VehicleID = %s, want EV-10001", resp.VehicleID)
	}
	if resp.Zone != "south-jakarta" {
		t.Errorf("Zone = %s, want south-jakarta", resp.Zone)
	}
	if resp.DurationHours != 3 {
		t.Errorf("DurationHours = %d, want 3", resp.DurationHours)
	}
	if resp.Currency != "IDR" {
		t.Errorf("Currency = %s, want IDR", resp.Currency)
	}
	if resp.TotalPrice <= 0 {
		t.Errorf("TotalPrice = %.0f, want > 0", resp.TotalPrice)
	}
	if resp.Breakdown.BaseRatePerHour != 6250 {
		t.Errorf("BaseRatePerHour = %.0f, want 6250", resp.Breakdown.BaseRatePerHour)
	}
	if resp.Breakdown.DemandMultiplier <= 0 {
		t.Errorf("DemandMultiplier = %.2f, want > 0", resp.Breakdown.DemandMultiplier)
	}
	if resp.Breakdown.ZoneSurgeFactor <= 0 {
		t.Errorf("ZoneSurgeFactor = %.2f, want > 0", resp.Breakdown.ZoneSurgeFactor)
	}
	if resp.Breakdown.BatteryDiscountFactor <= 0 {
		t.Errorf("BatteryDiscountFactor = %.2f, want > 0", resp.Breakdown.BatteryDiscountFactor)
	}
	if resp.CalculatedAt == "" {
		t.Error("CalculatedAt should not be empty")
	}
}

func TestCalculatePrice_VehicleNotFound(t *testing.T) {
	cfg := defaultConfig()

	vehicleRepo := &mockVehicleRepo{
		getVehicleFn: func(ctx context.Context, id string) (*pricing.Vehicle, error) {
			return nil, nil
		},
	}
	zoneRepo := &mockZoneRepo{
		getZoneFn: func(ctx context.Context, zone string) (*pricing.Zone, error) {
			return &pricing.Zone{Name: "south-jakarta", Utilization: 85}, nil
		},
	}
	configStore := &mockConfigStore{
		getActiveFn: func(ctx context.Context) (*pricing.PricingConfig, error) {
			return &cfg, nil
		},
	}
	auditRepo := &mockAuditRepo{
		recordFn: func(ctx context.Context, entry *pricing.AuditEntry) error {
			return nil
		},
	}

	uc := newUseCase(vehicleRepo, zoneRepo, configStore, auditRepo)

	_, err := uc.CalculatePrice(context.Background(), pricing.PricingInput{
		VehicleID:     "EV-10001",
		Zone:          "south-jakarta",
		DurationHours: 3,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Error message should contain "vehicle" and "not found"
	errMsg := err.Error()
	if !contains(errMsg, "vehicle") || !contains(errMsg, "not found") {
		t.Errorf("error = %q, want message containing 'vehicle' and 'not found'", errMsg)
	}
}

func TestCalculatePrice_ZoneNotFound(t *testing.T) {
	cfg := defaultConfig()

	vehicleRepo := &mockVehicleRepo{
		getVehicleFn: func(ctx context.Context, id string) (*pricing.Vehicle, error) {
			return &pricing.Vehicle{ID: "EV-10001", SoC: 35}, nil
		},
	}
	zoneRepo := &mockZoneRepo{
		getZoneFn: func(ctx context.Context, zone string) (*pricing.Zone, error) {
			return nil, nil
		},
	}
	configStore := &mockConfigStore{
		getActiveFn: func(ctx context.Context) (*pricing.PricingConfig, error) {
			return &cfg, nil
		},
	}
	auditRepo := &mockAuditRepo{
		recordFn: func(ctx context.Context, entry *pricing.AuditEntry) error {
			return nil
		},
	}

	uc := newUseCase(vehicleRepo, zoneRepo, configStore, auditRepo)

	_, err := uc.CalculatePrice(context.Background(), pricing.PricingInput{
		VehicleID:     "EV-10001",
		Zone:          "south-jakarta",
		DurationHours: 3,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetBreakdown_Success(t *testing.T) {
	cfg := defaultConfig()

	vehicleRepo := &mockVehicleRepo{
		getVehicleFn: func(ctx context.Context, id string) (*pricing.Vehicle, error) {
			return &pricing.Vehicle{ID: "EV-10001", Zone: "south-jakarta", SoC: 35, Model: "E1"}, nil
		},
	}
	zoneRepo := &mockZoneRepo{
		getZoneFn: func(ctx context.Context, zone string) (*pricing.Zone, error) {
			return &pricing.Zone{Name: "south-jakarta", Code: "SJ", Utilization: 85}, nil
		},
	}
	configStore := &mockConfigStore{
		getActiveFn: func(ctx context.Context) (*pricing.PricingConfig, error) {
			return &cfg, nil
		},
	}
	auditRepo := &mockAuditRepo{
		recordFn: func(ctx context.Context, entry *pricing.AuditEntry) error {
			return nil
		},
	}

	uc := newUseCase(vehicleRepo, zoneRepo, configStore, auditRepo)

	resp, err := uc.GetBreakdown(context.Background(), pricing.PricingInput{
		VehicleID:     "EV-10001",
		Zone:          "south-jakarta",
		DurationHours: 3,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.CalculationSteps) != 5 {
		t.Fatalf("CalculationSteps count = %d, want 5", len(resp.CalculationSteps))
	}

	expectedSteps := []string{
		"base_rate_per_hour",
		"demand_multiplier",
		"zone_surge_factor",
		"battery_discount_factor",
		"duration_hours",
	}
	for i, want := range expectedSteps {
		if resp.CalculationSteps[i].Step != want {
			t.Errorf("step[%d] = %s, want %s", i, resp.CalculationSteps[i].Step, want)
		}
	}

	if resp.Formula != "base_rate × demand_multiplier × zone_surge × battery_discount × duration" {
		t.Errorf("Formula = %s, want expected formula", resp.Formula)
	}
	if resp.TotalPrice <= 0 {
		t.Errorf("TotalPrice = %.0f, want > 0", resp.TotalPrice)
	}
}

func TestUpdateConfig_Validation(t *testing.T) {
	uc := newUseCase(&mockVehicleRepo{}, &mockZoneRepo{}, &mockConfigStore{}, &mockAuditRepo{})

	t.Run("base_price_per_hour <= 0", func(t *testing.T) {
		_, err := uc.UpdateConfig(context.Background(), 0, "IDR", 2.0, nil, nil, nil, "admin")
		if err == nil {
			t.Fatal("expected error for base_price_per_hour <= 0, got nil")
		}
		if err.Error() != "base_price_per_hour must be > 0" {
			t.Errorf("error = %q, want %q", err.Error(), "base_price_per_hour must be > 0")
		}
	})

	t.Run("surge_cap < 1.0", func(t *testing.T) {
		_, err := uc.UpdateConfig(context.Background(), 6250, "IDR", 0.5, nil, nil, nil, "admin")
		if err == nil {
			t.Fatal("expected error for surge_cap < 1.0, got nil")
		}
		if err.Error() != "surge_cap_multiplier must be >= 1.0" {
			t.Errorf("error = %q, want %q", err.Error(), "surge_cap_multiplier must be >= 1.0")
		}
	})
}

func TestUpdateConfig_Success(t *testing.T) {
	existingCfg := defaultConfig() // version 1
	savedCfg := (*pricing.PricingConfig)(nil)

	configStore := &mockConfigStore{
		getActiveFn: func(ctx context.Context) (*pricing.PricingConfig, error) {
			return &existingCfg, nil
		},
		saveFn: func(ctx context.Context, config *pricing.PricingConfig, changedBy string) error {
			savedCfg = config
			return nil
		},
	}

	uc := newUseCase(&mockVehicleRepo{}, &mockZoneRepo{}, configStore, &mockAuditRepo{})

	demandJSON := json.RawMessage(`{"default": 1.0, "rules": [{"days": [1,2,3,4,5], "hours": [17,18,19], "multiplier": 1.3}]}`)
	zoneJSON := json.RawMessage(`{"thresholds": [{"max_utilization": 50, "factor": 1.0}, {"max_utilization": 80, "factor": 1.2}, {"max_utilization": 100, "factor": 1.5}]}`)
	batteryJSON := json.RawMessage(`{"thresholds": [{"max_soc": 40, "discount_factor": 0.85}, {"max_soc": 60, "discount_factor": 0.92}, {"max_soc": 80, "discount_factor": 0.97}, {"max_soc": 100, "discount_factor": 1.0}]}`)

	cfg, err := uc.UpdateConfig(context.Background(), 7000, "IDR", 2.5, demandJSON, zoneJSON, batteryJSON, "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Version != 2 {
		t.Errorf("Version = %d, want 2", cfg.Version)
	}
	if cfg.BasePricePerHour != 7000 {
		t.Errorf("BasePricePerHour = %.0f, want 7000", cfg.BasePricePerHour)
	}
	if cfg.Currency != "IDR" {
		t.Errorf("Currency = %s, want IDR", cfg.Currency)
	}
	if cfg.SurgeCap != 2.5 {
		t.Errorf("SurgeCap = %.1f, want 2.5", cfg.SurgeCap)
	}
	if savedCfg == nil {
		t.Fatal("Save was not called")
	}
	if savedCfg.Version != 2 {
		t.Errorf("saved config Version = %d, want 2", savedCfg.Version)
	}
}

func TestGetAuditLogs(t *testing.T) {
	entries := []pricing.AuditEntry{
		{ID: "a1", VehicleID: "EV-1", Zone: "south-jakarta", DurationHours: 1, FinalPrice: 5000},
		{ID: "a2", VehicleID: "EV-2", Zone: "north-jakarta", DurationHours: 2, FinalPrice: 10000},
	}

	auditRepo := &mockAuditRepo{
		listFn: func(ctx context.Context, page, pageSize int, vehicleID, zone string) ([]pricing.AuditEntry, int, error) {
			return entries, 2, nil
		},
	}

	uc := newUseCase(&mockVehicleRepo{}, &mockZoneRepo{}, &mockConfigStore{}, auditRepo)

	resp, err := uc.GetAuditLogs(context.Background(), 1, 10, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Total != 2 {
		t.Errorf("Total = %d, want 2", resp.Total)
	}
	if resp.Page != 1 {
		t.Errorf("Page = %d, want 1", resp.Page)
	}
	if resp.PageSize != 10 {
		t.Errorf("PageSize = %d, want 10", resp.PageSize)
	}
	if resp.TotalPages != 1 {
		t.Errorf("TotalPages = %d, want 1", resp.TotalPages)
	}

	data, ok := resp.Data.([]pricing.AuditEntry)
	if !ok {
		t.Fatalf("Data is not []pricing.AuditEntry")
	}
	if len(data) != 2 {
		t.Errorf("Data length = %d, want 2", len(data))
	}
}

func TestListZones(t *testing.T) {
	zones := []pricing.Zone{
		{Name: "south-jakarta", Code: "SJ", Utilization: 85},
		{Name: "north-jakarta", Code: "NJ", Utilization: 60},
	}

	zoneRepo := &mockZoneRepo{
		listZonesFn: func(ctx context.Context) ([]pricing.Zone, error) {
			return zones, nil
		},
	}

	uc := newUseCase(&mockVehicleRepo{}, zoneRepo, &mockConfigStore{}, &mockAuditRepo{})

	result, err := uc.ListZones(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("zone count = %d, want 2", len(result))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
