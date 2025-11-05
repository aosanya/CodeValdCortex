package ai

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestParseAIResponse_ValidJSON(t *testing.T) {
	refiner := &IntroductionBuilder{
		logger: logrus.New(),
	}

	jsonResponse := `{
		"data": {
			"introduction": "This is a refined introduction that provides better context.",
			"goals": [],
			"work_items": [],
			"roles": [],
			"assignments": []
		},
		"explanation": "Added more detail about the agency's purpose and scope.",
		"changed": true,
		"changed_sections": ["introduction"]
	}`

	original := "This is the original introduction."

	refined, wasChanged, explanation, changedSections := refiner.parseAIResponse(jsonResponse, original)

	assert.Equal(t, "This is a refined introduction that provides better context.", refined)
	assert.True(t, wasChanged)
	assert.Equal(t, "Added more detail about the agency's purpose and scope.", explanation)
	assert.Equal(t, []string{"introduction"}, changedSections)
}

func TestParseAIResponse_NoChanges(t *testing.T) {
	refiner := &IntroductionBuilder{
		logger: logrus.New(),
	}

	original := "This is a good introduction."
	jsonResponse := `{
		"data": {
			"introduction": "This is a good introduction.",
			"goals": [],
			"work_items": [],
			"roles": [],
			"assignments": []
		},
		"explanation": "The introduction is already comprehensive and well-written.",
		"changed": false,
		"changed_sections": []
	}`

	refined, wasChanged, explanation, changedSections := refiner.parseAIResponse(jsonResponse, original)

	assert.Equal(t, original, refined)
	assert.False(t, wasChanged)
	assert.Equal(t, "The introduction is already comprehensive and well-written.", explanation)
	assert.Empty(t, changedSections)
}

func TestParseAIResponse_JSONWithExtraText(t *testing.T) {
	refiner := &IntroductionBuilder{
		logger: logrus.New(),
	}

	responseWithExtra := `Here's the refined introduction:

	{
		"data": {
			"introduction": "Updated introduction with better context.",
			"goals": [],
			"work_items": [],
			"roles": [],
			"assignments": []
		},
		"explanation": "Improved clarity and structure.",
		"changed": true,
		"changed_sections": ["introduction"]
	}

	I hope this helps!`

	original := "Original introduction."

	refined, wasChanged, explanation, changedSections := refiner.parseAIResponse(responseWithExtra, original)

	assert.Equal(t, "Updated introduction with better context.", refined)
	assert.True(t, wasChanged)
	assert.Equal(t, "Improved clarity and structure.", explanation)
	assert.Equal(t, []string{"introduction"}, changedSections)
}

func TestParseAIResponse_InvalidJSON(t *testing.T) {
	refiner := &IntroductionBuilder{
		logger: logrus.New(),
	}

	invalidResponse := "This is not JSON at all."
	original := "Original introduction."

	refined, wasChanged, explanation, changedSections := refiner.parseAIResponse(invalidResponse, original)

	assert.Equal(t, original, refined)
	assert.False(t, wasChanged)
	assert.Contains(t, explanation, "Could not parse AI response")
	assert.Empty(t, changedSections)
}

func TestParseAIResponse_EmptyRefinedIntroduction(t *testing.T) {
	refiner := &IntroductionBuilder{
		logger: logrus.New(),
	}

	jsonResponse := `{
		"data": {
			"introduction": "",
			"goals": [],
			"work_items": [],
			"roles": [],
			"assignments": []
		},
		"explanation": "Could not improve the introduction.",
		"changed": false,
		"changed_sections": []
	}`

	original := "Original introduction."

	refined, wasChanged, explanation, changedSections := refiner.parseAIResponse(jsonResponse, original)

	assert.Equal(t, original, refined)
	assert.False(t, wasChanged)
	assert.Contains(t, explanation, "AI returned empty introduction")
	assert.Empty(t, changedSections)
}

func TestParseAIResponse_IdenticalContent(t *testing.T) {
	refiner := &IntroductionBuilder{
		logger: logrus.New(),
	}

	original := "This is the introduction."
	jsonResponse := `{
		"data": {
			"introduction": "This is the introduction.",
			"goals": [],
			"work_items": [],
			"roles": [],
			"assignments": []
		},
		"explanation": "No changes needed.",
		"changed": true,
		"changed_sections": ["introduction"]
	}`

	refined, wasChanged, explanation, changedSections := refiner.parseAIResponse(jsonResponse, original)

	assert.Equal(t, original, refined)
	assert.False(t, wasChanged) // Should detect that content is identical
	assert.Equal(t, "No changes needed.", explanation)
	assert.Empty(t, changedSections) // Should be empty when content is identical
}
