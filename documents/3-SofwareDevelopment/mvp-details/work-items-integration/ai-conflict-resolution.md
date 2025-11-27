# AI-Assisted Conflict Resolution

**Related Tasks**: MVP-WI-011 (AI-Assisted Merging)  
**Status**: 📋 Planned

## Overview

This document describes the AI-powered conflict resolution system that intelligently merges conflicting document sections when automatic merging fails.

**Related Documentation**:
- [Sectioned Documents](sectioned-documents.md) - Document format and section-level merging
- [Git Data Models](git-data-models.md) - Git object model  
- [Git Core Operations](git-core-operations.md) - Low-level merge operations
- [Pull Requests](pull-requests.md) - Code review workflow

---

## Workflow

```
┌─────────────────┐
│ Pull Request    │
│ (Draft → Main)  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Attempt Merge   │
│ (Section-Level) │
└────────┬────────┘
         │
         ├─ No Conflicts ──────────────────────────────┐
         │                                             │
         └─ Conflicts Detected ──────────┐            │
                                         ▼            ▼
                                 ┌─────────────────┐  ┌─────────────────┐
                                 │ AI Resolution   │  │ Auto-Merge      │
                                 │ (ConflictResolver)│  │ Complete        │
                                 └────────┬────────┘  └─────────────────┘
                                          │
                                          ▼
                                 ┌─────────────────┐
                                 │ Human Review    │
                                 │ (Approve/Reject)│
                                 └────────┬────────┘
                                          │
                                          ▼
                                 ┌─────────────────┐
                                 │ Merge Complete  │
                                 └─────────────────┘
```

---

## Conflict Detection & AI Integration

```go
package conflict

type ConflictResolver struct {
    gitOps   *gitops.GitOps
    aiClient *ai.Client
}

// ResolveConflicts attempts AI-assisted resolution
func (r *ConflictResolver) ResolveConflicts(pr *models.PullRequest) (*models.AIConflictResolution, error) {
    // 1. Attempt merge
    mergeResult, _ := r.gitOps.Merge(pr.RepoID, pr.SourceBranch, pr.TargetBranch)
    
    if mergeResult.Status != "conflict" {
        return nil, nil // No conflicts - auto-merged
    }
    
    // 2. For each conflict, get context
    resolutions := []SectionResolution{}
    
    for _, conflict := range mergeResult.Conflicts {
        // Get three versions
        baseContent, _ := r.gitOps.GetBlob(conflict.BaseSHA)
        sourceContent, _ := r.gitOps.GetBlob(conflict.SourceSHA)
        targetContent, _ := r.gitOps.GetBlob(conflict.TargetSHA)
        
        // 3. Ask AI to resolve
        resolution, confidence := r.resolveWithAI(conflict.Path, baseContent, sourceContent, targetContent)
        
        resolutions = append(resolutions, SectionResolution{
            SectionID:  conflict.Path,
            Resolution: resolution,
            Confidence: confidence,
        })
    }
    
    // 4. Create AI resolution record
    aiResolution := &models.AIConflictResolution{
        ConflictFiles: extractPaths(mergeResult.Conflicts),
        ProposedMerge: combineResolutions(resolutions),
        Confidence:    calculateOverallConfidence(resolutions),
        Reasoning:     generateReasoningReport(resolutions),
    }
    
    return aiResolution, nil
}

// resolveWithAI uses AI to merge conflicting sections
func (r *ConflictResolver) resolveWithAI(sectionID, base, source, target string) (string, float64) {
    prompt := fmt.Sprintf(`
You are an expert technical writer merging two versions of a document section.

**Section**: %s

**Base version**:
%s

**Version A (draft/source)**:
%s

**Version B (main/target)**:
%s

**Task**: Provide a merged version that intelligently combines both changes. Preserve intent from both edits. If changes conflict logically, favor Version B (main) unless Version A has clear improvements.

Output ONLY the merged markdown content, no explanations.
`, sectionID, base, source, target)
    
    response, _ := r.aiClient.Complete(ai.CompletionRequest{
        Model:       "gpt-4",
        Prompt:      prompt,
        Temperature: 0.3, // Low temperature for consistency
        MaxTokens:   2000,
    })
    
    // Calculate confidence based on similarity to both versions
    confidence := calculateConfidence(response.Text, source, target)
    
    return response.Text, confidence
}

// calculateConfidence estimates merge quality
func calculateConfidence(merged, source, target string) float64 {
    // Simple heuristic: ratio of preserved content from both
    sourcePreserved := levenshteinSimilarity(merged, source)
    targetPreserved := levenshteinSimilarity(merged, target)
    
    // High confidence if both are well-preserved
    return (sourcePreserved + targetPreserved) / 2.0
}
```

---

## AI Resolution Review UI

