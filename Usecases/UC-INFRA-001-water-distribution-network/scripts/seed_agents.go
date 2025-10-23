package main
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agent"
	"github.com/aosanya/CodeValdCortex/internal/config"
	"github.com/aosanya/CodeValdCortex/internal/database"
	"github.com/aosanya/CodeValdCortex/internal/registry"
)

// AgentSpec defines an agent to be created
type AgentSpec struct {
	ID       string
	Name     string
	Type     string
	Metadata map[string]string
}

func main() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  INFRA-007: Create Infrastructure Agent Instances")
	fmt.Println("  Creating 27 agents for water distribution network demo")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	dbClient, err := database.NewArangoClient(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbClient.Close()

	fmt.Printf("✓ Connected to ArangoDB: %s\n", cfg.Database.Database)

	// Create agent registry repository
	repo, err := registry.NewRepository(dbClient)
	if err != nil {
		log.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Define the 27 agent instances for the water distribution network
	agents := []AgentSpec{
		// Zone 1: North Zone (13 agents)
		// Pipes (5)
		{ID: "PIPE-001", Name: "Main Supply Line North", Type: "pipe", Metadata: map[string]string{
			"zone": "north", "material": "steel", "diameter": "600", "length": "1500",
			"pressure_rating": "16", "installation_date": "2018-03-15", "segment": "main",
		}},
		{ID: "PIPE-002", Name: "Distribution Line North-A", Type: "pipe", Metadata: map[string]string{
			"zone": "north", "material": "PVC", "diameter": "300", "length": "800",
			"pressure_rating": "10", "installation_date": "2020-06-10", "segment": "distribution",
		}},
		{ID: "PIPE-003", Name: "Distribution Line North-B", Type: "pipe", Metadata: map[string]string{
			"zone": "north", "material": "PVC", "diameter": "250", "length": "650",
			"pressure_rating": "10", "installation_date": "2020-06-10", "segment": "distribution",
		}},
		{ID: "PIPE-004", Name: "Service Line North-1", Type: "pipe", Metadata: map[string]string{
			"zone": "north", "material": "copper", "diameter": "100", "length": "200",
			"pressure_rating": "16", "installation_date": "2021-02-20", "segment": "service",
		}},
		{ID: "PIPE-005", Name: "Service Line North-2", Type: "pipe", Metadata: map[string]string{
			"zone": "north", "material": "copper", "diameter": "100", "length": "180",
			"pressure_rating": "16", "installation_date": "2021-02-20", "segment": "service",
		}},

		// Sensors (4)
		{ID: "SENSOR-001", Name: "Pressure Monitor North Main", Type: "sensor", Metadata: map[string]string{
			"zone": "north", "sensor_type": "pressure", "location_lat": "-1.2921", "location_lon": "36.8219",
			"installation_date": "2020-08-01", "calibration_date": "2025-09-15", "monitoring_point": "main_line",
		}},
		{ID: "SENSOR-002", Name: "Flow Meter North-A", Type: "sensor", Metadata: map[string]string{
			"zone": "north", "sensor_type": "flow", "location_lat": "-1.2935", "location_lon": "36.8225",
			"installation_date": "2020-08-01", "calibration_date": "2025-09-15", "monitoring_point": "distribution_a",
		}},
		{ID: "SENSOR-003", Name: "Pressure Monitor North-B", Type: "sensor", Metadata: map[string]string{
			"zone": "north", "sensor_type": "pressure", "location_lat": "-1.2940", "location_lon": "36.8230",
			"installation_date": "2020-08-01", "calibration_date": "2025-09-15", "monitoring_point": "distribution_b",
		}},
		{ID: "SENSOR-004", Name: "Quality Monitor North", Type: "sensor", Metadata: map[string]string{
			"zone": "north", "sensor_type": "quality", "location_lat": "-1.2928", "location_lon": "36.8222",
			"installation_date": "2021-03-10", "calibration_date": "2025-09-20", "monitoring_point": "main_line",
		}},

		// Pumps (2)
		{ID: "PUMP-001", Name: "Main Booster Pump North", Type: "pump", Metadata: map[string]string{
			"zone": "north", "pump_type": "centrifugal", "capacity": "150", "power_rating": "45",
			"installation_date": "2019-05-12", "efficiency_rating": "92", "location_lat": "-1.2920", "location_lon": "36.8218",
		}},
		{ID: "PUMP-002", Name: "Secondary Pump North", Type: "pump", Metadata: map[string]string{
			"zone": "north", "pump_type": "booster", "capacity": "80", "power_rating": "22",
			"installation_date": "2020-07-15", "efficiency_rating": "88", "location_lat": "-1.2932", "location_lon": "36.8224",
		}},

		// Valves (2)
		{ID: "VALVE-001", Name: "Isolation Valve North Main", Type: "valve", Metadata: map[string]string{
			"zone": "north", "valve_type": "gate", "size": "600", "automation": "motorized",
			"installation_date": "2019-05-20", "location_lat": "-1.2922", "location_lon": "36.8220",
		}},
		{ID: "VALVE-002", Name: "Control Valve North-A", Type: "valve", Metadata: map[string]string{
			"zone": "north", "valve_type": "butterfly", "size": "300", "automation": "manual",
			"installation_date": "2020-06-15", "location_lat": "-1.2936", "location_lon": "36.8226",
		}},

		// Zone 2: Central Zone (14 agents)
		// Pipes (5)
		{ID: "PIPE-006", Name: "Main Supply Line Central", Type: "pipe", Metadata: map[string]string{
			"zone": "central", "material": "steel", "diameter": "800", "length": "2000",
			"pressure_rating": "16", "installation_date": "2017-08-20", "segment": "main",
		}},
		{ID: "PIPE-007", Name: "Distribution Line Central-A", Type: "pipe", Metadata: map[string]string{
			"zone": "central", "material": "PVC", "diameter": "400", "length": "1000",
			"pressure_rating": "10", "installation_date": "2019-04-10", "segment": "distribution",
		}},
		{ID: "PIPE-008", Name: "Distribution Line Central-B", Type: "pipe", Metadata: map[string]string{
			"zone": "central", "material": "PVC", "diameter": "400", "length": "1100",
			"pressure_rating": "10", "installation_date": "2019-04-10", "segment": "distribution",
		}},
		{ID: "PIPE-009", Name: "Service Line Central-1", Type: "pipe", Metadata: map[string]string{
			"zone": "central", "material": "PVC", "diameter": "150", "length": "300",
			"pressure_rating": "10", "installation_date": "2020-11-05", "segment": "service",
		}},
		{ID: "PIPE-010", Name: "Service Line Central-2", Type: "pipe", Metadata: map[string]string{
			"zone": "central", "material": "PVC", "diameter": "150", "length": "320",
			"pressure_rating": "10", "installation_date": "2020-11-05", "segment": "service",
		}},

		// Sensors (4)
		{ID: "SENSOR-005", Name: "Pressure Monitor Central Main", Type: "sensor", Metadata: map[string]string{
			"zone": "central", "sensor_type": "pressure", "location_lat": "-1.2864", "location_lon": "36.8172",
			"installation_date": "2019-09-12", "calibration_date": "2025-08-20", "monitoring_point": "main_line",
		}},
		{ID: "SENSOR-006", Name: "Flow Meter Central-A", Type: "sensor", Metadata: map[string]string{
			"zone": "central", "sensor_type": "flow", "location_lat": "-1.2870", "location_lon": "36.8180",
			"installation_date": "2019-09-12", "calibration_date": "2025-08-20", "monitoring_point": "distribution_a",
		}},
		{ID: "SENSOR-007", Name: "Flow Meter Central-B", Type: "sensor", Metadata: map[string]string{
			"zone": "central", "sensor_type": "flow", "location_lat": "-1.2878", "location_lon": "36.8185",
			"installation_date": "2019-09-12", "calibration_date": "2025-08-20", "monitoring_point": "distribution_b",
		}},
		{ID: "SENSOR-008", Name: "Multi-Sensor Central", Type: "sensor", Metadata: map[string]string{
			"zone": "central", "sensor_type": "multi", "location_lat": "-1.2866", "location_lon": "36.8175",
			"installation_date": "2022-01-15", "calibration_date": "2025-09-01", "monitoring_point": "main_line",
		}},

		// Pump (1)
		{ID: "PUMP-003", Name: "High Capacity Pump Central", Type: "pump", Metadata: map[string]string{
			"zone": "central", "pump_type": "centrifugal", "capacity": "200", "power_rating": "55",
			"installation_date": "2018-09-05", "efficiency_rating": "94", "location_lat": "-1.2863", "location_lon": "36.8170",
		}},

		// Valves (4)
		{ID: "VALVE-003", Name: "Isolation Valve Central Main", Type: "valve", Metadata: map[string]string{
			"zone": "central", "valve_type": "gate", "size": "800", "automation": "motorized",
			"installation_date": "2018-09-10", "location_lat": "-1.2865", "location_lon": "36.8173",
		}},
		{ID: "VALVE-004", Name: "Control Valve Central-A", Type: "valve", Metadata: map[string]string{
			"zone": "central", "valve_type": "butterfly", "size": "400", "automation": "motorized",
			"installation_date": "2019-04-15", "location_lat": "-1.2871", "location_lon": "36.8181",
		}},
		{ID: "VALVE-005", Name: "Control Valve Central-B", Type: "valve", Metadata: map[string]string{
			"zone": "central", "valve_type": "butterfly", "size": "400", "automation": "motorized",
			"installation_date": "2019-04-15", "location_lat": "-1.2879", "location_lon": "36.8186",
		}},
		{ID: "VALVE-006", Name: "Pressure Reducing Valve Central", Type: "valve", Metadata: map[string]string{
			"zone": "central", "valve_type": "pressure_reducing", "size": "300", "automation": "automatic",
			"installation_date": "2020-02-10", "location_lat": "-1.2874", "location_lon": "36.8178",
		}},

		// Zone Coordinators (2)
		{ID: "ZONE-NORTH", Name: "North Zone Coordinator", Type: "zone_coordinator", Metadata: map[string]string{
			"zone_id": "north", "zone_name": "North District", "coverage_area_sqkm": "12.5",
			"managed_pipes": "5", "managed_sensors": "4", "managed_pumps": "2", "managed_valves": "2",
			"population_served": "15000", "coordinator_status": "active",
		}},
		{ID: "ZONE-CENTRAL", Name: "Central Zone Coordinator", Type: "zone_coordinator", Metadata: map[string]string{
			"zone_id": "central", "zone_name": "Central Business District", "coverage_area_sqkm": "18.3",
			"managed_pipes": "5", "managed_sensors": "4", "managed_pumps": "1", "managed_valves": "4",
			"population_served": "25000", "coordinator_status": "active",
		}},
	}

	fmt.Printf("Creating %d agent instances...\n\n", len(agents))

	created := 0
	failed := 0

	for _, spec := range agents {
		fmt.Printf("  [%s] %s (%s)...", spec.ID, spec.Name, spec.Type)

		// Create agent instance
		ag := &agent.Agent{
			ID:            spec.ID,
			Name:          spec.Name,
			Type:          spec.Type,
			State:         agent.StateCreated,
			Metadata:      spec.Metadata,
			Config:        agent.Config{
				MaxConcurrentTasks: 5,
				TaskQueueSize:      100,
				HeartbeatInterval:  30 * time.Second,
			},
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			LastHeartbeat: time.Now(),
		}

		// Save to database
		if err := repo.Create(ctx, ag); err != nil {
			fmt.Printf(" ✗ FAILED: %v\n", err)
			failed++
			continue
		}

		fmt.Printf(" ✓\n")
		created++
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  Results: %d created, %d failed\n", created, failed)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	if failed > 0 {
		os.Exit(1)
	}

	fmt.Println("✓ Agent seeding complete!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. View agents in Web UI: http://localhost:8083")
	fmt.Println("  2. Navigate to 'Agents' page to see all instances")
	fmt.Println("  3. Check agent types have the correct configurations")
	fmt.Println("  4. Proceed with INFRA-008: Initialize agent states with baseline data")
}
