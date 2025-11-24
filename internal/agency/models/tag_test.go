package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestIsValidTagType(t *testing.T) {
	tests := []struct {
		name     string
		tagType  TagType
		expected bool
	}{
		{"Valid: Release", TagTypeRelease, true},
		{"Valid: Snapshot", TagTypeSnapshot, true},
		{"Valid: Experimental", TagTypeExperimental, true},
		{"Valid: Checkpoint", TagTypeCheckpoint, true},
		{"Invalid: Empty", TagType(""), false},
		{"Invalid: Unknown", TagType("unknown"), false},
		{"Invalid: Invalid", TagType("invalid_type"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidTagType(tt.tagType)
			if result != tt.expected {
				t.Errorf("IsValidTagType(%s) = %v, want %v", tt.tagType, result, tt.expected)
			}
		})
	}
}

func TestValidTagTypes(t *testing.T) {
	types := ValidTagTypes()

	if len(types) != 4 {
		t.Errorf("Expected 4 tag types, got %d", len(types))
	}

	expectedTypes := map[TagType]bool{
		TagTypeRelease:      true,
		TagTypeSnapshot:     true,
		TagTypeExperimental: true,
		TagTypeCheckpoint:   true,
	}

	for _, tagType := range types {
		if !expectedTypes[tagType] {
			t.Errorf("Unexpected tag type: %s", tagType)
		}
	}
}

func TestAgencyTagJSONMarshaling(t *testing.T) {
	tag := AgencyTag{
		Key:         "test_agency_v1_0_0",
		ID:          "agency_tags/test_agency_v1_0_0",
		Rev:         "_abc123",
		AgencyID:    "agency_123",
		Name:        "v1.0.0",
		Version:     "1.0.0",
		Description: "Initial release",
		Type:        TagTypeRelease,
		SHA:         "abc123def456",
		Metadata: TagMetadata{
			GitCommit:   "main-abc123",
			BuildNumber: "42",
			Environment: "production",
		},
		CreatedAt: time.Now(),
		CreatedBy: "test-user",
	}

	// Marshal to JSON
	data, err := json.Marshal(tag)
	if err != nil {
		t.Fatalf("Failed to marshal tag: %v", err)
	}

	// Unmarshal back
	var unmarshaled AgencyTag
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal tag: %v", err)
	}

	// Verify fields
	if unmarshaled.Key != tag.Key {
		t.Errorf("Key mismatch: got %s, want %s", unmarshaled.Key, tag.Key)
	}
	if unmarshaled.AgencyID != tag.AgencyID {
		t.Errorf("AgencyID mismatch: got %s, want %s", unmarshaled.AgencyID, tag.AgencyID)
	}
	if unmarshaled.Name != tag.Name {
		t.Errorf("Name mismatch: got %s, want %s", unmarshaled.Name, tag.Name)
	}
	if unmarshaled.Type != tag.Type {
		t.Errorf("Type mismatch: got %s, want %s", unmarshaled.Type, tag.Type)
	}
	if unmarshaled.Metadata.GitCommit != tag.Metadata.GitCommit {
		t.Errorf("Metadata.GitCommit mismatch: got %s, want %s",
			unmarshaled.Metadata.GitCommit, tag.Metadata.GitCommit)
	}
}

func TestTagTypeJSONMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		tagType  TagType
		expected string
	}{
		{"Release", TagTypeRelease, `"release"`},
		{"Snapshot", TagTypeSnapshot, `"snapshot"`},
		{"Experimental", TagTypeExperimental, `"experimental"`},
		{"Checkpoint", TagTypeCheckpoint, `"checkpoint"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.tagType)
			if err != nil {
				t.Fatalf("Failed to marshal TagType: %v", err)
			}

			if string(data) != tt.expected {
				t.Errorf("JSON mismatch: got %s, want %s", string(data), tt.expected)
			}

			// Unmarshal
			var unmarshaled TagType
			err = json.Unmarshal([]byte(tt.expected), &unmarshaled)
			if err != nil {
				t.Fatalf("Failed to unmarshal TagType: %v", err)
			}

			if unmarshaled != tt.tagType {
				t.Errorf("TagType mismatch: got %s, want %s", unmarshaled, tt.tagType)
			}
		})
	}
}

func TestTagMetadataJSONMarshaling(t *testing.T) {
	metadata := TagMetadata{
		GitCommit:   "abc123",
		BuildNumber: "42",
		Environment: "staging",
		CustomFields: map[string]interface{}{
			"deployment_id": "dep-123",
			"region":        "us-west-2",
		},
	}

	// Marshal
	data, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("Failed to marshal TagMetadata: %v", err)
	}

	// Unmarshal
	var unmarshaled TagMetadata
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal TagMetadata: %v", err)
	}

	// Verify
	if unmarshaled.GitCommit != metadata.GitCommit {
		t.Errorf("GitCommit mismatch: got %s, want %s",
			unmarshaled.GitCommit, metadata.GitCommit)
	}
	if unmarshaled.BuildNumber != metadata.BuildNumber {
		t.Errorf("BuildNumber mismatch: got %s, want %s",
			unmarshaled.BuildNumber, metadata.BuildNumber)
	}
	if unmarshaled.CustomFields["deployment_id"] != metadata.CustomFields["deployment_id"] {
		t.Error("CustomFields mismatch")
	}
}
