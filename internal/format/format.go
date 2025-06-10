package format

import (
	"fmt"
	"strings"

	"backstage-nobl9/internal/errors"
)

// FormatPrompt formats a prompt message
func FormatPrompt(message string) string {
	return fmt.Sprintf("🤖 %s", message)
}

// FormatError formats an error message
func FormatError(err error) string {
	switch {
	case errors.IsValidationError(err):
		return fmt.Sprintf("❌ Validation error: %s", err.Error())
	case errors.IsNotFoundError(err):
		return fmt.Sprintf("🔍 Not found: %s", err.Error())
	case errors.IsConflictError(err):
		return fmt.Sprintf("⚠️ Conflict: %s", err.Error())
	case errors.IsInternalError(err):
		return fmt.Sprintf("💥 Internal error: %s", err.Error())
	default:
		return fmt.Sprintf("❌ Error: %s", err.Error())
	}
}

// FormatProgress formats a progress message
func FormatProgress(message string) string {
	return fmt.Sprintf("⏳ %s", message)
}

// FormatRoleAssign formats a role assignment message
func FormatRoleAssign(project string, users []string) string {
	return fmt.Sprintf("✅ Assigned roles in project '%s' for users: %s", project, strings.Join(users, ", "))
}

// FormatSuccess formats a success message
func FormatSuccess(message string) string {
	return fmt.Sprintf("✅ %s", message)
}

// FormatWarning formats a warning message
func FormatWarning(message string) string {
	return fmt.Sprintf("⚠️ %s", message)
} 