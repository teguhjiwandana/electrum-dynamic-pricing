package pricing

import (
	"context"
	"math"
	"time"
)

// Service is the pure domain service for pricing calculations.
// It has no I/O, no config parsing, no external dependencies —
// just business math with already-resolved typed inputs.
type Service struct{}

// NewService creates a new pricing domain service.
func NewService() *Service {
	return &Service{}
}

// Calculate executes the core pricing formula:
//
//	final = baseRate × demand × surge × battery × hours
//
// with a surge cap: final cannot exceed baseRate × surgeCap × hours.
// Returns the computed price with full factor breakdown.
func (s *Service) Calculate(
	ctx context.Context,
	input PricingInput,
	cfg PricingConfig,
	vehicle Vehicle,
	zone Zone,
	now time.Time,
) (*PricingOutput, error) {
	// Respect context cancellation.
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	demandMult := cfg.DemandRules.GetDemandMultiplier(now.Weekday(), now.Hour())
	zoneSurge := cfg.ZoneSurge.GetSurgeFactor(zone.Utilization)
	batteryDisc := cfg.BatteryDiscounts.GetDiscountFactor(vehicle.SoC)

	baseRate := cfg.BasePricePerHour
	dur := float64(input.DurationHours)

	// Core formula.
	uncapped := baseRate * demandMult * zoneSurge * batteryDisc * dur

	// Surge cap enforcement.
	capLimit := baseRate * cfg.SurgeCap * dur
	if uncapped > capLimit {
		uncapped = capLimit
	}

	total := math.Round(uncapped)

	return &PricingOutput{
		VehicleID:     input.VehicleID,
		Zone:          input.Zone,
		DurationHours: input.DurationHours,
		TotalPrice:    total,
		Currency:      cfg.Currency,
		Factors: PricingFactors{
			BaseRatePerHour:       baseRate,
			DemandMultiplier:      demandMult,
			ZoneSurgeFactor:       zoneSurge,
			BatteryDiscountFactor: batteryDisc,
		},
		CalculatedAt: now,
	}, nil
}
