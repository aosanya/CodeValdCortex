package work

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"text/template"
)

// DefaultTemplateRenderer implements TemplateRenderer using Go's text/template
type DefaultTemplateRenderer struct {
	templates map[string]*template.Template
	mu        sync.RWMutex
}

// NewTemplateRenderer creates a new template renderer with default templates
func NewTemplateRenderer() TemplateRenderer {
	renderer := &DefaultTemplateRenderer{
		templates: make(map[string]*template.Template),
	}

	// Register default templates
	renderer.registerDefaultTemplates()

	return renderer
}

// registerDefaultTemplates registers all built-in comment templates
func (r *DefaultTemplateRenderer) registerDefaultTemplates() {
	// Agent Started Template
	r.RegisterTemplate("agent_started", `🤖 **Agent Assigned**

- **Stage**: {{.ColumnName}}
- **Agent Type**: {{.AgentType}}
- **Agent ID**: `+"`{{.AgentID}}`"+`
- **Started**: {{.FormattedTime}}

The agent will now work on this issue. Updates will be posted here.`)

	// Progress Update Template
	r.RegisterTemplate("progress_update", `📊 **Progress Update**

**Agent**: `+"`{{.AgentID}}`"+`  
**Status**: {{.StatusMessage}}  
{{- if gt .ProgressPercentage 0}}
**Progress**: {{.ProgressPercentage}}%  
{{- end}}
**Updated**: {{.FormattedTime}}

{{.DetailedDescription}}`)

	// Task Completed Template
	r.RegisterTemplate("task_completed", `✅ **Task Completed**

**Agent**: `+"`{{.AgentID}}`"+`  
**Task**: {{.TaskName}}  
{{- if .TaskDuration}}
**Duration**: {{.TaskDuration}}  
{{- end}}
**Completed**: {{.FormattedTime}}

**Summary**:
{{.TaskSummary}}

{{- if .Deliverables}}

**Deliverables**:
{{.Deliverables}}
{{- end}}`)

	// Error Alert Template
	r.RegisterTemplate("error_alert", `❌ **Agent Error**

**Agent**: `+"`{{.AgentID}}`"+`  
**Error Type**: {{.ErrorType}}  
**Severity**: {{.Severity}}  
**Occurred**: {{.FormattedTime}}

**Error Details**:
`+"```"+`
{{.ErrorMessage}}
`+"```"+`

{{- if .StackTrace}}

**Stack Trace**:
`+"```"+`
{{.StackTrace}}
`+"```"+`
{{- end}}

{{- if .RemediationGuide}}

**Next Steps**: {{.RemediationGuide}}
{{- end}}`)

	// Milestone Completed Template
	r.RegisterTemplate("milestone_completed", `🎯 **Milestone Completed**

**Agent**: `+"`{{.AgentID}}`"+`  
**Workflow Stage**: {{.CurrentColumn}} → {{.NextColumn}}  
**Completed**: {{.FormattedTime}}

This issue is now moving to the next workflow stage: **{{.NextColumn}}**

{{.WorkSummary}}`)

	// Agent Healthy Template
	r.RegisterTemplate("agent_healthy", `✅ **Agent Running**

**Agent**: `+"`{{.AgentID}}`"+`  
**Status**: Healthy and operational  
**Updated**: {{.FormattedTime}}

The agent is now actively working on this issue.`)

	// Agent Degraded Template
	r.RegisterTemplate("agent_degraded", `⚠️ **Agent Degraded**

**Agent**: `+"`{{.AgentID}}`"+`  
**Status**: Experiencing issues but operational  
**Updated**: {{.FormattedTime}}

**Reason**: {{.StatusMessage}}

The agent is still working but may be slower than expected.`)

	// Agent Quarantined Template
	r.RegisterTemplate("agent_quarantined", `🔒 **Agent Quarantined**

**Agent**: `+"`{{.AgentID}}`"+`  
**Status**: Isolated due to policy violation  
**Occurred**: {{.FormattedTime}}

**Violation**: {{.ErrorMessage}}

The agent has been quarantined and requires manual review before it can resume work.`)

	// Agent Stopped Template
	r.RegisterTemplate("agent_stopped", `⏹️ **Agent Stopped**

**Agent**: `+"`{{.AgentID}}`"+`  
**Status**: Shutdown gracefully  
**Stopped**: {{.FormattedTime}}

{{- if .StatusMessage}}

**Reason**: {{.StatusMessage}}
{{- end}}`)

	// Task Running Template
	r.RegisterTemplate("task_running", `▶️ **Executing Task**

**Agent**: `+"`{{.AgentID}}`"+`  
**Task**: {{.TaskName}}  
**Started**: {{.FormattedTime}}

The agent is now executing this task.`)

	// Waiting I/O Template
	r.RegisterTemplate("waiting_io", `⏳ **Waiting for I/O**

**Agent**: `+"`{{.AgentID}}`"+`  
**Reason**: {{.StatusMessage}}  
**Updated**: {{.FormattedTime}}

The agent is waiting for an external I/O operation to complete.`)

	// Waiting HITL Template
	r.RegisterTemplate("waiting_hitl", `👤 **Human Approval Needed**

**Agent**: `+"`{{.AgentID}}`"+`  
**Context**: {{.StatusMessage}}  
**Requested**: {{.FormattedTime}}

The agent requires human approval to proceed. Please review and approve or reject.`)

	// Task Failed Template
	r.RegisterTemplate("task_failed", `❌ **Task Failed**

**Agent**: `+"`{{.AgentID}}`"+`  
**Task**: {{.TaskName}}  
**Failed**: {{.FormattedTime}}

**Error**:
`+"```"+`
{{.ErrorMessage}}
`+"```"+`

{{- if .RemediationGuide}}

**Recovery**: {{.RemediationGuide}}
{{- end}}`)
}

// Render renders a template with the provided data
func (r *DefaultTemplateRenderer) Render(ctx context.Context, templateName string, data *CommentTemplateData) (string, error) {
	r.mu.RLock()
	tmpl, exists := r.templates[templateName]
	r.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("template not found: %s", templateName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}

	return buf.String(), nil
}

// RegisterTemplate registers a new template or overwrites an existing one
func (r *DefaultTemplateRenderer) RegisterTemplate(name string, templateStr string) error {
	tmpl, err := template.New(name).Parse(templateStr)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	r.mu.Lock()
	r.templates[name] = tmpl
	r.mu.Unlock()

	return nil
}

// GetTemplate retrieves a template by name
func (r *DefaultTemplateRenderer) GetTemplate(name string) (string, error) {
	r.mu.RLock()
	_, exists := r.templates[name]
	r.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("template not found: %s", name)
	}

	// Return the raw template definition
	// Note: Go's text/template doesn't provide direct access to the original string,
	// so this is a simplified implementation that returns the name
	return name, nil
}
