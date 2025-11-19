package work

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestDefaultTemplateRenderer_Render(t *testing.T) {
	renderer := NewTemplateRenderer()
	testTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	ctx := context.Background()

	tests := []struct {
		name         string
		templateName string
		data         CommentTemplateData
		wantContains []string
		wantErr      bool
	}{
		{
			name:         "agent_started template",
			templateName: "agent_started",
			data: CommentTemplateData{
				AgentID:       "agent-123",
				AgentType:     "TestAgent",
				ColumnName:    "In Progress",
				Timestamp:     testTime,
				FormattedTime: "2024-01-01 10:00:00",
			},
			wantContains: []string{"🤖", "Agent Assigned", "agent-123", "TestAgent", "In Progress"},
			wantErr:      false,
		},
		{
			name:         "agent_healthy template",
			templateName: "agent_healthy",
			data: CommentTemplateData{
				AgentID:       "agent-123",
				AgentType:     "TestAgent",
				Timestamp:     testTime,
				FormattedTime: "2024-01-01 10:00:00",
			},
			wantContains: []string{"✅", "Agent Running", "agent-123"},
			wantErr:      false,
		},
		{
			name:         "agent_degraded template with details",
			templateName: "agent_degraded",
			data: CommentTemplateData{
				AgentID:       "agent-123",
				AgentType:     "TestAgent",
				StatusMessage: "High memory usage detected",
				Timestamp:     testTime,
				FormattedTime: "2024-01-01 10:00:00",
			},
			wantContains: []string{"⚠️", "Agent Degraded", "High memory usage detected"},
			wantErr:      false,
		},
		{
			name:         "agent_quarantined template with error",
			templateName: "agent_quarantined",
			data: CommentTemplateData{
				AgentID:       "agent-123",
				AgentType:     "TestAgent",
				ErrorMessage:  "Critical failure detected",
				Timestamp:     testTime,
				FormattedTime: "2024-01-01 10:00:00",
			},
			wantContains: []string{"Agent Quarantined", "Critical failure detected"},
			wantErr:      false,
		},
		{
			name:         "agent_stopped template",
			templateName: "agent_stopped",
			data: CommentTemplateData{
				AgentID:       "agent-123",
				AgentType:     "TestAgent",
				Timestamp:     testTime,
				FormattedTime: "2024-01-01 10:00:00",
			},
			wantContains: []string{"⏹️", "Agent Stopped", "agent-123"},
			wantErr:      false,
		},
		{
			name:         "task_running template with task details",
			templateName: "task_running",
			data: CommentTemplateData{
				AgentID:       "agent-123",
				TaskName:      "DataProcessing",
				Timestamp:     testTime,
				FormattedTime: "2024-01-01 10:00:00",
			},
			wantContains: []string{"▶️", "Executing Task", "DataProcessing"},
			wantErr:      false,
		},
		{
			name:         "task_completed template with summary",
			templateName: "task_completed",
			data: CommentTemplateData{
				AgentID:       "agent-123",
				TaskName:      "DataProcessing",
				TaskSummary:   "Processed 1000 records successfully",
				Timestamp:     testTime,
				FormattedTime: "2024-01-01 10:00:00",
			},
			wantContains: []string{"✅", "Task Completed", "Processed 1000 records"},
			wantErr:      false,
		},
		{
			name:         "task_failed template with error",
			templateName: "task_failed",
			data: CommentTemplateData{
				AgentID:       "agent-123",
				TaskName:      "DataProcessing",
				ErrorMessage:  "Database connection timeout",
				Timestamp:     testTime,
				FormattedTime: "2024-01-01 10:00:00",
			},
			wantContains: []string{"❌", "Task Failed", "Database connection timeout"},
			wantErr:      false,
		},
		{
			name:         "waiting_io template",
			templateName: "waiting_io",
			data: CommentTemplateData{
				AgentID:       "agent-123",
				StatusMessage: "Waiting for external API response",
				Timestamp:     testTime,
				FormattedTime: "2024-01-01 10:00:00",
			},
			wantContains: []string{"⏳", "Waiting for I/O", "Waiting for external API response"},
			wantErr:      false,
		},
		{
			name:         "waiting_hitl template",
			templateName: "waiting_hitl",
			data: CommentTemplateData{
				AgentID:       "agent-123",
				StatusMessage: "Approval required for deployment",
				Timestamp:     testTime,
				FormattedTime: "2024-01-01 10:00:00",
			},
			wantContains: []string{"👤", "Human Approval Needed", "Approval required for deployment"},
			wantErr:      false,
		},
		{
			name:         "progress_update template with percentage",
			templateName: "progress_update",
			data: CommentTemplateData{
				AgentID:            "agent-123",
				ProgressPercentage: 75,
				StatusMessage:      "Data validation in progress",
				Timestamp:          testTime,
				FormattedTime:      "2024-01-01 10:00:00",
			},
			wantContains: []string{"📊", "Progress Update", "75%", "Data validation in progress"},
			wantErr:      false,
		},
		{
			name:         "milestone_completed template with summary",
			templateName: "milestone_completed",
			data: CommentTemplateData{
				AgentID:       "agent-123",
				WorkSummary:   "All integration tests passed",
				CurrentColumn: "In Progress",
				NextColumn:    "Testing",
				Timestamp:     testTime,
				FormattedTime: "2024-01-01 10:00:00",
			},
			wantContains: []string{"Milestone Completed", "All integration tests passed"},
			wantErr:      false,
		},
		{
			name:         "error_alert template with error",
			templateName: "error_alert",
			data: CommentTemplateData{
				AgentID:             "agent-123",
				ErrorMessage:        "Unexpected null pointer exception",
				DetailedDescription: "Stack trace available in logs",
				Timestamp:           testTime,
				FormattedTime:       "2024-01-01 10:00:00",
			},
			wantContains: []string{"❌", "Agent Error", "Unexpected null pointer exception"},
			wantErr:      false,
		},
		{
			name:         "unknown template",
			templateName: "nonexistent_template",
			data: CommentTemplateData{
				AgentID:       "agent-123",
				Timestamp:     testTime,
				FormattedTime: "2024-01-01 10:00:00",
			},
			wantContains: nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderer.Render(ctx, tt.templateName, &tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				for _, want := range tt.wantContains {
					if !strings.Contains(got, want) {
						t.Errorf("Render() output missing expected string %q\nGot:\n%s", want, got)
					}
				}
			}
		})
	}
}

