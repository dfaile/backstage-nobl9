package bot

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dfaile/backstage-nobl9/internal/command"
	"github.com/dfaile/backstage-nobl9/internal/errors"
	"github.com/dfaile/backstage-nobl9/internal/interactive"
	"github.com/dfaile/backstage-nobl9/internal/logging"
	"github.com/dfaile/backstage-nobl9/internal/nobl9"
	"github.com/dfaile/backstage-nobl9/internal/recovery"
)

// ConversationState represents the state of a conversation
type ConversationState struct {
	ProjectName        string
	ProjectDescription string
	Owner              string
	CreatedAt          time.Time
	UserRoles          map[string][]string // Map of user email to roles
	LastUpdated        time.Time
	Step               string
	PendingPrompt      interface{} // Can be *interactive.Prompt or *interactive.Confirmation
	CurrentStep        string
	RoleUser           string
	RoleType           string
}

// Bot represents the Nobl9 project bot
type Bot struct {
	nobl9Client *nobl9.Client
	logger      logging.Logger
	commands    *command.CommandRegistry
	state       map[string]*ConversationState
	mu          sync.RWMutex
}

// NewBot creates a new bot instance
func NewBot(nobl9Client *nobl9.Client, commands *command.CommandRegistry) *Bot {
	if commands == nil {
		commands = command.NewCommandRegistry()
		// Register default commands
		commands.Register(&command.Command{
			Name:        "help",
			Aliases:     []string{"h", "?"},
			Description: "Show available commands or help for a specific command",
			Usage:       "help [command]",
			Handler:     command.HelpCommand,
		})
		commands.Register(&command.Command{
			Name:        "create-project",
			Aliases:     []string{"create", "new"},
			Description: "Create a new Nobl9 project",
			Usage:       "create-project <name>",
			Handler:     command.CreateProjectCommand,
			Validate: func(args []string) error {
				if len(args) == 0 {
					return errors.NewValidationError("project name is required", nil)
				}
				return nil
			},
		})
		commands.Register(&command.Command{
			Name:        "assign-role",
			Aliases:     []string{"assign", "role"},
			Description: "Assign a role to a user in a project",
			Usage:       "assign-role <project> <user>",
			Handler:     command.AssignRoleCommand,
			Validate: func(args []string) error {
				if len(args) != 2 {
					return errors.NewValidationError("usage: assign-role <project> <user>", nil)
				}
				return nil
			},
		})
		commands.Register(&command.Command{
			Name:        "list-projects",
			Aliases:     []string{"list", "ls"},
			Description: "List available projects",
			Usage:       "list-projects",
			Handler:     command.ListProjectsCommand,
		})
	}

	logger, err := logging.NewLogger(logging.LevelInfo)
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}

	return &Bot{
		nobl9Client: nobl9Client,
		logger:      logger,
		commands:    commands,
		state:       make(map[string]*ConversationState),
	}
}

// Commands returns the command registry
func (b *Bot) Commands() *command.CommandRegistry {
	return b.commands
}

// SetCommands sets the command registry
func (b *Bot) SetCommands(commands *command.CommandRegistry) {
	b.commands = commands
}

// HandleMessage handles an incoming message and returns a response
func (b *Bot) HandleMessage(conversationID string, message string) (string, error) {
	ctx := context.WithValue(context.Background(), "conversation_id", conversationID)
	logger := b.logger.WithContext(ctx)

	state, exists := b.GetConversationState(conversationID)
	if !exists {
		logger.Error("Failed to get conversation state",
			logging.F("error", "conversation not found"),
		)
		return "", errors.NewNotFoundError("conversation not found", nil)
	}

	// Handle interactive responses
	if state.PendingPrompt != nil {
		response, err := b.handlePromptResponse(ctx, state, message)
		if err != nil {
			logger.Warn("Invalid response",
				logging.F("error", err),
				logging.F("message", message),
			)
			return fmt.Sprintf("Invalid response: %v. Please try again.", err), nil
		}
		return response, nil
	}

	// Handle commands
	cmd, args, err := command.ParseCommand(message)
	if err != nil {
		logger.Error("Failed to parse command",
			logging.F("error", err),
			logging.F("message", message),
		)
		return "", err
	}

	if cmd != nil {
		logger.Info("Handling command",
			logging.F("command", cmd.Name),
			logging.F("args", args),
		)
		return b.handleCommand(ctx, state, cmd, args)
	}

	// Handle default message
	return b.handleDefaultMessage(ctx, state, message)
}

