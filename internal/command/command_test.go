package command_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/dfaile/backstage-nobl9/internal/bot"
	"github.com/dfaile/backstage-nobl9/internal/errors"
	"github.com/dfaile/backstage-nobl9/internal/nobl9"
)

func TestCommandRegistry(t *testing.T) {
	registry := NewCommandRegistry()

	// Test empty registry
	if len(registry.List()) != 0 {
		t.Error("Expected empty registry")
	}

	// Test command registration
	cmd := &Command{
		Name:        "test",
		Description: "Test command",
		Handler:     func(*bot.Bot, string) (string, error) { return "test", nil },
	}
	registry.Register(cmd)

	// Test command retrieval
	if got, ok := registry.Get("test"); !ok || got != cmd {
		t.Error("Failed to retrieve registered command")
	}

	// Test command listing
	cmds := registry.List()
	if len(cmds) != 1 || cmds[0] != cmd {
		t.Error("Failed to list registered commands")
	}
}

func TestHelpCommand(t *testing.T) {
	b := bot.NewBot(nil, nil)
	registry := NewCommandRegistry()
	registry.Register(&Command{
		Name:        "test",
		Description: "Test command",
		Handler:     func(*bot.Bot, string) (string, error) { return "test", nil },
	})
	b.SetCommands(registry)

	response, err := HelpCommand(b, "")
	if err != nil {
		t.Errorf("HelpCommand failed: %v", err)
	}
	if response == "" {
		t.Error("Expected non-empty help response")
	}
}

func TestCreateProjectCommand(t *testing.T) {
	b := bot.NewBot(nil, nil)

	// Test missing project name
	_, err := CreateProjectCommand(b, "")
	if !errors.IsValidationError(err) {
		t.Error("Expected validation error for missing project name")
	}

	// Test valid project creation
	response, err := CreateProjectCommand(b, "test-project")
	if err != nil {
		t.Errorf("CreateProjectCommand failed: %v", err)
	}
	if response == "" {
		t.Error("Expected non-empty response")
	}
}

func TestAssignRoleCommand(t *testing.T) {
	mockClient := &mockNobl9Client{}
	b := bot.NewBot(mockClient, nil)

	// Test invalid arguments
	_, err := AssignRoleCommand(b, "project")
	if !errors.IsValidationError(err) {
		t.Error("Expected validation error for invalid arguments")
	}

	// Test user not found
	_, err = AssignRoleCommand(b, "project nonexistent-user")
	if !errors.IsNotFoundError(err) {
		t.Error("Expected not found error for nonexistent user")
	}

	// Test successful role assignment
	response, err := AssignRoleCommand(b, "project test-user")
	if err != nil {
		t.Errorf("AssignRoleCommand failed: %v", err)
	}
	if response == "" {
		t.Error("Expected non-empty response")
	}
}

func TestDefaultCommand(t *testing.T) {
	b := bot.NewBot(nil, nil)
	response, err := DefaultCommand(b, "")
	if err != nil {
		t.Errorf("DefaultCommand failed: %v", err)
	}
	if response == "" {
		t.Error("Expected non-empty response")
	}
}

// mockNobl9Client implements nobl9.Client for testing
type mockNobl9Client struct{}

func (m *mockNobl9Client) GetProject(name string) (*nobl9.Project, error) {
	return nil, nil
}

func (m *mockNobl9Client) CreateProject(name, description string) (*nobl9.Project, error) {
	return nil, nil
}

func (m *mockNobl9Client) ValidateProjectName(name string) (bool, error) {
	return true, nil
}

func (m *mockNobl9Client) ValidateUser(email string) (bool, error) {
	if email == "nonexistent-user" {
		return false, nil
	}
	return true, nil
}

