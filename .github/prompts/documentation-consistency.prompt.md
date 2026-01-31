# Documentation Consistency & Organization Checker

## Purpose
Perform systematic documentation consistency checks, identify outdated references, consolidate related files, and organize documentation structure for maintainability.

## Execution Workflow

### Step 1: Technology Stack Consistency Check

**Objective**: Verify all documentation reflects current technology decisions.

**Current Technology Stack** (Update this section when stack changes):
```yaml
Backend:
  Language: Go 1.21+
  Framework: Gin (REST APIs)
  Database: ArangoDB 3.11+
  Message Queue: Go channels + Redis
  
Frontend:
  Framework: Flutter 3.x (Dart)
  Project: CodeValdFortex
  Platforms: Web, iOS, Android, Desktop
  State Management: Riverpod 2.x
  HTTP Client: Dio 5.x
  Routing: go_router 13.x
  
Development UI (Temporary - To be removed):
  Framework: Templ + HTMX + Alpine.js
  Status: Will be removed via MVP-CLEANUP-001 through MVP-CLEANUP-014
  Note: NOT the primary frontend
```

**Actions**:
1. Search for outdated technology references:
   ```bash
   # Common outdated patterns to search for:
   - "React" (unless in archived files or Riverpod context)
   - "TypeScript frontend" (Flutter uses Dart)
   - "SPA" or "Single Page Application" (Flutter is not SPA)
   - "Templ" as primary UI (it's temporary)
   - "HTMX" as strategic choice (it's temporary)
   ```

2. For each match found:
   - **Context Check**: Determine if reference is appropriate (e.g., "reactive" is fine, "React framework" is not)
   - **Archive vs Update**: 
     - If entire document is outdated → Move to `archive/` folder with `-deprecated` suffix
     - If section is outdated → Update section to reflect current stack
     - If reference is historical context → Add note clarifying it's superseded
   
3. Create deprecation notices for archived major documents:
   ```markdown
   # [Original Document Title]
   
   **⚠️ DEPRECATION NOTICE**
   
   This document has been superseded by [new-document.md](new-document.md).
   
   **Reason**: [Brief explanation]
   
   **Archived Version**: [archive/original-document-deprecated.md](archive/original-document-deprecated.md)
   
   **Migration Date**: YYYY-MM-DD
   ```

### Step 2: Cross-Reference Validation

**Objective**: Ensure all internal document links point to current files.

**Actions**:
1. Extract all markdown links from documentation:
   ```bash
   grep -r "\[.*\](.*\.md)" documents/ --include="*.md"
   ```

2. For each link:
   - Verify target file exists
   - Check if target is in archive (dead link)
   - Validate link points to correct section (e.g., `#heading-anchor`)

3. Update broken links:
   - Replace links to archived files with current equivalents
   - Update README.md files that reference archived documents
   - Add redirect notices in archived files pointing to replacements

### Step 3: File Organization Analysis

**Objective**: Identify documentation that needs consolidation or subfolder organization.

**Organization Rules**:
- **≤ 2 files on topic**: Keep in current directory
- **3+ files on topic**: Create subfolder named after topic
- **500+ lines**: Consider splitting into smaller focused documents
- **Duplicate content**: Consolidate into single source of truth

**Actions**:
1. **Topic Detection**:
   ```bash
   # Group files by naming patterns
   ls documents/2-SoftwareDesignAndArchitecture/ | grep -E "^[a-z]+-" | cut -d'-' -f1 | sort | uniq -c
   
   # Example patterns:
   # - agency-*.md → "agency" topic
   # - a2a-*.md → "a2a" topic
   # - mvp-*.md → "mvp" topic
   ```

2. **Subfolder Creation** (when 3+ files share topic prefix):
   ```bash
   # Example for "agency" topic with 4+ files:
   mkdir -p documents/2-SoftwareDesignAndArchitecture/agency/
   mv agency-*.md documents/2-SoftwareDesignAndArchitecture/agency/
   # Create README.md in subfolder explaining organization
   ```

3. **Consolidation Analysis**:
   - Compare files with similar topics using semantic search
   - Identify overlapping content (≥30% similarity)
   - Propose consolidation strategy:
     - **Merge**: Combine into single comprehensive document
     - **Split**: Break large file into logical sections
     - **Reference**: Keep separate but add cross-references