func (b *Bot) handlePromptResponse(ctx context.Context, state *ConversationState, response string) (string, error) {
	logger := b.logger.WithContext(ctx)

	switch state.CurrentStep {
	case "project_name":
		// Validate project name with retry
		var available bool
		var validateErr error
		attempts := 0
		for {
			available, validateErr = b.ValidateProjectName(ctx, state.ProjectName, response)
			if validateErr == nil {
				break
			}
			if !recovery.ShouldRetry(validateErr, attempts) {
				logger.Error("Failed to validate project name",
					logging.F("error", validateErr),
					logging.F("attempts", attempts),
				)
				return "", validateErr
			}
			logger.Warn("Retrying project name validation",
				logging.F("error", validateErr),
				logging.F("attempts", attempts),
			)
			time.Sleep(recovery.GetRetryDelay(validateErr))
			attempts++
		}

		if !available {
			logger.Info("Project name not available",
				logging.F("project_name", response),
			)
			prompt := interactive.NewPrompt(
				"Project name is not available. Please choose another name:",
				nil,
				"",
			)
			state.PendingPrompt = prompt
			return prompt.Format(), nil
		}

		logger.Info("Project name validated",
			logging.F("project_name", response),
		)
		state.ProjectName = response
		state.CurrentStep = "project_description"

		// Prompt for project description
		prompt := interactive.NewPrompt(
			"Please provide a description for the project:",
			nil,
			"",
		)
		state.PendingPrompt = prompt
		return prompt.Format(), nil

	case "project_description":
		state.ProjectDescription = response
		state.CurrentStep = "confirm_creation"

		logger.Info("Project description received",
			logging.F("project_name", state.ProjectName),
			logging.F("description", response),
		)

		// Show confirmation
		confirm := interactive.NewConfirmation(
			fmt.Sprintf("Create project '%s' with description '%s'?", state.ProjectName, state.ProjectDescription),
			true,
		)
		state.PendingPrompt = confirm
		return confirm.Format(), nil

	case "confirm_creation":
		confirm, ok := state.PendingPrompt.(*interactive.Confirmation)
		if !ok {
			return "", fmt.Errorf("invalid prompt type: expected Confirmation")
		}

		confirmed, confirmErr := confirm.Validate(response)
		if confirmErr != nil {
			logger.Error("Invalid confirmation response",
				logging.F("error", confirmErr),
				logging.F("response", response),
			)
			return "", confirmErr
		}
		if !confirmed {
			logger.Info("Project creation cancelled",
				logging.F("project_name", state.ProjectName),
			)
			state.Reset()
			return "Project creation cancelled.", nil
		}

		// Create project with retry
		var projectErr error
		attempts := 0
		for {
			_, projectErr = b.CreateProject(state.ProjectName, state.ProjectDescription)
			if projectErr == nil {
				break
			}
			if !recovery.ShouldRetry(projectErr, attempts) {
				logger.Error("Failed to create project",
					logging.F("error", projectErr),
					logging.F("attempts", attempts),
				)
				return "", projectErr
			}
			logger.Warn("Retrying project creation",
				logging.F("error", projectErr),
				logging.F("attempts", attempts),
			)
			time.Sleep(recovery.GetRetryDelay(projectErr))
			attempts++
		}

		logger.Info("Project created",
			logging.F("project_name", state.ProjectName),
		)
		state.Reset()
		return "Project created successfully!", nil

	case "role_user":
		// Validate user with retry
		var exists bool
		var err error
		attempts := 0
		for {
			exists, err = b.ValidateUser(response)
			if err == nil {
				break
			}
			if !recovery.ShouldRetry(err, attempts) {
				logger.Error("Failed to validate user",
					logging.F("error", err),
					logging.F("attempts", attempts),
				)
				return "", err
			}
			logger.Warn("Retrying user validation",
				logging.F("error", err),
				logging.F("attempts", attempts),
			)
			time.Sleep(recovery.GetRetryDelay(err))
			attempts++
		}

		if !exists {
			logger.Info("User not found",
				logging.F("user", response),
			)
			state.PendingPrompt = interactive.NewPrompt(
				"User not found. Please enter a valid email:",
				nil,
				"",
			)
			return state.PendingPrompt.(*interactive.Prompt).Format(), nil
		}

		logger.Info("User validated",
			logging.F("user", response),
		)
		state.RoleUser = response
		state.CurrentStep = "role_type"

		// Prompt for role type
		state.PendingPrompt = interactive.NewPrompt(
			"Please select a role type:",
			[]string{"admin", "member", "viewer"},
			"member",
		)
		return state.PendingPrompt.(*interactive.Prompt).Format(), nil

	case "role_type":
		state.RoleType = response
		state.CurrentStep = "confirm_role"

		logger.Info("Role type received",
			logging.F("user", state.RoleUser),
			logging.F("role", response),
		)

		// Show confirmation
		confirm := interactive.NewConfirmation(
			fmt.Sprintf("Assign role '%s' to user '%s' in project '%s'?", state.RoleType, state.RoleUser, state.ProjectName),
			true,
		)
		state.PendingPrompt = confirm
		return confirm.Format(), nil

	case "confirm_role":
		confirm, ok := state.PendingPrompt.(*interactive.Confirmation)
		if !ok {
			return "", fmt.Errorf("invalid prompt type: expected Confirmation")
		}

		confirmed, confirmErr := confirm.Validate(response)
		if confirmErr != nil {
			logger.Error("Invalid confirmation response",
				logging.F("error", confirmErr),
				logging.F("response", response),
			)
			return "", confirmErr
		}
		if !confirmed {
			logger.Info("Role assignment cancelled",
				logging.F("user", state.RoleUser),
				logging.F("role", state.RoleType),
				logging.F("project", state.ProjectName),
			)
			state.Reset()
			return "Role assignment cancelled.", nil
		}

		// Assign role with retry
		var assignErr error
		attempts := 0
		for {
			assignErr = b.AssignRoles(state.ProjectName, []string{state.RoleUser})
			if assignErr == nil {
				break
			}
			if !recovery.ShouldRetry(assignErr, attempts) {
				logger.Error("Failed to assign role",
					logging.F("error", assignErr),
					logging.F("attempts", attempts),
				)
				return "", assignErr
			}
			logger.Warn("Retrying role assignment",
				logging.F("error", assignErr),
				logging.F("attempts", attempts),
			)
			time.Sleep(recovery.GetRetryDelay(assignErr))
			attempts++
		}

		logger.Info("Role assigned",
			logging.F("user", state.RoleUser),
			logging.F("role", state.RoleType),
			logging.F("project", state.ProjectName),
		)
		state.Reset()
		return "Role assigned successfully!", nil

	default:
		return "", fmt.Errorf("unknown step: %s", state.CurrentStep)
	}
}

