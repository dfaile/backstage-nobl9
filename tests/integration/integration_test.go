package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"backstage-nobl9/internal/bot"
	"backstage-nobl9/internal/nobl9"
	"backstage-nobl9/internal/logging"
)

var (
	apiKey = os.Getenv("NOBL9_API_KEY")
	org    = os.Getenv("NOBL9_ORG")
	baseURL = os.Getenv("NOBL9_BASE_URL")
)

func TestMain(m *testing.M) {
	if apiKey == "" || org == "" || baseURL == "" {
		os.Exit(0) // Skip tests if environment variables are not set
	}
	os.Exit(m.Run())
}

func setupTest(t *testing.T) (*bot.Bot, func()) {
	logger, err := logging.NewLogger(logging.LevelDebug)
	require.NoError(t, err)

	client, err := nobl9.NewClient(apiKey, org, baseURL)
	require.NoError(t, err)

	bot := bot.NewBot(client, logger)

	// Cleanup function to delete test projects
	cleanup := func() {
		ctx := context.Background()
		projects := []string{"test-project", "new-project"}
		for _, name := range projects {
			_ = client.DeleteProject(ctx, name)
		}
	}

	return bot, cleanup
}

func TestProjectCreation(t *testing.T) {
	bot, cleanup := setupTest(t)
	defer cleanup()

	// Test project creation flow
	conv, err := bot.StartConversation("test-user")
	require.NoError(t, err)

	// Test command-based creation
	response, err := bot.HandleMessage(conv.ID, "/create test-project")
	require.NoError(t, err)
	assert.Contains(t, response, "Please provide a display name for the project")

	response, err = bot.HandleMessage(conv.ID, "Test Project")
	require.NoError(t, err)
	assert.Contains(t, response, "Project created successfully")

	// Test interactive creation
	response, err = bot.HandleMessage(conv.ID, "I want to create a project")
	require.NoError(t, err)
	assert.Contains(t, response, "What would you like to name your project?")

	response, err = bot.HandleMessage(conv.ID, "new-project")
	require.NoError(t, err)
	assert.Contains(t, response, "Please provide a display name for the project")

	response, err = bot.HandleMessage(conv.ID, "New Project")
	require.NoError(t, err)
	assert.Contains(t, response, "Project created successfully")
}

func TestRoleManagement(t *testing.T) {
	bot, cleanup := setupTest(t)
	defer cleanup()

	// Create a test project first
	conv, err := bot.StartConversation("test-user")
	require.NoError(t, err)

	response, err := bot.HandleMessage(conv.ID, "/create test-project")
	require.NoError(t, err)
	response, err = bot.HandleMessage(conv.ID, "Test Project")
	require.NoError(t, err)

	// Test role assignment
	response, err = bot.HandleMessage(conv.ID, "/assign test-project test@example.com")
	require.NoError(t, err)
	assert.Contains(t, response, "Please specify the roles to assign")

	response, err = bot.HandleMessage(conv.ID, "admin,member")
	require.NoError(t, err)
	assert.Contains(t, response, "Roles assigned successfully")

	// Test interactive role assignment
	response, err = bot.HandleMessage(conv.ID, "I want to assign roles")
	require.NoError(t, err)
	assert.Contains(t, response, "Which project would you like to assign roles for?")

	response, err = bot.HandleMessage(conv.ID, "test-project")
	require.NoError(t, err)
	assert.Contains(t, response, "Which user would you like to assign roles to?")

	response, err = bot.HandleMessage(conv.ID, "another@example.com")
	require.NoError(t, err)
	assert.Contains(t, response, "Please specify the roles to assign")

	response, err = bot.HandleMessage(conv.ID, "member")
	require.NoError(t, err)
	assert.Contains(t, response, "Roles assigned successfully")
}

func TestErrorHandling(t *testing.T) {
	bot, cleanup := setupTest(t)
	defer cleanup()

	conv, err := bot.StartConversation("test-user")
	require.NoError(t, err)

	// Test invalid command
	response, err := bot.HandleMessage(conv.ID, "/invalid")
	require.NoError(t, err)
	assert.Contains(t, response, "Unknown command")

	// Test invalid project name
	response, err = bot.HandleMessage(conv.ID, "/create invalid project")
	require.NoError(t, err)
	assert.Contains(t, response, "Invalid project name")

	// Test duplicate project
	response, err = bot.HandleMessage(conv.ID, "/create test-project")
	require.NoError(t, err)
	response, err = bot.HandleMessage(conv.ID, "Test Project")
	require.NoError(t, err)

	response, err = bot.HandleMessage(conv.ID, "/create test-project")
	require.NoError(t, err)
	response, err = bot.HandleMessage(conv.ID, "Test Project")
	require.NoError(t, err)
	assert.Contains(t, response, "Project already exists")

	// Test invalid user email
	response, err = bot.HandleMessage(conv.ID, "/assign test-project invalid-email")
	require.NoError(t, err)
	assert.Contains(t, response, "Invalid email address")
}

func TestRateLimiting(t *testing.T) {
	bot, cleanup := setupTest(t)
	defer cleanup()

	conv, err := bot.StartConversation("test-user")
	require.NoError(t, err)

	// Send multiple requests in quick succession
	for i := 0; i < 10; i++ {
		response, err := bot.HandleMessage(conv.ID, "/list")
		require.NoError(t, err)
		assert.NotEmpty(t, response)
		time.Sleep(100 * time.Millisecond)
	}
}

func TestInteractiveFeatures(t *testing.T) {
	bot, cleanup := setupTest(t)
	defer cleanup()

	conv, err := bot.StartConversation("test-user")
	require.NoError(t, err)

	// Test help command
	response, err := bot.HandleMessage(conv.ID, "/help")
	require.NoError(t, err)
	assert.Contains(t, response, "Available commands")

	// Test list command
	response, err = bot.HandleMessage(conv.ID, "/list")
	require.NoError(t, err)
	assert.Contains(t, response, "No projects found")

	// Test conversation timeout
	time.Sleep(30 * time.Minute)
	response, err = bot.HandleMessage(conv.ID, "/list")
	require.NoError(t, err)
	assert.Contains(t, response, "Conversation expired")
} 