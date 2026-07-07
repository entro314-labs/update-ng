package updater

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

// TUIRunner is a runner that streams output for TUI display
type TUIRunner struct {
	FS      fs.FS
	Logger  *zap.Logger
	Verbose bool
	SendMsg func(interface{}) // Callback to send messages to TUI
}

// RunScriptStreaming executes a script and streams output line by line
func (r *TUIRunner) RunScriptStreaming(scriptName, description string) ScriptResult {
	start := time.Now()

	// Read script content
	scriptPath := filepath.Join("bin", scriptName)
	data, err := fs.ReadFile(r.FS, scriptPath)
	if err != nil {
		return ScriptResult{
			Name:        scriptName,
			Status:      StatusError,
			Duration:    time.Since(start),
			Error:       fmt.Errorf("failed to read script: %w", err),
			Description: description,
		}
	}

	// Create a secure temp directory
	tmpDir, err := os.MkdirTemp("", "update-ng-*")
	if err != nil {
		return ScriptResult{
			Name:        scriptName,
			Status:      StatusError,
			Duration:    time.Since(start),
			Error:       fmt.Errorf("failed to create temp dir: %w", err),
			Description: description,
		}
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, scriptName)
	if err := os.WriteFile(tmpFile, data, 0700); err != nil {
		return ScriptResult{
			Name:        scriptName,
			Status:      StatusError,
			Duration:    time.Since(start),
			Error:       fmt.Errorf("failed to write temp script: %w", err),
			Description: description,
		}
	}

	// Run the script with streaming output
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/bin/sh", tmpFile)
	cmd.Stdin = os.Stdin

	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return ScriptResult{
			Name:        scriptName,
			Status:      StatusError,
			Duration:    time.Since(start),
			Error:       fmt.Errorf("failed to create stdout pipe: %w", err),
			Description: description,
		}
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return ScriptResult{
			Name:        scriptName,
			Status:      StatusError,
			Duration:    time.Since(start),
			Error:       fmt.Errorf("failed to create stderr pipe: %w", err),
			Description: description,
		}
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return ScriptResult{
			Name:        scriptName,
			Status:      StatusError,
			Duration:    time.Since(start),
			Error:       fmt.Errorf("failed to start command: %w", err),
			Description: description,
		}
	}

	// Collect all output
	var allOutput strings.Builder

	// Stream stdout
	go r.streamOutput(stdout, scriptName, &allOutput)
	// Stream stderr
	go r.streamOutput(stderr, scriptName, &allOutput)

	// Wait for command to complete
	err = cmd.Wait()
	duration := time.Since(start)
	outputStr := allOutput.String()

	// Ensure minimum execution time for visibility
	if duration < 300*time.Millisecond {
		time.Sleep(300*time.Millisecond - duration)
		duration = time.Since(start)
	}

	// Log the result
	logFields := []zap.Field{
		zap.String("script", scriptName),
		zap.Duration("duration", duration),
		zap.String("output", outputStr),
	}

	if err != nil {
		// Enhanced error categorization
		if isToolNotAvailable(outputStr) {
			r.Logger.Warn("Script skipped - tool not available", logFields...)
			return ScriptResult{
				Name:        scriptName,
				Status:      StatusSkipped,
				Duration:    duration,
				Output:      outputStr,
				Description: description,
			}
		}

		if isExternallyManaged(outputStr) {
			r.Logger.Info("Script skipped - externally managed", logFields...)
			return ScriptResult{
				Name:        scriptName,
				Status:      StatusSkipped,
				Duration:    duration,
				Output:      outputStr,
				Description: description,
			}
		}

		if isAlreadyUpToDate(outputStr) {
			r.Logger.Info("Script completed - already up to date", logFields...)
			return ScriptResult{
				Name:        scriptName,
				Status:      StatusSuccess,
				Duration:    duration,
				Output:      outputStr,
				Description: description,
			}
		}

		logFields = append(logFields, zap.Error(err))
		r.Logger.Error("Script failed", logFields...)
		return ScriptResult{
			Name:        scriptName,
			Status:      StatusError,
			Duration:    duration,
			Output:      outputStr,
			Error:       err,
			Description: description,
		}
	}

	r.Logger.Info("Script finished successfully", logFields...)
	return ScriptResult{
		Name:        scriptName,
		Status:      StatusSuccess,
		Duration:    duration,
		Output:      outputStr,
		Description: description,
	}
}

func (r *TUIRunner) streamOutput(reader io.Reader, scriptName string, allOutput *strings.Builder) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		allOutput.WriteString(line + "\n")

		// Send output to TUI if callback is set
		if r.SendMsg != nil {
			r.SendMsg(ScriptOutput{
				Name:   scriptName,
				Line:   line,
				IsLast: false,
			})
		}
	}
}
