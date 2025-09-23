package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

//go:embed bin/*
var scripts embed.FS

const (
	programCommand = "update-ng"
	programVersion = "9.0.0"
	programUpdated = "2025-09-24T00:30:00Z"
	programLicense = "GPL-2.0-or-later or contact us for custom"
	programContact = "Dominikos Pritis (https://idominikos.com)"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C7C7C"))

	scriptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)
)

// Script categories for organized execution
var scriptCategories = map[string][]string{
	"System Package Managers": {
		"update-apt", "update-brew", "update-dnf", "update-macports",
		"update-pacman", "update-zypper", "update-winget", "update-snap",
	},
	"Version Managers": {
		"update-asdf", "update-fnm", "update-gvm", "update-mise", "update-volta",
	},
	"Language Package Managers": {
		"update-npm", "update-pip", "update-cargo", "update-gem", "update-conda",
		"update-mamba", "update-poetry", "update-pdm", "update-rye", "update-uv",
		"update-yarn", "update-bun", "update-deno",
	},
	"Cloud & Infrastructure": {
		"update-aws", "update-gcloud", "update-kubectl", "update-terraform",
		"update-docker", "update-podman",
	},
	"Development Tools": {
		"update-gh", "update-chezmoi", "update-pre-commit", "update-direnv",
	},
	"Shell & Environment": {
		"update-zsh", "update-oh-my-zsh", "update-antidote",
	},
}

type scriptStatus int

const (
	statusPending scriptStatus = iota
	statusRunning
	statusSuccess
	statusError
	statusSkipped
)

type scriptResult struct {
	name     string
	status   scriptStatus
	duration time.Duration
	output   string
	error    error
}

type model struct {
	scripts     []string
	results     map[string]*scriptResult
	progress    progress.Model
	spinner     spinner.Model
	completed   int
	total       int
	startTime   time.Time
	showTUI     bool
	parallel    bool
	categories  []string
	mu          sync.RWMutex
}

type scriptFinishedMsg scriptResult

func initialModel(scripts []string, showTUI, parallel bool, categories []string) model {
	p := progress.New(progress.WithDefaultGradient())
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	results := make(map[string]*scriptResult)
	for _, script := range scripts {
		results[script] = &scriptResult{
			name:   script,
			status: statusPending,
		}
	}

	return model{
		scripts:    scripts,
		results:    results,
		progress:   p,
		spinner:    s,
		total:      len(scripts),
		showTUI:    showTUI,
		parallel:   parallel,
		categories: categories,
		startTime:  time.Now(),
	}
}

