package bot

import (
	"bufio"
	"context"
	"fmt"
	"os"
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

	logger, err := logging.NewLogger(logging.LevelWarn)
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

	// Get or create conversation state
	state, exists := b.GetConversationState(conversationID)
	if !exists {
		// Initialize new conversation state
		state = &ConversationState{
			UserRoles:   make(map[string][]string),
			LastUpdated: time.Now(),
		}
		b.mu.Lock()
		b.state[conversationID] = state
		b.mu.Unlock()
		
		// Welcome new user
		if strings.TrimSpace(message) == "" || message == "help" || message == "start" {
			return b.getWelcomeMessage(), nil
		}
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

	// Handle help command specifically
	if strings.ToLower(strings.TrimSpace(message)) == "help" {
		return b.getHelpMessage(), nil
	}

	// Handle commands
	if cmd, args := b.parseCommand(message); cmd != nil {
		logger.Info("Handling command",
			logging.F("command", cmd.Name),
			logging.F("args", args),
		)
		return b.handleCommand(ctx, state, cmd, args)
	}

	// Handle default message
	return b.handleNaturalLanguage(message), nil
}

// parseCommand parses user input into a command and arguments
func (b *Bot) parseCommand(input string) (*command.Command, []string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, nil
	}

	// Split into fields
	fields := strings.Fields(input)
	if len(fields) == 0 {
		return nil, nil
	}

	// Get command name (with or without /)
	cmdName := fields[0]
	if strings.HasPrefix(cmdName, "/") {
		cmdName = strings.TrimPrefix(cmdName, "/")
	}
	
	args := fields[1:]

	// Look up the command in the registry
	cmd, exists := b.commands.Get(cmdName)
	if !exists {
		return nil, nil
	}

	return cmd, args
}

// getWelcomeMessage returns a friendly welcome message
func (b *Bot) getWelcomeMessage() string {
	return `👋 Hello! I'm your Nobl9 Project Bot. I can help you:

🏗️  **Create new projects** - Just say "create project" or "new project"
👥 **Assign user roles** - Say "assign role" to manage user permissions
📋 **List projects** - Say "list projects" to see available projects

**Quick start:**
• Type "create project" to create a new Nobl9 project
• Type "help" anytime to see this message
• Type "quit" or "exit" to leave

What would you like to do?`
}

// getHelpMessage returns detailed help information
func (b *Bot) getHelpMessage() string {
	return `🤖 **Nobl9 Project Bot Help**

**Available Commands:**
• **create-project** (or "create", "new") - Create a new Nobl9 project
• **assign-role** (or "assign", "role") - Assign roles to users
• **list-projects** (or "list", "ls") - List available projects
• **help** - Show this help message

**Natural Language:**
You can also try saying things like:
• "I want to create a new project"
• "Create a project called my-service"
• "Help me make a project"
• "Assign roles to users"

**Examples:**
• create-project my-awesome-service
• assign-role my-project user@example.com

Type anything to get started!`
}

// handleNaturalLanguage tries to understand natural language input
func (b *Bot) handleNaturalLanguage(message string) string {
	msg := strings.ToLower(strings.TrimSpace(message))
	
	// Project creation keywords
	if strings.Contains(msg, "create") && (strings.Contains(msg, "project") || strings.Contains(msg, "new")) {
		return `Great! Let's create a new project. 

You can use the command: **create-project <name>**

For example:
• create-project my-service
• create-project analytics-dashboard

Or just type "create-project" and I'll guide you through it step by step.`
	}
	
	// Role assignment keywords
	if strings.Contains(msg, "assign") || strings.Contains(msg, "role") || strings.Contains(msg, "user") {
		return `I can help you assign roles to users!

Use the command: **assign-role <project> <user-email>**

For example:
• assign-role my-project user@example.com

Available roles include:
• Project Owner
• Project Editor  
• Project User`
	}
	
	// List projects keywords
	if strings.Contains(msg, "list") || strings.Contains(msg, "show") || strings.Contains(msg, "projects") {
		return `To see all available projects, use: **list-projects**`
	}
	
	// Default helpful response
	return fmt.Sprintf(`I'm not sure what you mean by "%s". 

Here are some things you can try:
• **create-project** - Create a new project
• **assign-role** - Assign user roles
• **list-projects** - List available projects  
• **help** - Show detailed help

Or try describing what you want to do in your own words!`, message)
}