func (b *Bot) handleCommand(ctx context.Context, state *ConversationState, cmd *command.Command, args []string) (string, error) {
	logger := b.logger.WithContext(ctx)

	switch cmd.Name {
	case "help":
		return command.HelpCommand(b, args)

	case "create-project":
		logger.Info("Starting project creation")
		state.Reset()
		state.CurrentStep = "project_name"
		prompt := interactive.NewPrompt(
			"Please enter a project name:",
			nil,
			"",
		)
		state.PendingPrompt = prompt
		return prompt.Format(), nil

	case "assign-role":
		if len(args) < 1 {
			logger.Warn("Missing project name in assign role command")
			return "Please specify a project name.", nil
		}
		logger.Info("Starting role assignment",
			logging.F("project", args[0]),
		)
		state.Reset()
		state.ProjectName = args[0]
		state.CurrentStep = "role_user"
		prompt := interactive.NewPrompt(
			"Please enter the user's email:",
			nil,
			"",
		)
		state.PendingPrompt = prompt
		return prompt.Format(), nil

	case "list-projects":
		return command.ListProjectsCommand(b, args)

	default:
		return command.DefaultCommand(b, args)
	}
}

func (b *Bot) handleDefaultMessage(ctx context.Context, state *ConversationState, message string) (string, error) {
	logger := b.logger.WithContext(ctx)

	switch state.CurrentStep {
	case "role_user":
		// Validate user with retry
		var exists bool
		var validateErr error
		attempts := 0
		for {
			exists, validateErr = b.ValidateUser(message)
			if validateErr == nil {
				break
			}
			if !recovery.ShouldRetry(validateErr, attempts) {
				logger.Error("Failed to validate user",
					logging.F("error", validateErr),
					logging.F("attempts", attempts),
				)
				return "", validateErr
			}
			logger.Warn("Retrying user validation",
				logging.F("error", validateErr),
				logging.F("attempts", attempts),
			)
			time.Sleep(recovery.GetRetryDelay(validateErr))
			attempts++
		}

		if !exists {
			logger.Info("User not found",
				logging.F("user", message),
			)
			prompt := interactive.NewPrompt(
				"User not found. Please enter a valid email:",
				nil,
				"",
			)
			state.PendingPrompt = prompt
			return prompt.Format(), nil
		}

		state.RoleUser = message
		state.CurrentStep = "role_type"

		// Prompt for role type
		prompt := interactive.NewPrompt(
			"Please select a role type:",
			[]string{"admin", "editor", "viewer"},
			"viewer",
		)
		state.PendingPrompt = prompt
		return prompt.Format(), nil

	case "role_type":
		// Validate role type
		validRoles := []string{"admin", "editor", "viewer"}
		valid := false
		for _, role := range validRoles {
			if strings.EqualFold(message, role) {
				valid = true
				state.RoleType = role
				break
			}
		}

		if !valid {
			logger.Info("Invalid role type",
				logging.F("role", message),
			)
			prompt := interactive.NewPrompt(
				"Invalid role type. Please select from: admin, editor, viewer",
				validRoles,
				"viewer",
			)
			state.PendingPrompt = prompt
			return prompt.Format(), nil
		}

		// Show confirmation
		confirm := interactive.NewConfirmation(
			fmt.Sprintf("Assign role '%s' to user '%s' in project '%s'?", state.RoleType, state.RoleUser, state.ProjectName),
			true,
		)
		state.PendingPrompt = confirm
		return confirm.Format(), nil

	case "confirm_role":
		confirm, ok := state.PendingPrompt.(*interactive.Confirmation)
		if !ok {
			return "", fmt.Errorf("invalid prompt type: expected Confirmation")
		}

		confirmed, confirmErr := confirm.Validate(message)
		if confirmErr != nil {
			logger.Error("Invalid confirmation response",
				logging.F("error", confirmErr),
				logging.F("response", message),
			)
			return "", confirmErr
		}
		if !confirmed {
			logger.Info("Role assignment cancelled",
				logging.F("project", state.ProjectName),
				logging.F("user", state.RoleUser),
			)
			state.Reset()
			return "Role assignment cancelled.", nil
		}

		// Assign role with retry
		var assignErr error
		attempts := 0
		for {
			assignErr = b.AssignRoles(state.ProjectName, []string{state.RoleUser})
			if assignErr == nil {
				break
			}
			if !recovery.ShouldRetry(assignErr, attempts) {
				logger.Error("Failed to assign role",
					logging.F("error", assignErr),
					logging.F("attempts", attempts),
				)
				return "", assignErr
			}
			logger.Warn("Retrying role assignment",
				logging.F("error", assignErr),
				logging.F("attempts", attempts),
			)
			time.Sleep(recovery.GetRetryDelay(assignErr))
			attempts++
		}

		logger.Info("Role assigned",
			logging.F("project", state.ProjectName),
			logging.F("user", state.RoleUser),
			logging.F("role", state.RoleType),
		)
		state.Reset()
		return "Role assigned successfully!", nil

	default:
		return "", fmt.Errorf("unknown step: %s", state.CurrentStep)
	}
}

