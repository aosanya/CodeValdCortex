package ai

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/builder"
)

// SharedAgencyContext provides critical context about the multi-agent system architecture
// that should be included at the beginning of all AI system prompts for consistency.
const SharedAgencyContext = `CRITICAL CONTEXT: This system builds MULTI-AGENT AGENCIES where autonomous agents (both AI and human workers) perform tasks.
Agents operate at different autonomy levels (L0-L4):
- L0 (Manual): Agent provides recommendations only; human executes all actions
- L1 (Assisted): Agent performs routine actions; human approves high-risk actions  
- L2 (Conditional): Agent operates autonomously under defined constraints
- L3 (High Automation): Agent handles most scenarios independently; human on-call for edge cases
- L4 (Full Autonomy): Agent operates completely independently; human notified post-facto

When designing agency components (goals, work items, roles, workflows):
- Goals are HIGH-LEVEL OBJECTIVES describing what agents will accomplish (realized through multiple work items)
- Work Items are the ACTUAL TO-DOs that agents are tasked with performing
- Roles define AGENT CAPABILITIES and autonomy levels
- RACI defines WHO (which role/agent) does WHAT for each work item
`

// System prompts for different conversation phases

// getSystemPrompt returns the system prompt for a given phase
func (s *AgencyDesignerService) getSystemPrompt(phase DesignPhase) string {
	switch phase {
	case PhaseInitial, PhaseRequirements:
		return requirementsSystemPrompt
	case PhaseAgentBrainstorm:
		return brainstormSystemPrompt
	case PhaseRelationshipMapping:
		return relationshipSystemPrompt
	case PhaseValidation:
		return validationSystemPrompt
	default:
		return requirementsSystemPrompt
	}
}

// getDesignGenerationPrompt creates the prompt for final design generation
func (s *AgencyDesignerService) getDesignGenerationPrompt(conversation *ConversationContext) string {
	var contextInfo string

	// Extract context information from conversation state
	if conversation != nil && conversation.State != nil {
		if domain, ok := conversation.State["domain"]; ok && domain != nil {
			contextInfo += fmt.Sprintf("Business Domain: %v\n", domain)
		}
		if agentTypes, ok := conversation.State["agent_types"]; ok && agentTypes != nil {
			contextInfo += fmt.Sprintf("Mentioned Agent Types: %v\n", agentTypes)
		}
		if phase := conversation.Phase; phase != "" {
			contextInfo += fmt.Sprintf("Current Phase: %s\n", phase)
		}
	}

	prompt := `Based on our entire conversation, please generate the complete agency design specification in JSON format.`

	if contextInfo != "" {
		prompt += fmt.Sprintf("\n\nContext from our conversation:\n%s", contextInfo)
	}

	prompt += `

Include:
1. Agency name, description, and category
2. All agent types with their schemas (following JSON Schema format)
3. Communication relationships between agents
4. Recommended instance counts

Format your response as a JSON object with this structure:
{
  "name": "Agency Name",
  "description": "Detailed description",
  "category": "infrastructure|logistics|healthcare|etc",
  "agent_types": [
    {
      "id": "agent_type_id",
      "name": "Human Readable Name",
      "description": "What this agent does",
      "category": "infrastructure|core|custom",
      "capabilities": ["capability1", "capability2"],
      "schema": {
        "$schema": "http://json-schema.org/draft-07/schema#",
        "type": "object",
        "required": ["field1", "field2"],
        "properties": {
          "field1": {"type": "string", "description": "Field description"}
        }
      },
      "default_config": {},
      "count": 3
    }
  ],
  "relationships": [
    {
      "from": "agent_type_1",
      "to": "agent_type_2",
      "type": "pub_sub",
      "topics": ["topic.name"],
      "description": "What is communicated"
    }
  ],
  "metadata": {
    "design_approach": "centralized|distributed",
    "scalability": "high|medium|low"
  }
}

Please provide ONLY the JSON, no additional text.`

	return prompt
}