4. **Update Index Files**:
   - Update parent folder README.md with subfolder references
   - Maintain table of contents accuracy
   - Add navigation breadcrumbs if needed

### Step 4: File Size Compliance

**Objective**: Ensure documents remain maintainable and scannable.

**Size Guidelines**:
- **Ideal**: 200-500 lines per document
- **Warning**: 500-1000 lines (consider splitting)
- **Action Required**: 1000+ lines (must split or justify)

**Actions**:
1. **Scan for oversized files**:
   ```bash
   find documents/ -name "*.md" -exec wc -l {} + | sort -n | awk '$1 > 500 {print}'
   ```

2. **For each large file** (500+ lines):
   - Analyze document structure and logical sections
   - Determine if splitting is beneficial:
     - **Split if**: Multiple distinct topics covered
     - **Keep if**: Single cohesive narrative, splitting would harm comprehension
   
3. **Splitting Strategy**:
   ```markdown
   # Original: large-document.md (1200 lines)
   
   # After split:
   large-topic/
   ├── README.md (overview, 150 lines)
   ├── section-1-foundation.md (400 lines)
   ├── section-2-implementation.md (500 lines)
   └── section-3-deployment.md (300 lines)
   ```

4. **Update References**:
   - Add clear README.md in new folder explaining structure
   - Update parent folder index
   - Add "See Also" sections for related documents

### Step 5: Naming Convention Compliance

**Objective**: Ensure consistent, discoverable file naming.

**Naming Standards**:
```yaml
Architecture Documents:
  Pattern: "kebab-case-descriptive-name.md"
  Examples: "backend-architecture.md", "a2a-protocol-integration.md"
  Avoid: CamelCase, snake_case, abbreviations without context

MVP Details:
  Pattern: "MVP-XXX.md" or "MVP-XXX-descriptive-name.md"
  Examples: "MVP-015.md", "MVP-A2A-000_a2a_go_sdk_integration.md"
  
Use Case Documentation:
  Pattern: "UC-ABBR-NNN-short-name/"
  Examples: "UC-INFRA-001-water-distribution-network/"

Coding Sessions:
  Pattern: "MVP-XXX_descriptive_name.md" or "TASK-NNN_description.md"
  Examples: "MVP-043_ai-status-chat-refresh-fix.md"

Archive Files:
  Pattern: "original-name-deprecated.md" or "original-name-YYYYMMDD-deprecated.md"
  Examples: "frontend-architecture-react-deprecated.md"
```

**Actions**:
1. Scan for naming violations
2. Propose renames for clarity
3. Update all references after rename
4. Maintain backward compatibility notes in README

### Step 6: Content Duplication Detection

**Objective**: Identify and consolidate duplicate or near-duplicate content.

**Actions**:
1. **Exact Duplication**:
   ```bash
   # Find files with identical content
   find documents/ -name "*.md" -type f -exec md5sum {} + | sort | uniq -w32 -D
   ```

2. **Near Duplication** (use semantic search):
   - Compare documents in same domain
   - Calculate similarity score
   - If similarity > 70%: Propose consolidation

3. **Consolidation Strategy**:
   - **Single Source of Truth**: Choose canonical document
   - **Archive Duplicates**: Move to archive with redirect notice
   - **Cross-Reference**: Add links from related documents

### Step 7: Generate Consistency Report

**Objective**: Document findings and actions taken.

