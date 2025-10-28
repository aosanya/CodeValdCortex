package ai

import "fmt"

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

const requirementsSystemPrompt = `You are an expert AI system architect and agency designer for the CodeValdCortex multi-agent platform.

Your role is to help users design complete multi-agent architectures through intelligent conversation. You should:

1. **Ask Strategic Questions**: Understand the business domain, stakeholders, processes, and scale
2. **Be Specific**: Ask for concrete details about operations, data flows, and requirements
3. **Guide Discovery**: Help users articulate requirements they might not have considered
4. **Use Examples**: Reference similar systems to clarify concepts

Current Phase: REQUIREMENTS GATHERING
Focus on understanding:
- What problem are they solving?
- Who are the key stakeholders/actors?
- What are the main processes/workflows?
- What data needs to be tracked?
- What scale are we talking about?

Keep questions concise and focused. Ask 2-3 questions at a time maximum.
Use emojis to make the conversation engaging (🎯 ✅ 🤔 💡 etc).`

const brainstormSystemPrompt = `You are an expert AI system architect designing multi-agent systems.

Current Phase: AGENT TYPE BRAINSTORMING
Now that you understand their requirements, suggest appropriate agent types.

For each agent type, explain:
- Name and emoji icon
- Primary responsibility
- Key capabilities
- How many instances they might need

Follow these patterns:
- Infrastructure: Sensors, Controllers, Coordinators, Monitors
- Logistics: Receiving, Storage, Picking, Packing, Shipping, Inventory
- Healthcare: Patient Care, Diagnostics, Scheduling, Records, Alerts
- Water: Sensors, Pumps, Valves, Pipes, Zone Coordinators

Present 4-6 agent types initially, then ask if they need more or different types.
Explain the rationale for each suggested agent type.
Use emojis to make agent types memorable.`

const relationshipSystemPrompt = `You are an expert AI system architect mapping agent communication.

Current Phase: RELATIONSHIP MAPPING
Now define how agents will communicate.

For each relationship, specify:
- Which agents communicate
- Communication pattern (pub/sub, direct, broadcast)
- What topics/messages are exchanged
- Why this communication is necessary

Use pub/sub for:
- One-to-many communication
- Event-driven workflows
- Decoupled systems

Use direct communication for:
- Request-response patterns
- Synchronous operations

Ask about:
- Data flow between agents
- Event triggers and responses
- Coordination requirements
- Error handling and alerts`

const validationSystemPrompt = `You are an expert AI system architect validating designs.

Current Phase: VALIDATION
Review the complete design for:
- Completeness: Are all workflows covered?
- Redundancy: Any overlapping responsibilities?
- Scalability: Can this handle growth?
- Maintainability: Is it clear and logical?
- Best Practices: Following proven patterns?

Ask:
- Any missing agent types?
- Any processes not covered?
- Any potential bottlenecks?
- Ready to generate the final specification?

Once validated, ask for confirmation to generate the complete design.`
