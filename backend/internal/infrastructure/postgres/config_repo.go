package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/electrum/dynamic-pricing-engine/internal/domain/pricing"
)

// configRepo implements pricing.ConfigStore.
type configRepo struct{}

// NewConfigRepo returns a ConfigStore backed by PostgreSQL.
func NewConfigRepo() pricing.ConfigStore {
	return &configRepo{}
}

func (r *configRepo) GetActive(ctx context.Context) (*pricing.PricingConfig, error) {
	if Pool == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	row := Pool.QueryRow(ctx, `
		SELECT base_price_per_hour, currency, surge_cap_multiplier,
		       demand_multipliers, zone_surge_config, battery_discount_tiers,
		       version, created_at, updated_at
		FROM pricing_config
		ORDER BY version DESC
		LIMIT 1
	`)

	var cfg pricing.PricingConfig
	var demandRaw, surgeRaw, batteryRaw []byte
	var createdAt, updatedAt time.Time

	err := row.Scan(
		&cfg.BasePricePerHour, &cfg.Currency, &cfg.SurgeCap,
		&demandRaw, &surgeRaw, &batteryRaw,
		&cfg.Version, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get active config: %w", err)
	}

	cfg.CreatedAt = createdAt
	cfg.UpdatedAt = updatedAt

	// Eager parse: JSONB → typed domain value objects
	if err := json.Unmarshal(demandRaw, &cfg.DemandRules); err != nil {
		return nil, fmt.Errorf("parse demand rules: %w", err)
	}
	if err := json.Unmarshal(surgeRaw, &cfg.ZoneSurge); err != nil {
		return nil, fmt.Errorf("parse zone surge: %w", err)
	}
	if err := json.Unmarshal(batteryRaw, &cfg.BatteryDiscounts); err != nil {
		return nil, fmt.Errorf("parse battery discounts: %w", err)
	}

	return &cfg, nil
}

func (r *configRepo) Save(ctx context.Context, cfg *pricing.PricingConfig, changedBy string) error {
	if Pool == nil {
		return fmt.Errorf("database not initialized")
	}

	demandJSON, _ := json.Marshal(cfg.DemandRules)
	surgeJSON, _ := json.Marshal(cfg.ZoneSurge)
	batteryJSON, _ := json.Marshal(cfg.BatteryDiscounts)

	_, err := Pool.Exec(ctx, `
		INSERT INTO pricing_config
			(base_price_per_hour, currency, surge_cap_multiplier,
			 demand_multipliers, zone_surge_config, battery_discount_tiers,
			 version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
	`, cfg.BasePricePerHour, cfg.Currency, cfg.SurgeCap,
		demandJSON, surgeJSON, batteryJSON,
		cfg.Version, cfg.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	// History record
	_, err = Pool.Exec(ctx, `
		INSERT INTO pricing_config_history
			(config_id, version, base_price_per_hour, currency,
			 surge_cap_multiplier, demand_multipliers, zone_surge_config,
			 battery_discount_tiers, changed_by, changed_at)
		SELECT id, $1, $2, $3, $4, $5, $6, $7, $8, NOW()
		FROM pricing_config
		WHERE version = $1
	`, cfg.Version, cfg.BasePricePerHour, cfg.Currency, cfg.SurgeCap,
		demandJSON, surgeJSON, batteryJSON, changedBy,
	)
	if err != nil {
		return fmt.Errorf("save config history: %w", err)
	}

	return nil
}

func (r *configRepo) GetHistory(ctx context.Context, page, pageSize int) ([]pricing.PricingConfig, int, error) {
	if Pool == nil {
		return nil, 0, fmt.Errorf("database not initialized")
	}

	// Count total
	var total int
	err := Pool.QueryRow(ctx, `SELECT COUNT(*) FROM pricing_config_history`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count config history: %w", err)
	}

	offset := (page - 1) * pageSize
	rows, err := Pool.Query(ctx, `
		SELECT version, base_price_per_hour, currency, surge_cap_multiplier,
		       demand_multipliers, zone_surge_config, battery_discount_tiers,
		       changed_at
		FROM pricing_config_history
		ORDER BY changed_at DESC
		LIMIT $1 OFFSET $2
	`, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("get config history: %w", err)
	}
	defer rows.Close()

	var history []pricing.PricingConfig
	for rows.Next() {
		var cfg pricing.PricingConfig
		var demandRaw, surgeRaw, batteryRaw []byte
		var changedAt time.Time

		if err := rows.Scan(
			&cfg.Version, &cfg.BasePricePerHour, &cfg.Currency, &cfg.SurgeCap,
			&demandRaw, &surgeRaw, &batteryRaw, &changedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan config history: %w", err)
		}

		cfg.CreatedAt = changedAt
		cfg.UpdatedAt = changedAt

		json.Unmarshal(demandRaw, &cfg.DemandRules)
		json.Unmarshal(surgeRaw, &cfg.ZoneSurge)
		json.Unmarshal(batteryRaw, &cfg.BatteryDiscounts)

		history = append(history, cfg)
	}

	return history, total, rows.Err()
}
