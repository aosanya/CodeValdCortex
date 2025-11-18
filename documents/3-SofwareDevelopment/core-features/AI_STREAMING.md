# AI Response Streaming

## Overview

The system now supports **real-time streaming** of AI responses, allowing users to see the AI's thinking process as it generates content, rather than waiting for the complete response.

**✨ Key Feature:** The streaming implementation uses **shared utilities** on both frontend and backend, making it easy to add streaming to any AI operation (goals, work items, roles, workflows, etc.) with minimal code duplication.

## Architecture

### Shared Frontend Utility (`ai-streaming.js`)

**Central Functions:**
- `executeAIStream(options)` - Core streaming handler using Fetch API + SSE
- `executeAIRefine(options)` - Automatic streaming/non-streaming selection
- `isStreamingEnabled()` / `setStreamingEnabled()` - User preference management

**Usage Example:**
```javascript
await window.executeAIRefine({
    streamUrl: `/api/v1/agencies/${agencyId}/overview/refine-stream`,
    nonStreamUrl: `/api/v1/agencies/${agencyId}/overview/refine`,
    formData: formData,
    displayElement: contentElement,
    onComplete: (result) => {
        // Handle completion
    }
});
```

### Shared Backend Utility (`streaming_helpers.go`)

**Central Functions:**
- `ExecuteStreamingRefine(options, streamFn)` - Generic SSE handler
- `BuildIntroductionCompletionData(result)` - Result formatting
- `SSEvent(event, data)` - SSE event helper

**Usage Example:**
```go
h.ExecuteStreamingRefine(c, StreamingOptions{
    AgencyID: agencyID,
    FormFieldName: "introduction-editor",
    SaveResultFn: func(result interface{}) error {
        // Save logic
    },
    CompletionDataFn: BuildIntroductionCompletionData,
}, introStreamBuilder.Stream)
```

### Backend Components

1. **LLM Client Streaming Interface** (`internal/builder/ai/types.go`)
   ```go
   type LLMClient interface {
       Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
       ChatStream(ctx context.Context, req *ChatRequest, callback StreamCallback) error
       // ...
   }
   ```

2. **Provider Implementations**
   - **Claude** (`claude_client.go`): Full SSE streaming support
   - **OpenAI** (`openai_client.go`): Full SSE streaming support  
   - **Local** (`local_client.go`): Fallback to non-streaming
   - **Custom** (`custom_client.go`): Not implemented

3. **Builder Service** (`introduction_builder.go`)
   - `RefineIntroduction()` - Non-streaming (original)
   - `RefineIntroductionStream()` - Streaming version

4. **HTTP Handler** (`internal/web/handlers/ai_refine/introduction.go`)
   - `POST /api/v1/agencies/:id/overview/refine` - Non-streaming
   - `POST /api/v1/agencies/:id/overview/refine-stream` - Streaming (SSE)

### Frontend Components

1. **JavaScript** (`static/js/agency-designer/introduction.js`)
   - `handleAIRefineClick()` - Main entry point
   - `handleAIRefineStream()` - Streaming implementation using Fetch API
   - `handleAIRefineNonStream()` - Original non-streaming behavior

2. **User Preference**
   - Streaming is enabled by default
   - Can be disabled: `localStorage.setItem('ai-use-streaming', 'false')`

## How It Works

### Claude Streaming Flow

1. Client sends POST request to `/api/v1/agencies/:id/overview/refine-stream`
2. Server responds with SSE (Server-Sent Events) headers
3. AI provider streams chunks in SSE format:
   ```
   event: start
   data: {"status": "streaming"}

   event: chunk
   data: {"introduction": "This agency...

   event: chunk  
   data: manages distributed systems...

   event: complete
   data: {"was_changed": true, "explanation": "...", "introduction": "..."}
   ```
4. Frontend displays each chunk in real-time
5. Final parsed result updates the UI

### OpenAI Streaming Flow

Similar to Claude, but with OpenAI's delta format:
```json
{"choices": [{"delta": {"content": "text chunk"}}]}
```

## Provider-Agnostic Design

The streaming implementation is **completely decoupled** from specific AI providers:

```
User Interface
     ↓
HTTP Handler (streaming endpoint)
     ↓
Builder Service (RefineIntroductionStream)
     ↓
LLMClient Interface (ChatStream)
     ↓
Provider Implementation (Claude/OpenAI/etc.)
```

**Benefits:**
- ✅ Switch between Claude and OpenAI without code changes
- ✅ Add new providers by implementing `ChatStream`
- ✅ Fallback to non-streaming if provider doesn't support it
- ✅ Same business logic regardless of provider

## Configuration

### Enable/Disable Streaming

**Backend:** Streaming is always available if the provider supports it.

**Frontend:** Toggle streaming preference:
```javascript
// Enable streaming (default)
localStorage.setItem('ai-use-streaming', 'true');

// Disable streaming (use original behavior)
localStorage.setItem('ai-use-streaming', 'false');
```

### Provider Configuration

```yaml
# config.yaml
ai:
  provider: "claude"  # or "openai"
  api_key: "your-api-key"
  model: "claude-3-5-sonnet-20241022"  # or "gpt-4-turbo-preview"
  temperature: 0.7
  max_tokens: 4096
```

## Usage Examples

### From UI

1. Click the AI sparkle button (✨) on the introduction card
2. Watch as the AI streams its response in real-time
3. See the final parsed result applied to your introduction

