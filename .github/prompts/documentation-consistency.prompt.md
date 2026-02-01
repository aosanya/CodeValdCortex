# Documentation Consistency & Organization Checker

## Purpose
Perform systematic documentation consistency checks through **one question at a time**, identifying outdated references, consolidating related files, and organizing documentation structure for maintainability.

---

## Instructions for AI Assistant

Conduct a comprehensive documentation consistency analysis through **iterative single-question exploration**. Ask ONE question at a time, wait for the response, then decide whether to:
- **🔍 DEEPER**: Go deeper into the same topic with follow-up questions
- **📝 NOTE**: Record an issue/gap for later action
- **➡️ NEXT**: Move to the next consistency check area
- **📊 REVIEW**: Summarize findings and determine next steps

The goal is to systematically check documentation consistency one area at a time rather than overwhelming with batch operations.

---

## Current Technology Stack (Reference)

**Update this section when stack changes:**

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

---

## Question-by-Question Consistency Check Process

## Question-by-Question Consistency Check Process

### Session Initiation

When starting a documentation consistency check:

1. **State the scope** - Which documentation area are we checking?
2. **Scan quickly** - Get overview of file structure and sizes
3. **Ask the first question** - Start with highest priority check
4. **Wait for user input** - Get confirmation or additional context before proceeding

### Question Flow

**After each answer, explicitly choose one of these paths:**

- 🔍 **DEEPER**: "Let me examine this area more closely..."
  - Investigate specific files flagged
  - Check related documents
  - Verify cross-references

- 📝 **NOTE**: "I'll note this inconsistency: [description]..."
  - Record issue for action list
  - Mark files needing updates
  - Continue to different check

- ➡️ **NEXT**: "Moving to [new consistency check area]..."
  - Current check complete
  - Proceed to next question category
  - Maintain systematic progress

- 📊 **REVIEW**: "Let me summarize what we've found..."
  - List inconsistencies discovered
  - Identify files to archive/update
  - Propose actions

---

## Question Categories (Execute in Order)

### Question 1: Technology Stack Consistency

**Objective**: Verify all documentation reflects current technology decisions.

**Question Format**:
```
🔍 [Technology Stack Check]

**Question**: Are there any references to [outdated technology] in the 
current documentation?

**Search Pattern**: [specific grep pattern or file locations]

**Context**: We migrated from [old tech] to [new tech], need to verify 
no outdated references remain.

**What I'm Looking For**: File paths with outdated references, or 
confirmation that area is clean.
```

