package format_test

import (
	"testing"

	"github.com/dfaile/backstage-nobl9/internal/errors"
	"github.com/dfaile/backstage-nobl9/internal/nobl9"
)

func TestFormatMessage(t *testing.T) {
	template := &MessageTemplate{
		Type:    TypeHelp,
		Content: "Hello, {{.Name}}!",
		Data:    struct{ Name string }{"World"},
	}

	message, err := FormatMessage(template)
	if err != nil {
		t.Errorf("FormatMessage failed: %v", err)
	}
	if message != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", message)
	}
}

func TestFormatError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "validation error",
			err:      errors.NewValidationError("invalid input"),
			expected: "❌ Validation error: invalid input",
		},
		{
			name:     "not found error",
			err:      errors.NewNotFoundError("project not found"),
			expected: "🔍 Not found: project not found",
		},
		{
			name:     "conflict error",
			err:      errors.NewConflictError("project already exists"),
			expected: "⚠️ Conflict: project already exists",
		},
		{
			name:     "internal error",
			err:      errors.NewInternalError("internal server error"),
			expected: "💥 Internal error: internal server error",
		},
		{
			name:     "unknown error",
			err:      errors.New("unknown error"),
			expected: "❌ Error: unknown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := FormatError(tt.err)
			if message != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, message)
			}
		})
	}
}

func TestFormatSuccess(t *testing.T) {
	message := FormatSuccess("Operation completed")
	expected := "✅ Operation completed"
	if message != expected {
		t.Errorf("Expected '%s', got '%s'", expected, message)
	}
}

func TestFormatHelp(t *testing.T) {
	commands := []struct {
		Name        string
		Description string
	}{
		{"help", "Show available commands"},
		{"create-project", "Create a new project"},
	}

	message := FormatHelp(commands)
	if message == "" {
		t.Error("Expected non-empty help message")
	}
}

func TestFormatProjectCreate(t *testing.T) {
	project := &nobl9.Project{
		Name:        "test-project",
		Description: "Test project",
		Owner:       "test@example.com",
	}

	message := FormatProjectCreate(project)
	if message == "" {
		t.Error("Expected non-empty project creation message")
	}
}

func TestFormatRoleAssign(t *testing.T) {
	project := "test-project"
	users := []string{"user1@example.com", "user2@example.com"}

	message := FormatRoleAssign(project, users)
	if message == "" {
		t.Error("Expected non-empty role assignment message")
	}
}

func TestFormatPrompt(t *testing.T) {
	message := FormatPrompt("Please provide a description")
	expected := "❓ Please provide a description"
	if message != expected {
		t.Errorf("Expected '%s', got '%s'", expected, message)
	}
}

func TestFormatProgress(t *testing.T) {
	message := FormatProgress("Processing...")
	expected := "⏳ Processing..."
	if message != expected {
		t.Errorf("Expected '%s', got '%s'", expected, message)
	}
}

func TestFormatList(t *testing.T) {
	tests := []struct {
		name     string
		items    []string
		expected string
	}{
		{
			name:     "empty list",
			items:    []string{},
			expected: "No items found.",
		},
		{
			name:     "single item",
			items:    []string{"item1"},
			expected: "1. item1\n",
		},
		{
			name:     "multiple items",
			items:    []string{"item1", "item2", "item3"},
			expected: "1. item1\n2. item2\n3. item3\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := FormatList(tt.items)
			if message != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, message)
			}
		})
	}
}

func TestFormatTable(t *testing.T) {
	tests := []struct {
		name     string
		headers  []string
		rows     [][]string
		expected string
	}{
		{
			name:     "empty table",
			headers:  []string{"Name", "Age"},
			rows:     [][]string{},
			expected: "No data available.",
		},
		{
			name:    "single row",
			headers: []string{"Name", "Age"},
			rows: [][]string{
				{"John", "30"},
			},
			expected: "Name  Age  \n------  -----  \nJohn  30  \n",
		},
		{
			name:    "multiple rows",
			headers: []string{"Name", "Age", "City"},
			rows: [][]string{
				{"John", "30", "New York"},
				{"Jane", "25", "London"},
			},
			expected: "Name  Age  City      \n------  -----  ----------  \nJohn  30  New York  \nJane  25  London  \n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := FormatTable(tt.headers, tt.rows)
			if message != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, message)
			}
		})
	}
} 