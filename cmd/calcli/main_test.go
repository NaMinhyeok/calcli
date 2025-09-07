package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantStdout string
		wantStderr string
		wantExit   int
	}{
		{
			name:       "version flag",
			args:       []string{"--version"},
			wantStdout: "calcli v0.0.3",
			wantExit:   0,
		},
		{
			name:       "help flag",
			args:       []string{"--help"},
			wantStderr: "Usage:",
			wantExit:   0,
		},
		{
			name:       "no arguments shows help",
			args:       []string{},
			wantStderr: "Usage:",
			wantExit:   0,
		},
		{
			name:       "list command",
			args:       []string{"list"},
			wantStdout: "Team Standup",
			wantExit:   0,
		},
		{
			name:       "new command",
			args:       []string{"new"},
			wantStdout: "Event 'New Event' created successfully",
			wantExit:   0,
		},
		{
			name:       "calendars command",
			args:       []string{"calendars"},
			wantStdout: "home:",
			wantExit:   0,
		},
		{
			name:       "search command",
			args:       []string{"search", "Event"},
			wantStdout: "Test Event",
			wantExit:   0,
		},
		{
			name:       "import command requires file",
			args:       []string{"import"},
			wantStderr: "Usage:",
			wantExit:   1, // go run returns 1 even if os.Exit(2)
		},
		{
			name:       "unknown command",
			args:       []string{"unknown"},
			wantStderr: "Unknown command: unknown",
			wantExit:   1, // go run returns 1 even if os.Exit(2)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup isolated config/calendars for commands that need it
			cfgPath, _, cleanup := setupTestEnv(t)
			defer cleanup()

			cmd := exec.Command("go", append([]string{"run", "main.go"}, tt.args...)...)
			cmd.Dir = "."
			cmd.Env = append(os.Environ(), "CALCLI_CONFIG="+cfgPath)

			stdout, stderr, exitCode := runCommand(cmd)

			if exitCode != tt.wantExit {
				t.Errorf("exit code = %d, want %d", exitCode, tt.wantExit)
			}

			if tt.wantStdout != "" && !strings.Contains(stdout, tt.wantStdout) {
				t.Errorf("stdout = %q, want to contain %q", stdout, tt.wantStdout)
			}

			if tt.wantStderr != "" && !strings.Contains(stderr, tt.wantStderr) {
				t.Errorf("stderr = %q, want to contain %q", stderr, tt.wantStderr)
			}
		})
	}
}

func runCommand(cmd *exec.Cmd) (stdout, stderr string, exitCode int) {
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	stdout = strings.TrimSpace(outBuf.String())
	stderr = strings.TrimSpace(errBuf.String())

	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			exitCode = exitError.ExitCode()
		}
	}

	return stdout, stderr, exitCode
}

// setupTestEnv creates a temporary config pointing to a temporary calendar
// directory and writes a couple of sample ICS files for deterministic tests.
func setupTestEnv(t *testing.T) (configPath string, calendarPath string, cleanup func()) {
	t.Helper()

	base := t.TempDir()
	calDir := filepath.Join(base, "home")
	if err := os.MkdirAll(calDir, 0o755); err != nil {
		t.Fatalf("failed to create calendar dir: %v", err)
	}

	// Write two sample ICS files
	ics1 := "" +
		"BEGIN:VCALENDAR\n" +
		"VERSION:2.0\n" +
		"PRODID:-//calcli//test//EN\n" +
		"BEGIN:VEVENT\n" +
		"UID:uid-1\n" +
		"SUMMARY:Team Standup\n" +
		"DTSTART:20250830T100000Z\n" +
		"DTEND:20250830T103000Z\n" +
		"END:VEVENT\n" +
		"END:VCALENDAR\n"
	if err := os.WriteFile(filepath.Join(calDir, "event1.ics"), []byte(ics1), 0o644); err != nil {
		t.Fatalf("failed to write ICS1: %v", err)
	}

	ics2 := "" +
		"BEGIN:VCALENDAR\n" +
		"VERSION:2.0\n" +
		"PRODID:-//calcli//test//EN\n" +
		"BEGIN:VEVENT\n" +
		"UID:uid-2\n" +
		"SUMMARY:Test Event\n" +
		"DTSTART:20250831T120000Z\n" +
		"DTEND:20250831T130000Z\n" +
		"END:VEVENT\n" +
		"END:VCALENDAR\n"
	if err := os.WriteFile(filepath.Join(calDir, "event2.ics"), []byte(ics2), 0o644); err != nil {
		t.Fatalf("failed to write ICS2: %v", err)
	}

	// Config JSON targeting our temp calendar
	cfg := []byte("{\n" +
		"  \"calendars\": {\n" +
		"    \"home\": {\n" +
		"      \"path\": \"" + escapeForJSON(calDir) + "\",\n" +
		"      \"color\": \"blue\",\n" +
		"      \"readonly\": false\n" +
		"    }\n" +
		"  },\n" +
		"  \"defaults\": {\n" +
		"    \"defaultCalendar\": \"home\"\n" +
		"  }\n" +
		"}\n")
	cfgPath := filepath.Join(base, "config.json")
	if err := os.WriteFile(cfgPath, cfg, 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	return cfgPath, calDir, func() { _ = os.RemoveAll(base) }
}

// escapeForJSON is a tiny helper to escape backslashes in Windows-style paths.
func escapeForJSON(s string) string {
	return strings.ReplaceAll(s, "\\", "\\\\")
}
