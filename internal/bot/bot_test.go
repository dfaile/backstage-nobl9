package bot

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"backstage-nobl9/internal/command"
	"backstage-nobl9/internal/errors"
	"backstage-nobl9/internal/interactive"
	"backstage-nobl9/internal/logging"
	"backstage-nobl9/internal/nobl9"
)

// mockNobl9Client is a mock implementation of the Nobl9 client
type mockNobl9Client struct {
	mock.Mock
	validateProjectNameAttempts int
	createProjectAttempts        int
	validateUserAttempts         int
	assignRolesAttempts          int
}

func (m *mockNobl9Client) GetProject(ctx context.Context, name string) (*nobl9.Project, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*nobl9.Project), args.Error(1)
}

func (m *mockNobl9Client) CreateProject(ctx context.Context, name, description string) (*nobl9.Project, error) {
	args := m.Called(ctx, name, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*nobl9.Project), args.Error(1)
}

func (m *mockNobl9Client) ValidateProjectName(ctx context.Context, name string) (bool, string, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.String(1), args.Error(2)
}

func (m *mockNobl9Client) ValidateUser(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *mockNobl9Client) GetUserRoles(ctx context.Context, projectName, userEmail string) ([]string, error) {
	args := m.Called(ctx, projectName, userEmail)
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockNobl9Client) ValidateRoles(ctx context.Context, projectName, userEmail string, newRoles []string) (bool, []string, error) {
	args := m.Called(ctx, projectName, userEmail, newRoles)
	return args.Bool(0), args.Get(1).([]string), args.Error(2)
}

func (m *mockNobl9Client) AssignRoles(ctx context.Context, projectName string, assignments map[string][]string) error {
	args := m.Called(ctx, projectName, assignments)
	return args.Error(0)
}

type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Debug(msg string, fields ...logging.Field) {
	m.Called(msg, fields)
}

func (m *mockLogger) Info(msg string, fields ...logging.Field) {
	m.Called(msg, fields)
}

func (m *mockLogger) Warn(msg string, fields ...logging.Field) {
	m.Called(msg, fields)
}

func (m *mockLogger) Error(msg string, fields ...logging.Field) {
	m.Called(msg, fields)
}

func (m *mockLogger) WithFields(fields ...logging.Field) logging.Logger {
	args := m.Called(fields)
	return args.Get(0).(logging.Logger)
}

func (m *mockLogger) WithContext(ctx context.Context) logging.Logger {
	args := m.Called(ctx)
	return args.Get(0).(logging.Logger)
}

func TestNewBot(t *testing.T) {
	mockClient := &nobl9.MockClient{}
	mockLogger := &mockLogger{}

	bot := &Bot{
		nobl9Client: mockClient,
		logger:      mockLogger,
		state:       make(map[string]*ConversationState),
	}

	assert.NotNil(t, bot)
	assert.Equal(t, mockClient, bot.nobl9Client)
	assert.Equal(t, mockLogger, bot.logger)
	assert.NotNil(t, bot.state)
}

func TestHandleMessage(t *testing.T) {
	client := &mockNobl9Client{}
	b := NewBot(client, nil)

	// Test unknown command
	response, err := b.HandleMessage("thread1", "unknown-command")
	if err != nil {
		t.Errorf("HandleMessage failed: %v", err)
	}
	if response == "" {
		t.Error("Expected non-empty response for unknown command")
	}

	// Test help command
	response, err = b.HandleMessage("thread1", "help")
	if err != nil {
		t.Errorf("HandleMessage failed: %v", err)
	}
	if response == "" {
		t.Error("Expected non-empty help response")
	}

	// Test create-project command
	response, err = b.HandleMessage("thread1", "create-project test-project")
	if err != nil {
		t.Errorf("HandleMessage failed: %v", err)
	}
	if response == "" {
		t.Error("Expected non-empty response for project creation")
	}

	// Test conversation flow
	response, err = b.HandleMessage("thread1", "This is a project description")
	if err != nil {
		t.Errorf("HandleMessage failed: %v", err)
	}
	if response == "" {
		t.Error("Expected non-empty response for project description")
	}
}

