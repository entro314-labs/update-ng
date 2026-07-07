package cmd

import (
	"fmt"
	"io/fs"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/entro314-labs/update-ng/internal/config"
	"github.com/entro314-labs/update-ng/internal/ui"
	"github.com/entro314-labs/update-ng/internal/updater"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile   string
	ScriptsFS fs.FS // Scripts embed FS injected from main

	// Program metadata
	ProgramVersion = "1.0.4"
	ProgramUpdated = "2025-09-24T00:30:00Z"
	ProgramLicense = "MIT"
	ProgramContact = "Dominikos Pritis (https://idominikos.com)"
)

var rootCmd = &cobra.Command{
	Use:   "update-ng",
	Short: "Modern system updater with beautiful TUI",
	Long: `Update Command NG - A modern, fast system updater that manages
all your package managers, development tools, and system components
with a beautiful terminal interface.`,
	Version: ProgramVersion,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		config.InitConfig(cfgFile)
		config.InitLogger()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Viper is already bound in init(), we can read directly or via flags
		parallel := viper.GetBool("parallel")
		// verbose := viper.GetBool("verbose") // Can be used by runner
		only := viper.GetStringSlice("only")
		skip := viper.GetStringSlice("skip")
		exclude := viper.GetStringSlice("exclude")
		// categories := viper.GetStringSlice("categories") // Not currently implemented in filtering but in runner

		scripts, err := updater.GetAllScripts(ScriptsFS)
		if err != nil {
			return fmt.Errorf("failed to read scripts: %w", err)
		}

		// Filter scripts based on flags
		scripts = updater.FilterScripts(scripts, only, skip, exclude)

		if len(scripts) == 0 {
			fmt.Println("No scripts to run.")
			return nil
		}

		runner := &updater.Runner{
			FS:      ScriptsFS,
			Logger:  config.Logger,
			Verbose: viper.GetBool("verbose"),
		}

		if viper.GetBool("tui") {
			model := ui.NewModel(runner, scripts, parallel)
			p := tea.NewProgram(model)
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("failed to run TUI: %w", err)
			}
			return nil
		}

		runner.RunScripts(scripts, parallel)
		return nil
	},
}

func Execute(fs fs.FS) {
	ScriptsFS = fs
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Persistent flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/update-ng/config.yaml)")

	// Flags
	rootCmd.Flags().BoolP("tui", "t", true, "Show beautiful TUI interface")
	rootCmd.Flags().BoolP("parallel", "p", true, "Run scripts in parallel")
	rootCmd.Flags().BoolP("verbose", "v", true, "Show verbose output with detailed information")
	rootCmd.Flags().StringSliceP("only", "o", []string{}, "Only run scripts matching these patterns (supports exact names or partial matches)")
	rootCmd.Flags().StringSliceP("skip", "s", []string{}, "Skip scripts matching these patterns (supports exact names or partial matches)")
	rootCmd.Flags().StringSliceP("exclude", "e", []string{}, "Exclude specific scripts by exact name (e.g., 'update-cargo-projects')")
	rootCmd.Flags().StringSliceP("categories", "c", []string{}, "Only run scripts from these categories")
	rootCmd.Flags().String("log-file", "", "Path to log file (default is in user's config directory)")

	// Bind flags to viper
	viper.BindPFlag("tui", rootCmd.Flags().Lookup("tui"))
	viper.BindPFlag("parallel", rootCmd.Flags().Lookup("parallel"))
	viper.BindPFlag("verbose", rootCmd.Flags().Lookup("verbose"))
	viper.BindPFlag("only", rootCmd.Flags().Lookup("only"))
	viper.BindPFlag("skip", rootCmd.Flags().Lookup("skip"))
	viper.BindPFlag("exclude", rootCmd.Flags().Lookup("exclude"))
	viper.BindPFlag("categories", rootCmd.Flags().Lookup("categories"))
	viper.BindPFlag("log-file", rootCmd.Flags().Lookup("log-file"))

	// Version subcommand
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("update-ng version %s\n", ProgramVersion)
		fmt.Printf("Updated: %s\n", ProgramUpdated)
		fmt.Printf("License: %s\n", ProgramLicense)
		fmt.Printf("Contact: %s\n", ProgramContact)
	},
}