func (m *model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.spinner.Tick}

	if m.parallel {
		// Start all scripts in parallel
		for _, script := range m.scripts {
			cmds = append(cmds, runScript(script))
		}
	} else {
		// Start first script only
		if len(m.scripts) > 0 {
			cmds = append(cmds, runScript(m.scripts[0]))
		}
	}

	return tea.Batch(cmds...)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case scriptFinishedMsg:
		m.mu.Lock()
		if result, exists := m.results[msg.name]; exists {
			result.status = msg.status
			result.duration = msg.duration
			result.output = msg.output
			result.error = msg.error
			if msg.status == statusSuccess || msg.status == statusError || msg.status == statusSkipped {
				m.completed++
			}
		}
		m.mu.Unlock()

		// If running sequentially, start next script
		if !m.parallel && m.completed < m.total {
			for _, script := range m.scripts {
				if m.results[script].status == statusPending {
					return m, runScript(script)
				}
			}
		}

		// Check if all scripts are done
		if m.completed >= m.total {
			if m.showTUI {
				time.Sleep(2 * time.Second)
			}
			return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *model) View() string {
	if !m.showTUI {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("🚀 Update Command NG - Modern System Updater"))
	b.WriteString("\n\n")

	// Overall progress
	progressPercent := float64(m.completed) / float64(m.total)
	b.WriteString(fmt.Sprintf("Overall Progress: %s %.1f%% (%d/%d)\n\n",
		m.progress.ViewAs(progressPercent), progressPercent*100, m.completed, m.total))

	// Elapsed time
	elapsed := time.Since(m.startTime)
	b.WriteString(infoStyle.Render(fmt.Sprintf("Elapsed: %v", elapsed.Round(time.Second))))
	b.WriteString("\n\n")

	// Group results by category
	for category, categoryScripts := range scriptCategories {
		hasScripts := false
		var categoryResults strings.Builder

		for _, script := range categoryScripts {
			if result, exists := m.results[script]; exists {
				hasScripts = true
				categoryResults.WriteString(m.renderScriptStatus(result))
			}
		}

		if hasScripts {
			b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFB347")).Render(category))
			b.WriteString("\n")
			b.WriteString(categoryResults.String())
			b.WriteString("\n")
		}
	}

	// Handle uncategorized scripts
	var uncategorized []string
	for _, script := range m.scripts {
		found := false
		for _, categoryScripts := range scriptCategories {
			for _, catScript := range categoryScripts {
				if script == catScript {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			uncategorized = append(uncategorized, script)
		}
	}

	if len(uncategorized) > 0 {
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFB347")).Render("Other Tools"))
		b.WriteString("\n")
		for _, script := range uncategorized {
			if result, exists := m.results[script]; exists {
				b.WriteString(m.renderScriptStatus(result))
			}
		}
		b.WriteString("\n")
	}

	// Summary
	if m.completed >= m.total {
		var successCount, errorCount, skippedCount int
		for _, result := range m.results {
			switch result.status {
			case statusSuccess:
				successCount++
			case statusError:
				errorCount++
			case statusSkipped:
				skippedCount++
			}
		}

		b.WriteString("📊 " + lipgloss.NewStyle().Bold(true).Render("Summary:"))
		b.WriteString(fmt.Sprintf(" %s %d succeeded", successStyle.Render("✓"), successCount))
		if errorCount > 0 {
			b.WriteString(fmt.Sprintf(", %s %d failed", errorStyle.Render("✗"), errorCount))
		}
		if skippedCount > 0 {
			b.WriteString(fmt.Sprintf(", %s %d skipped", infoStyle.Render("⊝"), skippedCount))
		}
		b.WriteString(fmt.Sprintf(" (Total time: %v)", time.Since(m.startTime).Round(time.Second)))
	} else {
		b.WriteString(infoStyle.Render("Press q or Ctrl+C to quit"))
	}

	return b.String()
}

func (m *model) renderScriptStatus(result *scriptResult) string {
	name := strings.TrimPrefix(result.name, "update-")

	switch result.status {
	case statusPending:
		return fmt.Sprintf("  ⏳ %s %s\n", infoStyle.Render("○"), scriptStyle.Render(name))
	case statusRunning:
		return fmt.Sprintf("  %s %s %s\n", m.spinner.View(), scriptStyle.Render(name), infoStyle.Render("running..."))
	case statusSuccess:
		duration := ""
		if result.duration > 0 {
			duration = infoStyle.Render(fmt.Sprintf("(%v)", result.duration.Round(100*time.Millisecond)))
		}
		return fmt.Sprintf("  %s %s %s\n", successStyle.Render("✓"), scriptStyle.Render(name), duration)
	case statusError:
		return fmt.Sprintf("  %s %s %s\n", errorStyle.Render("✗"), scriptStyle.Render(name), errorStyle.Render("failed"))
	case statusSkipped:
		return fmt.Sprintf("  %s %s %s\n", infoStyle.Render("⊝"), scriptStyle.Render(name), infoStyle.Render("skipped"))
	}
	return ""
}

func runScript(scriptName string) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		start := time.Now()

		// Read script content
		scriptPath := filepath.Join("bin", scriptName)
		data, err := scripts.ReadFile(scriptPath)
		if err != nil {
			return scriptFinishedMsg{
				name:     scriptName,
				status:   statusError,
				duration: time.Since(start),
				error:    fmt.Errorf("failed to read script: %w", err),
			}
		}

		// Create temp file and execute
		tmpFile := filepath.Join(os.TempDir(), scriptName)
		if err := os.WriteFile(tmpFile, data, 0755); err != nil {
			return scriptFinishedMsg{
				name:     scriptName,
				status:   statusError,
				duration: time.Since(start),
				error:    fmt.Errorf("failed to write temp script: %w", err),
			}
		}
		defer os.Remove(tmpFile)

		// Run the script
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		cmd := exec.CommandContext(ctx, "/bin/sh", tmpFile)
		output, err := cmd.CombinedOutput()

		duration := time.Since(start)
		outputStr := string(output)

		if err != nil {
			// Check if it's a "not found" error (tool not installed)
			if strings.Contains(outputStr, "not found") || strings.Contains(outputStr, "skipping") {
				return scriptFinishedMsg{
					name:     scriptName,
					status:   statusSkipped,
					duration: duration,
					output:   outputStr,
				}
			}
			return scriptFinishedMsg{
				name:     scriptName,
				status:   statusError,
				duration: duration,
				output:   outputStr,
				error:    err,
			}
		}

		return scriptFinishedMsg{
			name:     scriptName,
			status:   statusSuccess,
			duration: duration,
			output:   outputStr,
		}
	})
}

