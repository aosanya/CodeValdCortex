package giteawebhook

import (
	"testing"

	"github.com/aosanya/CodeValdCortex/internal/infrastructure/work"
	"github.com/stretchr/testify/assert"
)

func TestNewAPIClientAdapter(t *testing.T) {
	// Create a mock client (nil for this test)
	var mockClient Client

	adapter := NewAPIClientAdapter(mockClient, "https://gitea.example.com")

	assert.NotNil(t, adapter)

	// Verify it implements the work.APIClient interface
	var _ work.APIClient = adapter
}

func TestAPIClientAdapter_ImplementsInterfaces(t *testing.T) {
	var mockClient Client
	adapter := NewAPIClientAdapter(mockClient, "https://gitea.example.com")

	// Verify it implements all sub-interfaces
	var _ work.IssueClient = adapter
	var _ work.PullRequestClient = adapter
	var _ work.MilestoneClient = adapter
	var _ work.CommentClient = adapter
	var _ work.LabelClient = adapter
	var _ work.RepositoryClient = adapter
}
