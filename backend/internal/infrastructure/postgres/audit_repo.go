package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/electrum/dynamic-pricing-engine/internal/domain/pricing"
)

// auditRepo implements pricing.AuditRecorder.
type auditRepo struct{}

// NewAuditRepo returns an AuditRecorder backed by PostgreSQL.
func NewAuditRepo() pricing.AuditRecorder {
	return &auditRepo{}
}

func (r *auditRepo) Record(ctx context.Context, entry *pricing.AuditEntry) error {
	if Pool == nil {
		return fmt.Errorf("database not initialized")
	}

	inputJSON, _ := json.Marshal(entry.InputData)
	factorsJSON, _ := json.Marshal(entry.Factors)

	_, err := Pool.Exec(ctx, `
		INSERT INTO audit_log
			(id, timestamp, vehicle_id, zone, duration_hours,
			 input_data, factors_applied, final_price, config_version, signature)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`,
		entry.ID, entry.Timestamp, entry.VehicleID, entry.Zone,
		entry.DurationHours, inputJSON, factorsJSON,
		entry.FinalPrice, entry.ConfigVersion, entry.Signature,
	)
	if err != nil {
		return fmt.Errorf("record audit: %w", err)
	}

	return nil
}

func (r *auditRepo) List(ctx context.Context, page, pageSize int, vehicleID, zone string) ([]pricing.AuditEntry, int, error) {
	if Pool == nil {
		return nil, 0, fmt.Errorf("database not initialized")
	}

	// Count
	countSQL := `SELECT COUNT(*) FROM audit_log WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if vehicleID != "" {
		countSQL += fmt.Sprintf(` AND vehicle_id = $%d`, argIdx)
		args = append(args, vehicleID)
		argIdx++
	}
	if zone != "" {
		countSQL += fmt.Sprintf(` AND zone = $%d`, argIdx)
		args = append(args, zone)
		argIdx++
	}

	var total int
	if err := Pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count audit logs: %w", err)
	}

	// Query
	offset := (page - 1) * pageSize
	querySQL := `
		SELECT id, timestamp, vehicle_id, zone, duration_hours,
		       input_data, factors_applied, final_price, config_version, signature
		FROM audit_log
		WHERE 1=1
	`

	queryArgs := []interface{}{}
	qIdx := 1
	if vehicleID != "" {
		querySQL += fmt.Sprintf(` AND vehicle_id = $%d`, qIdx)
		queryArgs = append(queryArgs, vehicleID)
		qIdx++
	}
	if zone != "" {
		querySQL += fmt.Sprintf(` AND zone = $%d`, qIdx)
		queryArgs = append(queryArgs, zone)
		qIdx++
	}

	querySQL += fmt.Sprintf(` ORDER BY timestamp DESC LIMIT $%d OFFSET $%d`, qIdx, qIdx+1)
	queryArgs = append(queryArgs, pageSize, offset)

	rows, err := Pool.Query(ctx, querySQL, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list audit logs: %w", err)
	}
	defer rows.Close()

	var entries []pricing.AuditEntry
	for rows.Next() {
		var e pricing.AuditEntry
		var inputJSON, factorsJSON []byte
		var ts time.Time

		if err := rows.Scan(
			&e.ID, &ts, &e.VehicleID, &e.Zone, &e.DurationHours,
			&inputJSON, &factorsJSON, &e.FinalPrice,
			&e.ConfigVersion, &e.Signature,
		); err != nil {
			return nil, 0, fmt.Errorf("scan audit log: %w", err)
		}

		e.Timestamp = ts
		json.Unmarshal(inputJSON, &e.InputData)
		json.Unmarshal(factorsJSON, &e.Factors)
		entries = append(entries, e)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	_ = totalPages // included in paginated response wrapper upstream

	return entries, total, nil
}
