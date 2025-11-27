# File Explorer Implementation

**Related Tasks**: MVP-WI-006 (File Explorer API)  
**Status**: ✅ Complete (2025-11-26)

## Overview

The file explorer provides a familiar file/folder interface that completely abstracts Git mechanics from end users. Users interact with files and directories without needing to understand commits, branches, or merges.

---

## File Index (Fast Path Lookup)

For performance, we maintain a denormalized index of the current file tree state:

```go
type FileIndex struct {
    Key  string `json:"_key"`  // File path: "/documents/requirements.md"
    
    // Location
    RepoID   string `json:"repo_id"`
    Path     string `json:"path"`
    Name     string `json:"name"`
    ParentPath string `json:"parent_path"`  // "/documents"
    
    // Current state (on main branch)
    BlobSHA  string `json:"blob_sha"`
    Size     int    `json:"size"`
    MimeType string `json:"mime_type"`
    
    // Metadata
    UpdatedBy string    `json:"updated_by"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

**Why File Index?**
- ⚡ Fast directory listings (no tree traversal)
- ⚡ Quick file lookups by path
- ⚡ File metadata without Git operations
- ⚡ Supports search and filtering

**Consistency**:
- Updated on every commit to main branch
- Rebuilt from Git history if corrupted
- Single source of truth is still Git objects

---

## File Browser API

### List Directory Contents

```http
GET /api/v1/agencies/{agencyID}/instances/{instanceID}/files?path=/documents

Response:
{
  "path": "/documents",
  "entries": [
    {
      "name": "requirements",
      "type": "directory",
      "path": "/documents/requirements"
    },
    {
      "name": "README.md",
      "type": "file",
      "path": "/documents/README.md",
      "size": 1024,
      "mime_type": "text/markdown",
      "updated_by": "user-alice",
      "updated_at": "2025-11-26T10:00:00Z"
    }
  ]
}
```

### Read File Content

```http
GET /api/v1/agencies/{agencyID}/instances/{instanceID}/files/content?path=/documents/README.md

Response:
{
  "path": "/documents/README.md",
  "content": "# Project Documentation\n\nThis directory contains...",
  "sha": "abc123",
  "size": 1024,
  "updated_by": "user-alice",
  "updated_at": "2025-11-26T10:00:00Z"
}
```

### Update File (Creates Commit)

```http
PUT /api/v1/agencies/{agencyID}/instances/{instanceID}/files
Content-Type: application/json

{
  "path": "/documents/README.md",
  "content": "# Updated Documentation\n\nNew content...",
  "message": "Update README",  // Optional commit message
  "author": "user-alice"
}

Backend workflow:
1. Create blob with new content
2. Update tree to reference new blob
3. Create commit
4. Update branch ref
5. Update file index
```

### Create File

```http
POST /api/v1/agencies/{agencyID}/instances/{instanceID}/files
Content-Type: application/json

{
  "path": "/documents/new-file.md",
  "content": "# New File\n\nContent...",
  "author": "user-alice"
}
```

### Delete File

```http
DELETE /api/v1/agencies/{agencyID}/instances/{instanceID}/files?path=/documents/old-file.md
```

### Create Directory

```http
POST /api/v1/agencies/{agencyID}/instances/{instanceID}/files/directory
Content-Type: application/json

{
  "path": "/documents/new-folder",
  "author": "user-alice"
}
```

### Rebuild File Index

```http
POST /api/v1/agencies/{agencyID}/instances/{instanceID}/files/rebuild

Response:
{
  "message": "File index rebuilt successfully",
  "files_indexed": 42
}
```

---

<!-- MVP-WI-006 -->
## Implementation Status (MVP-WI-006)