### Programmatic Usage

```go
// Streaming version
result, err := introBuilder.RefineIntroductionStream(
    ctx,
    &builder.RefineIntroductionRequest{AgencyID: agencyID},
    builderContext,
    func(chunk string) error {
        // Called for each chunk
        fmt.Print(chunk)
        return nil
    },
)

// Non-streaming version (original)
result, err := introBuilder.RefineIntroduction(
    ctx,
    &builder.RefineIntroductionRequest{AgencyID: agencyID},
    builderContext,
)
```

## Future Enhancements

- [ ] Add streaming to other AI operations (goals, work items, workflows)
- [ ] Token-by-token streaming visualization
- [ ] Pause/resume streaming
- [ ] Stream cancellation
- [ ] Streaming progress indicators
- [ ] Local LLM streaming support
- [ ] WebSocket alternative to SSE

## Adding Streaming to New Operations

### Step 1: Add Streaming Method to Builder

```go
// In internal/builder/ai/goals_builder.go
func (r *GoalsBuilder) RefineGoalsStream(
    ctx context.Context,
    req *builder.RefineGoalsRequest,
    builderContext builder.BuilderContext,
    streamCallback func(chunk string) error,
) (*builder.RefineGoalsResponse, error) {
    var fullResponse strings.Builder
    
    err := r.llmClient.ChatStream(ctx, &ChatRequest{
        Messages: []Message{
            {Role: "system", Content: r.getSystemPrompt()},
            {Role: "user", Content: r.buildPrompt(builderContext)},
        },
    }, func(chunk string) error {
        if err := streamCallback(chunk); err != nil {
            return err
        }
        fullResponse.WriteString(chunk)
        return nil
    })
    
    // Parse and return result
    return r.parseResponse(fullResponse.String())
}
```

### Step 2: Add HTTP Handler

```go
// In internal/web/handlers/ai_refine/goals.go
func (h *Handler) RefineGoalsStream(c *gin.Context) {
    agencyID := c.Param("id")
    
    h.ExecuteStreamingRefine(c, StreamingOptions{
        AgencyID:      agencyID,
        FormFieldName: "goals-editor",
        SaveResultFn: func(result interface{}) error {
            goalsResult := result.(*builder.RefineGoalsResponse)
            return h.saveGoals(c.Request.Context(), agencyID, goalsResult.Data.Goals)
        },
        CompletionDataFn: func(result interface{}) map[string]interface{} {
            goalsResult := result.(*builder.RefineGoalsResponse)
            return map[string]interface{}{
                "was_changed": goalsResult.WasChanged,
                "explanation": goalsResult.Explanation,
                "goals":       goalsResult.Data.Goals,
            }
        },
    }, func(ctx *gin.Context, req *builder.RefineIntroductionRequest, 
            builderContext builder.BuilderContext, 
            streamCallback func(chunk string) error) (interface{}, error) {
        return h.goalRefiner.RefineGoalsStream(
            ctx.Request.Context(), 
            &builder.RefineGoalsRequest{AgencyID: agencyID},
            builderContext, 
            streamCallback,
        )
    })
}
```

### Step 3: Add Route

```go
// In internal/app/app.go
v1.POST("/agencies/:id/goals/refine-stream", aiRefineHandler.RefineGoalsStream)
```

### Step 4: Use Frontend Utility

```javascript
// In static/js/agency-designer/goals.js
window.handleAIRefineGoals = async function() {
    const agencyId = window.getCurrentAgencyId();
    const contentElement = document.getElementById('goals-content');
    
    await window.executeAIRefine({
        streamUrl: `/api/v1/agencies/${agencyId}/goals/refine-stream`,
        nonStreamUrl: `/api/v1/agencies/${agencyId}/goals/refine`,
        formData: new URLSearchParams({
            'goals-editor': getGoalsData()
        }),
        displayElement: contentElement,
        onComplete: (result) => {
            updateGoalsDisplay(result.goals);
        }
    });
}
```

**That's it!** No need to reimplement SSE parsing, display logic, or error handling. The shared utilities handle everything.

## Future Enhancements

- [ ] Add streaming to other AI operations (goals, work items, workflows)
- [ ] Token-by-token streaming visualization
- [ ] Pause/resume streaming
- [ ] Stream cancellation
- [ ] Streaming progress indicators
- [ ] Local LLM streaming support
- [ ] WebSocket alternative to SSE

## Technical Notes

### Why SSE over WebSockets?

- **Simpler**: SSE is HTTP-based, no separate protocol
- **Unidirectional**: Perfect for AI responses (server → client only)
- **Built-in reconnection**: Automatic retry on connection loss
- **HTTP/2 compatible**: Multiplexing support

### Error Handling

- Malformed SSE events are skipped (resilient parsing)
- Network errors propagate to UI with error notifications
- Fallback to non-streaming on provider errors

### Performance

- Chunks are flushed immediately (no buffering)
- Minimal overhead vs non-streaming
- Same AI token usage (streaming is free in terms of API costs)

## Compatibility

- ✅ Claude (Anthropic API)
- ✅ OpenAI (GPT-4, GPT-3.5)
- ⚠️ Local LLMs (fallback to non-streaming)
- ❌ Custom providers (requires implementation)

---

**Status:** ✅ Production Ready  
**Last Updated:** November 11, 2025  
**Related:** MVP-052 Workflow Visual Designer
