package fileindex

import "time"

// FileIndex represents a file's current state for fast lookups
// Stored in file_index collection for O(1) path → blob resolution
type FileIndex struct {
	// ArangoDB fields
	Key string `json:"_key"` // File path: "/documents/requirements.md"
	ID  string `json:"_id"`  // Full document ID
	Rev string `json:"_rev"` // Document revision

	// Location
	RepoID     string `json:"repo_id"`     // Repository identifier
	Path       string `json:"path"`        // Full path: "/documents/requirements.md"
	Name       string `json:"name"`        // File/folder name: "requirements.md"
	ParentPath string `json:"parent_path"` // Parent directory: "/documents"
	IsDir      bool   `json:"is_dir"`      // True for directories

	// Current state (on main branch)
	BlobSHA  string `json:"blob_sha,omitempty"`  // Empty for directories
	TreeSHA  string `json:"tree_sha,omitempty"`  // Empty for files
	Size     int    `json:"size"`                // File size in bytes
	MimeType string `json:"mime_type,omitempty"` // MIME type for files

	// Metadata
	UpdatedBy string    `json:"updated_by"` // Last modifier (user or agent)
	UpdatedAt time.Time `json:"updated_at"` // Last modification timestamp
	Created   time.Time `json:"created"`    // Creation timestamp
}

// DirectoryEntry represents a file or folder in a directory listing
type DirectoryEntry struct {
	Name      string    `json:"name"`       // File/folder name
	Path      string    `json:"path"`       // Full path
	Type      string    `json:"type"`       // "file" or "directory"
	Size      int       `json:"size"`       // Size in bytes (0 for directories)
	MimeType  string    `json:"mime_type"`  // MIME type for files
	UpdatedBy string    `json:"updated_by"` // Last modifier
	UpdatedAt time.Time `json:"updated_at"` // Last modification
	BlobSHA   string    `json:"blob_sha"`   // Git blob SHA (for files)
	TreeSHA   string    `json:"tree_sha"`   // Git tree SHA (for directories)
}

// FileContent represents file content with metadata
type FileContent struct {
	Path      string    `json:"path"`       // Full path
	Content   string    `json:"content"`    // File content (plain text)
	SHA       string    `json:"sha"`        // Blob SHA
	Size      int       `json:"size"`       // Size in bytes
	MimeType  string    `json:"mime_type"`  // MIME type
	UpdatedBy string    `json:"updated_by"` // Last modifier
	UpdatedAt time.Time `json:"updated_at"` // Last modification
}
