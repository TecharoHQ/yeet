package internal

import (
	"testing"

	"github.com/TecharoHQ/yeet/internal/yeet"
)

func TestGitVersion(t *testing.T) {
	for _, tt := range []struct {
		name  string
		input string
		want  string
	}{
		{
			name: "base test",
		},
		{
			name:  "with version starts with v",
			input: "v1.0.0",
			want:  "1.0.0",
		},
		{
			name:  "with version without v",
			input: "1.0.0",
			want:  "1.0.0",
		},
		{
			name:  "with version with v and -",
			input: "v1.0.0-abc123",
			want:  "1.0.0-abc123",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			yeet.ForceGitVersion = &tt.input
			got := GitVersion()

			if tt.input != "" {
				if got != tt.want {
					t.Errorf("GitVersion() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
