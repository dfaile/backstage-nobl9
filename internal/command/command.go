package command

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/dfaile/backstage-nobl9/internal/errors"
)

// BotCommander defines the minimal interface for bot operations needed by commands
// (add methods as needed for command handlers)
type BotCommander interface {
	Commands() *CommandRegistry
	StartConversation(projectName string) error
	ValidateUser(email string) (bool, error)
	AssignRoles(project string, users []string) error
}

// Command represents a bot command
type Command struct {
	Name        string
	Aliases     []string
	Description string
	Usage       string
	Handler     func(BotCommander, []string) (string, error)
	Validate    func([]string) error
}

// CommandRegistry manages available commands
type CommandRegistry struct {
	commands map[string]*Command
	aliases  map[string]*Command
}

// NewCommandRegistry creates a new command registry
func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		commands: make(map[string]*Command),
		aliases:  make(map[string]*Command),
	}
}

// Register adds a command to the registry
func (r *CommandRegistry) Register(cmd *Command) {
	r.commands[cmd.Name] = cmd
	for _, alias := range cmd.Aliases {
		r.aliases[alias] = cmd
	}
}

// Get retrieves a command by name or alias
func (r *CommandRegistry) Get(name string) (*Command, bool) {
	if cmd, ok := r.commands[name]; ok {
		return cmd, true
	}
	if cmd, ok := r.aliases[name]; ok {
		return cmd, true
	}
	return nil, false
}

// List returns all registered commands
func (r *CommandRegistry) List() []*Command {
	cmds := make([]*Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		cmds = append(cmds, cmd)
	}
	return cmds
}

// FormatHelp formats the help message for available commands
func FormatHelp(commands []*Command) string {
	var sb strings.Builder
	w := tabwriter.NewWriter(&sb, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "Available commands:")
	fmt.Fprintln(w, "------------------")
	for _, cmd := range commands {
		fmt.Fprintf(w, "/%s\t%s\n", cmd.Name, cmd.Description)
		if cmd.Usage != "" {
			fmt.Fprintf(w, "  Usage: %s\n", cmd.Usage)
		}
	}
	w.Flush()

	return sb.String()
}

// FormatCommandHelp formats detailed help for a specific command
func FormatCommandHelp(cmd *Command) string {
	var sb strings.Builder
	w := tabwriter.NewWriter(&sb, 0, 0, 2, ' ', 0)

	fmt.Fprintf(w, "Command: /%s\n", cmd.Name)
	if len(cmd.Aliases) > 0 {
		fmt.Fprintf(w, "Aliases: %s\n", strings.Join(cmd.Aliases, ", "))
	}
	fmt.Fprintf(w, "Description: %s\n", cmd.Description)
	if cmd.Usage != "" {
		fmt.Fprintf(w, "Usage: %s\n", cmd.Usage)
	}
	w.Flush()

	return sb.String()
}

// HelpCommand provides information about available commands
func HelpCommand(b BotCommander, args []string) (string, error) {
	if len(args) > 0 {
		cmd, ok := b.Commands().Get(args[0])
		if !ok {
			return "", errors.NewValidationError(fmt.Sprintf("unknown command: %s", args[0]), nil)
		}
		return FormatCommandHelp(cmd), nil
	}

	return FormatHelp(b.Commands().List()), nil
}

// CreateProjectCommand handles project creation
func CreateProjectCommand(b BotCommander, args []string) (string, error) {
	if len(args) == 0 {
		return "", errors.NewValidationError("project name is required", nil)
	}

	// Start project creation conversation
	err := b.StartConversation(args[0])
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Starting project creation for '%s'. Please provide a description.", args[0]), nil
}

// AssignRoleCommand handles role assignment
func AssignRoleCommand(b BotCommander, args []string) (string, error) {
	if len(args) != 2 {
		return "", errors.NewValidationError("usage: assign-role <project> <user>", nil)
	}

	project, user := args[0], args[1]

	// Validate user
	exists, err := b.ValidateUser(user)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", errors.NewNotFoundError("user not found", nil)
	}

	// Assign role
	err = b.AssignRoles(project, []string{user})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("✅ Assigned roles in project '%s' for user: %s", project, user), nil
}

// ListProjectsCommand shows available projects
func ListProjectsCommand(b BotCommander, args []string) (string, error) {
	// TODO: Implement project listing when SDK is available
	return "⏳ Project listing not yet implemented", nil
}

// DefaultCommand handles unrecognized commands
func DefaultCommand(b BotCommander, args []string) (string, error) {
	return "❌ Error: unknown command. Type 'help' for available commands.", nil
}

// ParseCommand parses a command string into a Command struct
func ParseCommand(input string) (*Command, []string, error) {
	input = strings.TrimSpace(input)
	if input == "" || !strings.HasPrefix(input, "/") {
		return nil, nil, nil
	}

	fields := strings.Fields(input)
	if len(fields) == 0 {
		return nil, nil, errors.NewValidationError("invalid command format", nil)
	}

	name := strings.TrimPrefix(fields[0], "/")
	args := fields[1:]

	return &Command{
		Name: name,
	}, args, nil
} 