func TestHandleConversationMessage(t *testing.T) {
	client := &mockNobl9Client{}
	b := NewBot(client, nil)

	// Start a conversation
	_, err := b.StartConversation("test-project")
	if err != nil {
		t.Fatalf("Failed to start conversation: %v", err)
	}

	// Test invalid state
	_, err = b.handleConversationMessage("thread1", &ConversationState{
		ProjectName: "test-project",
		Step:        "invalid",
	}, "description")
	if !errors.IsInternalError(err) {
		t.Error("Expected internal error for invalid state")
	}

	// Test valid conversation flow
	response, err := b.handleConversationMessage("thread1", &ConversationState{
		ProjectName: "test-project",
		Step:        "description",
	}, "Test project description")
	if err != nil {
		t.Errorf("handleConversationMessage failed: %v", err)
	}
	if response == "" {
		t.Error("Expected non-empty response for project creation")
	}
}

func TestConversationLifecycle(t *testing.T) {
	mockClient := &nobl9.MockClient{}
	mockLogger := &mockLogger{}
	mockLogger.On("Info", "Started conversation", mock.Anything).Return()
	mockLogger.On("Info", "Ended conversation", mock.Anything).Return()

	bot := &Bot{
		nobl9Client: mockClient,
		logger:      mockLogger,
		state:       make(map[string]*ConversationState),
	}

	// Test starting conversation
	err := bot.StartConversation("test-conv")
	assert.NoError(t, err)
	state, err := bot.GetConversationState("test-conv")
	assert.NoError(t, err)
	assert.NotNil(t, state)
	assert.WithinDuration(t, time.Now(), state.LastUpdated, time.Second)

	// Test ending conversation
	err = bot.EndConversation("test-conv")
	assert.NoError(t, err)
	_, err = bot.GetConversationState("test-conv")
	assert.Error(t, err)
	assert.True(t, errors.IsNotFoundError(err))

	mockLogger.AssertExpectations(t)
}

func TestDuplicateConversation(t *testing.T) {
	mockClient := new(mockNobl9Client)
	bot := NewBot(mockClient, nil)
	ctx := context.Background()
	conversationID := "test-conversation"

	// Start first conversation
	err := bot.StartConversation(ctx, conversationID)
	assert.NoError(t, err)

	// Try to start duplicate conversation
	err = bot.StartConversation(ctx, conversationID)
	assert.Error(t, err)
	assert.True(t, errors.IsConflictError(err))
}

func TestNonexistentConversation(t *testing.T) {
	mockClient := new(mockNobl9Client)
	bot := NewBot(mockClient, nil)
	ctx := context.Background()
	conversationID := "nonexistent"

	// Test getting state of nonexistent conversation
	_, err := bot.GetConversationState(ctx, conversationID)
	assert.Error(t, err)
	assert.True(t, errors.IsNotFoundError(err))

	// Test updating state of nonexistent conversation
	err = bot.UpdateConversationState(ctx, conversationID, func(state *ConversationState) error {
		return nil
	})
	assert.Error(t, err)
	assert.True(t, errors.IsNotFoundError(err))

	// Test ending nonexistent conversation
	err = bot.EndConversation(ctx, conversationID)
	assert.Error(t, err)
	assert.True(t, errors.IsNotFoundError(err))
}

func TestProjectCreation(t *testing.T) {
	mockClient := &nobl9.MockClient{}
	mockLogger := &mockLogger{}
	mockLogger.On("Info", "Started conversation", mock.Anything).Return()
	mockLogger.On("Info", "Handling command", mock.Anything).Return()
	mockLogger.On("Info", "Starting project creation", mock.Anything).Return()
	mockLogger.On("Info", "Project name validated", mock.Anything).Return()
	mockLogger.On("Info", "Project description received", mock.Anything).Return()
	mockLogger.On("Info", "Project created", mock.Anything).Return()

	bot := &Bot{
		nobl9Client: mockClient,
		logger:      mockLogger,
		state:       make(map[string]*ConversationState),
	}

	err := bot.StartConversation("test-conv")
	assert.NoError(t, err)

	// Test project name validation
	mockClient.On("ValidateProjectName", "test-project").Return(true, nil)
	response, err := bot.HandleMessage("test-conv", "/create")
	assert.NoError(t, err)
	assert.Contains(t, response, "Please enter a project name")

	response, err = bot.HandleMessage("test-conv", "test-project")
	assert.NoError(t, err)
	assert.Contains(t, response, "Please provide a description")

	// Test project creation
	mockClient.On("CreateProject", "test-project", "test description").Return("owner@example.com", nil)
	response, err = bot.HandleMessage("test-conv", "test description")
	assert.NoError(t, err)
	assert.Contains(t, response, "Create project")

	response, err = bot.HandleMessage("test-conv", "yes")
	assert.NoError(t, err)
	assert.Contains(t, response, "Project created successfully")

	mockLogger.AssertExpectations(t)
}

