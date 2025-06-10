package nobl9

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// Project represents a Nobl9 project
type Project struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Owner       string    `json:"owner"`
	CreatedAt   time.Time `json:"created_at"`
}

// Role represents a Nobl9 role
type Role struct {
	Name        string
	Description string
	Permissions []string
}

// UserRole represents a user's role in a project
type UserRole struct {
	UserEmail string
	Roles     []string
}

// Client represents a Nobl9 API client
type Client struct {
	httpClient *http.Client
	baseURL    string
	org        string
}

// RateLimiter interface for handling rate limiting
type RateLimiter interface {
	Wait(ctx context.Context) error
	Success()
	Failure()
}

// NewClient creates a new Nobl9 client
func NewClient(clientID, clientSecret, org, baseURL string) (*Client, error) {
	if baseURL == "" {
		baseURL = "https://app.nobl9.com"
	}

	// Create OAuth2 config
	config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     fmt.Sprintf("%s/oauth/token", baseURL),
		Scopes:       []string{"api"},
	}

	// Create HTTP client with OAuth2 transport
	ctx := context.Background()
	httpClient := config.Client(ctx)

	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
		org:        org,
	}, nil
}

// GetProject retrieves a project by name
func (c *Client) GetProject(ctx context.Context, name string) (*Project, error) {
	url := fmt.Sprintf("%s/api/v1/orgs/%s/projects/%s", c.baseURL, c.org, name)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var project Project
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &project, nil
}

// CreateProject creates a new project
func (c *Client) CreateProject(ctx context.Context, name, description string) (*Project, error) {
	url := fmt.Sprintf("%s/api/v1/orgs/%s/projects", c.baseURL, c.org)
	
	project := Project{
		Name:        name,
		Description: description,
	}

	data, err := json.Marshal(project)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal project: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var createdProject Project
	if err := json.NewDecoder(resp.Body).Decode(&createdProject); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &createdProject, nil
}

// ValidateProjectName checks if a project name is valid and available
func (c *Client) ValidateProjectName(ctx context.Context, name string) (bool, string, error) {
	project, err := c.GetProject(ctx, name)
	if err != nil {
		return false, "", err
	}

	if project == nil {
		return true, "", nil
	}

	return false, project.Owner, nil
}

// ValidateUser checks if a user exists in Nobl9
func (c *Client) ValidateUser(ctx context.Context, email string) (bool, error) {
	url := fmt.Sprintf("%s/api/v1/orgs/%s/users/%s", c.baseURL, c.org, email)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to validate user: %w", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// GetUserRoles retrieves a user's roles in a project
func (c *Client) GetUserRoles(ctx context.Context, projectName, userEmail string) ([]string, error) {
	url := fmt.Sprintf("%s/api/v1/orgs/%s/projects/%s/users/%s/roles", c.baseURL, c.org, projectName, userEmail)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var roles []string
	if err := json.NewDecoder(resp.Body).Decode(&roles); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return roles, nil
}

// ValidateRoles checks if the roles are valid and not redundant
func (c *Client) ValidateRoles(ctx context.Context, projectName, userEmail string, newRoles []string) (bool, []string, error) {
	existingRoles, err := c.GetUserRoles(ctx, projectName, userEmail)
	if err != nil {
		return false, nil, fmt.Errorf("failed to get existing roles: %w", err)
	}

	// Check for redundant roles
	redundantRoles := make([]string, 0)
	for _, newRole := range newRoles {
		for _, existingRole := range existingRoles {
			if newRole == existingRole {
				redundantRoles = append(redundantRoles, newRole)
			}
		}
	}

	// If there are redundant roles, return them
	if len(redundantRoles) > 0 {
		return false, redundantRoles, nil
	}

	return true, nil, nil
}

// AssignRoles assigns roles to users in a project
func (c *Client) AssignRoles(ctx context.Context, projectName string, assignments map[string][]string) error {
	for userEmail, roles := range assignments {
		// Check if user exists
		exists, err := c.ValidateUser(ctx, userEmail)
		if err != nil {
			return fmt.Errorf("failed to validate user %s: %w", userEmail, err)
		}
		if !exists {
			return fmt.Errorf("user %s does not exist", userEmail)
		}

		// Check for redundant roles
		valid, redundant, err := c.ValidateRoles(ctx, projectName, userEmail, roles)
		if err != nil {
			return fmt.Errorf("failed to validate roles for user %s: %w", userEmail, err)
		}
		if !valid {
			return fmt.Errorf("redundant roles found for user %s: %v", userEmail, redundant)
		}

		// Assign roles
		url := fmt.Sprintf("%s/api/v1/orgs/%s/projects/%s/users/%s/roles", c.baseURL, c.org, projectName, userEmail)
		
		data, err := json.Marshal(roles)
		if err != nil {
			return fmt.Errorf("failed to marshal roles: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, "PUT", url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to assign roles: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
	}

	return nil
} 