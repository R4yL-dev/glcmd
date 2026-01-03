package logger

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
