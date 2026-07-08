package pricing_test

import (
	"context"
	"testing"
	"time"

	"github.com/electrum/dynamic-pricing-engine/internal/domain/pricing"
)

func makeConfig(baseRate, surgeCap float64) pricing.PricingConfig {
	return pricing.PricingConfig{
		BasePricePerHour: baseRate,
		Currency:         "IDR",
		SurgeCap:         surgeCap,
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
	}
}

// ---------------------------------------------------------------------------
// Engine tests
// ---------------------------------------------------------------------------

func TestCalculate_PRDExample(t *testing.T) {
	svc := pricing.NewService()
	cfg := makeConfig(6250, 2.0)

	// PRD example: weekday 5-7PM, utilization 85%, SoC 35%, 3 hours
	// Expected: 6250 * 1.3 * 1.5 * 0.85 * 3 = 31078.125 → 31078
	tuesday := time.Date(2026, 7, 7, 17, 30, 0, 0, time.UTC) // Tuesday 5:30PM

	input := pricing.PricingInput{
		VehicleID:     "EV-12345",
		Zone:          "south-jakarta",
		DurationHours: 3,
	}
	vehicle := pricing.Vehicle{ID: "EV-12345", Zone: "south-jakarta", SoC: 35, Model: "E1"}
	zone := pricing.Zone{Name: "south-jakarta", Utilization: 85}

	output, err := svc.Calculate(context.Background(), input, cfg, vehicle, zone, tuesday)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := float64(31078)
	if output.TotalPrice != expected {
		t.Errorf("total = %.0f, want %.0f", output.TotalPrice, expected)
	}

	if output.Factors.DemandMultiplier != 1.3 {
		t.Errorf("demand = %.2f, want 1.3", output.Factors.DemandMultiplier)
	}
	if output.Factors.ZoneSurgeFactor != 1.5 {
		t.Errorf("surge = %.2f, want 1.5", output.Factors.ZoneSurgeFactor)
	}
	if output.Factors.BatteryDiscountFactor != 0.85 {
		t.Errorf("battery = %.2f, want 0.85", output.Factors.BatteryDiscountFactor)
	}
}

func TestCalculate_SurgeCapEnforcement(t *testing.T) {
	svc := pricing.NewService()
	cfg := makeConfig(6250, 2.0) // cap = 2x

	// Extreme case: highest demand, highest surge, no discount, 10 hours
	// Uncapped: 6250 * 1.3 * 1.5 * 1.0 * 10 = 121875
	// Cap: 6250 * 2.0 * 10 = 125000
	// Result should be capped at 125000 (same as uncapped in this edge case — let's test a tighter cap)
	cfg.SurgeCap = 1.5 // cap 1.5x
	// Uncapped: 121875, Cap: 6250 * 1.5 * 10 = 93750

	tuesday := time.Date(2026, 7, 7, 18, 0, 0, 0, time.UTC)
	input := pricing.PricingInput{VehicleID: "EV-X", Zone: "hot-zone", DurationHours: 10}
	vehicle := pricing.Vehicle{ID: "EV-X", SoC: 100}
	zone := pricing.Zone{Name: "hot-zone", Utilization: 95}

	output, err := svc.Calculate(context.Background(), input, cfg, vehicle, zone, tuesday)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.TotalPrice > 93750 {
		t.Errorf("total = %.0f, want ≤ 93750 (surge cap)", output.TotalPrice)
	}
}

func TestCalculate_SurgeCapNotTriggered(t *testing.T) {
	svc := pricing.NewService()
	cfg := makeConfig(6250, 3.0) // high cap

	// Low demand (midnight), low zone, high discount, 1 hour
	// Uncapped: 6250 * 0.8 * 1.0 * 0.85 * 1 = 4250
	// Cap: 6250 * 3.0 * 1 = 18750 — not triggered
	tuesdayMidnight := time.Date(2026, 7, 7, 2, 0, 0, 0, time.UTC)
	input := pricing.PricingInput{VehicleID: "EV-Y", Zone: "quiet", DurationHours: 1}
	vehicle := pricing.Vehicle{ID: "EV-Y", SoC: 35}
	zone := pricing.Zone{Name: "quiet", Utilization: 30}

	output, err := svc.Calculate(context.Background(), input, cfg, vehicle, zone, tuesdayMidnight)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.TotalPrice != 4250 {
		t.Errorf("total = %.0f, want 4250", output.TotalPrice)
	}
}