// StartConversation starts a new conversation
func (b *Bot) StartConversation(projectName string) error {
	ctx := context.Background()
	
	// Check if project exists
	isValid, _, err := b.nobl9Client.ValidateProjectName(ctx, projectName)
	if err != nil {
		return fmt.Errorf("failed to validate project name: %w", err)
	}
	
	if !isValid {
		return errors.NewConflictError("project already exists", nil)
	}
	
	// Create new conversation state
	state := &ConversationState{
		ProjectName: projectName,
		Step:        "description",
		CreatedAt:   time.Now(),
	}

	b.state[projectName] = state
	return nil
}

// GetConversationState retrieves the state for a conversation
func (b *Bot) GetConversationState(threadID string) (*ConversationState, bool) {
	b.mu.RLock()
	state, exists := b.state[threadID]
	b.mu.RUnlock()
	
	if !exists {
		// Create a new state if it doesn't exist
		b.mu.Lock()
		defer b.mu.Unlock()
		
		// Double-check in case another goroutine created it
		state, exists = b.state[threadID]
		if !exists {
			state = &ConversationState{
				CreatedAt:    time.Now(),
				LastUpdated:  time.Now(),
				UserRoles:    make(map[string][]string),
				CurrentStep:  "",
				PendingPrompt: nil,
			}
			b.state[threadID] = state
		}
	}
	
	return state, true
}

