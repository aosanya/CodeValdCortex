package main

import "fmt"

// UC-INFRA-001: Water Distribution Network
//
// This use case runs the CodeValdCortex framework with use case-specific configuration.
// The framework automatically loads agent types from config/agents/*.json

func main() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  UC-INFRA-001: Water Distribution Network")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("This use case must be run through the framework.")
	fmt.Println()
	fmt.Println("Use the start script instead:")
	fmt.Println("  ./start.sh")
	fmt.Println()
	fmt.Println("Or run manually:")
	fmt.Println("  cd /workspaces/CodeValdCortex")
	fmt.Println("  export USECASE_CONFIG_DIR=$(pwd)/usecases/UC-INFRA-001-water-distribution-network")
	fmt.Println("  export $(cat usecases/UC-INFRA-001-water-distribution-network/.env | xargs)")
	fmt.Println("  ./bin/codevaldcortex")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
}