func TestRoleManagement(t *testing.T) {
	mockClient := &nobl9.MockClient{}
	mockLogger := &mockLogger{}
	mockLogger.On("Info", "Started conversation", mock.Anything).Return()
	mockLogger.On("Info", "Handling command", mock.Anything).Return()
	mockLogger.On("Info", "Starting role assignment", mock.Anything).Return()
	mockLogger.On("Info", "User validated", mock.Anything).Return()
	mockLogger.On("Info", "Role type received", mock.Anything).Return()
	mockLogger.On("Info", "Role assigned", mock.Anything).Return()

	bot := &Bot{
		nobl9Client: mockClient,
		logger:      mockLogger,
		state:       make(map[string]*ConversationState),
	}

	err := bot.StartConversation("test-conv")
	assert.NoError(t, err)

	// Test user validation
	mockClient.On("ValidateUser", "user@example.com").Return(true, nil)
	response, err := bot.HandleMessage("test-conv", "/assign test-project")
	assert.NoError(t, err)
	assert.Contains(t, response, "Please enter the user's email")

	response, err = bot.HandleMessage("test-conv", "user@example.com")
	assert.NoError(t, err)
	assert.Contains(t, response, "Please select a role type")

	// Test role assignment
	mockClient.On("AssignRoles", "test-project", "user@example.com", "member").Return(nil)
	response, err = bot.HandleMessage("test-conv", "member")
	assert.NoError(t, err)
	assert.Contains(t, response, "Assign role")

	response, err = bot.HandleMessage("test-conv", "yes")
	assert.NoError(t, err)
	assert.Contains(t, response, "Role assigned successfully")

	mockLogger.AssertExpectations(t)
}

func TestErrorHandling(t *testing.T) {
	mockClient := &nobl9.MockClient{}
	mockLogger := &mockLogger{}
	mockLogger.On("Info", "Started conversation", mock.Anything).Return()
	mockLogger.On("Error", "Failed to get conversation state", mock.Anything).Return()
	mockLogger.On("Warn", "Invalid response", mock.Anything).Return()
	mockLogger.On("Error", "Failed to validate project name", mock.Anything).Return()
	mockLogger.On("Warn", "Retrying project name validation", mock.Anything).Return()
	mockLogger.On("Error", "Failed to create project", mock.Anything).Return()
	mockLogger.On("Warn", "Retrying project creation", mock.Anything).Return()
	mockLogger.On("Error", "Failed to validate user", mock.Anything).Return()
	mockLogger.On("Warn", "Retrying user validation", mock.Anything).Return()
	mockLogger.On("Error", "Failed to assign role", mock.Anything).Return()
	mockLogger.On("Warn", "Retrying role assignment", mock.Anything).Return()
	mockLogger.On("Warn", "Received message while in conversation", mock.Anything).Return()

	bot := &Bot{
		nobl9Client: mockClient,
		logger:      mockLogger,
		state:       make(map[string]*ConversationState),
	}

	// Test conversation not found
	_, err := bot.HandleMessage("non-existent", "test")
	assert.Error(t, err)
	assert.True(t, errors.IsNotFoundError(err))

	// Test invalid response
	err = bot.StartConversation("test-conv")
	assert.NoError(t, err)
	response, err := bot.HandleMessage("test-conv", "/create")
	assert.NoError(t, err)
	assert.Contains(t, response, "Please enter a project name")

	response, err = bot.HandleMessage("test-conv", "invalid@email")
	assert.NoError(t, err)
	assert.Contains(t, response, "Invalid response")

	// Test project name validation error
	mockClient.On("ValidateProjectName", "test-project").Return(false, errors.NewInternalError("validation error"))
	response, err = bot.HandleMessage("test-conv", "test-project")
	assert.NoError(t, err)
	assert.Contains(t, response, "Project name is not available")

	// Test project creation error
	mockClient.On("CreateProject", "test-project", "test description").Return("", errors.NewInternalError("creation error"))
	response, err = bot.HandleMessage("test-conv", "test description")
	assert.NoError(t, err)
	assert.Contains(t, response, "Create project")

	response, err = bot.HandleMessage("test-conv", "yes")
	assert.Error(t, err)
	assert.True(t, errors.IsInternalError(err))

	// Test user validation error
	mockClient.On("ValidateUser", "user@example.com").Return(false, errors.NewInternalError("validation error"))
	response, err = bot.HandleMessage("test-conv", "/assign test-project")
	assert.NoError(t, err)
	assert.Contains(t, response, "Please enter the user's email")

	response, err = bot.HandleMessage("test-conv", "user@example.com")
	assert.NoError(t, err)
	assert.Contains(t, response, "User not found")

	// Test role assignment error
	mockClient.On("AssignRoles", "test-project", "user@example.com", "member").Return(errors.NewInternalError("assignment error"))
	response, err = bot.HandleMessage("test-conv", "member")
	assert.NoError(t, err)
	assert.Contains(t, response, "Assign role")

	response, err = bot.HandleMessage("test-conv", "yes")
	assert.Error(t, err)
	assert.True(t, errors.IsInternalError(err))

	// Test invalid state
	response, err = bot.HandleMessage("test-conv", "test")
	assert.NoError(t, err)
	assert.Contains(t, response, "Please complete the current step")

	mockLogger.AssertExpectations(t)
}

