package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// UnitOfWork defines the interface for transaction management.
type UnitOfWork interface {
	// ExecuteInTransaction executes a function within a database transaction.
	// If the function returns an error, the transaction is rolled back.
	// Otherwise, the transaction is committed.
	ExecuteInTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
}

// GORMUnitOfWork is the GORM implementation of UnitOfWork.
type GORMUnitOfWork struct {
	db *gorm.DB
}

// NewUnitOfWork creates a new UnitOfWork.
func NewUnitOfWork(db *gorm.DB) *GORMUnitOfWork {
	return &GORMUnitOfWork{db: db}
}

// ExecuteInTransaction executes a function within a database transaction.
//
// The transaction is stored in the context and can be retrieved by repositories
// using the txOrDefault helper function.
//
// If the function returns an error, the transaction is rolled back.
// If the function succeeds, the transaction is committed.
func (uow *GORMUnitOfWork) ExecuteInTransaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	// Begin transaction
	tx := uow.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Create a new context with the transaction
	txCtx := context.WithValue(ctx, txKey, tx)

	// Execute the function
	err := fn(txCtx)
	if err != nil {
		// Rollback on error
		if rbErr := tx.Rollback().Error; rbErr != nil {
			return fmt.Errorf("failed to rollback transaction after error %v: %w", err, rbErr)
		}
		return err
	}

	// Commit on success
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
