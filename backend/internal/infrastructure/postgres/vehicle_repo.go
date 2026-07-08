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
