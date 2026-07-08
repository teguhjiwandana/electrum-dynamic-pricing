package pricing

import (
	"time"
)

// ---------------------------------------------------------------------------
// Demand multiplier rules (time-based)
// ---------------------------------------------------------------------------

// DemandMultipliers configures time-of-day / day-of-week pricing factors.
// Rules are evaluated in order; the first matching rule wins.
// If no rule matches, Default is returned.
type DemandMultipliers struct {
	Default float64      `json:"default"`
	Rules   []DemandRule `json:"rules"`
}

// DemandRule is one time-slot rule. An empty Hours slice means "any hour".
type DemandRule struct {
	Days       []int   `json:"days"`   // Go Weekday: 0=Sun, 6=Sat
	Hours      []int   `json:"hours"`  // 0-23; empty = all hours
	Multiplier float64 `json:"multiplier"`
}

// GetDemandMultiplier returns the demand factor for the given weekday and hour.
func (dm DemandMultipliers) GetDemandMultiplier(dow time.Weekday, hour int) float64 {
	for _, r := range dm.Rules {
		if !matchDay(r.Days, dow) {
			continue
		}
		if len(r.Hours) > 0 && !matchHour(r.Hours, hour) {
			continue
		}
		return r.Multiplier
	}
	return dm.Default
}

func matchDay(days []int, dow time.Weekday) bool {
	for _, d := range days {
		if time.Weekday(d) == dow {
			return true
		}
	}
	return false
}

func matchHour(hours []int, hour int) bool {
	for _, h := range hours {
		if h == hour {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Zone surge config
// ---------------------------------------------------------------------------

// ZoneSurgeConfig maps utilization ranges to surge factors.
// Thresholds MUST be sorted by MaxUtilization ascending.
type ZoneSurgeConfig struct {
	Thresholds []ZoneSurgeThreshold `json:"thresholds"`
}

// ZoneSurgeThreshold pairs a max utilization with its surge factor.
type ZoneSurgeThreshold struct {
	MaxUtilization float64 `json:"max_utilization"`
	Factor         float64 `json:"factor"`
}

// GetSurgeFactor returns the surge factor for the given fleet utilization %.
// Falls back to 1.0 if thresholds are empty.
func (zsc ZoneSurgeConfig) GetSurgeFactor(utilization float64) float64 {
	for _, t := range zsc.Thresholds {
		if utilization <= t.MaxUtilization {
			return t.Factor
		}
	}
	if len(zsc.Thresholds) > 0 {
		return zsc.Thresholds[len(zsc.Thresholds)-1].Factor
	}
	return 1.0
}

// ---------------------------------------------------------------------------
// Battery discount tiers
// ---------------------------------------------------------------------------

// BatteryDiscountTiers maps State of Charge (SoC) ranges to discount factors.
// Thresholds MUST be sorted by MaxSOC ascending.
type BatteryDiscountTiers struct {
	Thresholds []BatteryDiscountThreshold `json:"thresholds"`
}

// BatteryDiscountThreshold pairs a max SoC with its discount factor.
type BatteryDiscountThreshold struct {
	MaxSOC         float64 `json:"max_soc"`
	DiscountFactor float64 `json:"discount_factor"`
}

// GetDiscountFactor returns the discount factor for the given vehicle SoC %.
// Falls back to 1.0 if thresholds are empty.
func (bdt BatteryDiscountTiers) GetDiscountFactor(soc float64) float64 {
	for _, t := range bdt.Thresholds {
		if soc <= t.MaxSOC {
			return t.DiscountFactor
		}
	}
	if len(bdt.Thresholds) > 0 {
		return bdt.Thresholds[len(bdt.Thresholds)-1].DiscountFactor
	}
	return 1.0
}