**Report Format**:
```markdown
# Documentation Consistency Report
**Date**: YYYY-MM-DD
**Scope**: /workspaces/CodeValdCortex/documents

## Technology Stack Issues
- [X] React references: 12 found → 10 archived, 2 updated
- [X] TypeScript frontend: 5 found → all updated to Flutter/Dart
- [ ] Templ primary UI: 3 found → 2 need context clarification

## Broken Links
- [X] react-migration-plan.md: 3 references → updated to flutter-migration-plan.md
- [ ] MVP-RM-* files: 15 references → need review (in archived folder)

## File Organization
- [X] Created agency-operation-framework/ (8 files consolidated)
- [X] Created flutter-designs/ in Fortex (6 files)
- [ ] Consider: mvp-details/ subfolder structure (40+ files)

## File Sizes
- [ ] backend-architecture.md: 1992 lines → Consider splitting
- [ ] a2a-protocol-integration.md: 1831 lines → Consider splitting
- [X] frontend-architecture.md: 1586 lines → Archived (outdated)

## Duplicates Found
- [X] react-migration-plan.md (duplicate found and archived)
- [ ] MVP-015 content appears in 3 places → Need consolidation

## Actions Required
1. Review mvp-details/ organization (40+ files, needs topic subfolders)
2. Split large architecture documents (1500+ lines)
3. Clarify Templ references (temporary vs historical context)
4. Update use case mvp.md files with current stack references

## Files Archived (this session)
1. frontend-architecture-react-deprecated.md
2. react-migration-plan-deprecated.md
3. react-migration-plan-duplicate.md
4. FRONTEND_IMPLEMENTATION_SUMMARY-templ-deprecated.md
5. frontend-architecture-updated-templ-deprecated.md
6. mvp-details/archive-react-migration-deprecated/ (folder)

## Files Created
1. frontend-architecture.md (deprecation notice)
2. flutter-migration-plan.md (current strategy)

## Files Updated
1. 2-SoftwareDesignAndArchitecture/README.md
2. 2-general-architecture.md
3. 3-SofwareDevelopment/introduction.md
```

## Special Case: Use Case MVP Files

**Objective**: Update use case-specific mvp.md files to reflect current platform architecture.

**Use Case Files to Check**:
```bash
find usecases/ -name "mvp.md" -type f
```

**Update Strategy for Each Use Case**:

1. **Technology Stack Section**:
   - Remove React/TypeScript references
   - Update to reflect Flutter frontend (if UI is mentioned)
   - Ensure Templ is marked as temporary development UI only

2. **Architecture Diagrams**:
   - Update frontend layer: (React/TS) → (Flutter - Fortex)
   - Add note: "(Templ - temporary dev UI)" if Templ is shown

3. **Dependencies Section**:
   - Check for MVP-RM-* references (React migration) → Remove or mark deprecated
   - Update to MVP-FL-* references if frontend dependencies exist
   - Verify all MVP-CLEANUP-* dependencies are properly noted

4. **Implementation Notes**:
   - Clarify separation: Backend (Cortex) vs Frontend (Fortex)
   - Note that use cases are configuration-only (per usecase-architecture.md)
   - Remove any custom frontend implementation details

**Example Update for UC-INFRA-001**:
```markdown
## Technology Integration

### Backend (CodeValdCortex)
- **Framework**: Go + Gin
- **Database**: ArangoDB (graph + document storage)
- **Communication**: Change streams + message passing

### Frontend (CodeValdFortex)
- **Framework**: Flutter (cross-platform)
- **UI Components**: Material Design 3
- **State Management**: Riverpod
- **Note**: Templ-based dev UI exists temporarily, will be replaced

### Development Status
- ~~MVP-RM-001~~ ✅ (Deprecated - React migration cancelled)
- MVP-FL-003: Routing & Navigation → Required for dashboard
- MVP-CLEANUP-001 through 014: Remove Templ (blocked by Fortex MVP-FL-* tasks)
```

## Automation Recommendations

**Scripts to Create**:
1. `scripts/check-documentation-consistency.sh`:
   - Grep for outdated technology references
   - Validate all internal links
   - Report file size violations

2. `scripts/organize-documentation.sh`:
   - Detect topic groups (3+ files)
   - Propose subfolder organization
   - Generate consolidation reports

3. `scripts/update-use-case-mvps.sh`:
   - Batch update all use case mvp.md files
   - Replace deprecated references
   - Standardize technology stack sections

## Integration with Workflow Prompts

**Reference from start-task.prompt.md and finish-task.prompt.md**:
- After task completion, run documentation consistency checks
- Verify no new outdated references were introduced
- Update cross-references if documentation structure changed
- Maintain consistency report in session notes

## Success Criteria

- ✅ Zero references to outdated technologies (React, TypeScript frontend) in current docs
- ✅ All archived files have clear deprecation notices
- ✅ No broken internal links in active documentation
- ✅ Topics with 3+ files organized in subfolders
- ✅ No files exceed 1500 lines without justification
- ✅ All use case mvp.md files reflect current architecture
- ✅ Consistency report generated and reviewed
