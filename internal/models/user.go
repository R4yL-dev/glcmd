package models

import "time"

// UserPreferences represents user preferences and account information.
// Source: /user → data.user
type UserPreferences struct {
	UserID               string    // id: Unique user ID
	FirstName            string    // firstName: User's first name
	LastName             string    // lastName: User's last name
	Email                string    // email: User's email address
	Country              string    // country: ISO 2-letter country code (e.g., "CH")
	AccountType          string    // accountType: Account type ("pat" = patient)
	DateOfBirth          time.Time // dateOfBirth: Unix timestamp converted to time.Time
	Created              time.Time // created: Account creation timestamp
	LastLogin            time.Time // lastLogin: Last login timestamp

	// Display preferences
	UILanguage           string    // uiLanguage: UI language code ("fr", "en", etc.)
	CommunicationLanguage string   // communicationLanguage: Communication language
	UnitOfMeasure        int       // uom: Unit of measurement (0=mmol/L, 1=mg/dL)
	DateFormat           int       // dateFormat: Preferred date format
	TimeFormat           int       // timeFormat: Preferred time format (2 = 24h?)

	// Additional metadata
	EmailDays            []int     // emailDay: Days for email notifications
}
