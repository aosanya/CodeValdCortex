# Collaborative Editing & Conflict Resolution

**Related Tasks**: MVP-WI-007 (Pull Requests), MVP-WI-008 (Kanban Board)  
**Status**: 📋 Planned

## Overview

This document describes advanced features for collaborative document editing, including sectioned documents with granular merging and AI-assisted conflict resolution.

---

## Sectioned Document Integration

### Why Sectioned Documents?

**Problem**: Traditional file-level merging causes conflicts when multiple people edit different parts of the same document.

**Solution**: Split documents into logical sections with independent versioning.

**Benefits**:
- ✅ Reduced merge conflicts (section-level granularity)
- ✅ Parallel editing by multiple agents/users
- ✅ Preserve section metadata (author, timestamp)
- ✅ Clear audit trail per section

---

### Document Format

Documents stored as plain Markdown with YAML frontmatter:

```markdown
---
id: requirements-doc-001
type: requirements_document
sections:
  - id: introduction
    order: 1
  - id: functional-requirements
    order: 2
  - id: non-functional-requirements
    order: 3
---

<!-- section: introduction -->
# Introduction

This document outlines the requirements for...

<!-- section: functional-requirements -->
# Functional Requirements

## Authentication
- Users must be able to login with email/password
- Support OAuth 2.0

## Authorization
- Role-based access control
- Permissions per resource

<!-- section: non-functional-requirements -->
# Non-Functional Requirements

## Performance
- API response time < 200ms
- Support 1000 concurrent users
```

**Format Specification**:
- **YAML Frontmatter**: Document metadata (ID, type, section list)
- **Section Markers**: `<!-- section: {section-id} -->` comment tags
- **Section Content**: Markdown content between markers
- **Section Order**: Defined in frontmatter for reconstruction

---

### Document Parsing

```go
package sectiondocs

type SectionedDocument struct {
    Metadata map[string]interface{}
    Sections map[string]string // section_id → markdown content
}

// ParseDocument extracts sections from Markdown
func ParseDocument(content string) (*SectionedDocument, error) {
    doc := &SectionedDocument{
        Sections: make(map[string]string),
    }
    
    // 1. Extract YAML frontmatter
    parts := strings.Split(content, "---")
    if len(parts) >= 3 {
        yaml.Unmarshal([]byte(parts[1]), &doc.Metadata)
        content = strings.Join(parts[2:], "---")
    }
    
    // 2. Split by section markers
    sections := strings.Split(content, "<!-- section:")
    
    for _, section := range sections[1:] { // Skip preamble
        parts := strings.SplitN(section, "-->", 2)
        if len(parts) == 2 {
            sectionID := strings.TrimSpace(parts[0])
            sectionContent := strings.TrimSpace(parts[1])
            doc.Sections[sectionID] = sectionContent
        }
    }
    
    return doc, nil
}

// ReconstructDocument rebuilds Markdown from sections
func ReconstructDocument(sections map[string]string, metadata map[string]interface{}) string {
    var buf bytes.Buffer
    
    // 1. Write YAML frontmatter
    buf.WriteString("---\n")
    yamlData, _ := yaml.Marshal(metadata)
    buf.Write(yamlData)
    buf.WriteString("---\n\n")
    
    // 2. Get section order from metadata
    sectionList := metadata["sections"].([]interface{})
    
    // 3. Write sections in order
    for _, s := range sectionList {
        sectionMap := s.(map[string]interface{})
        sectionID := sectionMap["id"].(string)
        
        buf.WriteString(fmt.Sprintf("<!-- section: %s -->\n", sectionID))
        buf.WriteString(sections[sectionID])
        buf.WriteString("\n\n")
    }
    
    return buf.String()
}
```

---

### Section-Level Three-Way Merge

```go
package gitops

// MergeSectionedDocument performs intelligent section-level merge
func (g *GitOps) MergeSectionedDocument(baseSHA, sourceSHA, targetSHA string) (string, []Conflict, error) {
    // 1. Get file contents
    baseContent, _ := g.GetBlob(baseSHA)
    sourceContent, _ := g.GetBlob(sourceSHA)
    targetContent, _ := g.GetBlob(targetSHA)
    
    // 2. Parse into sections
    baseDoc, _ := sectiondocs.ParseDocument(baseContent)
    sourceDoc, _ := sectiondocs.ParseDocument(sourceContent)
    targetDoc, _ := sectiondocs.ParseDocument(targetContent)
    
    // 3. Merge at section level
    merged := make(map[string]string)
    conflicts := []Conflict{}
    
    allSections := getAllSectionIDs(baseDoc, sourceDoc, targetDoc)
    
    for _, sectionID := range allSections {
        baseText := baseDoc.Sections[sectionID]
        sourceText := sourceDoc.Sections[sectionID]
        targetText := targetDoc.Sections[sectionID]
        
        // Case 1: Both versions identical
        if sourceText == targetText {
            merged[sectionID] = sourceText
            continue
        }
        
        // Case 2: Only source modified
        if baseText == targetText && sourceText != baseText {
            merged[sectionID] = sourceText
            continue
        }
        
        // Case 3: Only target modified
        if baseText == sourceText && targetText != baseText {
            merged[sectionID] = targetText
            continue
        }
        
        // Case 4: Both modified - CONFLICT on this section
        conflicts = append(conflicts, Conflict{
            Path:      fmt.Sprintf("section:%s", sectionID),
            BaseSHA:   hashString(baseText),
            SourceSHA: hashString(sourceText),
            TargetSHA: hashString(targetText),
        })
    }
    
    if len(conflicts) > 0 {
        return "", conflicts, nil
    }
    
    // 4. Reconstruct document
    mergedContent := sectiondocs.ReconstructDocument(merged, baseDoc.Metadata)
    mergedSHA, _ := g.WriteBlob("", mergedContent)
    
    return mergedSHA, nil, nil
}

// Helper to get all unique section IDs
func getAllSectionIDs(base, source, target *sectiondocs.SectionedDocument) []string {
    seen := make(map[string]bool)
    var ids []string
    
    for id := range base.Sections {
        if !seen[id] {
            ids = append(ids, id)
            seen[id] = true
        }
    }
    for id := range source.Sections {
        if !seen[id] {
            ids = append(ids, id)
            seen[id] = true
        }
    }
    for id := range target.Sections {
        if !seen[id] {
            ids = append(ids, id)
            seen[id] = true
        }
    }
    
    return ids
}
```

**Merge Algorithm**:
1. Parse all three versions into sections
2. For each section ID:
   - If source == target: Use either version
   - If base == target: Use source (only source changed)
   - If base == source: Use target (only target changed)
   - Otherwise: **CONFLICT** (both modified)
3. Reconstruct document from merged sections

---

## AI-Assisted Conflict Resolution

### Workflow

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

### Conflict Detection & AI Integration

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

### AI Resolution Review UI

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

### AI Resolution Data Model

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

- [Git Operations](git-operations.md) - Low-level Git implementation
- [Pull Requests](pull-requests.md) - Code review workflow
- [File Explorer](file-explorer.md) - File browsing and editing UI
- [AI Policy Layer](../../2-SoftwareDesignAndArchitecture/ai-policy-layer.md) - AI governance
