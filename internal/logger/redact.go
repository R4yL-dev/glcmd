package logger

import "regexp"

var uuidRegex = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)

// RedactPath masks UUIDs in URL paths to prevent logging sensitive IDs.
func RedactPath(path string) string {
	return uuidRegex.ReplaceAllString(path, "***")
}

// RedactSensitive masks sensitive data for logging.
// All sensitive values (credentials, tokens, IDs) are completely masked
// to prevent any data leakage in logs, even at DEBUG level.
func RedactSensitive(value string) string {
	if value == "" {
		return ""
	}
	return "***"
}

// RedactEmail masks email addresses while keeping the domain visible.
// This function is not currently used but provided for future flexibility.
func RedactEmail(email string) string {
	if email == "" {
		return ""
	}
	return "***@***"
}