func TestCalculate_ContextCancelled(t *testing.T) {
	svc := pricing.NewService()
	cfg := makeConfig(6250, 2.0)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tuesday := time.Date(2026, 7, 7, 17, 0, 0, 0, time.UTC)
	input := pricing.PricingInput{VehicleID: "EV-Z", Zone: "anywhere", DurationHours: 1}
	vehicle := pricing.Vehicle{ID: "EV-Z", SoC: 50}
	zone := pricing.Zone{Name: "anywhere", Utilization: 50}

	_, err := svc.Calculate(ctx, input, cfg, vehicle, zone, tuesday)
	if err == nil {
		t.Error("expected context cancellation error, got nil")
	}
}

// ---------------------------------------------------------------------------
// Factor tests
// ---------------------------------------------------------------------------

func TestGetDemandMultiplier(t *testing.T) {
	dm := pricing.DemandMultipliers{
		Default: 1.0,
		Rules: []pricing.DemandRule{
			{Days: []int{1, 2, 3, 4, 5}, Hours: []int{17, 18, 19}, Multiplier: 1.3},
			{Days: []int{1, 2, 3, 4, 5}, Hours: []int{0, 1, 2, 3, 4}, Multiplier: 0.8},
		},
	}

	tests := []struct {
		name string
		dow  time.Weekday
		hour int
		want float64
	}{
		{"weekday peak", time.Tuesday, 18, 1.3},
		{"weekday off-peak", time.Wednesday, 2, 0.8},
		{"weekend default", time.Sunday, 15, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dm.GetDemandMultiplier(tt.dow, tt.hour)
			if got != tt.want {
				t.Errorf("demand = %.2f, want %.2f", got, tt.want)
			}
		})
	}
}

func TestGetZoneSurgeFactor(t *testing.T) {
	zsc := pricing.ZoneSurgeConfig{
		Thresholds: []pricing.ZoneSurgeThreshold{
			{MaxUtilization: 50, Factor: 1.0},
			{MaxUtilization: 80, Factor: 1.2},
			{MaxUtilization: 100, Factor: 1.5},
		},
	}

	tests := []struct {
		name        string
		utilization float64
		want        float64
	}{
		{"low", 30, 1.0},
		{"mid", 65, 1.2},
		{"high", 90, 1.5},
		{"boundary", 50, 1.0},
		{"max", 100, 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := zsc.GetSurgeFactor(tt.utilization)
			if got != tt.want {
				t.Errorf("surge = %.2f, want %.2f", got, tt.want)
			}
		})
	}
}

func TestGetBatteryDiscountFactor(t *testing.T) {
	bdt := pricing.BatteryDiscountTiers{
		Thresholds: []pricing.BatteryDiscountThreshold{
			{MaxSOC: 40, DiscountFactor: 0.85},
			{MaxSOC: 60, DiscountFactor: 0.92},
			{MaxSOC: 80, DiscountFactor: 0.97},
			{MaxSOC: 100, DiscountFactor: 1.0},
		},
	}

	tests := []struct {
		name string
		soc  float64
		want float64
	}{
		{"low battery", 15, 0.85},
		{"mid battery", 55, 0.92},
		{"high battery", 90, 1.0},
		{"full", 100, 1.0},
		{"boundary", 40, 0.85},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bdt.GetDiscountFactor(tt.soc)
			if got != tt.want {
				t.Errorf("discount = %.2f, want %.2f", got, tt.want)
			}
		})
	}
}
