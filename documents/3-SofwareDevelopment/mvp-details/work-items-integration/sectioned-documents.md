# Sectioned Documents & Granular Merging

**Related Tasks**: MVP-WI-007 (Pull Requests), MVP-WI-011 (AI-Assisted Merging)  
**Status**: 📋 Planned

## Overview

This document describes the sectioned document format that enables granular merging by splitting documents into logical sections with independent versioning, reducing merge conflicts during collaborative editing.

**Related Documentation**:
- [AI Conflict Resolution](ai-conflict-resolution.md) - AI-assisted merge strategies for conflicts
- [Git Data Models](git-data-models.md) - Git object model
- [Git Core Operations](git-core-operations.md) - Low-level merge operations
- [Pull Requests](pull-requests.md) - Code review workflow

---

## Why Sectioned Documents?

**Problem**: Traditional file-level merging causes conflicts when multiple people edit different parts of the same document.

**Solution**: Split documents into logical sections with independent versioning.

**Benefits**:
- ✅ Reduced merge conflicts (section-level granularity)
- ✅ Parallel editing by multiple agents/users
- ✅ Preserve section metadata (author, timestamp)
- ✅ Clear audit trail per section

---

## Document Format

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

## Document Parsing

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

## Section-Level Three-Way Merge

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
   - Otherwise: **CONFLICT** (both modified) → Send to [AI Resolution](ai-conflict-resolution.md)
3. Reconstruct document from merged sections

---

## Related Documentation

- [AI Conflict Resolution](ai-conflict-resolution.md) - AI-powered conflict resolution workflow
- [Git Operations](git-operations.md) - Low-level Git merge implementation
- [Pull Requests](pull-requests.md) - Code review and merge workflow
- [Git-Based Document System](git-based-document-system.md) - Architecture overview
