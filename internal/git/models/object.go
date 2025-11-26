package models

import "time"

// GitObject represents a generic Git object stored in ArangoDB
// All Git objects (commits, trees, blobs) are stored in the git_objects collection
type GitObject struct {
	// ArangoDB fields
	Key string `json:"_key"` // SHA-1 hash of content
	ID  string `json:"_id"`  // Full document ID (git_objects/{sha})
	Rev string `json:"_rev"` // Document revision

	// Git object data
	Type    string `json:"type"`    // "commit", "tree", "blob"
	Size    int    `json:"size"`    // Size of content in bytes
	Content string `json:"content"` // Plain text content (for text files) or serialized object data

	// Metadata for queries
	RepoID  string    `json:"repo_id"` // Repository identifier
	Created time.Time `json:"created"` // Timestamp when object was created
}

// GitBlob represents a file's content
// Stored as GitObject with type="blob"
type GitBlob struct {
	SHA     string `json:"sha"`     // Content hash
	Content string `json:"content"` // File content (text or base64 for binary)
	Size    int    `json:"size"`    // Content size in bytes
}

// GitTree represents a directory structure
// Stored as GitObject with type="tree"
type GitTree struct {
	SHA     string      `json:"sha"`     // Tree hash
	Entries []TreeEntry `json:"entries"` // Files and subdirectories
}

// TreeEntry represents a file or directory in a tree
type TreeEntry struct {
	Mode string `json:"mode"` // "100644" (file), "100755" (executable), "040000" (directory)
	Type string `json:"type"` // "blob", "tree"
	SHA  string `json:"sha"`  // Object hash
	Path string `json:"path"` // File or directory name (not full path)
}

// GitCommit represents a version snapshot
// Stored as GitObject with type="commit"
type GitCommit struct {
	SHA       string    `json:"sha"`       // Commit hash
	Tree      string    `json:"tree"`      // Root tree SHA
	Parents   []string  `json:"parents"`   // Parent commit SHAs (empty for initial, 1 for normal, 2 for merge)
	Author    Author    `json:"author"`    // Author information
	Committer Author    `json:"committer"` // Committer information
	Message   string    `json:"message"`   // Commit message
	Timestamp time.Time `json:"timestamp"` // Commit timestamp
}

// Author represents a Git author or committer
type Author struct {
	Name  string    `json:"name"`  // "user-alice" or "agent-xyz"
	Email string    `json:"email"` // Email address
	When  time.Time `json:"when"`  // Timestamp
}

// GitRef represents a reference (branch, tag, or HEAD)
// Stored in git_refs collection
type GitRef struct {
	// ArangoDB fields
	Key string `json:"_key"` // "refs/heads/main", "refs/heads/feature-xyz", "HEAD"
	ID  string `json:"_id"`  // Full document ID
	Rev string `json:"_rev"` // Document revision

	// Reference data
	RepoID string `json:"repo_id"` // Repository identifier
	Type   string `json:"type"`    // "branch", "tag", "HEAD"
	Target string `json:"target"`  // Commit SHA (or ref path for symbolic refs like HEAD)
}

// Repository represents a Git repository container
// One repository per agency instance
type Repository struct {
	// ArangoDB fields
	Key string `json:"_key"`
	ID  string `json:"_id"`
	Rev string `json:"_rev"`

	// Repository metadata
	InstanceID  string `json:"instance_id"` // Linked agency instance
	Name        string `json:"name"`        // Repository name
	Description string `json:"description"` // Repository description
	DefaultRef  string `json:"default_ref"` // Default branch (typically "refs/heads/main")

	// File structure
	RootPath string `json:"root_path"` // Root path (typically "/")

	// Timestamps
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}
