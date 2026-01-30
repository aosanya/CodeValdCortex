package prompts

// FlexibleIntroductionSystemPrompt defines the AI behavior for flexible introduction operations
const FlexibleIntroductionSystemPrompt = `You are an expert agency introduction designer specializing in creating clear, structured, and professional agency documentation.

YOUR ROLE:
- Generate well-structured agency introduction sections based on user input
- Refine existing sections to improve clarity, professionalism, and effectiveness
- Ensure content is specific, actionable, and aligned with the agency's purpose
- Maintain consistency with the provided template structure

SECTION TYPES YOU WORK WITH:
1. TEXT: Single paragraph or multi-paragraph text content
2. LIST: Bullet points or numbered items (use "items" array in content)
3. NESTED: Hierarchical sections with subsections (use "sections" array in content)
4. TABLE: Tabular data with headers and rows (use "headers" and "rows" arrays in content)

OUTPUT REQUIREMENTS:
- Always return valid JSON matching the required schema
- Be concise and professional - avoid marketing fluff
- Focus on specific capabilities and value propositions
- Use clear, jargon-free language unless domain-specific terms are necessary
- Ensure all content is actionable and informative

FORBIDDEN BEHAVIORS:
❌ Generic placeholder text like "Insert description here"
❌ Overly promotional or salesy language
❌ Vague statements without specific details
❌ Inconsistent terminology across sections
❌ Conversational tone ("Let me help you...", "I think...")

CORRECT BEHAVIORS:
✓ Specific, concrete descriptions
✓ Professional, third-person perspective
✓ Consistent terminology and structure
✓ Clear value propositions
✓ Actionable information

You are a JSON API that generates or refines introduction sections. Return ONLY valid JSON.`

// GenerateIntroductionPromptTemplate creates a prompt for generating a complete introduction
func GenerateIntroductionPromptTemplate(template, keywords string, agencyContext map[string]interface{}) string {
	return `Generate a complete agency introduction using the following information:

**TEMPLATE:** ` + template + `

**KEYWORDS:** ` + keywords + `

**AGENCY CONTEXT:**
` + formatAgencyContext(agencyContext) + `

**TASK:**
Generate all sections for the specified template with content that:
1. Incorporates the provided keywords naturally
2. Reflects the agency's context and purpose
3. Is specific, professional, and actionable
4. Follows the template structure exactly

**OUTPUT FORMAT:**
Return a JSON object with this structure:
{
  "sections": [
    {
      "id": "unique-uuid",
      "type": "text|list|nested|table",
      "title": "Section Title",
      "content": { /* type-appropriate content */ },
      "order": 1,
      "required": true|false
    }
    // ... more sections
  ],
  "confidence": 0.0-1.0,
  "explanation": "Brief explanation of design choices"
}

Return ONLY the JSON object. No additional text.`
}

// RefineSectionPromptTemplate creates a prompt for refining a specific section
func RefineSectionPromptTemplate(sectionType, currentContent, refinementRequest string, agencyContext map[string]interface{}) string {
	return `Refine the following agency introduction section based on the user's request.

**SECTION TYPE:** ` + sectionType + `

**CURRENT CONTENT:**
` + currentContent + `

**REFINEMENT REQUEST:**
` + refinementRequest + `

**AGENCY CONTEXT:**
` + formatAgencyContext(agencyContext) + `

**TASK:**
Modify the section content according to the refinement request while:
1. Maintaining the section type and structure
2. Keeping the content professional and specific
3. Ensuring consistency with agency context
4. Improving clarity and effectiveness

**OUTPUT FORMAT:**
Return a JSON object with this structure:
{
  "content": { /* refined content matching section type */ },
  "changed": true|false,
  "explanation": "What was changed and why",
  "confidence": 0.0-1.0
}

Return ONLY the JSON object. No additional text.`
}

// GenerateSectionPromptTemplate creates a prompt for generating a single new section
func GenerateSectionPromptTemplate(sectionType, title, description string, agencyContext map[string]interface{}) string {
	return `Generate a new agency introduction section with the following specifications:

**SECTION TYPE:** ` + sectionType + `

**SECTION TITLE:** ` + title + `

**DESCRIPTION:** ` + description + `

**AGENCY CONTEXT:**
` + formatAgencyContext(agencyContext) + `

**TASK:**
Create content for this section that:
1. Matches the specified type and title
2. Fulfills the description requirements
3. Is professional, specific, and actionable
4. Aligns with the agency context

**OUTPUT FORMAT:**
Return a JSON object with this structure:
{
  "content": { /* content matching section type */ },
  "explanation": "Brief explanation of the generated content",
  "confidence": 0.0-1.0
}

Return ONLY the JSON object. No additional text.`
}

// formatAgencyContext converts agency context map to formatted string
func formatAgencyContext(context map[string]interface{}) string {
	if context == nil || len(context) == 0 {
		return "No additional context provided"
	}

	result := ""
	if name, ok := context["name"].(string); ok && name != "" {
		result += "Agency Name: " + name + "\n"
	}
	if description, ok := context["description"].(string); ok && description != "" {
		result += "Description: " + description + "\n"
	}
	if domain, ok := context["domain"].(string); ok && domain != "" {
		result += "Domain: " + domain + "\n"
	}
	if purpose, ok := context["purpose"].(string); ok && purpose != "" {
		result += "Purpose: " + purpose + "\n"
	}

	if result == "" {
		return "No additional context provided"
	}
	return result
}

// TemplateDescriptions provides context about each template type
const TemplateDescriptions = `
TEMPLATE: genesis
Description: Comprehensive 6-section introduction covering all aspects
Sections:
  1. Overview (TEXT): High-level purpose and mission
  2. Key Capabilities (LIST): Core features and strengths
  3. Target Audience (TEXT): Who the agency serves
  4. Core Principles (NESTED): Guiding values and principles
  5. Service Offerings (TABLE): Services with descriptions and timelines
  6. Getting Started (TEXT): How to begin working with the agency

TEMPLATE: minimal
Description: Streamlined 3-section introduction for quick setup
Sections:
  1. About (TEXT): Brief introduction and purpose
  2. What We Offer (LIST): Key services or capabilities
  3. Contact (TEXT): How to get in touch

TEMPLATE: custom
Description: User-defined structure with any combination of sections
Sections: Defined by user requirements
`

// ExampleOutputs provides reference examples for AI
const ExampleOutputs = `
EXAMPLE - TEXT SECTION:
{
  "content": {
    "text": "This agency specializes in automated software development, managing the complete lifecycle from requirements gathering through deployment. It coordinates multiple AI agents to handle coding, testing, documentation, and infrastructure tasks."
  }
}

EXAMPLE - LIST SECTION:
{
  "content": {
    "items": [
      "Automated code generation and refactoring",
      "Comprehensive test suite creation",
      "Real-time documentation generation",
      "CI/CD pipeline management"
    ]
  }
}

EXAMPLE - NESTED SECTION:
{
  "content": {
    "sections": [
      {
        "title": "Quality First",
        "content": "Every deliverable undergoes automated testing and code review to ensure production-ready quality."
      },
      {
        "title": "Transparency",
        "content": "All decisions, changes, and progress are logged and accessible through comprehensive audit trails."
      }
    ]
  }
}

EXAMPLE - TABLE SECTION:
{
  "content": {
    "headers": ["Service", "Description", "Typical Duration"],
    "rows": [
      ["Code Review", "Automated analysis and suggestions", "< 1 hour"],
      ["Feature Development", "Full implementation with tests", "1-3 days"],
      ["Refactoring", "Code quality improvements", "4-8 hours"]
    ]
  }
}
`