**Common Searches**:
- "React" (unless in archived files or Riverpod context)
- "TypeScript frontend" (Flutter uses Dart)
- "SPA" or "Single Page Application" (Flutter is not SPA)
- "Templ" as primary UI (it's temporary)
- "HTMX" as strategic choice (it's temporary)

**For Each Match Found**:
- **🔍 DEEPER**: Read file to determine context
- **📝 NOTE**: Is it outdated? Archive or update?
- **➡️ NEXT**: No issues found, proceed to next check

---

### Question 2: Cross-Reference Validation

**Objective**: Ensure all internal document links point to current files.

**Question Format**:
```
🔍 [Link Validation Check]

**Question**: Do all links in [specific file or folder] point to existing, 
non-archived documents?

**Context**: Need to verify README and index files have current references.

**What I'm Looking For**: Broken links, links to archived files, or 
confirmation all links are valid.
```

**Actions Per Broken Link**:
- **🔍 DEEPER**: Find the correct replacement document
- **📝 NOTE**: Add to update list
- **➡️ NEXT**: All links valid, move to next file/folder

---

### Question 3: File Organization Analysis

**Objective**: Identify documentation that needs consolidation or subfolder organization.

**Question Format**:
```
🔍 [File Organization Check]

**Question**: Are there 3 or more files in [folder] that share a common 
topic prefix (e.g., "agency-*.md", "a2a-*.md")?

**Context**: Files with shared topics should be organized in subfolders 
for better discoverability.

**What I'm Looking For**: File groupings that exceed organization threshold.
```

**Organization Rules**:
- **≤ 2 files on topic**: Keep in current directory
- **3+ files on topic**: Create subfolder named after topic
- **500+ lines**: Consider splitting into smaller focused documents
- **Duplicate content**: Consolidate into single source of truth

**For Each Topic Group**:
- **🔍 DEEPER**: Should these be merged or kept in subfolder?
- **📝 NOTE**: Add to organization action list
- **➡️ NEXT**: No groupings found, continue

---

### Question 4: File Size Compliance

**Objective**: Ensure documents remain maintainable and scannable.

**Question Format**:
```
🔍 [File Size Check]

**Question**: Are there any .md files in [folder] exceeding 500 lines?

**Context**: Large files should be split for better maintainability.

**What I'm Looking For**: Files over 500 lines (warning) or 1000+ lines 
(action required).
```

**Size Guidelines**:
- **Ideal**: 200-500 lines per document
- **Warning**: 500-1000 lines (consider splitting)
- **Action Required**: 1000+ lines (must split or justify)

**For Each Large File**:
- **🔍 DEEPER**: Analyze structure - can it be split logically?
- **📝 NOTE**: Add to refactoring list with split strategy
- **➡️ NEXT**: All files within guidelines

---

### Question 5: Naming Convention Compliance

**Objective**: Ensure consistent, discoverable file naming.

**Question Format**:
```
🔍 [Naming Convention Check]

**Question**: Do all files in [folder] follow the naming convention 
pattern [expected-pattern.md]?

**Context**: Consistent naming improves discoverability and automation.

**What I'm Looking For**: Files violating naming conventions.
```

**Naming Standards**:
```yaml
Architecture Documents:
  Pattern: "kebab-case-descriptive-name.md"
  Examples: "backend-architecture.md", "a2a-protocol-integration.md"
  
MVP Details:
  Pattern: "MVP-XXX.md" or "MVP-XXX-descriptive-name.md"
  
Use Case Documentation:
  Pattern: "UC-ABBR-NNN-short-name/"
  
Coding Sessions:
  Pattern: "MVP-XXX_descriptive_name.md" or "TASK-NNN_description.md"
  
Archive Files:
  Pattern: "original-name-deprecated.md"
```

**For Each Violation**:
- **🔍 DEEPER**: What's the correct name per convention?
- **📝 NOTE**: Add to rename action list
- **➡️ NEXT**: All names compliant

---

### Question 6: Content Duplication Detection

**Objective**: Identify and consolidate duplicate or near-duplicate content.

**Question Format**:
```
🔍 [Duplication Check]

**Question**: Are there multiple files in [folder] covering the same topic 
or with similar content?

**Context**: Duplicate content creates maintenance burden and confusion.

**What I'm Looking For**: Files with overlapping purpose or >70% similar content.
```

**For Each Potential Duplicate**:
- **🔍 DEEPER**: Compare files to verify duplication level
- **📝 NOTE**: Decide merge strategy or keep with cross-references
- **➡️ NEXT**: No duplicates detected

---

### Question 7: Use Case MVP Files Check

**Objective**: Update use case-specific mvp.md files to reflect current platform architecture.

**Question Format**:
```
🔍 [Use Case Consistency Check]

**Question**: Does [use-case]/mvp.md reference the current technology stack 
(Flutter frontend, Templ as temporary)?

**Context**: Use case documentation should reflect actual platform state.

**What I'm Looking For**: Outdated tech stack references, deprecated 
dependency links, incorrect frontend descriptions.
```

**Update Strategy**:
1. Technology Stack Section - Remove React/TypeScript references
2. Architecture Diagrams - Update frontend layer labels
3. Dependencies Section - Remove MVP-RM-* (React migration) references
4. Implementation Notes - Clarify Backend (Cortex) vs Frontend (Fortex) separation

**For Each Use Case File**:
- **🔍 DEEPER**: What specific sections need updates?
- **📝 NOTE**: Add to use case update action list
- **➡️ NEXT**: Use case file is current

---

### Question 8: Production Readiness - Security & Authentication

**Objective**: Verify security documentation and implementation completeness for production deployment.

**Question Format**:
```
🔍 [Security Production Readiness Check]

**Question**: Is there comprehensive documentation covering authentication, 
authorization, secret management, and security hardening for production?

**Context**: Production systems require robust security measures to protect 
user data and prevent unauthorized access.

**What I'm Looking For**: Documentation gaps in:
- Authentication mechanisms (JWT, OAuth, API keys)
- Authorization/RBAC implementation
- Secrets management (environment variables, vaults)
- TLS/HTTPS configuration
- API rate limiting and throttling
- Input validation and sanitization
- Security headers and CORS policies
- Audit logging for security events
```

**Production Security Checklist**:
- ✅ Authentication flow documented
- ✅ Authorization/permissions model defined
- ✅ Secret rotation strategy documented
- ✅ Security testing procedures defined
- ✅ Incident response plan exists
- ✅ Data encryption at rest/in transit documented
- ✅ Vulnerability scanning process defined
- ✅ Security compliance requirements addressed

**For Each Gap**:
- **🔍 DEEPER**: Check implementation files for undocumented security features
- **📝 NOTE**: Add missing documentation to action list
- **➡️ NEXT**: Security documentation complete

---

### Question 9: Production Readiness - Monitoring & Observability

**Objective**: Ensure monitoring, logging, and alerting are production-ready.

**Question Format**:
```
🔍 [Monitoring Production Readiness Check]

**Question**: Is there documentation for production monitoring, logging 
infrastructure, metrics collection, and alerting strategies?

**Context**: Production systems require comprehensive observability to 
detect and resolve issues quickly.

**What I'm Looking For**: Documentation gaps in:
- Metrics collection (Prometheus, custom metrics)
- Logging infrastructure (structured logging, log aggregation)
- Distributed tracing (if microservices)
- Alerting rules and escalation policies
- Dashboard configurations
- SLI/SLO/SLA definitions
- Performance monitoring
- Error tracking and reporting
```

**Production Observability Checklist**:
- ✅ Metrics endpoints documented
- ✅ Log format and retention policies defined
- ✅ Critical alerts documented (SLIs)
- ✅ Dashboard designs specified
- ✅ On-call procedures documented
- ✅ Runbook for common issues exists
- ✅ Performance baselines established
- ✅ Error budget policy defined

**For Each Gap**:
- **🔍 DEEPER**: Check deployments/prometheus.yml and implementation
- **📝 NOTE**: Add missing observability documentation
- **➡️ NEXT**: Monitoring documentation complete

---

### Question 10: Production Readiness - Deployment & Infrastructure

**Objective**: Verify deployment procedures, infrastructure configuration, and disaster recovery plans.

**Question Format**:
```
🔍 [Deployment Production Readiness Check]

**Question**: Is there complete documentation for deployment processes, 
infrastructure as code, scaling strategies, and disaster recovery?

**Context**: Production deployments require reliable, repeatable processes 
and recovery mechanisms.

**What I'm Looking For**: Documentation gaps in:
- CI/CD pipeline configuration
- Infrastructure as Code (Terraform, k8s manifests)
- Environment configuration (dev/staging/prod)
- Database migration procedures
- Rollback procedures
- Scaling strategies (horizontal/vertical)
- Backup and restore procedures
- Disaster recovery plan (RTO/RPO)
- Blue-green or canary deployment strategy
```

**Production Deployment Checklist**:
- ✅ CI/CD pipeline documented
- ✅ Environment variables catalog exists
- ✅ Database migration runbook exists
- ✅ Rollback procedures documented
- ✅ Backup schedule and testing documented
- ✅ Infrastructure diagrams current
- ✅ Scaling thresholds defined
- ✅ DR plan tested and documented

**For Each Gap**:
- **🔍 DEEPER**: Check docker-compose.yml, Dockerfile, deployments/
- **📝 NOTE**: Add missing deployment documentation
- **➡️ NEXT**: Deployment documentation complete

---

### Question 11: Production Readiness - Data Management & Compliance

**Objective**: Ensure data handling, privacy, and compliance requirements are documented.

**Question Format**:
```
🔍 [Data Management Production Readiness Check]

**Question**: Is there documentation covering data models, database schemas, 
data retention policies, and regulatory compliance requirements?

**Context**: Production systems must handle data responsibly and comply 
with regulations (GDPR, CCPA, etc.).

**What I'm Looking For**: Documentation gaps in:
- Database schema documentation (ArangoDB collections)
- Data retention and archival policies
- PII (Personally Identifiable Information) handling
- GDPR/CCPA compliance procedures
- Data backup and recovery testing
- Database performance optimization
- Migration and upgrade procedures
- Data validation rules
```

**Production Data Management Checklist**:
- ✅ Schema documentation current
- ✅ Data retention policies defined
- ✅ PII handling documented
- ✅ Compliance requirements addressed
- ✅ Backup verification procedures exist
- ✅ Performance tuning guidelines documented
- ✅ Data migration tested
- ✅ Data access controls documented

**For Each Gap**:
- **🔍 DEEPER**: Check internal/database/ and compliance docs
- **📝 NOTE**: Add missing data management documentation
- **➡️ NEXT**: Data documentation complete

---

### Question 12: Production Readiness - API Documentation & Versioning

**Objective**: Verify API documentation is complete and production-ready for external consumers.

**Question Format**:
```
🔍 [API Production Readiness Check]

**Question**: Is there comprehensive API documentation including endpoints, 
request/response schemas, error codes, rate limits, and versioning strategy?

**Context**: Production APIs must be well-documented for developers and 
support teams.

**What I'm Looking For**: Documentation gaps in:
- OpenAPI/Swagger specification
- Authentication requirements per endpoint
- Request/response examples
- Error code catalog with resolution steps
- Rate limiting and quota documentation
- API versioning strategy (v1, v2, etc.)
- Deprecation policy and timeline
- Breaking change communication plan
```

**Production API Documentation Checklist**:
- ✅ API specification (OpenAPI/Swagger) exists
- ✅ Authentication per endpoint documented
- ✅ All endpoints have examples
- ✅ Error codes documented with meanings
- ✅ Rate limits clearly specified
- ✅ Versioning strategy documented
- ✅ Deprecation policy defined
- ✅ API changelog maintained

**For Each Gap**:
- **🔍 DEEPER**: Check api/ and internal/api/ folders
- **📝 NOTE**: Add missing API documentation
- **➡️ NEXT**: API documentation complete

---

### Question 13: Production Readiness - Testing & Quality Assurance

**Objective**: Ensure testing coverage and quality gates are production-ready.

**Question Format**:
```
🔍 [Testing Production Readiness Check]

**Question**: Is there documentation for test coverage requirements, testing 
strategies, and quality gates for production releases?

**Context**: Production code requires comprehensive testing to ensure 
reliability and prevent regressions.

**What I'm Looking For**: Documentation gaps in:
- Unit test coverage requirements (minimum %)
- Integration test strategy
- End-to-end test scenarios
- Performance/load testing procedures
- Security testing (SAST/DAST)
- Regression test suite
- Test data management
- Quality gates for CI/CD
```

**Production Testing Checklist**:
- ✅ Test coverage targets defined (e.g., 80%+)
- ✅ Integration test strategy documented
- ✅ E2E test scenarios identified
- ✅ Performance benchmarks established
- ✅ Security testing integrated
- ✅ Test data management documented
- ✅ CI/CD quality gates configured
- ✅ Testing runbook exists

**For Each Gap**:
- **🔍 DEEPER**: Check test/ folder and CI configuration
- **📝 NOTE**: Add missing testing documentation
- **➡️ NEXT**: Testing documentation complete

---

### Question 14: Production Readiness - Operations & Support

**Objective**: Verify operational runbooks, support procedures, and maintenance documentation exist.

**Question Format**:
```
🔍 [Operations Production Readiness Check]

**Question**: Is there documentation for operational procedures, troubleshooting 
guides, support escalation, and maintenance windows?

**Context**: Production systems require clear operational procedures for 
support teams.

**What I'm Looking For**: Documentation gaps in:
- Operational runbooks for common tasks
- Troubleshooting guides with decision trees
- Support escalation matrix
- Maintenance window procedures
- Health check endpoints documentation
- System dependencies map
- Contact information for critical services
- Change management process
```

**Production Operations Checklist**:
- ✅ Runbooks for common operations exist
- ✅ Troubleshooting guides documented
- ✅ Escalation procedures defined
- ✅ Maintenance procedures documented
- ✅ Health check endpoints listed
- ✅ System dependency map current
- ✅ Emergency contacts documented
- ✅ Change management process defined

**For Each Gap**:
- **🔍 DEEPER**: Check documents/4-QA/ and operations docs
- **📝 NOTE**: Add missing operational documentation
- **➡️ NEXT**: Operations documentation complete

---

## Response Processing Guidelines

**After receiving each answer:**

1. **Acknowledge**: "✅ Understood - [brief summary of finding]"
2. **Decide Path**: Choose DEEPER/NOTE/NEXT/REVIEW
3. **State Choice**: "🔍 Going deeper..." / "📝 Noting issue..." / "➡️ Moving to..." / "📊 Reviewing..."
4. **Execute**: Ask next question OR provide review summary

**Example Flow**:
```
AI: 🔍 [Technology Stack Check]
    
    Question: Are there references to "React" in 
    /documents/2-SoftwareDesignAndArchitecture/?
    
    Context: We migrated to Flutter, need to verify no outdated references.
    
    What I'm Looking For: Files containing "React" references.

User: Found 3 files: frontend-architecture.md, react-migration-plan.md, 
      README.md

AI: ✅ Found 3 files with React references.
    
    🔍 Going deeper - checking if these are already archived or current:
    
    Question: Have frontend-architecture.md and react-migration-plan.md 
    already been archived, or are they still active documents?

User: frontend-architecture.md is now just a deprecation notice. 
      react-migration-plan.md is still in main folder.

AI: ✅ Understood - one is deprecation notice (good), one needs archiving.
    
    📝 Noting issue: react-migration-plan.md needs to be archived to 
    archive/react-migration-plan-deprecated.md
    
    ➡️ Moving to README.md check:
    
    Question: What React references exist in README.md?
```

---

## Issue Tracking During Session

**Maintain a running list visible every 3-5 questions:**

### 🚨 Inconsistencies Found
- 📝 **[File]**: Outdated tech reference - [specific issue]
- 📝 **[File]**: Broken link - [link target]
- 📝 **[Folder]**: Needs subfolder organization - [topic group]

### ✅ Verified Clean
- ✅ **[Area]**: No issues found
- ✅ **[File]**: Already compliant

### 🔄 Actions Required
- 🔧 Archive: [list of files]
- 🔧 Update: [list of files needing edits]
- 🔧 Organize: [folders needing restructure]
- 🔧 Rename: [files needing rename]

---

## Periodic Review Format

**Every 5-7 questions, provide progress summary:**

```
📊 **CONSISTENCY CHECK - Progress Review**

**Areas Checked:**
✅ Technology Stack (2-SoftwareDesignAndArchitecture/) - 3 issues found
✅ Cross-References (README files) - 2 broken links
⏸️ File Organization - Not yet checked
⏸️ File Sizes - Not yet checked

**Issues Identified:**
📝 react-migration-plan.md needs archiving
📝 README.md has 2 React references to update
📝 introduction.md references "React Developer" role

**Files to Archive:**
- react-migration-plan.md → archive/react-migration-plan-deprecated.md

**Files to Update:**
- README.md (2 locations)
- introduction.md (1 location)

**Remaining Checks:**
- File organization analysis
- File size compliance
- Naming convention check
- Duplication detection
- Use case mvp.md updates

**Next Steps**: Continue with file organization check, or address found 
issues first?
```

---

## Completion Criteria

**The consistency check session is complete when:**

- ✅ All 14 question categories have been systematically checked
- ✅ All inconsistencies have been identified and noted
- ✅ Action list is complete (archive/update/organize/rename)
- ✅ Production readiness gaps documented
- ✅ User confirms readiness to execute actions

**Final Deliverable**:
```markdown
# Documentation Consistency & Production Readiness Report
**Date**: YYYY-MM-DD
**Scope**: [folders checked]

## Summary
- Files Scanned: XX
- Issues Found: XX
- Files to Archive: XX
- Files to Update: XX
- Reorganizations Needed: XX
- Production Gaps Identified: XX

## Technology Stack Issues
[List with file paths and specific line numbers/content]

## Broken Links
[List with file paths and broken link targets]

## File Organization
[List topic groups needing subfolder creation]

## File Sizes
[List files exceeding thresholds]

## Naming Violations
[List files needing rename]

## Duplicates Found
[List duplicate/near-duplicate content]

## Use Case Updates
[List use case mvp.md files needing updates]

## Production Readiness Gaps

### Security & Authentication
[List missing security documentation and implementation gaps]

### Monitoring & Observability
[List missing monitoring, logging, and alerting documentation]

### Deployment & Infrastructure
[List missing deployment and DR documentation]

### Data Management & Compliance
[List missing data policies and compliance documentation]

### API Documentation & Versioning
[List missing API documentation]

### Testing & Quality Assurance
[List missing test coverage and quality gate documentation]

### Operations & Support
[List missing runbooks and operational procedures]

## Recommended Actions
1. Archive: [commands to execute]
2. Update: [files and specific changes]
3. Organize: [folder restructure commands]
4. Rename: [rename commands]
5. Production Readiness: [critical documentation to create]

## Verification Steps
[How to confirm consistency after changes]

## Production Launch Blockers
**Critical Gaps** (Must be resolved before production):
- [ ] [Critical security gap]
- [ ] [Critical monitoring gap]
- [ ] [Critical deployment gap]

**High Priority** (Should be resolved before production):
- [ ] [High priority gap]

**Medium Priority** (Can be resolved post-launch with mitigation):
- [ ] [Medium priority gap]
```

---

## Example Question Sequence

**Session Start:**
```
AI: We're performing a documentation consistency check for 
    /workspaces/CodeValdCortex/documents/

    I've scanned the folder structure. Starting with Technology Stack Check.

    🔍 [Technology Stack Check - Question 1]
    
    **Question**: Are there any files in 
    documents/2-SoftwareDesignAndArchitecture/ that reference "React" 
    (excluding archive/ folder)?
    
    **Context**: We migrated from React to Flutter for the frontend. Need 
    to verify no outdated references remain in active documentation.
    
    **What I'm Looking For**: File paths containing "React" references, 
    or confirmation that the area is clean.
```

**After First Answer:**
```
User: Yes, found in frontend-architecture.md and README.md

AI: ✅ Found 2 files with React references.
    
    🔍 Going deeper into frontend-architecture.md:
    
    **Question**: Is frontend-architecture.md a current document describing 
    React architecture, or a deprecation notice pointing to the Flutter plan?
    
    **Context**: Need to determine if this is already handled or needs archiving.
    
    **What I'm Looking For**: File content/purpose - is it outdated content 
    or a redirect notice?
```

**After Progress (Question 5):**
```
AI: 📊 **PROGRESS CHECK**
    
    **Completed Checks:**
    ✅ Technology Stack - 3 issues identified
    ✅ Cross-References - 1 broken link found
    
    **Issues Found:**
    📝 react-migration-plan.md - needs archiving
    📝 README.md - 2 React references to update
    📝 docs link broken - points to archived file
    
    **Next Area**: File Organization Analysis
    
    Continue with organization check, or would you like to review/address 
    issues first?
```

---

## Success Criteria

**Documentation Consistency:**
- ✅ Zero references to outdated technologies in active docs
- ✅ All archived files have clear deprecation notices
- ✅ No broken internal links
- ✅ Topics with 3+ files organized in subfolders
- ✅ No files exceed 1500 lines without justification
- ✅ All use case mvp.md files reflect current architecture
- ✅ Comprehensive consistency report generated

**Production Readiness:**
- ✅ Security documentation complete (auth, secrets, hardening)
- ✅ Monitoring & alerting documented with SLIs/SLOs
- ✅ Deployment procedures and DR plans documented
- ✅ Data management and compliance requirements addressed
- ✅ API documentation production-ready (OpenAPI/Swagger)
- ✅ Testing coverage and quality gates defined
- ✅ Operational runbooks and troubleshooting guides exist
- ✅ No critical blockers for production deployment
- ✅ Production readiness checklist 100% complete
