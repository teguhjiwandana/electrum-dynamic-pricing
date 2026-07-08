package postgres

import (
	"context"
	"fmt"

	"github.com/electrum/dynamic-pricing-engine/internal/domain/pricing"
)

// zoneRepo implements pricing.ZoneLookup.
type zoneRepo struct{}

// NewZoneRepo returns a ZoneLookup backed by PostgreSQL.
func NewZoneRepo() pricing.ZoneLookup {
	return &zoneRepo{}
}

func (r *zoneRepo) GetZone(ctx context.Context, zoneName string) (*pricing.Zone, error) {
	if Pool == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	row := Pool.QueryRow(ctx,
		`SELECT name, zone, utilization FROM zone_utilization WHERE zone = $1`,
		zoneName,
	)

	var z pricing.Zone
	err := row.Scan(&z.Name, &z.Code, &z.Utilization)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil // not found
		}
		return nil, fmt.Errorf("get zone %s: %w", zoneName, err)
	}

	return &z, nil
}

func (r *zoneRepo) ListZones(ctx context.Context) ([]pricing.Zone, error) {
	if Pool == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := Pool.Query(ctx,
		`SELECT name, zone, utilization FROM zone_utilization ORDER BY name`,
	)
	if err != nil {
		return nil, fmt.Errorf("list zones: %w", err)
	}
	defer rows.Close()

	var zones []pricing.Zone
	for rows.Next() {
		var z pricing.Zone
		if err := rows.Scan(&z.Name, &z.Code, &z.Utilization); err != nil {
			return nil, fmt.Errorf("scan zone: %w", err)
		}
		zones = append(zones, z)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return zones, nil
}