func (b *Bot) handlePromptResponse(ctx context.Context, state *ConversationState, response string) (string, error) {
	logger := b.logger.WithContext(ctx)

	switch state.CurrentStep {
	case "project_selection":
		state.ProjectName = response
		state.CurrentStep = "role_user"
		
		logger.Info("Project selected for role assignment",
			logging.F("project", response),
		)
		
		prompt := interactive.NewPrompt(
			fmt.Sprintf("Please enter the user's email for project '%s':", response),
			nil,
			"",
		)
		state.PendingPrompt = prompt
		return prompt.Format(), nil
		
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

	// Special handling for commands that need interactive flows
	switch cmd.Name {
	case "create-project":
		if len(args) == 0 {
			// Start interactive flow
			logger.Info("Starting interactive project creation")
			state.Reset()
			state.CurrentStep = "project_name"
			prompt := interactive.NewPrompt(
				"Please enter a project name:",
				nil,
				"",
			)
			state.PendingPrompt = prompt
			return prompt.Format(), nil
		} else {
			// Project name provided, go to description step
			logger.Info("Starting project creation with name", logging.F("name", args[0]))
			state.Reset()
			state.ProjectName = args[0]
			state.CurrentStep = "project_description"
			prompt := interactive.NewPrompt(
				fmt.Sprintf("Please provide a description for project '%s':", args[0]),
				nil,
				"",
			)
			state.PendingPrompt = prompt
			return prompt.Format(), nil
		}

	case "assign-role":
		if len(args) == 0 {
			// Start interactive flow from project selection
			logger.Info("Starting interactive role assignment")
			state.Reset()
			state.CurrentStep = "project_selection"
			prompt := interactive.NewPrompt(
				"Please enter the project name:",
				nil,
				"",
			)
			state.PendingPrompt = prompt
			return prompt.Format(), nil
		} else if len(args) == 1 {
			// Project provided, ask for user
			logger.Info("Starting role assignment for project", logging.F("project", args[0]))
			state.Reset()
			state.ProjectName = args[0]
			state.CurrentStep = "role_user"
			prompt := interactive.NewPrompt(
				fmt.Sprintf("Please enter the user's email for project '%s':", args[0]),
				nil,
				"",
			)
			state.PendingPrompt = prompt
			return prompt.Format(), nil
		} else {
			// Both project and user provided, ask for role
			logger.Info("Starting role selection", 
				logging.F("project", args[0]),
				logging.F("user", args[1]),
			)
			state.Reset()
			state.ProjectName = args[0]
			state.RoleUser = args[1]
			state.CurrentStep = "role_type"
			prompt := interactive.NewPrompt(
				fmt.Sprintf("Please select a role for user '%s' in project '%s':", args[1], args[0]),
				[]string{"admin", "member", "viewer"},
				"member",
			)
			state.PendingPrompt = prompt
			return prompt.Format(), nil
		}

	case "list-projects":
		return command.ListProjectsCommand(b, args)

	default:
		// For other commands, call the handler directly
		if cmd.Handler != nil {
			return cmd.Handler(b, args)
		}
		return fmt.Sprintf("Command '%s' is not implemented yet", cmd.Name), nil
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
	
	// Register all commands
	commandRegistry.Register(&command.Command{
		Name:        "help",
		Aliases:     []string{"h", "?"},
		Description: "Show available commands or help for a specific command",
		Usage:       "help [command]",
		Handler:     command.HelpCommand,
	})
	
	commandRegistry.Register(&command.Command{
		Name:        "create-project",
		Aliases:     []string{"create", "new"},
		Description: "Create a new Nobl9 project",
		Usage:       "create-project [name]",
		Handler:     command.CreateProjectCommand,
	})
	
	commandRegistry.Register(&command.Command{
		Name:        "assign-role",
		Aliases:     []string{"assign", "role"},
		Description: "Assign a role to a user in a project",
		Usage:       "assign-role [project] [user]",
		Handler:     command.AssignRoleCommand,
	})
	
	commandRegistry.Register(&command.Command{
		Name:        "list-projects",
		Aliases:     []string{"list", "ls"},
		Description: "List available projects",
		Usage:       "list-projects",
		Handler:     command.ListProjectsCommand,
	})
	
	logger, err := logging.NewLogger(logging.LevelWarn)
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
	fmt.Println(b.getWelcomeMessage())
	fmt.Println()
	
	// Initialize conversation state for CLI
	response, err := b.HandleMessage("cli", "start")
	if err == nil && response != "" {
		// Don't print the welcome message twice
	}
	
	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		select {
		case <-ctx.Done():
			fmt.Println("\n👋 Thanks for using Nobl9 Project Bot! Goodbye!")
			return nil
		default:
			fmt.Print("> ")
			
			if !scanner.Scan() {
				if scanner.Err() != nil {
					fmt.Printf("❌ Error reading input: %v\n", scanner.Err())
				}
				continue
			}
			
			input := strings.TrimSpace(scanner.Text())
			if input == "" {
				continue
			}
			
			if input == "quit" || input == "exit" {
				fmt.Println("👋 Thanks for using Nobl9 Project Bot! Goodbye!")
				return nil
			}
			
			response, err := b.HandleMessage("cli", input)
			if err != nil {
				fmt.Printf("❌ Error: %v\n\n", err)
			} else {
				fmt.Printf("%s\n\n", response)
			}
		}
	}
} 