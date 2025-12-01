package validation

import (
	"strings"
	"testing"

	"github.com/aosanya/CodeValdCortex/internal/agency/models"
)

func TestDeliverableValidator_ValidateNode(t *testing.T) {
	validator := NewDeliverableValidator()

	tests := []struct {
		name      string
		node      models.DeliverableNode
		depth     int
		wantError bool
		errorType error
	}{
		{
			name: "valid file node",
			node: models.DeliverableNode{
				ID:                 "file-1",
				Name:               "readme",
				Description:        "Project readme file",
				Type:               models.DeliverableTypeFile,
				FileExtension:      ".md",
				PromptInstructions: "Create a comprehensive README",
			},
			depth:     1,
			wantError: false,
		},
		{
			name: "valid folder node",
			node: models.DeliverableNode{
				ID:                 "folder-1",
				Name:               "documentation",
				Description:        "Documentation folder",
				Type:               models.DeliverableTypeFolder,
				PromptInstructions: "All documentation files",
			},
			depth:     1,
			wantError: false,
		},
		{
			name: "empty name - should fail",
			node: models.DeliverableNode{
				ID:            "file-2",
				Name:          "",
				Type:          models.DeliverableTypeFile,
				FileExtension: ".md",
			},
			depth:     1,
			wantError: true,
			errorType: ErrEmptyName,
		},
		{
			name: "file without extension - should fail",
			node: models.DeliverableNode{
				ID:   "file-3",
				Name: "readme",
				Type: models.DeliverableTypeFile,
			},
			depth:     1,
			wantError: true,
			errorType: ErrMissingFileExtension,
		},
		{
			name: "folder with extension - should fail",
			node: models.DeliverableNode{
				ID:            "folder-2",
				Name:          "docs",
				Type:          models.DeliverableTypeFolder,
				FileExtension: ".md",
			},
			depth:     1,
			wantError: true,
			errorType: ErrFolderWithExtension,
		},
		{
			name: "file with children - should fail",
			node: models.DeliverableNode{
				ID:            "file-4",
				Name:          "readme",
				Type:          models.DeliverableTypeFile,
				FileExtension: ".md",
				Children: []models.DeliverableNode{
					{
						ID:            "child-1",
						Name:          "child",
						Type:          models.DeliverableTypeFile,
						FileExtension: ".md",
					},
				},
			},
			depth:     1,
			wantError: true,
			errorType: ErrFileWithChildren,
		},
		{
			name: "invalid file extension - should fail",
			node: models.DeliverableNode{
				ID:            "file-5",
				Name:          "script",
				Type:          models.DeliverableTypeFile,
				FileExtension: ".js",
			},
			depth:     1,
			wantError: true,
			errorType: ErrInvalidFileExtension,
		},
		{
			name: "excessive depth - should fail",
			node: models.DeliverableNode{
				ID:            "file-6",
				Name:          "deep",
				Type:          models.DeliverableTypeFile,
				FileExtension: ".md",
			},
			depth:     models.MaxNestingDepth + 1,
			wantError: true,
			errorType: ErrExcessiveNesting,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateNode(&tt.node, tt.depth)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error but got nil")
					return
				}
				// Optionally check error type if specified
				if tt.errorType != nil && !isErrorType(err, tt.errorType) {
					t.Errorf("expected error type %v, got %v", tt.errorType, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestDeliverableValidator_ValidateTree(t *testing.T) {
	validator := NewDeliverableValidator()

	tests := []struct {
		name      string
		nodes     []models.DeliverableNode
		wantError bool
	}{
		{
			name:      "empty tree - valid",
			nodes:     []models.DeliverableNode{},
			wantError: false,
		},
		{
			name: "valid tree with nested structure",
			nodes: []models.DeliverableNode{
				{
					ID:                 "root-1",
					Name:               "docs",
					Type:               models.DeliverableTypeFolder,
					PromptInstructions: "Documentation folder",
					Children: []models.DeliverableNode{
						{
							ID:                 "file-1",
							Name:               "readme",
							Type:               models.DeliverableTypeFile,
							FileExtension:      ".md",
							PromptInstructions: "Project README",
						},
						{
							ID:   "folder-1",
							Name: "guides",
							Type: models.DeliverableTypeFolder,
							Children: []models.DeliverableNode{
								{
									ID:            "file-2",
									Name:          "installation",
									Type:          models.DeliverableTypeFile,
									FileExtension: ".md",
								},
							},
						},
					},
				},
			},
			wantError: false,
		},
		{
			name: "duplicate IDs - should fail",
			nodes: []models.DeliverableNode{
				{
					ID:            "file-1",
					Name:          "first",
					Type:          models.DeliverableTypeFile,
					FileExtension: ".md",
				},
				{
					ID:            "file-1", // Duplicate ID
					Name:          "second",
					Type:          models.DeliverableTypeFile,
					FileExtension: ".md",
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTree(tt.nodes)

			if tt.wantError && err == nil {
				t.Errorf("expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestDeliverableNode_ComputePath(t *testing.T) {
	tests := []struct {
		name       string
		node       models.DeliverableNode
		parentPath string
		want       string
	}{
		{
			name: "file at root",
			node: models.DeliverableNode{
				Name:          "readme",
				Type:          models.DeliverableTypeFile,
				FileExtension: ".md",
			},
			parentPath: "",
			want:       "readme.md",
		},
		{
			name: "folder at root",
			node: models.DeliverableNode{
				Name: "docs",
				Type: models.DeliverableTypeFolder,
			},
			parentPath: "",
			want:       "docs",
		},
		{
			name: "file in subfolder",
			node: models.DeliverableNode{
				Name:          "installation",
				Type:          models.DeliverableTypeFile,
				FileExtension: ".md",
			},
			parentPath: "docs/guides",
			want:       "docs/guides/installation.md",
		},
		{
			name: "nested folder",
			node: models.DeliverableNode{
				Name: "api",
				Type: models.DeliverableTypeFolder,
			},
			parentPath: "docs",
			want:       "docs/api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.node.ComputePath(tt.parentPath)
			if got != tt.want {
				t.Errorf("ComputePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeliverableNode_GetDepth(t *testing.T) {
	tests := []struct {
		name string
		node models.DeliverableNode
		want int
	}{
		{
			name: "single node",
			node: models.DeliverableNode{
				Name: "readme",
				Type: models.DeliverableTypeFile,
			},
			want: 1,
		},
		{
			name: "folder with one level of children",
			node: models.DeliverableNode{
				Name: "docs",
				Type: models.DeliverableTypeFolder,
				Children: []models.DeliverableNode{
					{Name: "file1", Type: models.DeliverableTypeFile},
					{Name: "file2", Type: models.DeliverableTypeFile},
				},
			},
			want: 2,
		},
		{
			name: "nested folders",
			node: models.DeliverableNode{
				Name: "root",
				Type: models.DeliverableTypeFolder,
				Children: []models.DeliverableNode{
					{
						Name: "level1",
						Type: models.DeliverableTypeFolder,
						Children: []models.DeliverableNode{
							{
								Name: "level2",
								Type: models.DeliverableTypeFolder,
								Children: []models.DeliverableNode{
									{Name: "file", Type: models.DeliverableTypeFile},
								},
							},
						},
					},
				},
			},
			want: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.node.GetDepth()
			if got != tt.want {
				t.Errorf("GetDepth() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to check if an error matches a specific error type
func isErrorType(err, target error) bool {
	if err == nil || target == nil {
		return false
	}
	// Use errors.Is or check if the error wraps the target
	return strings.Contains(err.Error(), target.Error())
}
