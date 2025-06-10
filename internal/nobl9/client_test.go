package nobl9_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/dfaile/backstage-nobl9/internal/nobl9"
)

// MockRateLimiter is a mock implementation of RateLimiter
type MockRateLimiter struct {
	mock.Mock
}

func (m *MockRateLimiter) Wait(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockRateLimiter) Success() {
	m.Called()
}

func (m *MockRateLimiter) Failure() {
	m.Called()
}

func TestNewClient(t *testing.T) {
	client, err := NewClient("test-key", "test-org", "https://api.nobl9.com")
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "test-key", client.apiKey)
	assert.Equal(t, "test-org", client.organization)
	assert.Equal(t, "https://api.nobl9.com", client.baseURL)
	assert.NotNil(t, client.rateLimiter)
}

func TestGetProject(t *testing.T) {
	client, err := NewClient("test-key", "test-org", "https://api.nobl9.com")
	require.NoError(t, err)

	// Test successful project retrieval
	project, err := client.GetProject(context.Background(), "test-project")
	require.NoError(t, err)
	assert.Equal(t, "test-project", project.Name)
	assert.Equal(t, "Test Project", project.DisplayName)
	assert.Equal(t, "test-org", project.Organization)

	// Test non-existent project
	_, err = client.GetProject(context.Background(), "non-existent")
	assert.Error(t, err)
	assert.IsType(t, &ProjectNotFoundError{}, err)
}

func TestCreateProject(t *testing.T) {
	client, err := NewClient("test-key", "test-org", "https://api.nobl9.com")
	require.NoError(t, err)

	// Test successful project creation
	project, err := client.CreateProject(context.Background(), "new-project", "New Project")
	require.NoError(t, err)
	assert.Equal(t, "new-project", project.Name)
	assert.Equal(t, "New Project", project.DisplayName)
	assert.Equal(t, "test-org", project.Organization)

	// Test duplicate project
	_, err = client.CreateProject(context.Background(), "new-project", "New Project")
	assert.Error(t, err)
	assert.IsType(t, &ProjectAlreadyExistsError{}, err)
}

func TestValidateProjectName(t *testing.T) {
	client, err := NewClient("test-key", "test-org", "https://api.nobl9.com")
	require.NoError(t, err)

	tests := []struct {
		name     string
		valid    bool
		errorMsg string
	}{
		{"valid-project", true, ""},
		{"invalid project", false, "project name contains invalid characters"},
		{"", false, "project name cannot be empty"},
		{"a", false, "project name must be at least 2 characters"},
		{"a".repeat(64), false, "project name must be at most 63 characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.ValidateProjectName(tt.name)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			}
		})
	}
}

func TestValidateUser(t *testing.T) {
	client, err := NewClient("test-key", "test-org", "https://api.nobl9.com")
	require.NoError(t, err)

	// Test valid user
	valid, err := client.ValidateUser(context.Background(), "user@example.com")
	require.NoError(t, err)
	assert.True(t, valid)

	// Test invalid user
	valid, err = client.ValidateUser(context.Background(), "invalid@example.com")
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestGetUserRoles(t *testing.T) {
	rateLimiter := new(MockRateLimiter)
	rateLimiter.On("Wait", mock.Anything).Return(nil)
	rateLimiter.On("Success").Return()

	client := NewClient("test-key", "test-org", "https://api.nobl9.com", rateLimiter)
	roles, err := client.GetUserRoles(context.Background(), "test-project", "user@example.com")

	assert.NoError(t, err)
	assert.Equal(t, []string{"Project User"}, roles)
}

func TestValidateRoles(t *testing.T) {
	rateLimiter := new(MockRateLimiter)
	rateLimiter.On("Wait", mock.Anything).Return(nil)
	rateLimiter.On("Success").Return()

	client := NewClient("test-key", "test-org", "https://api.nobl9.com", rateLimiter)

	// Test valid roles
	valid, redundant, err := client.ValidateRoles(context.Background(), "test-project", "user@example.com", []string{"Project Admin"})
	assert.NoError(t, err)
	assert.True(t, valid)
	assert.Empty(t, redundant)

	// Test redundant roles
	valid, redundant, err = client.ValidateRoles(context.Background(), "test-project", "user@example.com", []string{"Project User"})
	assert.NoError(t, err)
	assert.False(t, valid)
	assert.Equal(t, []string{"Project User"}, redundant)
}

func TestAssignRoles(t *testing.T) {
	client, err := NewClient("test-key", "test-org", "https://api.nobl9.com")
	require.NoError(t, err)

	// Test successful role assignment
	err = client.AssignRoles(context.Background(), "test-project", "user@example.com", []string{"admin"})
	require.NoError(t, err)

	// Test invalid role
	err = client.AssignRoles(context.Background(), "test-project", "user@example.com", []string{"invalid-role"})
	assert.Error(t, err)
	assert.IsType(t, &InvalidRoleError{}, err)

	// Test non-existent project
	err = client.AssignRoles(context.Background(), "non-existent", "user@example.com", []string{"admin"})
	assert.Error(t, err)
	assert.IsType(t, &ProjectNotFoundError{}, err)
}

func TestRateLimiting(t *testing.T) {
	client, err := NewClient("test-key", "test-org", "https://api.nobl9.com")
	require.NoError(t, err)

	mockLimiter := &MockRateLimiter{}
	client.rateLimiter = mockLimiter

	// Test rate limiting is called
	_, err = client.GetProject(context.Background(), "test-project")
	require.NoError(t, err)
	assert.True(t, mockLimiter.waitCalled)
}

func TestErrorTypes(t *testing.T) {
	client, err := NewClient("test-key", "test-org", "https://api.nobl9.com")
	require.NoError(t, err)

	tests := []struct {
		name     string
		apiError error
		expected error
	}{
		{
			"ProjectNotFound",
			&APIError{Code: "NOT_FOUND", Message: "Project not found"},
			&ProjectNotFoundError{},
		},
		{
			"ProjectAlreadyExists",
			&APIError{Code: "CONFLICT", Message: "Project already exists"},
			&ProjectAlreadyExistsError{},
		},
		{
			"InvalidRole",
			&APIError{Code: "INVALID_ARGUMENT", Message: "Invalid role"},
			&InvalidRoleError{},
		},
		{
			"RateLimitExceeded",
			&APIError{Code: "RESOURCE_EXHAUSTED", Message: "Rate limit exceeded"},
			&RateLimitExceededError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.handleAPIError(tt.apiError)
			assert.IsType(t, tt.expected, err)
		})
	}
} 