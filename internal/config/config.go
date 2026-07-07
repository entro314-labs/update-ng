package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func InitConfig(cfgFile string) {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find config directory.
		configDir, err := os.UserConfigDir()
		cobra.CheckErr(err)
		updateNgConfigDir := filepath.Join(configDir, "update-ng")

		// Search config in user's config directory.
		viper.AddConfigPath(updateNgConfigDir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")

		// For backward compatibility, also search for old config file in home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.AddConfigPath(home)
		viper.SetConfigName(".update-ng")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	} else {
		// Just log if it's not a "not found" error, or ignore if we want optional config
		// For now, silently ignore missing config as it's optional
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// It was found but some other error occurred
			fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
		}
	}
}

func InitLogger() {
	logPath := viper.GetString("log-file")
	if logPath == "" {
		// Default log path
		configDir, err := os.UserConfigDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not get user config directory: %v\n", err)
			os.Exit(1)
		}
		logPath = filepath.Join(configDir, "update-ng", "update-ng.log")
	}

	// Ensure log directory exists
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Could not create log directory: %v\n", err)
		os.Exit(1)
	}

	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoder := zapcore.NewJSONEncoder(config)
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open log file: %v\n", err)
		os.Exit(1)
	}

	writer := zapcore.AddSync(logFile)
	defaultLogLevel := zapcore.InfoLevel
	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, writer, defaultLogLevel),
	)

	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}
