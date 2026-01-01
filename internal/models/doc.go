// Package models contains data structures for LibreView API entities.
//
// These structures represent the various data models discovered through
// API exploration (see exploration/COMPLETE_API_DOCUMENTATION.md).
//
// Structures:
//   - SensorConfig: Glucose sensor information and metadata
//   - UserPreferences: User account and display preferences
//   - DeviceInfo: Patient device configuration and thresholds
//   - GlucoseTargets: Global glucose target thresholds for statistics
//
// Note: GlucoseMeasurement is defined in internal/glucosemeasurement package.
//
// These models are used by the Storage interface to persist data from
// the LibreView API for historical tracking and statistics.
package models
