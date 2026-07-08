package auth

import (
	"context"
	"fmt"

	"github.com/electrum/dynamic-pricing-engine/internal/infrastructure/postgres"
)

// userRepo implements UserLookup backed by PostgreSQL.
type userRepo struct{}

// NewUserRepo returns a UserLookup implementation.
func NewUserRepo() UserLookup {
	return &userRepo{}
}

func (r *userRepo) GetByUsername(ctx context.Context, username string) (*User, error) {
	if postgres.Pool == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	row := postgres.Pool.QueryRow(ctx,
		`SELECT id, username, password_hash, role FROM users WHERE username = $1`,
		username,
	)

	var u User
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("get user %s: %w", username, err)
	}

	return &u, nil
}