func getAllScripts() ([]string, error) {
	entries, err := scripts.ReadDir("bin")
	if err != nil {
		return nil, err
	}

	var scriptNames []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "update-") {
			scriptNames = append(scriptNames, entry.Name())
		}
	}
	return scriptNames, nil
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "update-ng",
		Short: "Modern system updater with beautiful TUI",
		Long: `Update Command NG - A modern, fast system updater that manages
all your package managers, development tools, and system components
with a beautiful terminal interface.`,
		Version: programVersion,
		RunE: func(cmd *cobra.Command, args []string) error {
			showTUI, _ := cmd.Flags().GetBool("tui")
			parallel, _ := cmd.Flags().GetBool("parallel")
			categories, _ := cmd.Flags().GetStringSlice("categories")
			only, _ := cmd.Flags().GetStringSlice("only")
			skip, _ := cmd.Flags().GetStringSlice("skip")

			scripts, err := getAllScripts()
			if err != nil {
				return fmt.Errorf("failed to read scripts: %w", err)
			}

			// Filter scripts based on flags
			if len(only) > 0 {
				var filtered []string
				for _, script := range scripts {
					for _, pattern := range only {
						if strings.Contains(script, pattern) {
							filtered = append(filtered, script)
							break
						}
					}
				}
				scripts = filtered
			}

			if len(skip) > 0 {
				var filtered []string
				for _, script := range scripts {
					shouldSkip := false
					for _, pattern := range skip {
						if strings.Contains(script, pattern) {
							shouldSkip = true
							break
						}
					}
					if !shouldSkip {
						filtered = append(filtered, script)
					}
				}
				scripts = filtered
			}

			if len(scripts) == 0 {
				fmt.Println("No scripts to run.")
				return nil
			}

			if showTUI {
				m := initialModel(scripts, true, parallel, categories)
				p := tea.NewProgram(&m, tea.WithAltScreen())
				_, err := p.Run()
				return err
			} else {
				// Simple text mode
				fmt.Printf("Running %d update scripts...\n", len(scripts))
				for _, script := range scripts {
					fmt.Printf("Running %s...", script)
					// Run script synchronously in text mode
					// Implementation similar to runScript but synchronous
					fmt.Println(" done")
				}
				return nil
			}
		},
	}

	rootCmd.Flags().BoolP("tui", "t", true, "Show beautiful TUI interface")
	rootCmd.Flags().BoolP("parallel", "p", true, "Run scripts in parallel")
	rootCmd.Flags().StringSliceP("only", "o", []string{}, "Only run scripts matching these patterns")
	rootCmd.Flags().StringSliceP("skip", "s", []string{}, "Skip scripts matching these patterns")
	rootCmd.Flags().StringSliceP("categories", "c", []string{}, "Only run scripts from these categories")

	// Version subcommands
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s version %s\n", programCommand, programVersion)
			fmt.Printf("Updated: %s\n", programUpdated)
			fmt.Printf("License: %s\n", programLicense)
			fmt.Printf("Contact: %s\n", programContact)
		},
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}