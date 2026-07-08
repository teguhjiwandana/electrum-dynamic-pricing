package postgres

import (
	"context"
	"fmt"

	"github.com/electrum/dynamic-pricing-engine/internal/domain/pricing"
)

// vehicleRepo implements pricing.VehicleLookup.
type vehicleRepo struct{}

// NewVehicleRepo returns a VehicleLookup backed by PostgreSQL.
func NewVehicleRepo() pricing.VehicleLookup {
	return &vehicleRepo{}
}

func (r *vehicleRepo) GetVehicle(ctx context.Context, vehicleID string) (*pricing.Vehicle, error) {
	if Pool == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	row := Pool.QueryRow(ctx,
		`SELECT id, zone, soc, model FROM vehicles WHERE id = $1`,
		vehicleID,
	)

	var v pricing.Vehicle
	err := row.Scan(&v.ID, &v.Zone, &v.SoC, &v.Model)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil // not found
		}
		return nil, fmt.Errorf("get vehicle %s: %w", vehicleID, err)
	}

	return &v, nil
}

func (r *vehicleRepo) ListVehicles(ctx context.Context) ([]pricing.Vehicle, error) {
	if Pool == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := Pool.Query(ctx,
		`SELECT id, zone, soc, model FROM vehicles ORDER BY id`,
	)
	if err != nil {
		return nil, fmt.Errorf("list vehicles: %w", err)
	}
	defer rows.Close()

	var vehicles []pricing.Vehicle
	for rows.Next() {
		var v pricing.Vehicle
		if err := rows.Scan(&v.ID, &v.Zone, &v.SoC, &v.Model); err != nil {
			return nil, fmt.Errorf("scan vehicle: %w", err)
		}
		vehicles = append(vehicles, v)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return vehicles, nil
}