func TestInteractiveProjectCreation(t *testing.T) {
	mockClient := &nobl9.MockClient{}
	mockLogger := &mockLogger{}
	mockLogger.On("Info", "Started conversation", mock.Anything).Return()
	mockLogger.On("Info", "Handling command", mock.Anything).Return()
	mockLogger.On("Info", "Starting project creation", mock.Anything).Return()
	mockLogger.On("Info", "Project name validated", mock.Anything).Return()
	mockLogger.On("Info", "Project description received", mock.Anything).Return()
	mockLogger.On("Info", "Project created", mock.Anything).Return()

	bot := &Bot{
		nobl9Client: mockClient,
		logger:      mockLogger,
		state:       make(map[string]*ConversationState),
	}

	err := bot.StartConversation("test-conv")
	assert.NoError(t, err)

	// Test project creation flow
	mockClient.On("ValidateProjectName", "test-project").Return(true, nil)
	mockClient.On("CreateProject", "test-project", "test description").Return("owner@example.com", nil)

	response, err := bot.HandleMessage("test-conv", "/create")
	assert.NoError(t, err)
	assert.Contains(t, response, "Please enter a project name")

	response, err = bot.HandleMessage("test-conv", "test-project")
	assert.NoError(t, err)
	assert.Contains(t, response, "Please provide a description")

	response, err = bot.HandleMessage("test-conv", "test description")
	assert.NoError(t, err)
	assert.Contains(t, response, "Create project")

	response, err = bot.HandleMessage("test-conv", "yes")
	assert.NoError(t, err)
	assert.Contains(t, response, "Project created successfully")

	mockLogger.AssertExpectations(t)
}

func TestInteractiveRoleAssignment(t *testing.T) {
	mockClient := &nobl9.MockClient{}
	mockLogger := &mockLogger{}
	mockLogger.On("Info", "Started conversation", mock.Anything).Return()
	mockLogger.On("Info", "Handling command", mock.Anything).Return()
	mockLogger.On("Info", "Starting role assignment", mock.Anything).Return()
	mockLogger.On("Info", "User validated", mock.Anything).Return()
	mockLogger.On("Info", "Role type received", mock.Anything).Return()
	mockLogger.On("Info", "Role assigned", mock.Anything).Return()

	bot := &Bot{
		nobl9Client: mockClient,
		logger:      mockLogger,
		state:       make(map[string]*ConversationState),
	}

	err := bot.StartConversation("test-conv")
	assert.NoError(t, err)

	// Test role assignment flow
	mockClient.On("ValidateUser", "user@example.com").Return(true, nil)
	mockClient.On("AssignRoles", "test-project", "user@example.com", "member").Return(nil)

	response, err := bot.HandleMessage("test-conv", "/assign test-project")
	assert.NoError(t, err)
	assert.Contains(t, response, "Please enter the user's email")

	response, err = bot.HandleMessage("test-conv", "user@example.com")
	assert.NoError(t, err)
	assert.Contains(t, response, "Please select a role type")

	response, err = bot.HandleMessage("test-conv", "member")
	assert.NoError(t, err)
	assert.Contains(t, response, "Assign role")

	response, err = bot.HandleMessage("test-conv", "yes")
	assert.NoError(t, err)
	assert.Contains(t, response, "Role assigned successfully")

	mockLogger.AssertExpectations(t)
}

