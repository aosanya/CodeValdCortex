package ai_refine

import (
	"encoding/json"
	"fmt"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/aosanya/CodeValdCortex/internal/builder"
	"github.com/aosanya/CodeValdCortex/internal/builder/ai"
	"github.com/gin-gonic/gin"
)

// StreamingOptions contains configuration for a streaming AI operation
type StreamingOptions struct {
	AgencyID         string
	FormFieldName    string // e.g., "introduction-editor", "goal-description"
	UserRequestField string // Optional user request from form
	BuilderContextFn func() (builder.BuilderContext, error)
	StreamCallbackFn func(chunk string) error
	SaveResultFn     func(result interface{}) error
	CompletionDataFn func(result interface{}) map[string]interface{}
}

// StreamingHandlerFunc is a function that executes a streaming AI operation
type StreamingHandlerFunc func(
	ctx *gin.Context,
	req *builder.RefineIntroductionRequest,
	builderContext builder.BuilderContext,
	streamCallback func(chunk string) error,
) (interface{}, error)

// ExecuteStreamingRefine is a generic handler for AI streaming operations
// It handles SSE setup, context building, streaming, and result saving
func (h *Handler) ExecuteStreamingRefine(c *gin.Context, options StreamingOptions, streamFn StreamingHandlerFunc) {
	h.logger.WithField("agency_id", options.AgencyID).Info("Processing streaming AI refinement request")

	// Get agency context
	ag, err := h.agencyService.GetAgency(c.Request.Context(), options.AgencyID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch agency")
		c.SSEvent("error", `{"error": "Agency not found"}`)
		return
	}

	// Get current value from form
	currentValue := c.PostForm(options.FormFieldName)

	// Get user request
	userRequest := c.PostForm("user-request")
	if userRequest == "" && options.UserRequestField != "" {
		userRequest = c.PostForm(options.UserRequestField)
	}
	if userRequest == "" {
		userRequest = c.GetHeader("X-User-Request")
	}

	// Build AI context
	builderContextData, err := h.contextBuilder.BuildBuilderContext(c.Request.Context(), ag, currentValue, userRequest)
	if err != nil {
		h.logger.WithError(err).Error("Failed to build AI context data")
		c.SSEvent("error", `{"error": "Failed to build context"}`)
		return
	}

	// Set up SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // Disable nginx buffering

	// Flush headers
	c.Writer.Flush()

	// Send start event
	c.SSEvent("start", `{"status": "streaming"}`)
	c.Writer.Flush()

	// Call the streaming function
	result, err := streamFn(
		c,
		&builder.RefineIntroductionRequest{AgencyID: options.AgencyID},
		builderContextData,
		func(chunk string) error {
			// Send each chunk as an SSE event
			c.SSEvent("chunk", chunk)
			c.Writer.Flush()
			return nil
		},
	)

	if err != nil {
		h.logger.WithError(err).Error("Streaming refinement failed")
		c.SSEvent("error", fmt.Sprintf(`{"error": "%s"}`, err.Error()))
		return
	}

	// Save result if function provided
	if options.SaveResultFn != nil {
		if err := options.SaveResultFn(result); err != nil {
			h.logger.WithError(err).Error("Failed to save result")
			c.SSEvent("error", `{"error": "Failed to save changes"}`)
			return
		}
	}

	// Send completion event with metadata
	var completionData map[string]interface{}
	if options.CompletionDataFn != nil {
		completionData = options.CompletionDataFn(result)
	} else {
		// Default completion data
		completionData = map[string]interface{}{
			"complete": true,
		}
	}

	c.SSEvent("complete", completionData)
	c.Writer.Flush()
}

// IntroductionStreamBuilder wraps the introduction builder for streaming
type IntroductionStreamBuilder struct {
	builder *ai.IntroductionBuilder
}

// Stream executes the streaming introduction refinement
func (b *IntroductionStreamBuilder) Stream(
	ctx *gin.Context,
	req *builder.RefineIntroductionRequest,
	builderContext builder.BuilderContext,
	streamCallback func(chunk string) error,
) (interface{}, error) {
	return b.builder.RefineIntroductionStream(ctx.Request.Context(), req, builderContext, streamCallback)
}

// BuilderContextFromSpec is a helper to build context from specification
func (h *Handler) BuilderContextFromSpec(
	c *gin.Context,
	agencyID string,
	currentValue string,
	userRequest string,
) (builder.BuilderContext, error) {
	ag, err := h.agencyService.GetAgency(c.Request.Context(), agencyID)
	if err != nil {
		return builder.BuilderContext{}, fmt.Errorf("failed to fetch agency: %w", err)
	}

	return h.contextBuilder.BuildBuilderContext(c.Request.Context(), ag, currentValue, userRequest)
}

// SaveIntroduction saves the refined introduction
func (h *Handler) SaveIntroduction(c *gin.Context, agencyID string, spec *models.AgencySpecification, introduction string) error {
	needsSave := (introduction != spec.Introduction)
	if needsSave {
		_, err := h.agencyService.UpdateIntroduction(c.Request.Context(), agencyID, introduction, "ai-refine-stream")
		return err
	}
	return nil
}

// BuildIntroductionCompletionData creates the completion data for introduction streaming
func BuildIntroductionCompletionData(result interface{}) map[string]interface{} {
	introResult, ok := result.(*builder.RefineIntroductionResponse)
	if !ok {
		return map[string]interface{}{"error": "invalid result type"}
	}

	completionData := map[string]interface{}{
		"was_changed":      introResult.WasChanged,
		"explanation":      introResult.Explanation,
		"changed_sections": introResult.ChangedSections,
	}

	if introResult.Data != nil {
		completionData["introduction"] = introResult.Data.Introduction
	}

	return completionData
}

// SSEvent is a helper to send SSE events (if not using Gin's built-in method)
func SSEvent(c *gin.Context, event string, data interface{}) {
	var dataStr string
	switch v := data.(type) {
	case string:
		dataStr = v
	case map[string]interface{}:
		jsonData, _ := json.Marshal(v)
		dataStr = string(jsonData)
	default:
		jsonData, _ := json.Marshal(v)
		dataStr = string(jsonData)
	}

	c.Writer.WriteString(fmt.Sprintf("event: %s\n", event))
	c.Writer.WriteString(fmt.Sprintf("data: %s\n\n", dataStr))
	c.Writer.Flush()
}
