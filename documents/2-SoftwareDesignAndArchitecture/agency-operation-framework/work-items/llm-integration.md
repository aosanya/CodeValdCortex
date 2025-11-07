# LLM Integration

This document describes how LLM (Large Language Model) agents generate content for work items and integrate with the GitOps workflow.

## Overview

LLM agents execute work items by:
1. **Analyzing** Gitea issue requirements
2. **Loading** agency context from ArangoDB
3. **Generating** content (documents, code, proposals)
4. **Validating** outputs for correctness
5. **Creating** git commits with generated content

## LLM Agent Architecture

### Agent Configuration

```go
type LLMAgent struct {
    ID           string
    Name         string
    Model        string           // "gpt-4", "gpt-3.5-turbo", etc.
    Temperature  float64          // 0.0-1.0 (creativity)
    MaxTokens    int              // Max tokens per request
    Client       *openai.Client
    ContextDB    *arangodb.Database
    Budget       AgentBudget
}

type AgentBudget struct {
    MaxTokensPerDay     int
    MaxCostPerDay       float64   // USD
    MaxTokensPerRequest int
    UsedTokens          int
    UsedCost            float64
}
```

### Initialization

```go
func NewLLMAgent(config AgentConfig) (*LLMAgent, error) {
    client := openai.NewClient(config.APIKey)
    
    return &LLMAgent{
        ID:          config.ID,
        Name:        config.Name,
        Model:       config.Model,
        Temperature: config.Temperature,
        MaxTokens:   config.MaxTokens,
        Client:      client,
        ContextDB:   config.Database,
        Budget: AgentBudget{
            MaxTokensPerDay:     config.MaxTokensPerDay,
            MaxCostPerDay:       config.MaxCostPerDay,
            MaxTokensPerRequest: config.MaxTokensPerRequest,
        },
    }, nil
}
```

## Content Generation by Work Type

### Document Generation

```go
func (a *LLMAgent) GenerateDocument(ctx context.Context, issue *gitea.Issue) (string, error) {
    // 1. Load existing document (if updating)
    existingContent := a.loadExistingFile(issue.Metadata.FilePath)
    
    // 2. Load agency context
    agencyCtx := a.loadAgencyContext(issue.AgencyID)
    
    // 3. Build prompt
    prompt := fmt.Sprintf(`
You are a documentation expert for the %s agency.

**Task**: Update the following document based on issue requirements.

**File**: %s

**Current Content**:
---
%s
---

**Issue #%d**: %s

**Requirements**:
%s

**Instructions**:
1. Review the current content carefully
2. Update relevant sections based on requirements
3. Maintain existing structure and formatting
4. Add new sections if needed
5. Ensure technical accuracy
6. Return COMPLETE updated document in Markdown format

**Agency Context**:
- Goals: %s
- Current Status: %s

Generate the updated document:
`, 
        agencyCtx.Agency.Name,
        issue.Metadata.FilePath,
        existingContent,
        issue.Index,
        issue.Title,
        issue.Body,
        summarizeGoals(agencyCtx.Goals),
        agencyCtx.Agency.Status,
    )
    
    // 4. Call LLM
    response, err := a.complete(ctx, prompt, 4000)
    if err != nil {
        return "", fmt.Errorf("LLM generation failed: %w", err)
    }
    
    // 5. Validate Markdown
    if !isValidMarkdown(response) {
        return "", fmt.Errorf("invalid markdown generated")
    }
    
    return response, nil
}
```

### Code Generation

