package interactive_test

import (
	"testing"
	"time"
	"github.com/dfaile/backstage-nobl9/internal/interactive"
)

func TestPrompt(t *testing.T) {
	tests := []struct {
		name     string
		prompt   *interactive.Prompt
		response string
		want     string
		wantErr  bool
	}{
		{
			name: "valid option",
			prompt: interactive.NewPrompt("Choose an option", []string{
				"option1",
				"option2",
			}, "option1"),
			response: "option2",
			want:     "option2",
			wantErr:  false,
		},
		{
			name: "default option",
			prompt: interactive.NewPrompt("Choose an option", []string{
				"option1",
				"option2",
			}, "option1"),
			response: "",
			want:     "option1",
			wantErr:  false,
		},
		{
			name: "invalid option",
			prompt: interactive.NewPrompt("Choose an option", []string{
				"option1",
				"option2",
			}, "option1"),
			response: "invalid",
			want:     "",
			wantErr:  true,
		},
		{
			name:     "no options",
			prompt:   interactive.NewPrompt("Enter text", nil, ""),
			response: "any text",
			want:     "any text",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.prompt.Validate(tt.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("Prompt.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Prompt.Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfirmation(t *testing.T) {
	tests := []struct {
		name     string
		confirm  *interactive.Confirmation
		response string
		want     bool
		wantErr  bool
	}{
		{
			name:     "yes response",
			confirm:  interactive.NewConfirmation("Continue?", true),
			response: "yes",
			want:     true,
			wantErr:  false,
		},
		{
			name:     "no response",
			confirm:  interactive.NewConfirmation("Continue?", true),
			response: "no",
			want:     false,
			wantErr:  false,
		},
		{
			name:     "default yes",
			confirm:  interactive.NewConfirmation("Continue?", true),
			response: "",
			want:     true,
			wantErr:  false,
		},
		{
			name:     "default no",
			confirm:  interactive.NewConfirmation("Continue?", false),
			response: "",
			want:     false,
			wantErr:  false,
		},
		{
			name:     "invalid response",
			confirm:  interactive.NewConfirmation("Continue?", true),
			response: "invalid",
			want:     false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.confirm.Validate(tt.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("Confirmation.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Confirmation.Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProgress(t *testing.T) {
	progress := interactive.NewProgress("Processing", 100)

	// Test initial state
	message := progress.Format()
	if message == "" {
		t.Error("Expected non-empty progress message")
	}

	// Test progress update
	progress.Update(50)
	message = progress.Format()
	if message == "" {
		t.Error("Expected non-empty progress message after update")
	}

	// Test completion
	progress.Update(100)
	message = progress.Format()
	if message == "" {
		t.Error("Expected non-empty progress message at completion")
	}

	// Test custom update delay
	progress = interactive.NewProgress("Processing", 100).WithUpdateDelay(1 * time.Second)
	if progress.UpdateDelay != 1*time.Second {
		t.Error("Expected custom update delay")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "seconds",
			duration: 45 * time.Second,
			want:     "45s",
		},
		{
			name:     "minutes",
			duration: 5 * time.Minute,
			want:     "5m",
		},
		{
			name:     "hours",
			duration: 2 * time.Hour,
			want:     "2.0h",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.duration)
			if got != tt.want {
				t.Errorf("formatDuration() = %v, want %v", got, tt.want)
			}
		})
	}
} 