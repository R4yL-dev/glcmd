package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// UserPreferences represents user preferences and account information.
// Source: /user → data.user
type UserPreferences struct {
	// Database fields
	ID        uint      `gorm:"primaryKey" json:"-"`
	UpdatedAt time.Time `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`

	UserID               string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"userId"`               // id: Unique user ID
	FirstName            string    `gorm:"type:varchar(100)" json:"firstName"`                                 // firstName: User's first name
	LastName             string    `gorm:"type:varchar(100)" json:"lastName"`                                  // lastName: User's last name
	Email                string    `gorm:"type:varchar(255)" json:"email"`                                     // email: User's email address
	Country              string    `gorm:"type:varchar(2)" json:"country"`                                     // country: ISO 2-letter country code (e.g., "CH")
	AccountType          string    `gorm:"type:varchar(50)" json:"accountType"`                                // accountType: Account type ("pat" = patient)
	DateOfBirth          time.Time `gorm:"type:datetime" json:"dateOfBirth"`                                   // dateOfBirth: Unix timestamp converted to time.Time
	Created              time.Time `gorm:"type:datetime" json:"created"`                                       // created: Account creation timestamp
	LastLogin            time.Time `gorm:"type:datetime" json:"lastLogin"`                                     // lastLogin: Last login timestamp

	// Display preferences
	UILanguage           string    `gorm:"type:varchar(10)" json:"uiLanguage"`                                 // uiLanguage: UI language code ("fr", "en", etc.)
	CommunicationLanguage string   `gorm:"type:varchar(10)" json:"communicationLanguage"`                      // communicationLanguage: Communication language
	UnitOfMeasure        int       `gorm:"type:integer" json:"unitOfMeasure"`                                  // uom: Unit of measurement (0=mmol/L, 1=mg/dL)
	DateFormat           int       `gorm:"type:integer" json:"dateFormat"`                                     // dateFormat: Preferred date format
	TimeFormat           int       `gorm:"type:integer" json:"timeFormat"`                                     // timeFormat: Preferred time format (2 = 24h?)

	// Additional metadata
	EmailDays            IntArray  `gorm:"type:text" json:"emailDays"`                                         // emailDay: Days for email notifications (stored as JSON)
}

// TableName specifies the table name for GORM.
func (UserPreferences) TableName() string {
	return "user_preferences"
}

// IntArray is a custom type for storing []int as JSON in the database.
type IntArray []int

// Scan implements the sql.Scanner interface for reading from the database.
func (a *IntArray) Scan(value interface{}) error {
	if value == nil {
		*a = []int{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal IntArray value")
	}

	return json.Unmarshal(bytes, a)
}

// Value implements the driver.Valuer interface for writing to the database.
func (a IntArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "[]", nil
	}
	return json.Marshal(a)
}
