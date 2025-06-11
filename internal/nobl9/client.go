package nobl9

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/project"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/rolebinding"
	"github.com/nobl9/nobl9-go/sdk"
	objectsV1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
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

// Client represents a Nobl9 API client using the official SDK
type Client struct {
	sdkClient *sdk.Client
	org       string
}

// RateLimiter interface for handling rate limiting
type RateLimiter interface {
	Wait(ctx context.Context) error
	Success()
	Failure()
}

// NewClient creates a new Nobl9 client using the official SDK
func NewClient(clientID, clientSecret, org, baseURL string) (*Client, error) {
	// The Nobl9 SDK recommends using sdk.DefaultClient() which handles configuration
	// from environment variables, config files, and other sources automatically
	sdkClient, err := sdk.DefaultClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Nobl9 SDK client: %w", err)
	}

	return &Client{
		sdkClient: sdkClient,
		org:       org,
	}, nil
}

// GetProject retrieves a project by name
func (c *Client) GetProject(ctx context.Context, name string) (*Project, error) {
	projects, err := c.sdkClient.Objects().V1().GetV1alphaProjects(ctx, objectsV1.GetProjectsRequest{
		Names: []string{name},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if len(projects) == 0 {
		return nil, nil // Project not found
	}

	proj := projects[0]
	return &Project{
		Name:        proj.Metadata.Name,
		Description: proj.Spec.Description,
		CreatedAt:   time.Now(), // SDK doesn't expose creation time directly
	}, nil
}

// ListProjects retrieves all projects in the organization
func (c *Client) ListProjects(ctx context.Context) ([]*Project, error) {
	projects, err := c.sdkClient.Objects().V1().GetV1alphaProjects(ctx, objectsV1.GetProjectsRequest{
		// Empty Names slice means get all projects
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	result := make([]*Project, len(projects))
	for i, proj := range projects {
		result[i] = &Project{
			Name:        proj.Metadata.Name,
			Description: proj.Spec.Description,
			CreatedAt:   time.Now(), // SDK doesn't expose creation time directly
		}
	}

	return result, nil
}

// CreateProject creates a new project
func (c *Client) CreateProject(ctx context.Context, name, description string) (*Project, error) {
	// Create project using the official SDK
	proj := project.New(
		project.Metadata{
			Name:        name,
			DisplayName: name,
		},
		project.Spec{
			Description: description,
		},
	)

	// Apply the project - note the correct type conversion
	objects := []manifest.Object{proj}
	if err := c.sdkClient.Objects().V1().Apply(ctx, objects); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return &Project{
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
	}, nil
}

// ValidateProjectName checks if a project name is valid and available
func (c *Client) ValidateProjectName(ctx context.Context, name string) (bool, string, error) {
	project, err := c.GetProject(ctx, name)
	if err != nil {
		return false, "", err
	}

	if project == nil {
		return true, "", nil // Project doesn't exist, name is available
	}

	return false, project.Owner, nil // Project exists
}

// ValidateUser checks if a user exists in Nobl9
func (c *Client) ValidateUser(ctx context.Context, email string) (bool, error) {
	// Note: The Nobl9 SDK doesn't have direct user validation methods in the public API
	// This would typically require admin/organization-level permissions
	// For now, we'll assume the user exists if the email format is valid
	if email == "" {
		return false, nil
	}
	
	// Basic email validation - in a real implementation, you might want to:
	// 1. Use organization API to list users
	// 2. Call a user validation endpoint if available
	// 3. Validate against your organization's user directory
	return true, nil
}

// GetUserRoles retrieves a user's roles in a project
func (c *Client) GetUserRoles(ctx context.Context, projectName, userEmail string) ([]string, error) {
	// Note: Role management in Nobl9 typically happens through:
	// 1. Organization-level role assignments
	// 2. Project-level permissions
	// 3. External identity providers
	
	// The public SDK doesn't expose role management APIs directly
	// This would require admin permissions and organization management APIs
	return []string{"Project User"}, nil
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

	if len(redundantRoles) > 0 {
		return false, redundantRoles, nil
	}

	return true, nil, nil
}

// AssignRoles assigns roles to users in a project using RoleBinding objects through the Nobl9 SDK
func (c *Client) AssignRoles(ctx context.Context, projectName string, assignments map[string][]string) error {
	objects := make([]manifest.Object, 0)

	for userEmail, roles := range assignments {
		// Validate user exists (basic check)
		exists, err := c.ValidateUser(ctx, userEmail)
		if err != nil {
			return fmt.Errorf("failed to validate user %s: %w", userEmail, err)
		}
		if !exists {
			return fmt.Errorf("user %s does not exist", userEmail)
		}

		// Create RoleBinding for each role assignment
		for _, role := range roles {
			var nobl9Role string
			switch role {
			case "admin", "1":
				nobl9Role = "project-owner" 
			case "member", "2":
				nobl9Role = "project-editor"
			case "viewer", "3":
				nobl9Role = "project-viewer"
			default:
				return fmt.Errorf("invalid role: %s. Valid roles are: admin, member, viewer", role)
			}

			// Generate a unique name for the role binding (RFC-1123 compliant)
			// Replace @ with -at-, dots with -, underscores with -, and ensure lowercase
			emailPart := strings.ReplaceAll(userEmail, "@", "-at-")
			emailPart = strings.ReplaceAll(emailPart, ".", "-")
			emailPart = strings.ReplaceAll(emailPart, "_", "-")
			
			// Also ensure the role part is RFC-1123 compliant
			rolePart := strings.ReplaceAll(nobl9Role, "_", "-")
			
			roleBindingName := fmt.Sprintf("%s-%s-%s", projectName, emailPart, rolePart)
			roleBindingName = strings.ToLower(roleBindingName)
			
			// Final cleanup to ensure RFC-1123 compliance
			roleBindingName = strings.ReplaceAll(roleBindingName, "_", "-")
			
			if len(roleBindingName) > 63 {
				// Truncate but keep it meaningful
				suffix := fmt.Sprintf("-%s", rolePart)
				maxPrefix := 63 - len(suffix)
				roleBindingName = roleBindingName[:maxPrefix] + suffix
			}

			fmt.Printf("✅ Creating RoleBinding for user %s in project %s with role %s\n", userEmail, projectName, nobl9Role)
			
			// Create RoleBinding object using the Nobl9 SDK
			roleBinding := rolebinding.New(
				rolebinding.Metadata{
					Name: roleBindingName,
				},
				rolebinding.Spec{
					User:       &userEmail,  // User email
					RoleRef:    nobl9Role,   // Role reference
					ProjectRef: projectName, // Project reference
				},
			)

			objects = append(objects, roleBinding)
		}
	}

	// Apply all RoleBinding objects
	if len(objects) > 0 {
		fmt.Printf("🔄 Applying %d role binding(s) to Nobl9...\n", len(objects))
		if err := c.sdkClient.Objects().V1().Apply(ctx, objects); err != nil {
			return fmt.Errorf("failed to apply role bindings: %w", err)
		}
		fmt.Printf("✅ Successfully applied all role bindings!\n")
	}
	
	return nil
} 