```javascript
// static/js/conflict-resolution.js

class ConflictReviewPanel {
    constructor(prID) {
        this.prID = prID;
        this.loadConflicts();
    }
    
    async loadConflicts() {
        const pr = await fetch(`/api/pull-requests/${this.prID}`).then(r => r.json());
        
        if (pr.ai_resolution) {
            this.renderResolution(pr.ai_resolution);
        }
    }
    
    renderResolution(resolution) {
        const html = `
            <div class="ai-resolution">
                <div class="notification is-warning">
                    <p><strong>AI Conflict Resolution</strong></p>
                    <p>Confidence: ${(resolution.confidence * 100).toFixed(0)}%</p>
                    <p>${resolution.reasoning}</p>
                </div>
                
                <div class="columns">
                    <div class="column">
                        <h3>Conflicts Detected</h3>
                        <ul>
                            ${resolution.conflict_files.map(f => `<li>${f}</li>`).join('')}
                        </ul>
                    </div>
                    
                    <div class="column">
                        <h3>Proposed Merge</h3>
                        <pre><code>${escapeHTML(resolution.proposed_merge)}</code></pre>
                    </div>
                </div>
                
                <div class="buttons">
                    <button class="button is-success" onclick="applyAIResolution()">
                        Apply AI Resolution
                    </button>
                    <button class="button is-danger" onclick="rejectAIResolution()">
                        Reject & Resolve Manually
                    </button>
                </div>
            </div>
        `;
        
        document.getElementById('conflict-panel').innerHTML = html;
    }
}

async function applyAIResolution() {
    const prID = getCurrentPRID();
    
    await fetch(`/api/pull-requests/${prID}/apply-ai-resolution`, {
        method: 'POST'
    });
    
    // Refresh PR view
    window.location.reload();
}
```

---

## AI Resolution Data Model

```go
// internal/models/pull_request.go

type AIConflictResolution struct {
    ConflictFiles []string  `json:"conflict_files"`   // Paths with conflicts
    ProposedMerge string    `json:"proposed_merge"`   // AI-generated merged content
    Confidence    float64   `json:"confidence"`       // 0.0-1.0 (merge quality estimate)
    Reasoning     string    `json:"reasoning"`        // AI explanation
    AppliedAt     time.Time `json:"applied_at,omitempty"` // If human approved
}

type PullRequest struct {
    // ... (existing fields)
    
    AIResolution *AIConflictResolution `json:"ai_resolution,omitempty"`
}
```

**Workflow**:
1. Merge fails with conflicts
2. AI generates resolution for each conflicting section
3. Human reviews AI proposal
4. Human approves → merge completes
5. Human rejects → manual conflict resolution

---

## Performance Considerations

### Section Caching

**Problem**: Parsing large documents on every merge is slow.

**Solution**: Cache parsed sections in memory/Redis.

```go
type SectionCache struct {
    cache *redis.Client
}

func (c *SectionCache) GetParsedDocument(sha string) (*sectiondocs.SectionedDocument, error) {
    // 1. Check cache
    cached, err := c.cache.Get(ctx, fmt.Sprintf("sections:%s", sha)).Result()
    if err == nil {
        var doc sectiondocs.SectionedDocument
        json.Unmarshal([]byte(cached), &doc)
        return &doc, nil
    }
    
    // 2. Parse document
    content, _ := gitops.GetBlob(sha)
    doc, _ := sectiondocs.ParseDocument(content)
    
    // 3. Cache result
    data, _ := json.Marshal(doc)
    c.cache.Set(ctx, fmt.Sprintf("sections:%s", sha), data, 1*time.Hour)
    
    return doc, nil
}
```

### AI Rate Limiting

**Problem**: Too many AI API calls during high conflict periods.

**Solutions**:
- Batch conflict resolutions (resolve multiple sections in one prompt)
- Use cheaper models for high-confidence cases (GPT-3.5 vs GPT-4)
- Cache AI resolutions by content hash
- Fallback to manual resolution if quota exceeded

---

## Future Enhancements

1. **Inline Conflict Markers**: Show `<<<<<<<` markers like Git for manual resolution
2. **Multi-Agent Collaboration**: Track which agent/user edited which section
3. **Section Locking**: Prevent concurrent edits to same section
4. **Diff Visualization**: Show side-by-side comparison in UI
5. **AI Training**: Learn from human conflict resolutions
6. **Custom Merge Strategies**: Per-document-type merge logic
7. **Real-Time Collaboration**: Operational Transform for live editing
8. **Version History Per Section**: Track section-level change history

---

## Related Documentation

- [Sectioned Documents](sectioned-documents.md) - Document format and section-level merging
- [Git Operations](git-operations.md) - Low-level Git implementation
- [Pull Requests](pull-requests.md) - Code review workflow
- [File Explorer](file-explorer.md) - File browsing and editing UI
- [AI Policy Layer](../../2-SoftwareDesignAndArchitecture/ai-policy-layer.md) - AI governance
