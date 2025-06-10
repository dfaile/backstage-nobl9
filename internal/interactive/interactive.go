package interactive

import (
	"fmt"
	"strings"
	"time"

	"backstage-nobl9/internal/format"
)

// Prompt represents an interactive prompt
type Prompt struct {
	Message     string
	Options     []string
	Default     string
	Timeout     time.Duration
	MaxAttempts int
}

// NewPrompt creates a new prompt
func NewPrompt(message string, options []string, defaultOption string) *Prompt {
	return &Prompt{
		Message:     message,
		Options:     options,
		Default:     defaultOption,
		Timeout:     5 * time.Minute,
		MaxAttempts: 3,
	}
}

// WithTimeout sets the prompt timeout
func (p *Prompt) WithTimeout(timeout time.Duration) *Prompt {
	p.Timeout = timeout
	return p
}

// WithMaxAttempts sets the maximum number of attempts
func (p *Prompt) WithMaxAttempts(attempts int) *Prompt {
	p.MaxAttempts = attempts
	return p
}

// Format formats the prompt message
func (p *Prompt) Format() string {
	var prompt strings.Builder
	prompt.WriteString(format.FormatPrompt(p.Message))
	prompt.WriteString("\n\n")

	if len(p.Options) > 0 {
		prompt.WriteString("Options:\n")
		for i, option := range p.Options {
			if option == p.Default {
				prompt.WriteString(fmt.Sprintf("%d. %s (default)\n", i+1, option))
			} else {
				prompt.WriteString(fmt.Sprintf("%d. %s\n", i+1, option))
			}
		}
	}

	return prompt.String()
}

// Validate validates the user's response
func (p *Prompt) Validate(response string) (string, error) {
	response = strings.TrimSpace(response)

	// Handle empty response with default
	if response == "" && p.Default != "" {
		return p.Default, nil
	}

	// Check if response is a valid option
	if len(p.Options) > 0 {
		for _, option := range p.Options {
			if strings.EqualFold(response, option) {
				return option, nil
			}
		}
		return "", fmt.Errorf("invalid option: %s", response)
	}

	return response, nil
}

// Confirmation represents a yes/no confirmation dialog
type Confirmation struct {
	Message     string
	Default     bool
	Timeout     time.Duration
	MaxAttempts int
}

// NewConfirmation creates a new confirmation dialog
func NewConfirmation(message string, defaultYes bool) *Confirmation {
	return &Confirmation{
		Message:     message,
		Default:     defaultYes,
		Timeout:     5 * time.Minute,
		MaxAttempts: 3,
	}
}

// WithTimeout sets the confirmation timeout
func (c *Confirmation) WithTimeout(timeout time.Duration) *Confirmation {
	c.Timeout = timeout
	return c
}

// WithMaxAttempts sets the maximum number of attempts
func (c *Confirmation) WithMaxAttempts(attempts int) *Confirmation {
	c.MaxAttempts = attempts
	return c
}

// Format formats the confirmation message
func (c *Confirmation) Format() string {
	var prompt strings.Builder
	prompt.WriteString(format.FormatPrompt(c.Message))
	prompt.WriteString("\n\n")

	if c.Default {
		prompt.WriteString("(Y/n) ")
	} else {
		prompt.WriteString("(y/N) ")
	}

	return prompt.String()
}

// Validate validates the user's response
func (c *Confirmation) Validate(response string) (bool, error) {
	response = strings.TrimSpace(strings.ToLower(response))

	// Handle empty response with default
	if response == "" {
		return c.Default, nil
	}

	// Check for yes/no responses
	switch response {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		return false, fmt.Errorf("invalid response: %s", response)
	}
}

// Progress represents a progress indicator
type Progress struct {
	Message     string
	Total       int
	Current     int
	StartTime   time.Time
	UpdateDelay time.Duration
}

// NewProgress creates a new progress indicator
func NewProgress(message string, total int) *Progress {
	return &Progress{
		Message:     message,
		Total:       total,
		Current:     0,
		StartTime:   time.Now(),
		UpdateDelay: 500 * time.Millisecond,
	}
}

// WithUpdateDelay sets the update delay
func (p *Progress) WithUpdateDelay(delay time.Duration) *Progress {
	p.UpdateDelay = delay
	return p
}

// Update updates the progress
func (p *Progress) Update(current int) {
	p.Current = current
}

// Format formats the progress message
func (p *Progress) Format() string {
	var progress strings.Builder
	progress.WriteString(format.FormatProgress(p.Message))
	progress.WriteString("\n\n")

	// Calculate progress percentage
	percentage := 0
	if p.Total > 0 {
		percentage = (p.Current * 100) / p.Total
	}

	// Calculate elapsed time
	elapsed := time.Since(p.StartTime)

	// Calculate estimated time remaining
	var remaining time.Duration
	if p.Current > 0 {
		remaining = (elapsed * time.Duration(p.Total-p.Current)) / time.Duration(p.Current)
	}

	// Build progress bar
	barWidth := 20
	filled := (percentage * barWidth) / 100
	empty := barWidth - filled

	progress.WriteString("[")
	progress.WriteString(strings.Repeat("=", filled))
	progress.WriteString(strings.Repeat(" ", empty))
	progress.WriteString("] ")

	progress.WriteString(fmt.Sprintf("%d%% (%d/%d)", percentage, p.Current, p.Total))
	progress.WriteString(fmt.Sprintf(" - Elapsed: %s", formatDuration(elapsed)))
	if remaining > 0 {
		progress.WriteString(fmt.Sprintf(" - Remaining: %s", formatDuration(remaining)))
	}

	return progress.String()
}

// formatDuration formats a duration in a human-readable format
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
} 