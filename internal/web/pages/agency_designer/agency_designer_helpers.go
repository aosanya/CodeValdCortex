package agency_designer

import (
	"fmt"
	"html"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/ai"
)

// formatAIMessage formats AI message content with basic markdown-like formatting
func formatAIMessage(content string) string {
	// Escape HTML
	content = html.EscapeString(content)

	// Convert markdown-style formatting to HTML
	// Bold: **text** -> <strong>text</strong>
	content = convertMarkdownBold(content)

	// Italic: *text* -> <em>text</em>
	content = convertMarkdownItalic(content)

	// Lists: - item -> <li>item</li>
	content = convertMarkdownLists(content)

	// Line breaks
	content = strings.ReplaceAll(content, "\n", "<br/>")

	return content
}

// convertMarkdownBold converts **text** to <strong>text</strong>
func convertMarkdownBold(content string) string {
	// Simple regex replacement for bold
	for strings.Contains(content, "**") {
		firstIdx := strings.Index(content, "**")
		if firstIdx == -1 {
			break
		}
		secondIdx := strings.Index(content[firstIdx+2:], "**")
		if secondIdx == -1 {
			break
		}
		secondIdx += firstIdx + 2

		before := content[:firstIdx]
		text := content[firstIdx+2 : secondIdx]
		after := content[secondIdx+2:]

		content = before + "<strong>" + text + "</strong>" + after
	}
	return content
}

// convertMarkdownItalic converts *text* to <em>text</em>
func convertMarkdownItalic(content string) string {
	// Handle single asterisks for italic (but not part of **)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		parts := strings.Split(line, "*")
		if len(parts) > 1 && len(parts)%2 == 1 {
			for j := 1; j < len(parts); j += 2 {
				// Only convert if not part of <strong> tag
				if !strings.Contains(parts[j-1], "<strong>") && !strings.Contains(parts[j+1], "</strong>") {
					parts[j] = "<em>" + parts[j] + "</em>"
				} else {
					// Put the asterisk back
					parts[j] = "*" + parts[j] + "*"
				}
			}
			lines[i] = strings.Join(parts, "")
		}
	}
	return strings.Join(lines, "\n")
}

// convertMarkdownLists converts bullet points to HTML lists
func convertMarkdownLists(content string) string {
	lines := strings.Split(content, "\n")
	var result strings.Builder
	inList := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			if !inList {
				result.WriteString("<ul>")
				inList = true
			}
			item := strings.TrimPrefix(strings.TrimPrefix(trimmed, "- "), "* ")
			result.WriteString("<li>")
			result.WriteString(item)
			result.WriteString("</li>")
		} else {
			if inList {
				result.WriteString("</ul>")
				inList = false
			}
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	if inList {
		result.WriteString("</ul>")
	}

	return result.String()
}

// getMessageEndpoint returns the appropriate API endpoint for sending messages
func getMessageEndpoint(agencyID string, conversation *ai.ConversationContext) string {
	if conversation == nil || conversation.ID == "" {
		// Start new conversation
		return fmt.Sprintf("/api/v1/agencies/%s/designer/conversations", agencyID)
	}
	// Continue existing conversation
	return fmt.Sprintf("/api/v1/conversations/%s/messages", conversation.ID)
}

// formatValue formats a value for display
func formatValue(value interface{}) string {
	return fmt.Sprintf("%v", value)
}

// getPhaseDisplay returns a human-readable phase name
func getPhaseDisplay(phase ai.DesignPhase) string {
	switch phase {
	case ai.PhaseInitial:
		return "Initial"
	case ai.PhaseRequirements:
		return "Requirements Gathering"
	case ai.PhaseAgentBrainstorm:
		return "Agent Brainstorming"
	case ai.PhaseRelationshipMapping:
		return "Relationship Mapping"
	case ai.PhaseValidation:
		return "Validation"
	case ai.PhaseComplete:
		return "Complete"
	default:
		return "Unknown"
	}
}
