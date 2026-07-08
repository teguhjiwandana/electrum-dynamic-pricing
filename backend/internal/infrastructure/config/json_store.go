package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/electrum/dynamic-pricing-engine/internal/domain/pricing"
)

// JSONStore implements pricing.ConfigStore for file-based persistence.
// It also provides hot-reload via a background file watcher.
type JSONStore struct {
	mu         sync.RWMutex
	configPath string
	current    *pricing.PricingConfig
	modTime    time.Time
	version    int
}

// NewJSONStore creates a config store backed by a JSON file.
func NewJSONStore(configPath string) *JSONStore {
	return &JSONStore{configPath: configPath}
}

// Load reads the JSON file and eagerly parses it into a PricingConfig.
func (s *JSONStore) Load(ctx context.Context) error {
	info, err := os.Stat(s.configPath)
	if err != nil {
		return fmt.Errorf("config file stat: %w", err)
	}

	s.mu.RLock()
	same := info.ModTime().Equal(s.modTime) && s.current != nil
	s.mu.RUnlock()

	if same {
		return nil
	}

	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return fmt.Errorf("config file read: %w", err)
	}

	// Intermediate raw struct for JSON unmarshal
	var raw rawConfig

	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("config parse: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if content actually changed
	if s.current != nil && configsEqual(s.current, &raw) {
		s.modTime = info.ModTime()
		return nil
	}

	newVersion := s.version + 1
	if newVersion < 1 {
		newVersion = 1
	}

	now := time.Now()
	s.current = &pricing.PricingConfig{
		BasePricePerHour: raw.BasePricePerHour,
		Currency:         raw.Currency,
		SurgeCap:         raw.SurgeCap,
		DemandRules:      raw.DemandRules,
		ZoneSurge:        raw.ZoneSurge,
		BatteryDiscounts: raw.BatteryDiscounts,
		Version:          newVersion,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	s.version = newVersion
	s.modTime = info.ModTime()

	return nil
}

// UpdateFromRequest creates a new config version from the update request,
// persists to the JSON file, and updates the in-memory config.
func (s *JSONStore) UpdateFromRequest(ctx context.Context, basePrice float64, currency string, surgeCap float64, demandRules pricing.DemandMultipliers, zoneSurge pricing.ZoneSurgeConfig, batteryDiscounts pricing.BatteryDiscountTiers, changedBy string) (*pricing.PricingConfig, error) {
	if basePrice <= 0 {
		return nil, fmt.Errorf("base_price_per_hour must be > 0")
	}
	if surgeCap < 1.0 {
		return nil, fmt.Errorf("surge_cap_multiplier must be >= 1.0")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.current == nil {
		return nil, fmt.Errorf("no config loaded")
	}

	newVer := s.current.Version + 1
	cfg := &pricing.PricingConfig{
		BasePricePerHour: basePrice,
		Currency:         currency,
		SurgeCap:         surgeCap,
		DemandRules:      demandRules,
		ZoneSurge:        zoneSurge,
		BatteryDiscounts: batteryDiscounts,
		Version:          newVer,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Write to file
	if err := s.writeLocked(cfg); err != nil {
		return nil, err
	}

	s.current = cfg
	s.version = newVer
	return cfg, nil
}

// GetActive implements pricing.ConfigStore.
func (s *JSONStore) GetActive(ctx context.Context) (*pricing.PricingConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.current == nil {
		return nil, fmt.Errorf("no config loaded")
	}

	cpy := *s.current
	return &cpy, nil
}

// Save implements pricing.ConfigStore — delegates to DB store if available.
func (s *JSONStore) Save(ctx context.Context, cfg *pricing.PricingConfig, changedBy string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.writeLocked(cfg); err != nil {
		return err
	}

	s.current = cfg
	if cfg.Version > s.version {
		s.version = cfg.Version
	}
	return nil
}

// GetHistory implements pricing.ConfigStore — file store has no history.
func (s *JSONStore) GetHistory(ctx context.Context, page, pageSize int) ([]pricing.PricingConfig, int, error) {
	return nil, 0, nil
}

// Watcher polls the config file for changes and calls onReload on detection.
func (s *JSONStore) Watcher(ctx context.Context, interval time.Duration, onReload func(cfg *pricing.PricingConfig)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			info, err := os.Stat(s.configPath)
			if err != nil {
				continue
			}

			s.mu.RLock()
			changed := !info.ModTime().Equal(s.modTime)
			s.mu.RUnlock()

			if changed {
				if err := s.Load(ctx); err == nil && s.current != nil && onReload != nil {
					onReload(s.current)
				}
			}
		}
	}
}

func (s *JSONStore) writeLocked(cfg *pricing.PricingConfig) error {
	raw := rawConfig{
		BasePricePerHour: cfg.BasePricePerHour,
		Currency:         cfg.Currency,
		SurgeCap:         cfg.SurgeCap,
		DemandRules:      cfg.DemandRules,
		ZoneSurge:        cfg.ZoneSurge,
		BatteryDiscounts: cfg.BatteryDiscounts,
	}

	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(s.configPath, data, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	if info, err := os.Stat(s.configPath); err == nil {
		s.modTime = info.ModTime()
	}
	return nil
}

type rawConfig struct {
	BasePricePerHour float64                   `json:"base_price_per_hour"`
	Currency         string                    `json:"currency"`
	SurgeCap         float64                   `json:"surge_cap_multiplier"`
	DemandRules      pricing.DemandMultipliers `json:"demand_multipliers"`
	ZoneSurge        pricing.ZoneSurgeConfig   `json:"zone_surge_config"`
	BatteryDiscounts pricing.BatteryDiscountTiers `json:"battery_discount_tiers"`
}

func configsEqual(a *pricing.PricingConfig, raw *rawConfig) bool {
	return a.BasePricePerHour == raw.BasePricePerHour &&
		a.Currency == raw.Currency &&
		a.SurgeCap == raw.SurgeCap &&
		rulesEqual(a.DemandRules, raw.DemandRules) &&
		surgeEqual(a.ZoneSurge, raw.ZoneSurge) &&
		battEqual(a.BatteryDiscounts, raw.BatteryDiscounts)
}

func rulesEqual(a, b pricing.DemandMultipliers) bool { return a.Default == b.Default && len(a.Rules) == len(b.Rules) }
func surgeEqual(a, b pricing.ZoneSurgeConfig) bool   { return len(a.Thresholds) == len(b.Thresholds) }
func battEqual(a, b pricing.BatteryDiscountTiers) bool { return len(a.Thresholds) == len(b.Thresholds) }
