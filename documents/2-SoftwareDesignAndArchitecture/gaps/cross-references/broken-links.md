# Broken Links Issue

**Last Updated**: 2026-02-01  
**Links Affected**: 10+ identified  
**Status**: Open

## Issue Description

Multiple README files contain broken internal links pointing to:
- Non-existent files marked as "pending"
- Directories instead of files
- Incorrect paths
- Archived or moved files

**Impact**: 
- Poor user experience navigating documentation
- Confusion about what exists vs. planned
- Wasted time following dead links
- Loss of trust in documentation accuracy

## Broken Links Identified

### 2-SoftwareDesignAndArchitecture/README.md

| Line | Link | Issue | Recommended Fix |
|------|------|-------|----------------|
| 12 | `[backend-architecture.md](backend-architecture.md)` | Links to directory, not file | Change to `[Backend Architecture](backend-architecture/README.md)` |
| 16 | `[agency-operation-framework/agency-operations-framework.md]` | File doesn't exist | Should be `agency-operation-framework/README.md` |
| 18 | `[a2a-integration/go-sdk-integration.md]` | File doesn't exist | Remove or create file |

### 3-SofwareDevelopment/README.md

| Line | Link | Issue | Recommended Fix |
|------|------|-------|----------------|
| 22 | `[deployment.md](deployment.md)` | Marked as *(pending)*, link broken | Remove link or create placeholder |
| 23 | `[future-features.md](future-features.md)` | Marked as *(pending)*, link broken | Remove link or create placeholder |
| 24 | `[maintenance.md](maintenance.md)` | Marked as *(pending)*, link broken | Remove link or create placeholder |

### Other README Files

Additional broken links found in:
- `3-SofwareDevelopment/mvp-details/README.md`
- `2-SoftwareDesignAndArchitecture/agency-operation-framework/README.md`
- `3-SofwareDevelopment/research/workflow-designer/README.md`

## Categories of Broken Links

### 1. Links to Directories Instead of Files

**Problem**: Markdown links to directories don't work
```markdown
[Backend Architecture](backend-architecture.md)  ❌
```

**Solution**: Link to the README within the directory
```markdown
[Backend Architecture](backend-architecture/README.md)  ✅
```

**Files to Fix**:
- [ ] `2-SoftwareDesignAndArchitecture/README.md` line 12

### 2. Links to Non-Existent "Pending" Files

**Problem**: Documentation lists links to files not yet created
```markdown
- **[deployment.md](deployment.md)**: Production deployment... *(pending)*  ❌
```

**Option 1**: Remove the link, keep description
```markdown
- **deployment.md** *(pending)*: Production deployment...  ✅
```

**Option 2**: Create placeholder file
```markdown
- **[deployment.md](deployment.md)**: Production deployment...  ✅
```
With `deployment.md`:
```markdown
# Deployment Documentation

**Status**: 🚧 Pending

This documentation is planned but not yet written.

## Planned Content
- Production deployment procedures
- Rollback strategies
...
```

**Files to Fix**:
- [ ] `3-SofwareDevelopment/README.md` lines 22-24

### 3. Links to Moved/Renamed Files

**Problem**: File was moved/renamed but links not updated

**Solution**: 
- Use global search to find all references
- Update all links
- Consider adding redirect note in old location (if applicable)

**Files to Audit**:
- [ ] All references to `agency-operations-framework.md`
- [ ] All references to moved archive files

### 4. Relative Path Errors

**Problem**: Incorrect relative path calculation

**Example**:
```markdown
# In file: 3-SofwareDevelopment/mvp-details/README.md
[Backend Architecture](../../2-SoftwareDesignAndArchitecture/backend-architecture.md)  ❌
# Should be:
[Backend Architecture](../../2-SoftwareDesignAndArchitecture/backend-architecture/README.md)  ✅
```

## Recommended Solutions

### Approach 1: Fix Immediately (Preferred)

For production documentation:
1. **Remove broken links** to pending files
2. **Fix directory links** to point to README files
3. **Verify all paths** are correct
4. **Test all links** manually or with link checker

### Approach 2: Create Placeholders

For planned documentation:
1. **Create placeholder files** with "Pending" status
2. **Include planned content outline** 
3. **Add to backlog** for completion
4. **Links become valid** immediately

## Action Plan

### Immediate (This Week)

- [ ] Audit all README files for broken links
- [ ] Fix broken links in main README files:
  - [ ] `2-SoftwareDesignAndArchitecture/README.md`
  - [ ] `3-SofwareDevelopment/README.md`
  - [ ] `1-SoftwareRequirements/README.md`
  - [ ] `4-QA/README.md`

### High Priority (Next Week)

- [ ] Audit all subdirectory README files
- [ ] Fix broken cross-references in documentation
- [ ] Create link checker script or GitHub Action
- [ ] Document linking standards

### Ongoing

- [ ] Run link checker on documentation changes
- [ ] Update broken links as files are created
- [ ] Establish policy: no links to non-existent files without placeholder

## Link Validation Tools

### Manual Validation
```bash
# Find all markdown links in a file
grep -o '\[.*\](.*\.md)' file.md

# Check if linked file exists
# (requires custom script)
```

### Automated Validation Options

1. **markdown-link-check** (npm package)
   ```bash
   npx markdown-link-check documents/**/*.md
   ```

2. **GitHub Action**
   - Add workflow to check links on PR
   - Fail PR if broken links detected

3. **Pre-commit Hook**
   - Check links before commit
   - Warn on broken links

## Documentation Linking Standards

### Standard to Establish

1. **Never link to non-existent files** without placeholder
2. **Always link to README.md** for directories
3. **Use relative paths** consistently
4. **Test links** before committing
5. **Update links** when moving/renaming files

### Link Format Examples

```markdown
# ✅ Good Examples
[Feature Documentation](../features/README.md)
[Specific Document](./subdirectory/document.md)
[External Link](https://example.com)

# ❌ Bad Examples  
[Missing File](nonexistent.md)  ← Link to file that doesn't exist
[Directory](../directory)  ← Link to directory without README
[Feature](feature.md) *(pending)*  ← Linked file marked pending
```

## Resolution Checklist

- [ ] All broken links identified
- [ ] All broken links fixed or removed
- [ ] All directory links point to README files
- [ ] All "pending" links either have placeholders or links removed
- [ ] Link checker tool configured
- [ ] Linking standards documented
- [ ] Team trained on linking standards

## Resolution

(To be filled when resolved)

- **Resolved By**: 
- **Date**: 
- **PR/Commit**: 
- **Links Fixed**: 
- **Tool Implemented**: 
- **Notes**: 

## References

- Markdown link syntax: `[Text](path/to/file.md)`
- Relative path guide: `../` = parent directory, `./` = current directory
- Related: [File Organization Issues](../file-organization/misplaced-files.md)