```go
func (a *LLMAgent) GenerateCode(ctx context.Context, issue *gitea.Issue) (*CodeGeneration, error) {
    // Step 1: Create implementation plan
    plan, err := a.createImplementationPlan(ctx, issue)
    if err != nil {
        return nil, err
    }
    
    // Step 2: Generate code for each file
    var generatedFiles []GeneratedFile
    for _, fileSpec := range plan.Files {
        code, err := a.generateFile(ctx, issue, fileSpec)
        if err != nil {
            return nil, fmt.Errorf("failed to generate %s: %w", fileSpec.Path, err)
        }
        
        generatedFiles = append(generatedFiles, GeneratedFile{
            Path:    fileSpec.Path,
            Content: code,
        })
    }
    
    // Step 3: Generate tests
    var testFiles []GeneratedFile
    for _, testSpec := range plan.Tests {
        test, err := a.generateTest(ctx, issue, testSpec)
        if err != nil {
            return nil, fmt.Errorf("failed to generate test %s: %w", testSpec.Path, err)
        }
        
        testFiles = append(testFiles, GeneratedFile{
            Path:    testSpec.Path,
            Content: test,
        })
    }
    
    return &CodeGeneration{
        Plan:  plan,
        Files: generatedFiles,
        Tests: testFiles,
    }, nil
}

func (a *LLMAgent) createImplementationPlan(ctx context.Context, issue *gitea.Issue) (*ImplementationPlan, error) {
    prompt := fmt.Sprintf(`
You are an expert software architect.

**Task**: Create implementation plan for the following feature.

**Issue #%d**: %s

**Requirements**:
%s

**Instructions**:
Generate a JSON implementation plan with:
{
  "files": [
    {"path": "internal/api/handler.go", "action": "create|modify", "description": "..."},
    ...
  ],
  "tests": [
    {"path": "internal/api/handler_test.go", "description": "..."},
    ...
  ],
  "dependencies": ["go get github.com/pkg/errors"],
  "summary": "Brief description of approach"
}

Generate the plan:
`,
        issue.Index,
        issue.Title,
        issue.Body,
    )
    
    response, err := a.complete(ctx, prompt, 1000)
    if err != nil {
        return nil, err
    }
    
    var plan ImplementationPlan
    if err := json.Unmarshal([]byte(response), &plan); err != nil {
        return nil, fmt.Errorf("invalid plan JSON: %w", err)
    }
    
    return &plan, nil
}

func (a *LLMAgent) generateFile(ctx context.Context, issue *gitea.Issue, spec FileSpec) (string, error) {
    existingCode := ""
    if spec.Action == "modify" {
        existingCode = a.loadExistingFile(spec.Path)
    }
    
    prompt := fmt.Sprintf(`
You are an expert Go developer.

**Task**: %s file: %s

**Existing Code**:
---
%s
---

**Requirements**: %s

**Instructions**:
1. Write complete, working Go code
2. Include proper error handling
3. Add Godoc comments for all exported items
4. Follow Go best practices and idioms
5. Use standard library when possible
6. Include necessary imports

Generate the code:
`,
        spec.Action,
        spec.Path,
        existingCode,
        spec.Description,
    )
    
    code, err := a.complete(ctx, prompt, 3000)
    if err != nil {
        return "", err
    }
    
    // Validate Go syntax
    if err := validateGoSyntax(code); err != nil {
        return "", fmt.Errorf("invalid Go syntax: %w", err)
    }
    
    return code, nil
}
```

### Proposal Generation

```go
func (a *LLMAgent) GenerateProposal(ctx context.Context, issue *gitea.Issue) (string, error) {
    agencyCtx := a.loadAgencyContext(issue.AgencyID)
    
    prompt := fmt.Sprintf(`
You are a business strategist for %s.

**Task**: Update the proposal document.

**Issue #%d**: %s

**Current Proposal**:
---
%s
---

**Update Requirements**:
%s

**Agency Context**:
- Industry: %s
- Goals: %s
- Capabilities: %s

**Instructions**:
1. Update relevant sections of the proposal
2. Maintain professional business writing style
3. Include specific metrics and examples
4. Ensure alignment with agency goals
5. Return complete updated proposal

Generate the proposal:
`,
        agencyCtx.Agency.Name,
        issue.Index,
        issue.Title,
        a.loadExistingFile(issue.Metadata.FilePath),
        issue.Body,
        agencyCtx.Agency.Industry,
        summarizeGoals(agencyCtx.Goals),
        summarizeCapabilities(agencyCtx.Roles),
    )
    
    return a.complete(ctx, prompt, 4000)
}
```

### Analysis Generation

```go
func (a *LLMAgent) GenerateAnalysis(ctx context.Context, issue *gitea.Issue) (string, error) {
    // Load relevant graph data
    graphData := a.queryGraphData(issue)
    
    prompt := fmt.Sprintf(`
You are a technical analyst.

