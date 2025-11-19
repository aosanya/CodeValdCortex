# MVP-048: AI Policy Save and Retrieve Debug Fix

**Date:** November 19, 2025  
**Task ID:** MVP-048  
**Branch:** feature/MVP-048_ai_policy_foundation  
**Status:** ✅ Complete

## Overview

Fixed critical bug preventing AI policies from being saved and retrieved in the Agency Designer. The issue was a template interpolation error where the agency ID was not being properly passed from the templ template to JavaScript code.

## Problem Statement

Users were unable to save or retrieve AI policies through the Agency Designer interface. The browser console showed a 404 error, and backend logs revealed that the agency ID was being sent as the literal string `"{ currentAgency.ID }"` instead of the actual agency ID value.

### Error Messages
```
[Error] Failed to load resource: the server responded with a status of 404 (Not Found) (summary, line 0)
[Log] Policy summary response: 404
[Log] Policy summary not found
ERRO Repository GetSpecification failed agency_id="{ currentAgency.ID }" error="failed to get agency: agency not found: { currentAgency.ID }"
```

## Root Cause Analysis

The issue was in `/workspaces/CodeValdCortex/internal/web/pages/agency_designer/agency_designer_overview.templ`:

1. **Incorrect template interpolation in JavaScript**: The template had inline JavaScript with `{ currentAgency.ID }` which is not valid templ syntax inside `<script>` tags
2. **Missing data attribute**: The agency ID wasn't being passed via a data attribute for JavaScript to access

### Before (Broken Code)
```javascript
<script>
    async function loadPolicySummary() {
        const response = await fetch(`/api/agencies/{ currentAgency.ID }/policy/summary`);
        // ...
    }
</script>
```

## Solution Implemented

### 1. Added Data Attribute to Container
Modified the `AIPolicyContent` templ function to include the agency ID as a data attribute:

```go
templ AIPolicyContent(currentAgency *models.Agency) {
    <div class="overview-section" data-agency-id={ currentAgency.ID }>
        // ...
    </div>
}
```

### 2. Updated JavaScript to Read from Data Attribute
Modified the JavaScript to extract the agency ID from the DOM:

```javascript
<script>
    async function loadPolicySummary() {
        const aiPolicySection = document.getElementById('content-ai-policy');
        const agencyId = aiPolicySection.querySelector('.overview-section').dataset.agencyId;
        const response = await fetch(`/api/agencies/${agencyId}/policy/summary`);
        // ...
    }
</script>
```

### 3. Removed Compliance Frameworks UI (Bonus)
As requested, removed the compliance frameworks checkboxes section from the AI Policy Wizard:
- Removed SOC 2, GDPR, HIPAA, and ISO 27001 checkbox UI elements
- Backend logic remains intact for future use

## Debugging Process

Following `.github/prompts/debug.prompt.md`, added strategic debug prints to trace the issue:

### Backend Debug Points (ai_policy_handler.go)
- ShowPolicyWizard: Entry point, agency retrieval, policy conversion
- GetPolicy: Specification retrieval, policy existence check
- SavePolicy: JSON parsing, validation, conversion, persistence
- Conversion functions: Data transformation between models

### Frontend Debug Points (ai-policy-wizard.js)
- init(): Container detection, agency ID extraction, existing policy loading
- loadExistingPolicy(): Policy data processing, field initialization
- savePolicy(): Request preparation, response handling

All debug logs were systematically removed after identifying and fixing the root cause.

## Files Modified

### Template Files
- `/workspaces/CodeValdCortex/internal/web/pages/agency_designer/agency_designer_overview.templ`
  - Added `data-agency-id` attribute to overview-section div
  - Updated JavaScript to read agency ID from data attribute

- `/workspaces/CodeValdCortex/internal/web/pages/agency_designer/ai_policy_wizard.templ`
  - Removed compliance frameworks checkboxes section

### No Code Changes Required
The handler and JavaScript files already had correct logic. The issue was purely in the template interpolation.

## Testing & Validation

### Build Validation
```bash
✅ make build - Success
✅ go fmt ./... - Formatted 1 file
✅ go vet ./... - No errors
✅ templ generate - Complete, 0 updates
```

### Functional Testing
- ✅ Agency Designer loads correctly
- ✅ AI Policy wizard displays without errors
- ✅ Policy summary API endpoint uses correct agency ID
- ✅ No 404 errors in console
- ✅ Template variables properly interpolated

## Technical Highlights

### Best Practice: Data Attributes for Template-to-JavaScript Communication
This fix reinforces the pattern established in the codebase:
- Use templ syntax `{ variable }` for HTML attributes only
- Pass dynamic data to JavaScript via `data-*` attributes
- Read from DOM in JavaScript, never embed template variables in `<script>` blocks

### Examples in Codebase
```go
// ai_policy_wizard.templ
data-agency-id={ currentAgency.ID }
data-existing-policy={ templ.JSONString(existingPolicy) }

// workflow_designer.templ
data-agency-id={ agencyID }

// chat_panel.templ
data-agency-id={ agencyID }
```

## Dependencies & Impact

### Dependencies Resolved
This fix unblocks:
- AI Policy configuration in Agency Designer
- Policy summary display
- Policy CRUD operations

### No Breaking Changes
- All existing functionality preserved
- Backward compatible with existing agencies
- UI cleanup (removed compliance checkboxes) is non-breaking

## Lessons Learned

1. **Template Interpolation Rules**: Templ variables work in HTML attributes but not in JavaScript code blocks
2. **Debugging Strategy**: Strategic debug logging helped trace data flow from template → handler → service
3. **Data Attributes Pattern**: Consistently use data attributes for passing server-side data to client-side JavaScript
4. **Clean Debug Logs**: Always remove debug prints before merging (automated with sed in finish-task process)

## Next Steps

The AI Policy foundation is now stable. Related tasks that can now proceed:
- MVP-049: AI Policy Evaluation Engine
- MVP-050: Policy Violation Tracking
- MVP-051-053: Compliance Framework Agents (SOC2, GDPR, HIPAA, ISO27001)

## Commit History

```bash
# Implementation committed with debug fix
git add internal/web/pages/agency_designer/
git commit -m "Fix MVP-048: Correct template interpolation for agency ID in AI Policy

- Add data-agency-id attribute to policy section container
- Update JavaScript to read agency ID from data attribute
- Remove compliance frameworks checkboxes from wizard UI
- Fix 404 errors when saving/retrieving AI policies"
```

---

**Completion Checklist:**
- ✅ Root cause identified and fixed
- ✅ Debug logs added for troubleshooting
- ✅ All debug logs removed
- ✅ Code formatted (go fmt)
- ✅ No lint errors (go vet)
- ✅ Templates regenerated
- ✅ Build successful
- ✅ Functional testing complete
- ✅ Documentation updated
