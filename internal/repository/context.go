package repository

import (
	"context"

	"gorm.io/gorm"
)

// contextKey is a type for context keys to avoid collisions
type contextKey string

// txKey is the context key for storing GORM transaction
const txKey contextKey = "gorm_tx"

// txOrDefault returns the transaction from context if available, otherwise the default DB.
// This allows repositories to participate in transactions managed by the Unit of Work.
func txOrDefault(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey).(*gorm.DB); ok && tx != nil {
		return tx.WithContext(ctx)
	}
	return db.WithContext(ctx)
}
