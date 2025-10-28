package ai

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestParseAIResponse_ValidJSON(t *testing.T) {
	refiner := &IntroductionRefiner{
		logger: logrus.New(),
	}

	jsonResponse := `{
		"refined_introduction": "This is a refined introduction that provides better context.",
		"explanation": "Added more detail about the agency's purpose and scope.",
		"changed": true
	}`

	original := "This is the original introduction."

	refined, wasChanged, explanation := refiner.parseAIResponse(jsonResponse, original)

	assert.Equal(t, "This is a refined introduction that provides better context.", refined)
	assert.True(t, wasChanged)
	assert.Equal(t, "Added more detail about the agency's purpose and scope.", explanation)
}

func TestParseAIResponse_NoChanges(t *testing.T) {
	refiner := &IntroductionRefiner{
		logger: logrus.New(),
	}

	original := "This is a good introduction."
	jsonResponse := `{
		"refined_introduction": "This is a good introduction.",
		"explanation": "The introduction is already comprehensive and well-written.",
		"changed": false
	}`

	refined, wasChanged, explanation := refiner.parseAIResponse(jsonResponse, original)

	assert.Equal(t, original, refined)
	assert.False(t, wasChanged)
	assert.Equal(t, "The introduction is already comprehensive and well-written.", explanation)
}

func TestParseAIResponse_JSONWithExtraText(t *testing.T) {
	refiner := &IntroductionRefiner{
		logger: logrus.New(),
	}

	responseWithExtra := `Here's the refined introduction:

	{
		"refined_introduction": "Updated introduction with better context.",
		"explanation": "Improved clarity and structure.",
		"changed": true
	}

	I hope this helps!`

	original := "Original introduction."

	refined, wasChanged, explanation := refiner.parseAIResponse(responseWithExtra, original)

	assert.Equal(t, "Updated introduction with better context.", refined)
	assert.True(t, wasChanged)
	assert.Equal(t, "Improved clarity and structure.", explanation)
}

func TestParseAIResponse_InvalidJSON(t *testing.T) {
	refiner := &IntroductionRefiner{
		logger: logrus.New(),
	}

	invalidResponse := "This is not JSON at all."
	original := "Original introduction."

	refined, wasChanged, explanation := refiner.parseAIResponse(invalidResponse, original)

	assert.Equal(t, original, refined)
	assert.False(t, wasChanged)
	assert.Contains(t, explanation, "Could not parse AI response")
}

func TestParseAIResponse_EmptyRefinedIntroduction(t *testing.T) {
	refiner := &IntroductionRefiner{
		logger: logrus.New(),
	}

	jsonResponse := `{
		"refined_introduction": "",
		"explanation": "Could not improve the introduction.",
		"changed": false
	}`

	original := "Original introduction."

	refined, wasChanged, explanation := refiner.parseAIResponse(jsonResponse, original)

	assert.Equal(t, original, refined)
	assert.False(t, wasChanged)
	assert.Contains(t, explanation, "AI returned empty introduction")
}

func TestParseAIResponse_IdenticalContent(t *testing.T) {
	refiner := &IntroductionRefiner{
		logger: logrus.New(),
	}

	original := "This is the introduction."
	jsonResponse := `{
		"refined_introduction": "This is the introduction.",
		"explanation": "No changes needed.",
		"changed": true
	}`

	refined, wasChanged, explanation := refiner.parseAIResponse(jsonResponse, original)

	assert.Equal(t, original, refined)
	assert.False(t, wasChanged) // Should detect that content is identical
	assert.Equal(t, "No changes needed.", explanation)
}