**Task**: Perform analysis based on codebase data.

**Issue #%d**: %s

**Analysis Requirements**:
%s

**Codebase Data**:
%s

**Instructions**:
1. Analyze the provided data
2. Identify patterns and issues
3. Provide actionable recommendations
4. Use Markdown formatting
5. Include code examples where relevant

Generate the analysis report:
`,
        issue.Index,
        issue.Title,
        issue.Body,
        formatGraphData(graphData),
    )
    
    return a.complete(ctx, prompt, 3000)
}
```

## Agency Context Loading

### Context Structure

```go
type AgencyContext struct {
    Agency    *Agency
    Goals     []*Goal
    WorkItems []*WorkItem
    Roles     []*Role
    CodeGraph *CodeGraphSummary
}

type CodeGraphSummary struct {
    TotalFiles       int
    LanguageBreakdown map[string]int
    TopDependencies  []DependencyInfo
}
```

### Loading Context

```go
func (a *LLMAgent) loadAgencyContext(agencyID string) (*AgencyContext, error) {
    query := `
        LET agency = DOCUMENT(CONCAT('agencies/', @agencyID))
        
        LET goals = (
            FOR goal IN goals
                FILTER goal.agency_id == @agencyID
                FILTER goal.status == "active"
                RETURN goal
        )
        
        LET workItems = (
            FOR wi IN work_items
                FILTER wi.agency_id == @agencyID
                FILTER wi.status IN ["pending", "executing"]
                RETURN wi
        )
        
        LET roles = (
            FOR role IN roles
                FILTER role.agency_id == @agencyID
                RETURN role
        )
        
        LET codeStats = (
            FOR file IN git_objects
                FILTER file.repo_id == agency.repo_id
                FILTER file.type == "blob"
                COLLECT language = file.language WITH COUNT INTO count
                RETURN {language, count}
        )
        
        RETURN {
            agency: agency,
            goals: goals,
            work_items: workItems,
            roles: roles,
            code_stats: codeStats
        }
    `
    
    cursor, err := a.ContextDB.Query(context.Background(), query, map[string]interface{}{
        "agencyID": agencyID,
    })
    if err != nil {
        return nil, err
    }
    
    var ctx AgencyContext
    _, err = cursor.ReadDocument(context.Background(), &ctx)
    return &ctx, err
}
```

## LLM API Integration

### Completion Wrapper

```go
func (a *LLMAgent) complete(ctx context.Context, prompt string, maxTokens int) (string, error) {
    // Check budget
    if err := a.checkBudget(maxTokens); err != nil {
        return "", err
    }
    
    // Make API call
    req := openai.CompletionRequest{
        Model:       a.Model,
        Prompt:      prompt,
        MaxTokens:   maxTokens,
        Temperature: a.Temperature,
    }
    
    startTime := time.Now()
    resp, err := a.Client.CreateCompletion(ctx, req)
    if err != nil {
        return "", fmt.Errorf("OpenAI API error: %w", err)
    }
    duration := time.Since(startTime)
    
    // Track usage
    a.trackUsage(resp.Usage, duration)
    
    // Extract response
    if len(resp.Choices) == 0 {
        return "", fmt.Errorf("no completion choices returned")
    }
    
    return strings.TrimSpace(resp.Choices[0].Text), nil
}

func (a *LLMAgent) checkBudget(estimatedTokens int) error {
    if a.Budget.UsedTokens+estimatedTokens > a.Budget.MaxTokensPerDay {
        return fmt.Errorf("daily token budget exceeded")
    }
    
    estimatedCost := float64(estimatedTokens) * 0.00002 // Example pricing
    if a.Budget.UsedCost+estimatedCost > a.Budget.MaxCostPerDay {
        return fmt.Errorf("daily cost budget exceeded")
    }
    
    return nil
}

func (a *LLMAgent) trackUsage(usage openai.Usage, duration time.Duration) {
    a.Budget.UsedTokens += usage.TotalTokens
    a.Budget.UsedCost += calculateCost(usage, a.Model)
    
    // Store in ArangoDB for analytics
    a.ContextDB.Collection("llm_usage").CreateDocument(context.Background(), map[string]interface{}{
        "agent_id":        a.ID,
        "model":           a.Model,
        "prompt_tokens":   usage.PromptTokens,
        "completion_tokens": usage.CompletionTokens,
        "total_tokens":    usage.TotalTokens,
        "cost":            calculateCost(usage, a.Model),
        "duration_ms":     duration.Milliseconds(),
        "timestamp":       time.Now(),
    })
}
```