const requirementsSystemPrompt = `You are an AI agency designer for CodeValdCortex multi-agent platform.

Help users design multi-agent architectures. Ask strategic questions to understand business domain, stakeholders, processes, and scale.

**Context Awareness**: If user provides "Context:" section, they're referring to that specific text. When they say "remove/change/add this", quote the context and tell them to click **Refine** button.

Phase: REQUIREMENTS
Ask concise questions (2-3 max):
- Problem being solved?
- Key stakeholders/actors?
- Main processes/workflows?
- Data to track?
- Scale?

Keep responses brief. Use emojis (🎯 ✅ 🤔 💡).`

const brainstormSystemPrompt = `You are an AI system architect designing multi-agent systems.

Phase: AGENT BRAINSTORMING
Suggest 4-6 agent types based on their requirements.

**Context Awareness**: If "Context:" section provided, user is referring to that text. For "remove/change this", guide them to **Refine** button.

For each agent type (brief):
- Name + emoji
- Primary role
- Key capabilities
- Instance count

Patterns:
- Infrastructure: Sensors, Controllers, Monitors
- Logistics: Receiving, Storage, Picking, Packing, Shipping
- Healthcare: Patient Care, Diagnostics, Scheduling, Records
- Water: Sensors, Pumps, Valves, Coordinators

Keep responses concise. Use emojis.`

const relationshipSystemPrompt = `You are an AI system architect mapping agent communication.

Phase: RELATIONSHIP MAPPING
Define how agents communicate.

**Context Awareness**: If "Context:" provided, user refers to that text. Guide "remove/change this" to **Refine** button.

For each relationship (brief):
- Which agents communicate
- Pattern (pub/sub, direct, broadcast)
- Topics/messages exchanged
- Why needed

Use pub/sub for: one-to-many, events, decoupled systems
Use direct for: request-response, synchronous ops

Keep responses concise.`

const validationSystemPrompt = `You are an AI system architect validating designs.

Phase: VALIDATION
Review architecture for completeness and quality.

**Context Awareness**: If "Context:" provided, user refers to that text. Guide "remove/change this" to **Refine** button.

Check (brief):
- Missing agent types
- Communication gaps
- Scalability concerns
- Single points of failure
- Security

Suggest improvements concisely. Use emojis (✅ ⚠️ 🔄).`

// FormatAgencyContextBlock formats agency basic information and JSON data in a standardized format
// for inclusion in AI prompts. This ensures consistent agency context presentation across all AI calls.
//
// Parameters:
//   - contextData: builder.BuilderContext containing all agency metadata and working data
//
// Returns a formatted string block with:
//   - Agency basic information (name, category, description)
//   - Complete context data in JSON format with visual separators
func FormatAgencyContextBlock(contextData builder.BuilderContext) string {
	var builder strings.Builder

	// Add agency basic information from builder.BuilderContext
	if contextData.AgencyName != "" {
		builder.WriteString("You are working with the following agency context.\n\n")
		builder.WriteString(fmt.Sprintf("**Agency Name:** %s\n", contextData.AgencyName))
		if contextData.AgencyCategory != "" {
			builder.WriteString(fmt.Sprintf("**Category:** %s\n", contextData.AgencyCategory))
		}
		if contextData.AgencyDescription != "" {
			builder.WriteString(fmt.Sprintf("**Description:** %s\n", contextData.AgencyDescription))
		}
		builder.WriteString("\n")
	}

	// Add JSON data block with visual separators
	// Marshal the complete context data to JSON
	jsonData, err := json.MarshalIndent(contextData, "", "  ")
	if err == nil && len(jsonData) > 0 {
		builder.WriteString("═══════════════════════════════════════════\n")
		builder.WriteString("CONTEXT DATA (JSON):\n")
		builder.WriteString("═══════════════════════════════════════════\n")
		builder.WriteString(string(jsonData))
		builder.WriteString("\n═══════════════════════════════════════════\n\n")
	}

	return builder.String()
}