func (m *mockNobl9Client) AssignRoles(project string, users []string) error {
	return nil
}

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Command
		err      bool
	}{
		{
			name:  "Create command",
			input: "/create test-project",
			expected: &Command{
				Name: "create",
				Args: []string{"test-project"},
			},
			err: false,
		},
		{
			name:  "Assign command",
			input: "/assign test-project user@example.com",
			expected: &Command{
				Name: "assign",
				Args: []string{"test-project", "user@example.com"},
			},
			err: false,
		},
		{
			name:  "List command",
			input: "/list",
			expected: &Command{
				Name: "list",
				Args: []string{},
			},
			err: false,
		},
		{
			name:  "Help command",
			input: "/help",
			expected: &Command{
				Name: "help",
				Args: []string{},
			},
			err: false,
		},
		{
			name:  "Invalid command",
			input: "not a command",
			expected: nil,
			err: true,
		},
		{
			name:  "Empty command",
			input: "",
			expected: nil,
			err: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := ParseCommand(tt.input)
			if tt.err {
				assert.Error(t, err)
				assert.Nil(t, cmd)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected.Name, cmd.Name)
				assert.Equal(t, tt.expected.Args, cmd.Args)
			}
		})
	}
}

func TestValidateCommand(t *testing.T) {
	tests := []struct {
		name     string
		cmd      *Command
		expected error
	}{
		{
			name: "Valid create command",
			cmd: &Command{
				Name: "create",
				Args: []string{"test-project"},
			},
			expected: nil,
		},
		{
			name: "Invalid create command - missing args",
			cmd: &Command{
				Name: "create",
				Args: []string{},
			},
			expected: ErrMissingArgs,
		},
		{
			name: "Valid assign command",
			cmd: &Command{
				Name: "assign",
				Args: []string{"test-project", "user@example.com"},
			},
			expected: nil,
		},
		{
			name: "Invalid assign command - missing args",
			cmd: &Command{
				Name: "assign",
				Args: []string{"test-project"},
			},
			expected: ErrMissingArgs,
		},
		{
			name: "Valid list command",
			cmd: &Command{
				Name: "list",
				Args: []string{},
			},
			expected: nil,
		},
		{
			name: "Valid help command",
			cmd: &Command{
				Name: "help",
				Args: []string{},
			},
			expected: nil,
		},
		{
			name: "Invalid command name",
			cmd: &Command{
				Name: "invalid",
				Args: []string{},
			},
			expected: ErrUnknownCommand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCommand(tt.cmd)
			if tt.expected == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.expected, err)
			}
		})
	}
}

func TestFormatResponse(t *testing.T) {
	tests := []struct {
		name     string
		response *Response
		expected string
	}{
		{
			name: "Success response",
			response: &Response{
				Type:    ResponseTypeSuccess,
				Message: "Project created successfully",
			},
			expected: "✅ Project created successfully",
		},
		{
			name: "Error response",
			response: &Response{
				Type:    ResponseTypeError,
				Message: "Project already exists",
			},
			expected: "❌ Project already exists",
		},
		{
			name: "Info response",
			response: &Response{
				Type:    ResponseTypeInfo,
				Message: "Please provide a display name",
			},
			expected: "ℹ️ Please provide a display name",
		},
		{
			name: "Warning response",
			response: &Response{
				Type:    ResponseTypeWarning,
				Message: "Invalid project name",
			},
			expected: "⚠️ Invalid project name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted := FormatResponse(tt.response)
			assert.Equal(t, tt.expected, formatted)
		})
	}
}

func TestFormatList(t *testing.T) {
	tests := []struct {
		name     string
		items    []string
		expected string
	}{
		{
			name:     "Empty list",
			items:    []string{},
			expected: "No items found",
		},
		{
			name:     "Single item",
			items:    []string{"item1"},
			expected: "• item1",
		},
		{
			name:     "Multiple items",
			items:    []string{"item1", "item2", "item3"},
			expected: "• item1\n• item2\n• item3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted := FormatList(tt.items)
			assert.Equal(t, tt.expected, formatted)
		})
	}
}

func TestFormatHelp(t *testing.T) {
	expected := `Available commands:
• /create <project-name> - Create a new project
• /assign <project-name> <user-email> - Assign roles to a user
• /list - List all projects
• /help - Show this help message`

	formatted := FormatHelp()
	assert.Equal(t, expected, formatted)
} 