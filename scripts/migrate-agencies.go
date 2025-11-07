package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/aosanya/CodeValdCortex/internal/agency/arangodb"
	"github.com/aosanya/CodeValdCortex/internal/agency/models"
	"github.com/aosanya/CodeValdCortex/internal/agency/services"
	"github.com/aosanya/CodeValdCortex/internal/config"
	"github.com/aosanya/CodeValdCortex/internal/database"
)

func main() {
	fmt.Println("=== Agency Migration Script ===")
	fmt.Println("Discovering and importing use cases...")

	// Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	arangoClient, err := database.NewArangoClient(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	db := arangoClient.Database()

	// Create repository
	repo, err := arangodb.New(arangoClient.Client(), db)
	if err != nil {
		log.Fatalf("Failed to create repository: %v", err)
	}

	// Create service
	validator := agency.NewValidator()
	service := services.New(repo, validator)

	// Discover use cases
	usecasesDir := "/workspaces/CodeValdCortex/usecases"
	agencies, err := discoverUseCases(usecasesDir)
	if err != nil {
		log.Fatalf("Failed to discover use cases: %v", err)
	}

	fmt.Printf("Found %d use cases\n", len(agencies))

	// Import agencies
	ctx := context.Background()
	imported := 0
	skipped := 0

	for _, ag := range agencies {
		// Check if already exists
		exists, err := repo.Exists(ctx, ag.ID)
		if err != nil {
			log.Printf("Error checking existence of %s: %v", ag.ID, err)
			continue
		}

		if exists {
			fmt.Printf("Skipping %s (already exists)\n", ag.ID)
			skipped++
			continue
		}

		// Create agency
		if err := service.CreateAgency(ctx, ag); err != nil {
			log.Printf("Failed to import %s: %v", ag.ID, err)
			continue
		}

		fmt.Printf("Imported: %s - %s\n", ag.ID, ag.DisplayName)
		imported++
	}

	fmt.Println("\n=== Migration Complete ===")
	fmt.Printf("Imported: %d\n", imported)
	fmt.Printf("Skipped: %d\n", skipped)
	fmt.Printf("Total: %d\n", len(agencies))
}

// discoverUseCases scans the usecases directory and creates agency records
func discoverUseCases(rootDir string) ([]*models.Agency, error) {
	var agencies []*models.Agency

	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read usecases directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Parse use case folder name (e.g., UC-INFRA-001-water-distribution-network)
		dirName := entry.Name()
		if !strings.HasPrefix(dirName, "UC-") {
			continue
		}

		parts := strings.Split(dirName, "-")
		if len(parts) < 3 {
			log.Printf("Skipping invalid directory name: %s", dirName)
			continue
		}

		ucID := strings.Join(parts[:3], "-")  // UC-INFRA-001
		category := strings.ToLower(parts[1]) // infra
		namePart := strings.Join(parts[3:], " ")

		// Create agency record
		ag := &models.Agency{
			ID:          ucID,
			Name:        formatName(namePart),
			DisplayName: formatDisplayName(namePart),
			Description: fmt.Sprintf("Use case: %s", formatName(namePart)),
			Category:    category,
			Icon:        getIconForCategory(category),
			Status:      models.AgencyStatusActive,
			Metadata: models.AgencyMetadata{
				Roles:       []string{},
				TotalAgents: 0,
				Tags:        []string{category},
				APIEndpoint: fmt.Sprintf("/api/v1/agencies/%s", ucID),
			},
			Settings: models.AgencySettings{
				AutoStart:         false,
				MonitoringEnabled: true,
				DashboardEnabled:  true,
				VisualizerEnabled: true,
			},
			CreatedBy: "migration",
		}

		agencies = append(agencies, ag)
	}

	return agencies, nil
}

// formatName converts a hyphenated name to title case
func formatName(name string) string {
	words := strings.Split(name, " ")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

// formatDisplayName creates a display name with an icon
func formatDisplayName(name string) string {
	return formatName(name)
}

// getIconForCategory returns an appropriate emoji icon for a category
func getIconForCategory(category string) string {
	icons := map[string]string{
		"infra": "💧",
		"char":  "🤝",
		"comm":  "📡",
		"event": "📅",
		"fra":   "💰",
		"live":  "🌱",
		"log":   "📦",
		"ride":  "🚗",
		"track": "🚍",
		"wms":   "🏭",
	}

	if icon, ok := icons[category]; ok {
		return icon
	}
	return "📋"
}
