package models

import "time"

// AgencyTag represents an immutable snapshot of an agency
type AgencyTag struct {
	// ArangoDB fields
	Key string `json:"_key"`
	ID  string `json:"_id"`
	Rev string `json:"_rev,omitempty"`

	// Tag metadata
	AgencyID    string  `json:"agency_id"`
	Name        string  `json:"name"`              // Unique per agency
	Version     string  `json:"version,omitempty"` // Semantic version (optional)
	Description string  `json:"description"`
	Type        TagType `json:"type"`
	SHA         string  `json:"sha"` // Content hash (git-style)

	// Complete snapshot
	Snapshot AgencySnapshot `json:"snapshot"`

	// Additional metadata
	Metadata  TagMetadata `json:"metadata"`
	CreatedAt time.Time   `json:"created_at"`
	CreatedBy string      `json:"created_by"`
}

// TagType defines the purpose of the tag
type TagType string

const (
	TagTypeRelease      TagType = "release"      // Major release version
	TagTypeSnapshot     TagType = "snapshot"     // Point-in-time save
	TagTypeExperimental TagType = "experimental" // Testing variation
	TagTypeCheckpoint   TagType = "checkpoint"   // Design milestone
)

// TagMetadata for custom fields
type TagMetadata struct {
	GitCommit    string                 `json:"git_commit,omitempty"`
	BuildNumber  string                 `json:"build_number,omitempty"`
	Environment  string                 `json:"environment,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

// ValidTagTypes returns all valid tag types
func ValidTagTypes() []TagType {
	return []TagType{
		TagTypeRelease,
		TagTypeSnapshot,
		TagTypeExperimental,
		TagTypeCheckpoint,
	}
}

// IsValidTagType checks if tag type is valid
func IsValidTagType(t TagType) bool {
	for _, valid := range ValidTagTypes() {
		if t == valid {
			return true
		}
	}
	return false
}