func TestInteractiveValidation(t *testing.T) {
	client := &mockNobl9Client{}
	bot := NewBot(client)

	// Test invalid project name
	response, err := bot.HandleMessage("test-thread", "/create")
	if err != nil {
		t.Fatalf("Failed to start project creation: %v", err)
	}

	response, err = bot.HandleMessage("test-thread", "invalid-project")
	if err != nil {
		t.Fatalf("Failed to enter project name: %v", err)
	}
	if response == "" {
		t.Error("Expected non-empty response for invalid project name")
	}

	// Test invalid user email
	response, err = bot.HandleMessage("test-thread", "/assign test-project")
	if err != nil {
		t.Fatalf("Failed to start role assignment: %v", err)
	}

	response, err = bot.HandleMessage("test-thread", "invalid-email")
	if err != nil {
		t.Fatalf("Failed to enter user email: %v", err)
	}
	if response == "" {
		t.Error("Expected non-empty response for invalid user email")
	}

	// Test invalid role type
	response, err = bot.HandleMessage("test-thread", "invalid-role")
	if err != nil {
		t.Fatalf("Failed to enter role type: %v", err)
	}
	if response == "" {
		t.Error("Expected non-empty response for invalid role type")
	}
}

func TestInteractiveCancellation(t *testing.T) {
	client := &mockNobl9Client{}
	bot := NewBot(client)

	// Test project creation cancellation
	response, err := bot.HandleMessage("test-thread", "/create")
	if err != nil {
		t.Fatalf("Failed to start project creation: %v", err)
	}

	response, err = bot.HandleMessage("test-thread", "test-project")
	if err != nil {
		t.Fatalf("Failed to enter project name: %v", err)
	}

	response, err = bot.HandleMessage("test-thread", "Test description")
	if err != nil {
		t.Fatalf("Failed to enter project description: %v", err)
	}

	response, err = bot.HandleMessage("test-thread", "no")
	if err != nil {
		t.Fatalf("Failed to cancel project creation: %v", err)
	}
	if response == "" {
		t.Error("Expected non-empty response for cancellation")
	}

	// Verify state is reset
	state, err := bot.GetConversationState("test-thread")
	if err != nil {
		t.Fatalf("Failed to get conversation state: %v", err)
	}
	if state.CurrentStep != "" || state.PendingPrompt != nil {
		t.Error("Expected conversation state to be reset")
	}
}

func TestErrorRecovery(t *testing.T) {
	client := &mockNobl9Client{
		validateProjectNameAttempts: 2,
		createProjectAttempts:       2,
		validateUserAttempts:        2,
		assignRolesAttempts:         2,
	}
	bot := NewBot(client)

	// Test project name validation recovery
	response, err := bot.HandleMessage("test-thread", "/create")
	if err != nil {
		t.Fatalf("Failed to start project creation: %v", err)
	}

	response, err = bot.HandleMessage("test-thread", "test-project")
	if err != nil {
		t.Fatalf("Failed to validate project name: %v", err)
	}
	if response == "" {
		t.Error("Expected non-empty response after project name validation")
	}

	// Test project creation recovery
	response, err = bot.HandleMessage("test-thread", "Test description")
	if err != nil {
		t.Fatalf("Failed to enter project description: %v", err)
	}

	response, err = bot.HandleMessage("test-thread", "yes")
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}
	if response == "" {
		t.Error("Expected non-empty response after project creation")
	}

	// Test user validation recovery
	response, err = bot.HandleMessage("test-thread", "/assign test-project")
	if err != nil {
		t.Fatalf("Failed to start role assignment: %v", err)
	}

	response, err = bot.HandleMessage("test-thread", "user@example.com")
	if err != nil {
		t.Fatalf("Failed to validate user: %v", err)
	}
	if response == "" {
		t.Error("Expected non-empty response after user validation")
	}

	// Test role assignment recovery
	response, err = bot.HandleMessage("test-thread", "member")
	if err != nil {
		t.Fatalf("Failed to enter role type: %v", err)
	}

	response, err = bot.HandleMessage("test-thread", "yes")
	if err != nil {
		t.Fatalf("Failed to assign role: %v", err)
	}
	if response == "" {
		t.Error("Expected non-empty response after role assignment")
	}
}

