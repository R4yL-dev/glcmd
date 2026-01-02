package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/R4yL-dev/glcmd/internal/domain"
	"github.com/R4yL-dev/glcmd/internal/persistence"
)

// UserRepositoryGORM is the GORM implementation of UserRepository.
// This is a singleton repository - only one user record is expected.
type UserRepositoryGORM struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *gorm.DB) *UserRepositoryGORM {
	return &UserRepositoryGORM{db: db}
}

// Save creates or updates user preferences (singleton).
func (r *UserRepositoryGORM) Save(ctx context.Context, u *domain.UserPreferences) error {
	db := txOrDefault(ctx, r.db)

	// Upsert: Update all fields on conflict with user_id
	result := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"first_name", "last_name", "email", "country", "account_type",
			"date_of_birth", "created", "last_login", "ui_language",
			"communication_language", "unit_of_measure", "date_format",
			"time_format", "email_days", "updated_at",
		}),
	}).Create(u)

	return result.Error
}

// Find returns the user preferences (only one record expected).
func (r *UserRepositoryGORM) Find(ctx context.Context) (*domain.UserPreferences, error) {
	db := txOrDefault(ctx, r.db)

	var user domain.UserPreferences
	result := db.First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, persistence.ErrNotFound
		}
		return nil, result.Error
	}

	return &user, nil
}