func TestDefaultTemplateRenderer_RegisterTemplate(t *testing.T) {
	renderer := NewTemplateRenderer()
	ctx := context.Background()

	templateContent := "Test template: {{.AgentID}}"
	err := renderer.RegisterTemplate("custom_test", templateContent)
	if err != nil {
		t.Fatalf("RegisterTemplate() error = %v", err)
	}

	// Test rendering the custom template
	data := CommentTemplateData{
		AgentID:       "agent-999",
		Timestamp:     time.Now(),
		FormattedTime: "2024-01-01 10:00:00",
	}
	got, err := renderer.Render(ctx, "custom_test", &data)
	if err != nil {
		t.Errorf("Render() error = %v", err)
	}
	if !strings.Contains(got, "agent-999") {
		t.Errorf("Custom template not rendered correctly. Got: %s", got)
	}
}

func TestDefaultTemplateRenderer_RegisterTemplate_InvalidTemplate(t *testing.T) {
	renderer := NewTemplateRenderer()

	// Invalid template syntax
	invalidTemplate := "Test template: {{.AgentID"
	err := renderer.RegisterTemplate("invalid_test", invalidTemplate)
	if err == nil {
		t.Error("RegisterTemplate() expected error for invalid template, got nil")
	}
}

func TestDefaultTemplateRenderer_GetTemplate(t *testing.T) {
	renderer := NewTemplateRenderer()

	tests := []struct {
		name         string
		templateName string
		wantOk       bool
	}{
		{
			name:         "existing template - agent_started",
			templateName: "agent_started",
			wantOk:       true,
		},
		{
			name:         "existing template - task_completed",
			templateName: "task_completed",
			wantOk:       true,
		},
		{
			name:         "nonexistent template",
			templateName: "does_not_exist",
			wantOk:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := renderer.GetTemplate(tt.templateName)
			wantErr := !tt.wantOk
			if (err != nil) != wantErr {
				t.Errorf("GetTemplate() error = %v, wantErr %v", err, wantErr)
			}
			if tt.wantOk && tmpl == "" {
				t.Error("GetTemplate() returned empty template for existing template")
			}
		})
	}
}

func TestDefaultTemplateRenderer_ThreadSafety(t *testing.T) {
	renderer := NewTemplateRenderer()
	testTime := time.Now()
	ctx := context.Background()

	// Test concurrent rendering
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			data := CommentTemplateData{
				AgentID:       "agent-concurrent",
				Timestamp:     testTime,
				FormattedTime: "2024-01-01 10:00:00",
			}
			_, err := renderer.Render(ctx, "agent_started", &data)
			if err != nil {
				t.Errorf("Concurrent Render() error = %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