func TestErrorRecoveryFailure(t *testing.T) {
	client := &mockNobl9Client{
		validateProjectNameAttempts: 4, // Exceeds max attempts
		createProjectAttempts:       4,
		validateUserAttempts:        4,
		assignRolesAttempts:         4,
	}
	bot := NewBot(client)

	// Test project name validation failure
	response, err := bot.HandleMessage("test-thread", "/create")
	if err != nil {
		t.Fatalf("Failed to start project creation: %v", err)
	}

	_, err = bot.HandleMessage("test-thread", "test-project")
	if err == nil {
		t.Error("Expected error after exceeding max attempts")
	}

	// Test project creation failure
	response, err = bot.HandleMessage("test-thread", "/create")
	if err != nil {
		t.Fatalf("Failed to start project creation: %v", err)
	}

	response, err = bot.HandleMessage("test-thread", "test-project")
	if err != nil {
		t.Fatalf("Failed to enter project name: %v", err)
	}

	response, err = bot.HandleMessage("test-thread", "Test description")
	if err != nil {
		t.Fatalf("Failed to enter project description: %v", err)
	}

	_, err = bot.HandleMessage("test-thread", "yes")
	if err == nil {
		t.Error("Expected error after exceeding max attempts")
	}

	// Test user validation failure
	response, err = bot.HandleMessage("test-thread", "/assign test-project")
	if err != nil {
		t.Fatalf("Failed to start role assignment: %v", err)
	}

	_, err = bot.HandleMessage("test-thread", "user@example.com")
	if err == nil {
		t.Error("Expected error after exceeding max attempts")
	}

	// Test role assignment failure
	response, err = bot.HandleMessage("test-thread", "/assign test-project")
	if err != nil {
		t.Fatalf("Failed to start role assignment: %v", err)
	}

	response, err = bot.HandleMessage("test-thread", "user@example.com")
	if err != nil {
		t.Fatalf("Failed to enter user email: %v", err)
	}

	response, err = bot.HandleMessage("test-thread", "member")
	if err != nil {
		t.Fatalf("Failed to enter role type: %v", err)
	}

	_, err = bot.HandleMessage("test-thread", "yes")
	if err == nil {
		t.Error("Expected error after exceeding max attempts")
	}
}

func TestNewBot(t *testing.T) {
	logger, err := logging.NewLogger(logging.LevelInfo)
	require.NoError(t, err)

	client := &mockNobl9Client{}
	bot := NewBot(client, logger)

	assert.NotNil(t, bot)
	assert.Equal(t, client, bot.client)
	assert.Equal(t, logger, bot.logger)
	assert.NotNil(t, bot.conversations)
}

func TestConversationLifecycle(t *testing.T) {
	logger := &mockLogger{}
	client := &mockNobl9Client{}
	bot := NewBot(client, logger)

	// Test starting a conversation
	conv, err := bot.StartConversation("user123")
	require.NoError(t, err)
	assert.NotNil(t, conv)
	assert.Equal(t, "user123", conv.UserID)
	assert.Equal(t, ConversationStateWaitingForCommand, conv.State)
	assert.True(t, logger.infoCalled)
	assert.Equal(t, "user123", logger.fields["user_id"])

	// Test ending a conversation
	err = bot.EndConversation(conv.ID)
	require.NoError(t, err)
	_, err = bot.GetConversation(conv.ID)
	assert.Error(t, err)
	assert.True(t, logger.infoCalled)
}

func TestProjectCreation(t *testing.T) {
	logger := &mockLogger{}
	client := &mockNobl9Client{}
	bot := NewBot(client, logger)

	// Test project creation command
	conv, err := bot.StartConversation("user123")
	require.NoError(t, err)

	response, err := bot.HandleMessage(conv.ID, "/create test-project")
	require.NoError(t, err)
	assert.Contains(t, response, "Please provide a display name for the project")
	assert.Equal(t, ConversationStateWaitingForProjectDisplayName, conv.State)
	assert.True(t, logger.infoCalled)

	response, err = bot.HandleMessage(conv.ID, "Test Project")
	require.NoError(t, err)
	assert.Contains(t, response, "Project created successfully")
	assert.Equal(t, ConversationStateWaitingForCommand, conv.State)
	assert.True(t, client.createProjectCalled)
	assert.True(t, logger.infoCalled)
}