**Completed**: 2025-11-26  
**Task**: [MVP-WI-006 File Explorer API](../../../mvp_done.md#mvp-wi-006-file-explorer-api)  
**Coding Session**: [MVP-WI-006_file_explorer_api.md](../../../coding_sessions/MVP-WI-006_file_explorer_api.md)

The file explorer API has been implemented with agency-specific database isolation using the InstanceRepository pattern. This provides the foundation for Git-backed file management.

### Key Implementation Details

1. **Multi-Tenant Architecture**: Each agency has its own ArangoDB database (e.g., `UC-CHAR-001`, `UC-INFRA-001`) with isolated data
2. **Collections**: `git_artifacts` (file/folder metadata), `git_objects`, `git_refs`, `repositories`
3. **Repository Pattern**: InstanceRepository using `driver.Client` with `agencyDB` parameter on all methods
4. **Lazy Collection Creation**: Collections created on-demand during first write operation
5. **ArangoDB Key Generation**: Path sanitization for compliance (alphanumeric, dash, underscore only)

### Architecture Stack

```
Handler (files/handler.go)
    ↓ calls
Service (fileindex/service.go)  
    ↓ calls  
Repository (fileindex/repository.go)
    ↓ queries
ArangoDB (agency-specific database)
```

### Key Files Created

- `internal/git/fileindex/repository.go` - Data access layer with lazy collection creation (353 LOC)
- `internal/git/fileindex/service.go` - Business logic for file operations (663 LOC)
- `internal/web/handlers/files/handler.go` - REST API endpoints (522 LOC)
- `static/js/file-browser.js` - Client-side file explorer UI (438 LOC)

### API Endpoints Implemented

- `GET /agencies/{agencyID}/instances/{instanceID}/files/explorer` - File explorer UI
- `GET /agencies/{agencyID}/instances/{instanceID}/files` - List directory contents
- `GET /agencies/{agencyID}/instances/{instanceID}/files/content` - Read file content
- `POST /agencies/{agencyID}/instances/{instanceID}/files` - Create file
- `PUT /agencies/{agencyID}/instances/{instanceID}/files` - Update file
- `DELETE /agencies/{agencyID}/instances/{instanceID}/files` - Delete file
- `POST /agencies/{agencyID}/instances/{instanceID}/files/directory` - Create folder
- `POST /agencies/{agencyID}/instances/{instanceID}/files/rebuild` - Rebuild file index

### Technical Challenges Solved

1. **Collection Missing**: Implemented lazy creation on write errors instead of proactive checks
2. **Query Field Mismatches**: Corrected field names (`repository_id` → `repo_id`, `type` → `is_dir`)
3. **ArangoDB Key Validation**: Enhanced `makeKey()` function to sanitize special characters
4. **Path Construction Bug**: Fixed `getCurrentPath()` in JavaScript to read from URL parameter first

### Validation Results

✅ `go vet ./...` - 0 issues  
✅ `go fmt ./...` - no changes needed  
✅ `templ generate` - successful  
✅ `make build` - successful  
✅ All debug logs removed  
✅ Multi-agency testing (UC-CHAR-001, UC-INFRA-001)

<!-- /MVP-WI-006 -->

---

## User Experience

### File Explorer UI

The file explorer presents a familiar folder tree interface:

```
📁 Root
├── 📁 documents
│   ├── 📄 requirements.md
│   ├── 📄 architecture.md
│   └── 📁 research
│       └── 📄 findings.md
└── 📁 code
    ├── 📄 main.go
    └── 📄 README.md
```

**Features**:
- Breadcrumb navigation
- Create folder/file buttons
- File search and filtering
- Right-click context menu
- Drag-and-drop support (future)

### Behind the Scenes

When user clicks "Save" on a file:

1. **Frontend**: POST to `/files` endpoint with content
2. **Handler**: Validates input, extracts agency/instance context
3. **Service**: Calls repository to update file
4. **Repository**: 
   - Creates Git blob with content
   - Updates Git tree
   - Creates commit
   - Updates branch reference
   - Updates file index
5. **Response**: Returns success with new SHA

User never sees Git terminology - just "File saved successfully"

---

## Performance Optimization

### File Index Benefits

- **Directory listing**: O(1) with index vs O(n) tree traversal
- **File lookup**: O(1) with index vs O(log n) tree walk
- **Search**: Full-text search on file index
- **Metadata**: No Git operations for size/type/modified

### Caching Strategy

```go
type FileCache struct {
    // Cache frequently accessed blobs
    blobCache *lru.Cache  // SHA → content
    
    // Cache tree objects
    treeCache *lru.Cache  // SHA → tree
    
    // Cache file index queries
    listCache *lru.Cache  // path → []FileIndex
}
```

**Cache invalidation**:
- Blob cache: Never (content-addressable)
- Tree cache: Never (immutable)
- List cache: On commit to affected directory

### Lazy Loading

- Load directory contents on-demand
- Stream large file content
- Paginate file history
- Throttle Git operations

---

## Future Enhancements

1. **File Search**: Full-text search across all files
2. **File History**: View commit history for single file
3. **Blame View**: See who edited each line
4. **Diff View**: Visual diff between versions
5. **Bulk Operations**: Upload/download multiple files
6. **Permissions**: File-level access control
7. **Webhooks**: Notify on file changes
8. **Real-time Collaboration**: Multiple users editing simultaneously

---

## Related Documentation

- [Git Operations](git-operations.md) - Low-level Git implementation
- [Collaborative Editing](collaborative-editing.md) - Sectioned documents and AI merging
- [Pull Requests](pull-requests.md) - Code review workflow
- [Kanban Workflow](kanban-workflow.md) - Issue tracking integration
