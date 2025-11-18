---
mode: agent
---

# Complete and Merge Current MVP Task

Follow the **mandatory completion process** for MVP tasks:

## Completion Process (MANDATORY)

1. **Create detailed coding session document** in `coding_sessions/` format: `{TaskID}_{description}.md`
   - Document all implementation details, decisions, and validation results
   - Include technical highlights, files created/modified, and dependencies unblocked

2. **Update architecture documentation if necessary**
   - If task introduced new architectural patterns, update `documents/2-SoftwareDesignAndArchitecture/`
   - Update relevant sections:
     - `general-architecture.md` for overall architecture changes
     - `frontend-architecture.md` for UI/template patterns
     - `backend-architecture.md` for data layer changes
   - Add new architecture decision records if significant design choices were made
   - Document new services, handlers, or repositories added

3. **Add completed task to `mvp_done.md`** with completion date
   - Include summary, key deliverables, technical highlights, validation results
   - List dependencies unblocked by this completion

4. **Remove completed task from active `mvp.md` file**
   - Strike through the completed MVP-XXX in dependency lists (~~MVP-XXX~~)

5. **Update dependent task references**
   - Update all tasks that depended on this one to show ~~MVP-XXX~~

6. **ALWAYS remove all debug logs before merge (MANDATORY)**
   
   **Backend Go Logs**:
   - Search for and remove all debug `fmt.Printf()`, `fmt.Println()` statements
   - Remove MVP-XXX prefixed debug logs: `fmt.Printf("MVP-XXX-DEBUG:`, `fmt.Printf("MVP-XXX-TRACE:`, etc.
   - Remove emoji-prefixed debug logs: `🔍 DEBUG [`, `📊 BEFORE UPDATE`, `💾 Saved workflow`, `🔹 Workflow[`, etc.
   - Remove detailed trace logs with object dumps and state inspection
   - Search patterns to check:
     - `grep -r "fmt.Printf" internal/ cmd/` (should only show essential production logs)
     - `grep -r "🔍\|📊\|💾\|🔹\|✅\|⚠️" internal/ cmd/` (emoji indicators often mean debug logs)
     - `grep -r "DEBUG \[" internal/ cmd/`
   
   **Frontend JavaScript Logs**:
   - Search for and remove all `console.log()`, `console.warn()` statements in JavaScript files
   - Search patterns to check:
     - `grep -r "console.log" static/js/`
     - `grep -r "console.warn" static/js/`
   - Keep only `console.error()` for actual error handling
   
   **General Rules**:
   - Remove TODO comments that reference debug logging
   - Keep only essential production logging (errors, critical warnings)
   - **Test the application after removing logs to ensure nothing breaks**
   - This is MANDATORY - no debug logs should remain in merged code

7. **Prepare next task (if applicable)**
   - Identify the next priority task from `mvp.md`
   - Check if `documents/3-SofwareDevelopment/mvp-details/MVP-XXX.md` exists for next task
   - **If details file doesn't exist for next task:**
     - Search `documents/2-SoftwareDesignAndArchitecture/` for relevant context
     - Review similar tasks in `documents/3-SofwareDevelopment/coding_sessions/`
     - Create new `MVP-XXX.md` in `mvp-details/` folder using the template:
       ```markdown
       # MVP-XXX: [Task Title]
       
       ## Overview
       **Priority**: [P0/P1/P2]  
       **Effort**: [Low/Medium/High]  
       **Skills Required**: [List skills]  
       **Dependencies**: [MVP-XXX, MVP-YYY]  
       **Status**: Not Started
       
       ## Description
       [Detailed description from mvp.md or architecture docs]
       
       ## Objectives
       - [Key objective 1]
       - [Key objective 2]
       
       ## Requirements
       [Functional and technical requirements]
       
       ## Acceptance Criteria
       - [ ] [Criterion 1]
       - [ ] [Criterion 2]
       
       ## Technical Specifications
       [Implementation details, architecture decisions]
       ```

8. **Fix all linting issues before merge**
   - Run `go vet ./...` and fix ALL errors and warnings (must show 0 issues)
   - Run `gofmt -w .` or `go fmt ./...` to ensure consistent code formatting
   - Run `golangci-lint run` if configured in project
   - Generate templ files: `templ generate` and ensure no errors
   - Use IDE quick fixes or manual resolution for all diagnostics
   - Common issues to address:
     - Unused imports or variables
     - Error handling (all errors must be checked)
     - Shadowed variables
     - Inefficient string concatenation
     - Missing documentation comments (if enabled)

9. **Merge to main after testing validation**
   - Ensure all debug logs removed (no fmt.Printf/Println, no console.log)
   - Ensure `go vet ./...` shows 0 issues
   - Ensure `gofmt` or `go fmt` has been run
   - Ensure `templ generate` completes successfully
   - All tests passing: `go test ./...` (if applicable)
   - Performance requirements met (if applicable)

## Git Workflow

```bash
# Before merge - validation and cleanup
go vet ./...       # Fix ALL issues until output is clean
go fmt ./...       # Format all Go files
templ generate     # Generate templ templates
go test ./...      # Run tests (if applicable)

# CRITICAL: Commit implementation code FIRST
git add internal/ cmd/ pkg/ static/
git commit -m "Implement MVP-XXX: [Description]

- Key implementation detail 1
- Key implementation detail 2
- Remove all debug print/log statements
- Fix all lint issues
"

# Then commit documentation updates
git add documents/ .github/
git commit -m "Complete MVP-XXX: Update task tracking and documentation"

# Merge when complete and tested
git checkout main
git merge feature/MVP-XXX_description --no-ff -m "Merge MVP-XXX: [Description]"
git branch -d feature/MVP-XXX_description
```

## Success Criteria
- ✅ Coding session document created in `documents/3-SofwareDevelopment/coding_sessions/`
- ✅ Architecture documentation updated in `documents/2-SoftwareDesignAndArchitecture/` (if needed, really try, approach changes during implementation.)
- ✅ Entry added to `mvp_done.md` with date and full details
- ✅ Task removed from active `mvp.md`
- ✅ Dependencies updated with strikethrough
- ✅ Next task details file created in `mvp-details/` (if missing)
- ✅ **ALWAYS: All debug logs removed (no fmt.Printf/Println, no console.log)**
- ✅ **Implementation code committed before documentation**
- ✅ All linting issues resolved (go vet shows 0 errors/warnings)
- ✅ Code formatted with go fmt
- ✅ Templ templates generated successfully
- ✅ All tests pass: `go test ./...` (if applicable)
- ✅ Performance requirements met (if applicable)
- ✅ Merged to main and feature branch deleted