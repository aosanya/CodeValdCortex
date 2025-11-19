package giteawebhook

import (
"testing"
"time"

"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
tests := []struct {
name        string
config      ClientConfig
expectError bool
errorMsg    string
}{
{
name: "missing base URL",
config: ClientConfig{
Token: "test-token",
},
expectError: true,
errorMsg:    "base URL is required",
},
{
name: "missing token",
config: ClientConfig{
BaseURL: "https://gitea.example.com",
},
expectError: true,
errorMsg:    "API token is required",
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
client, err := NewClient(tt.config)

if tt.expectError {
assert.Error(t, err)
assert.Contains(t, err.Error(), tt.errorMsg)
assert.Nil(t, client)
}
})
}
}

// Note: Tests requiring actual Gitea client instances are deferred to integration tests
// because the Gitea SDK validates server connectivity on client creation.

func TestCreateIssueOptions(t *testing.T) {
opts := CreateIssueOptions{
Title:     "Test Issue",
Body:      "Test body",
Assignees: []string{"user1", "user2"},
Milestone: 1,
Labels:    []int64{10, 20},
Closed:    false,
}

assert.Equal(t, "Test Issue", opts.Title)
assert.Equal(t, "Test body", opts.Body)
assert.Len(t, opts.Assignees, 2)
assert.Equal(t, int64(1), opts.Milestone)
assert.Len(t, opts.Labels, 2)
assert.False(t, opts.Closed)
}

func TestUpdateIssueOptions(t *testing.T) {
title := "Updated Title"
body := "Updated body"
milestone := int64(2)
state := "closed"

opts := UpdateIssueOptions{
Title:     &title,
Body:      &body,
Assignees: []string{"user3"},
Milestone: &milestone,
State:     &state,
}

assert.NotNil(t, opts.Title)
assert.Equal(t, "Updated Title", *opts.Title)
assert.NotNil(t, opts.Body)
assert.Equal(t, "Updated body", *opts.Body)
assert.NotNil(t, opts.Milestone)
assert.Equal(t, int64(2), *opts.Milestone)
assert.NotNil(t, opts.State)
assert.Equal(t, "closed", *opts.State)
}

func TestListIssueOptions(t *testing.T) {
since := time.Now()
opts := ListIssueOptions{
State:     "open",
Labels:    []string{"bug", "urgent"},
Milestone: "v1.0",
Since:     &since,
Page:      1,
Limit:     50,
}

assert.Equal(t, "open", opts.State)
assert.Len(t, opts.Labels, 2)
assert.Equal(t, "v1.0", opts.Milestone)
assert.NotNil(t, opts.Since)
assert.Equal(t, 1, opts.Page)
assert.Equal(t, 50, opts.Limit)
}

func TestMergePullRequestOptions(t *testing.T) {
opts := MergePullRequestOptions{
Style:   "squash",
Message: "Squash and merge",
}

assert.Equal(t, "squash", opts.Style)
assert.Equal(t, "Squash and merge", opts.Message)
}

func TestCreateMilestoneOptions(t *testing.T) {
dueDate := time.Now().Add(30 * 24 * time.Hour)
opts := CreateMilestoneOptions{
Title:       "v1.0",
Description: "First release",
DueDate:     &dueDate,
State:       "open",
}

assert.Equal(t, "v1.0", opts.Title)
assert.Equal(t, "First release", opts.Description)
assert.NotNil(t, opts.DueDate)
assert.Equal(t, "open", opts.State)
}
