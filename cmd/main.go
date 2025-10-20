package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/aosanya/CodeValdCortex/internal/app"
	"github.com/aosanya/CodeValdCortex/internal/config"
	"github.com/sirupsen/logrus"
)

var (
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
)

func main() {
	var (
		configPath  = flag.String("config", "config.yaml", "Path to configuration file")
		showVersion = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	if *showVersion {
		fmt.Printf("CodeValdCortex\n")
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Build Time: %s\n", buildTime)
		fmt.Printf("Git Commit: %s\n", gitCommit)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load configuration")
	}

	// Initialize logger
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		logrus.WithError(err).Warn("Invalid log level, using info")
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	if cfg.LogFormat == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	logrus.WithFields(logrus.Fields{
		"version":    version,
		"build_time": buildTime,
		"git_commit": gitCommit,
	}).Info("Starting CodeValdCortex")

	// Initialize and start application
	application := app.New(cfg)
	if err := application.Run(); err != nil {
		logrus.WithError(err).Fatal("Application failed to start")
	}
}