// EndConversation ends a conversation
func (b *Bot) EndConversation(threadID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.state, threadID)
}

// CreateProject creates a new project
func (b *Bot) CreateProject(name, description string) (*nobl9.Project, error) {
	ctx := context.Background()
	return b.nobl9Client.CreateProject(ctx, name, description)
}

// ValidateUser checks if a user exists
func (b *Bot) ValidateUser(email string) (bool, error) {
	ctx := context.Background()
	return b.nobl9Client.ValidateUser(ctx, email)
}

// AssignRoles assigns roles to users in a project
func (b *Bot) AssignRoles(project string, users []string) error {
	ctx := context.Background()
	// Convert users slice to map with empty roles
	assignments := make(map[string][]string)
	for _, user := range users {
		assignments[user] = []string{"member"} // Default role
	}
	return b.nobl9Client.AssignRoles(ctx, project, assignments)
}

// ValidateProjectName checks if a project name is valid and available
func (b *Bot) ValidateProjectName(ctx context.Context, conversationID, name string) (bool, error) {
	isValid, _, err := b.nobl9Client.ValidateProjectName(ctx, name)
	return isValid, err
}

// UpdateConversationState updates the state of a conversation
func (b *Bot) UpdateConversationState(ctx context.Context, conversationID string, updateFn func(*ConversationState) error) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	state, exists := b.state[conversationID]
	if !exists {
		return errors.NewNotFoundError(fmt.Sprintf("conversation %s does not exist", conversationID), nil)
	}

	if err := updateFn(state); err != nil {
		return errors.NewInternalError("failed to update conversation state", err)
	}

	state.LastUpdated = time.Now()
	return nil
}

// GetUserRoles retrieves a user's roles in a project
func (b *Bot) GetUserRoles(ctx context.Context, conversationID, userEmail string) ([]string, error) {
	return b.nobl9Client.GetUserRoles(ctx, conversationID, userEmail)
}

// ValidateRoles checks if the roles are valid and not redundant
func (b *Bot) ValidateRoles(ctx context.Context, conversationID, userEmail string, newRoles []string) (bool, []string, error) {
	if conversationID == "" {
		return false, nil, fmt.Errorf("conversation ID is required")
	}
	
	state, exists := b.GetConversationState(conversationID)
	if !exists {
		return false, nil, fmt.Errorf("conversation not found")
	}
	
	if state.ProjectName == "" {
		return false, nil, fmt.Errorf("no project associated with conversation")
	}
	
	return b.nobl9Client.ValidateRoles(ctx, state.ProjectName, userEmail, newRoles)
}

// Ensure Bot implements command.BotCommander
var _ command.BotCommander = (*Bot)(nil)

func (s *ConversationState) Reset() {
	s.ProjectName = ""
	s.ProjectDescription = ""
	s.Owner = ""
	s.Step = ""
	s.PendingPrompt = nil
	s.CurrentStep = ""
	s.RoleUser = ""
	s.RoleType = ""
}

// New creates a new bot instance
func New(client *nobl9.Client) (*Bot, error) {
	commandRegistry := command.NewCommandRegistry()
	
	logger, err := logging.NewLogger(logging.LevelInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}
	
	return &Bot{
		nobl9Client: client,
		logger:      logger,
		commands:    commandRegistry,
		state:       make(map[string]*ConversationState),
	}, nil
}

// Start starts the bot and runs the interactive CLI
func (b *Bot) Start(ctx context.Context) error {
	fmt.Println("Welcome to Nobl9 Project Bot!")
	fmt.Println("Type 'help' for available commands or 'quit' to exit.")
	
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Bot shutting down...")
			return nil
		default:
			// Simple CLI interface for now
			fmt.Print("> ")
			var input string
			fmt.Scanln(&input)
			
			if input == "quit" || input == "exit" {
				return nil
			}
			
			// For now, just echo the input
			// In a real implementation, this would handle the message
			response, err := b.HandleMessage("cli", input)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println(response)
			}
		}
	}
} 