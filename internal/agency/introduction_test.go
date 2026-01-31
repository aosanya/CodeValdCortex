package agency_test

import (
	"testing"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/stretchr/testify/assert"
)

func TestAgencyIntroduction_Validation(t *testing.T) {
	tests := []struct {
		name      string
		intro     *agency.AgencyIntroduction
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid introduction with genesis template",
			intro: &agency.AgencyIntroduction{
				ID:        "intro/123",
				AgencyID:  "agency/456",
				Template:  "genesis",
				Sections:  []agency.IntroductionSection{},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantError: false,
		},
		{
			name: "invalid - missing agency ID",
			intro: &agency.AgencyIntroduction{
				ID:       "intro/123",
				Template: "genesis",
				Sections: []agency.IntroductionSection{},
			},
			wantError: true,
			errorMsg:  "AgencyID is required",
		},
		{
			name: "invalid - unknown template",
			intro: &agency.AgencyIntroduction{
				ID:       "intro/123",
				AgencyID: "agency/456",
				Template: "unknown",
				Sections: []agency.IntroductionSection{},
			},
			wantError: true,
			errorMsg:  "Invalid template",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.intro.Validate()
			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIntroductionSection_Validation(t *testing.T) {
	tests := []struct {
		name      string
		section   *agency.IntroductionSection
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid text section",
			section: &agency.IntroductionSection{
				ID:    "section/1",
				Type:  agency.SectionTypeText,
				Title: "Overview",
				Content: map[string]interface{}{
					"text": "This is our overview",
				},
				Order: 1,
			},
			wantError: false,
		},
		{
			name: "valid list section with items",
			section: &agency.IntroductionSection{
				ID:    "section/2",
				Type:  agency.SectionTypeList,
				Title: "Features",
				Content: map[string]interface{}{
					"items":      []interface{}{"Feature 1", "Feature 2"},
					"isNumbered": false,
				},
				Order: 2,
			},
			wantError: false,
		},
		{
			name: "invalid - missing title",
			section: &agency.IntroductionSection{
				ID:      "section/3",
				Type:    agency.SectionTypeText,
				Content: map[string]interface{}{},
				Order:   1,
			},
			wantError: true,
			errorMsg:  "Title is required",
		},
		{
			name: "invalid - unknown section type",
			section: &agency.IntroductionSection{
				ID:      "section/4",
				Type:    "unknown",
				Title:   "Test",
				Content: map[string]interface{}{},
				Order:   1,
			},
			wantError: true,
			errorMsg:  "Invalid section type",
		},
		{
			name: "invalid - text section missing content",
			section: &agency.IntroductionSection{
				ID:      "section/5",
				Type:    agency.SectionTypeText,
				Title:   "Empty",
				Content: map[string]interface{}{},
				Order:   1,
			},
			wantError: true,
			errorMsg:  "text is required",
		},
		{
			name: "invalid - list section missing items",
			section: &agency.IntroductionSection{
				ID:      "section/6",
				Type:    agency.SectionTypeList,
				Title:   "Empty List",
				Content: map[string]interface{}{},
				Order:   1,
			},
			wantError: true,
			errorMsg:  "items are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.section.Validate()
			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAgencyIntroduction_IsComplete(t *testing.T) {
	tests := []struct {
		name     string
		intro    *agency.AgencyIntroduction
		expected bool
	}{
		{
			name: "complete with all required sections",
			intro: &agency.AgencyIntroduction{
				Sections: []agency.IntroductionSection{
					{
						ID:         "section/1",
						Type:       agency.SectionTypeText,
						Title:      "Required Section",
						IsRequired: true,
						Content:    map[string]interface{}{"text": "Content"},
					},
					{
						ID:         "section/2",
						Type:       agency.SectionTypeText,
						Title:      "Optional Section",
						IsRequired: false,
						Content:    map[string]interface{}{"text": "Content"},
					},
				},
			},
			expected: true,
		},
		{
			name: "incomplete - missing required section content",
			intro: &agency.AgencyIntroduction{
				Sections: []agency.IntroductionSection{
					{
						ID:         "section/1",
						Type:       agency.SectionTypeText,
						Title:      "Required Section",
						IsRequired: true,
						Content:    map[string]interface{}{},
					},
				},
			},
			expected: false,
		},
		{
			name: "complete - no required sections",
			intro: &agency.AgencyIntroduction{
				Sections: []agency.IntroductionSection{
					{
						ID:         "section/1",
						Type:       agency.SectionTypeText,
						Title:      "Optional Section",
						IsRequired: false,
						Content:    map[string]interface{}{},
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.intro.IsComplete()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAgencyIntroduction_OrderedSections(t *testing.T) {
	intro := &agency.AgencyIntroduction{
		Sections: []agency.IntroductionSection{
			{ID: "section/3", Title: "Third", Order: 3},
			{ID: "section/1", Title: "First", Order: 1},
			{ID: "section/2", Title: "Second", Order: 2},
		},
	}

	ordered := intro.OrderedSections()

	assert.Len(t, ordered, 3)
	assert.Equal(t, "First", ordered[0].Title)
	assert.Equal(t, "Second", ordered[1].Title)
	assert.Equal(t, "Third", ordered[2].Title)
	assert.Equal(t, 1, ordered[0].Order)
	assert.Equal(t, 2, ordered[1].Order)
	assert.Equal(t, 3, ordered[2].Order)
}

func TestIntroductionSection_HasContent(t *testing.T) {
	tests := []struct {
		name     string
		section  agency.IntroductionSection
		expected bool
	}{
		{
			name: "text section with content",
			section: agency.IntroductionSection{
				Type:    agency.SectionTypeText,
				Content: map[string]interface{}{"text": "Some content"},
			},
			expected: true,
		},
		{
			name: "text section without content",
			section: agency.IntroductionSection{
				Type:    agency.SectionTypeText,
				Content: map[string]interface{}{"text": ""},
			},
			expected: false,
		},
		{
			name: "list section with items",
			section: agency.IntroductionSection{
				Type:    agency.SectionTypeList,
				Content: map[string]interface{}{"items": []interface{}{"Item 1", "Item 2"}},
			},
			expected: true,
		},
		{
			name: "list section without items",
			section: agency.IntroductionSection{
				Type:    agency.SectionTypeList,
				Content: map[string]interface{}{"items": []interface{}{}},
			},
			expected: false,
		},
		{
			name: "nested section with subsections",
			section: agency.IntroductionSection{
				Type: agency.SectionTypeNested,
				Content: map[string]interface{}{
					"subsections": []interface{}{
						map[string]interface{}{"title": "Sub 1"},
					},
				},
			},
			expected: true,
		},
		{
			name: "table section with rows",
			section: agency.IntroductionSection{
				Type: agency.SectionTypeTable,
				Content: map[string]interface{}{
					"rows": []interface{}{
						[]interface{}{"Cell 1", "Cell 2"},
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.section.HasContent()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetTemplateByName(t *testing.T) {
	tests := []struct {
		name         string
		templateName string
		expectError  bool
	}{
		{
			name:         "genesis template",
			templateName: "genesis",
			expectError:  false,
		},
		{
			name:         "minimal template",
			templateName: "minimal",
			expectError:  false,
		},
		{
			name:         "custom template",
			templateName: "custom",
			expectError:  false,
		},
		{
			name:         "unknown template",
			templateName: "unknown",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, err := agency.GetTemplateByName(tt.templateName)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, template)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, template)
				assert.Equal(t, tt.templateName, template.Name)

				// Verify template structure
				if tt.templateName == "genesis" {
					assert.Equal(t, 6, len(template.Sections))
				} else if tt.templateName == "minimal" {
					assert.Equal(t, 3, len(template.Sections))
				} else if tt.templateName == "custom" {
					assert.Equal(t, 0, len(template.Sections))
				}
			}
		})
	}
}
