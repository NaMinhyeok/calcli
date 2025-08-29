package main

import (
	"errors"
	"os/exec"
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
			wantExit:   1,
		},
		{
			name:       "unknown command",
			args:       []string{"unknown"},
			wantStderr: "Unknown command: unknown",
			wantExit:   1, // go run 에서는 1을 return 하지만, 빌드 후 실제 실행에서는 2를 return 한다
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("go", append([]string{"run", "main.go"}, tt.args...)...)
			cmd.Dir = "."

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
