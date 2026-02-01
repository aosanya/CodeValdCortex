# Oversized Files Issue

**Last Updated**: 2026-02-01  
**Files Affected**: 70+ files  
**Status**: Open

## Issue Description

Multiple documentation files exceed maintainability thresholds:
- **70+ files** exceed 500 lines (warning threshold)
- **20+ files** exceed 1,000 lines (action required threshold)
- **1 file** has 3,107 lines (critical violation)

**Impact**: 
- Difficult to navigate and find information
- Poor readability and comprehension
- Merge conflicts more likely
- Hard to maintain and update
- Violates documentation best practices

## Severity Levels

| Lines | Severity | Action |
|-------|----------|--------|
| < 500 | ✅ Good | None |
| 500-1000 | 🟡 Warning | Consider splitting |
| 1000-1500 | 🟠 High | Should split |
| 1500+ | 🔴 Critical | Must split |

## Critical Violations (>1500 lines)

| File | Lines | Priority | Recommended Action |
|------|-------|----------|-------------------|
| `3-SofwareDevelopment/mvp_done.md` | 3,107 | 🔴 **CRITICAL** | Split into monthly archives |
| `3-SofwareDevelopment/core-features.md` | 2,639 | 🔴 **CRITICAL** | Split by feature category |
| `3-SofwareDevelopment/research/workflow-designer-status-bar-properties.md` | 1,840 | 🔴 **HIGH** | Split into logical sections |
| `2-SoftwareDesignAndArchitecture/a2a-integration/protocol-specification.md` | 1,831 | 🔴 **HIGH** | Split by protocol sections |
| `2-SoftwareDesignAndArchitecture/agency-operation-framework/multi-tenancy-org-model.md` | 1,606 | 🔴 **HIGH** | Split into architecture + implementation |
| `2-SoftwareDesignAndArchitecture/archive/frontend-architecture-react-deprecated.md` | 1,585 | ⚪ **LOW** | Already archived, can ignore |
| `1-SoftwareRequirements/requirements/use-cases/UC-CHAR-001-tumaini.md` | 1,557 | 🟡 **MEDIUM** | Use case can be longer |
| `2-SoftwareDesignAndArchitecture/agency-operation-framework/work-items.md` | 1,556 | 🔴 **HIGH** | Split into spec + examples |
| `3-SofwareDevelopment/mvp-details/MVP-052-New-Design.md` | 1,555 | 🔴 **HIGH** | Split into design + implementation |

## High Priority (1000-1500 lines)

| File | Lines | Recommended Action |
|------|-------|-------------------|
| `1-SoftwareRequirements/requirements/use-cases/UC-EVENT-001-events.md` | 1,442 | Split by event category |
| `2-SoftwareDesignAndArchitecture/agency-publishing-tagging-architecture.md` | 1,359 | Split into publishing + tagging |
| `2-SoftwareDesignAndArchitecture/agency-operation-framework/agent-states-fsm.md` | 1,267 | Split into states + transitions |
| `3-SofwareDevelopment/css_consolidation_research.md` | 1,155 | Move to research/ or archive |
| Multiple other files... | 1,000-1,200 | Evaluate case-by-case |

## Recommended Splitting Strategy

### 1. mvp_done.md (3,107 lines) → Split by Month

**Current**: Single massive file tracking all completed MVPs

**Proposed**:
```
3-SofwareDevelopment/mvp-done/
├── README.md (index of all completed work)
├── 2025-11/
│   ├── MVP-001_description.md
│   ├── MVP-002_description.md
│   └── ...
├── 2025-12/
│   ├── MVP-015_description.md
│   └── ...
└── 2026-01/
    ├── MVP-052_description.md
    └── ...
```

**Benefits**:
- Easy to find recent work
- Natural chronological organization
- Each file < 500 lines
- Clear completion timeline

### 2. core-features.md (2,639 lines) → Split by Feature

**Current**: All core features in one file

**Proposed**:
```
3-SofwareDevelopment/core-features/
├── README.md (overview and index)
├── agency-management.md
├── agent-orchestration.md
├── work-items.md
├── workflow-designer.md
├── ai-integration.md
└── ...
```

### 3. protocol-specification.md (1,831 lines) → Split by Section

**Current**: Entire A2A protocol in one file

**Proposed**:
```
2-SoftwareDesignAndArchitecture/a2a-integration/
├── README.md (protocol overview)
├── protocol-specification.md (reduced to overview)
├── message-formats.md
├── discovery-registration.md
├── task-delegation.md
├── security-compliance.md
└── go-sdk-integration.md (already exists)
```

### 4. work-items.md (1,556 lines) → Split Spec from Examples

**Proposed**:
```
2-SoftwareDesignAndArchitecture/agency-operation-framework/
├── work-items.md (specification only, ~600 lines)
└── work-items-examples.md (examples and use cases, ~800 lines)
```

## Action Plan

### Phase 1: Critical Files (This Week)

- [ ] Split `mvp_done.md` into monthly archives
  - Create directory structure
  - Move entries by completion date
  - Create index README
  - Update links in other docs

- [ ] Split `core-features.md` by feature category
  - Create feature subdirectory
  - Extract each feature to separate file
  - Create index README
  - Update cross-references

### Phase 2: High Priority Files (Next 2 Weeks)

- [ ] Split `protocol-specification.md` by major sections
- [ ] Split `work-items.md` into spec + examples
- [ ] Split `MVP-052-New-Design.md` into design + implementation
- [ ] Review and split other 1000+ line files

### Phase 3: Medium Priority (Next Month)

- [ ] Review all 500-1000 line files
- [ ] Split where logical boundaries exist
- [ ] Document guidelines for future file sizes
- [ ] Add pre-commit hook to warn on large files

## Guidelines for Future

### New Documentation Rules

1. **Target**: 200-500 lines per document
2. **Warning**: 500-1000 lines (consider splitting)
3. **Maximum**: 1,000 lines (must justify or split)

### When to Split

Split when:
- File exceeds 500 lines AND has logical sections
- File covers multiple distinct topics
- File mixes specification with examples
- File combines multiple time periods (historical + current)
- File is difficult to navigate

### When NOT to Split

Keep together when:
- Content is tightly coupled
- Splitting would harm comprehension
- File is reference material (like API spec)
- File is use case documentation (can be longer)

### Exceptions

Allow longer files for:
- Complete use case specifications (up to 1,500 lines)
- Comprehensive API documentation (up to 1,200 lines)
- Tutorial/walkthrough documents (up to 1,000 lines)
- Must be justified and approved

## Validation Checklist

After splitting:
- [ ] All cross-references updated
- [ ] Index/README created for split directories
- [ ] No broken links
- [ ] Content remains coherent
- [ ] Each new file follows naming conventions
- [ ] Git history preserved (use `git mv` where possible)

## Resolution

(To be filled when resolved)

- **Resolved By**: 
- **Date**: 
- **PR/Commit**: 
- **Files Split**: 
- **Notes**: 

## References

- Documentation standard: Max 500 lines recommended
- Industry best practice: 300-500 lines per document
- Related: [Naming Violations](naming-violations.md)