### Cost Calculation

```go
func calculateCost(usage openai.Usage, model string) float64 {
    // Pricing as of Nov 2025 (example)
    pricing := map[string]struct{ prompt, completion float64 }{
        "gpt-4":          {0.03, 0.06},     // per 1K tokens
        "gpt-3.5-turbo":  {0.0015, 0.002},
    }
    
    rates, ok := pricing[model]
    if !ok {
        rates = pricing["gpt-3.5-turbo"] // default
    }
    
    promptCost := float64(usage.PromptTokens) / 1000.0 * rates.prompt
    completionCost := float64(usage.CompletionTokens) / 1000.0 * rates.completion
    
    return promptCost + completionCost
}
```

## Validation

### Syntax Validation

```go
func validateGoSyntax(code string) error {
    fset := token.NewFileSet()
    _, err := parser.ParseFile(fset, "generated.go", code, parser.AllErrors)
    if err != nil {
        return fmt.Errorf("syntax error: %w", err)
    }
    return nil
}

func isValidMarkdown(content string) bool {
    // Basic validation - check for common markdown patterns
    hasHeaders := strings.Contains(content, "#")
    notEmpty := strings.TrimSpace(content) != ""
    return hasHeaders && notEmpty
}
```

### Content Quality Checks

```go
func (a *LLMAgent) validateDocumentQuality(content string) error {
    // Check minimum length
    if len(content) < 100 {
        return fmt.Errorf("document too short")
    }
    
    // Check for incomplete sentences
    if strings.Contains(content, "...") && !strings.Contains(content, "```") {
        return fmt.Errorf("appears to contain truncated content")
    }
    
    // Check for code blocks in code
    lines := strings.Split(content, "\n")
    if lines[0] != "```" && strings.Contains(content, "func ") {
        // Go code should be in code blocks
        return fmt.Errorf("code should be in markdown code blocks")
    }
    
    return nil
}
```

## Retry Logic

```go
func (a *LLMAgent) completeWithRetry(ctx context.Context, prompt string, maxTokens int) (string, error) {
    maxRetries := 3
    backoff := time.Second
    
    for attempt := 0; attempt < maxRetries; attempt++ {
        response, err := a.complete(ctx, prompt, maxTokens)
        
        if err == nil {
            return response, nil
        }
        
        // Check if retryable error
        if isRateLimitError(err) || isTimeoutError(err) {
            time.Sleep(backoff)
            backoff *= 2
            continue
        }
        
        // Non-retryable error
        return "", err
    }
    
    return "", fmt.Errorf("max retries exceeded")
}

func isRateLimitError(err error) bool {
    return strings.Contains(err.Error(), "rate limit")
}

func isTimeoutError(err error) bool {
    return strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline")
}
```

## Metrics & Observability

### Usage Metrics

```go
type LLMMetrics struct {
    TotalRequests      int64
    TotalTokens        int64
    TotalCost          float64
    AvgResponseTime    time.Duration
    ErrorRate          float64
    ByModel            map[string]ModelMetrics
    ByWorkType         map[string]WorkTypeMetrics
}

type ModelMetrics struct {
    Requests int64
    Tokens   int64
    Cost     float64
}
```

### Query Usage

```go
// Daily usage by agent
query := `
    FOR usage IN llm_usage
        FILTER usage.timestamp >= DATE_SUBTRACT(DATE_NOW(), 1, "day")
        COLLECT agent = usage.agent_id
        AGGREGATE 
            total_requests = COUNT(),
            total_tokens = SUM(usage.total_tokens),
            total_cost = SUM(usage.cost)
        RETURN {
            agent,
            requests: total_requests,
            tokens: total_tokens,
            cost: total_cost
        }
`
```

---

**See Also**:
- [Work Item Types](./work-item-types.md) - Work type specifications
- [GitOps Workflow](./gitops-workflow.md) - How generated content is committed