func TestRoleManagement(t *testing.T) {
	logger := &mockLogger{}
	client := &mockNobl9Client{}
	bot := NewBot(client, logger)

	// Test role assignment command
	conv, err := bot.StartConversation("user123")
	require.NoError(t, err)

	response, err := bot.HandleMessage(conv.ID, "/assign test-project user@example.com")
	require.NoError(t, err)
	assert.Contains(t, response, "Please specify the roles to assign")
	assert.Equal(t, ConversationStateWaitingForRoles, conv.State)
	assert.True(t, logger.infoCalled)

	response, err = bot.HandleMessage(conv.ID, "admin,member")
	require.NoError(t, err)
	assert.Contains(t, response, "Roles assigned successfully")
	assert.Equal(t, ConversationStateWaitingForCommand, conv.State)
	assert.True(t, client.assignRolesCalled)
	assert.True(t, logger.infoCalled)
}

func TestErrorHandling(t *testing.T) {
	logger := &mockLogger{}
	client := &mockNobl9Client{}
	bot := NewBot(client, logger)

	// Test invalid command
	conv, err := bot.StartConversation("user123")
	require.NoError(t, err)

	response, err := bot.HandleMessage(conv.ID, "/invalid")
	require.NoError(t, err)
	assert.Contains(t, response, "Unknown command")
	assert.True(t, logger.warnCalled)

	// Test invalid project name
	response, err = bot.HandleMessage(conv.ID, "/create invalid project")
	require.NoError(t, err)
	assert.Contains(t, response, "Invalid project name")
	assert.True(t, logger.warnCalled)

	// Test invalid user email
	response, err = bot.HandleMessage(conv.ID, "/assign test-project invalid-email")
	require.NoError(t, err)
	assert.Contains(t, response, "Invalid email address")
	assert.True(t, logger.warnCalled)
}

func TestInteractiveProjectCreation(t *testing.T) {
	logger := &mockLogger{}
	client := &mockNobl9Client{}
	bot := NewBot(client, logger)

	conv, err := bot.StartConversation("user123")
	require.NoError(t, err)

	// Test interactive project creation
	response, err := bot.HandleMessage(conv.ID, "I want to create a project")
	require.NoError(t, err)
	assert.Contains(t, response, "What would you like to name your project?")
	assert.Equal(t, ConversationStateWaitingForProjectName, conv.State)
	assert.True(t, logger.infoCalled)

	response, err = bot.HandleMessage(conv.ID, "test-project")
	require.NoError(t, err)
	assert.Contains(t, response, "Please provide a display name for the project")
	assert.Equal(t, ConversationStateWaitingForProjectDisplayName, conv.State)
	assert.True(t, logger.infoCalled)

	response, err = bot.HandleMessage(conv.ID, "Test Project")
	require.NoError(t, err)
	assert.Contains(t, response, "Project created successfully")
	assert.Equal(t, ConversationStateWaitingForCommand, conv.State)
	assert.True(t, client.createProjectCalled)
	assert.True(t, logger.infoCalled)
}

func TestInteractiveRoleAssignment(t *testing.T) {
	logger := &mockLogger{}
	client := &mockNobl9Client{}
	bot := NewBot(client, logger)

	conv, err := bot.StartConversation("user123")
	require.NoError(t, err)

	// Test interactive role assignment
	response, err := bot.HandleMessage(conv.ID, "I want to assign roles")
	require.NoError(t, err)
	assert.Contains(t, response, "Which project would you like to assign roles for?")
	assert.Equal(t, ConversationStateWaitingForProjectName, conv.State)
	assert.True(t, logger.infoCalled)

	response, err = bot.HandleMessage(conv.ID, "test-project")
	require.NoError(t, err)
	assert.Contains(t, response, "Which user would you like to assign roles to?")
	assert.Equal(t, ConversationStateWaitingForUserEmail, conv.State)
	assert.True(t, logger.infoCalled)

	response, err = bot.HandleMessage(conv.ID, "user@example.com")
	require.NoError(t, err)
	assert.Contains(t, response, "Please specify the roles to assign")
	assert.Equal(t, ConversationStateWaitingForRoles, conv.State)
	assert.True(t, logger.infoCalled)

	response, err = bot.HandleMessage(conv.ID, "admin,member")
	require.NoError(t, err)
	assert.Contains(t, response, "Roles assigned successfully")
	assert.Equal(t, ConversationStateWaitingForCommand, conv.State)
	assert.True(t, client.assignRolesCalled)
	assert.True(t, logger.infoCalled)
} 