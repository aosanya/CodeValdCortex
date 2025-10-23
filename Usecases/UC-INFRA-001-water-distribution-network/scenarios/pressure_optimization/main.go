package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
)

const (
	baseURL = "http://localhost:8083/api/v1"

	// Agent IDs
	sensor1ID = "SENSOR-001"
	sensor2ID = "SENSOR-002"
	sensor3ID = "SENSOR-003"
	pump1ID   = "PUMP-001"
	pump2ID   = "PUMP-002"
	pump3ID   = "PUMP-003"
	coordID   = "COORD-NORTH"

	// Topics
	pressureReadingsTopic = "zone.north.pressure.readings"
	pumpAdjustmentTopic   = "zone.north.pump.adjustments"
	optimizationTopic     = "zone.north.optimization.status"
)

// Message represents a direct agent-to-agent message
type Message struct {
	FromAgentID string                 `json:"from_agent_id"`
	ToAgentID   string                 `json:"to_agent_id"`
	MessageType string                 `json:"message_type"`
	Payload     map[string]interface{} `json:"payload"`
	Priority    int                    `json:"priority"`
}

// PubSubMessage represents a publish/subscribe message
type PubSubMessage struct {
	PublisherAgentID   string                 `json:"publisher_agent_id"`
	PublisherAgentType string                 `json:"publisher_agent_type"`
	EventName          string                 `json:"event_name"`
	Payload            map[string]interface{} `json:"payload"`
	PublicationType    string                 `json:"publication_type,omitempty"`
	TTLSeconds         int                    `json:"ttl_seconds,omitempty"`
}

func main() {
	fmt.Println("=== Starting Pressure Optimization Scenario ===")

	// Load environment variables
	if err := godotenv.Load("../../.env"); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Wait for framework to be ready
	fmt.Println("Waiting for framework to be ready...")
	waitForFramework()

	// Run optimization cycle
	fmt.Println("\n🔄 Starting continuous pressure optimization cycle...")
	fmt.Println("   (Running 3 optimization iterations)")

	for i := 1; i <= 3; i++ {
		fmt.Printf("\n📊 === Optimization Cycle %d ===\n", i)

		// Step 1: Sensors publish pressure readings
		fmt.Println("\n📡 Step 1: Sensors publish pressure readings")
		simulateSensorReadings(i)

		time.Sleep(2 * time.Second)

		// Step 2: Pumps analyze and coordinate adjustments
		fmt.Println("\n⚙️  Step 2: Pumps coordinate pressure adjustments")
		simulatePumpCoordination(i)

		time.Sleep(2 * time.Second)

		// Step 3: Zone coordinator monitors and optimizes
		fmt.Println("\n📋 Step 3: Zone coordinator monitors system balance")
		simulateZoneOptimization(i)

		if i < 3 {
			fmt.Println("\n⏳ Waiting 3 seconds before next cycle...")
			time.Sleep(3 * time.Second)
		}
	}

	// Final summary
	fmt.Println("\n=== Pressure Optimization Scenario Complete ===")
	fmt.Println("\nSummary:")
	fmt.Println("✅ Completed 3 optimization cycles")
	fmt.Println("✅ Sensors continuously monitored pressure across 3 zones")
	fmt.Println("✅ Pumps coordinated to maintain optimal pressure (5.5-6.0 bar)")
	fmt.Println("✅ Zone coordinator balanced system for efficiency")
	fmt.Println("✅ Demonstrated real-time multi-agent pressure optimization")
	fmt.Println("\nThis demonstrates collaborative infrastructure optimization!")
}

func waitForFramework() {
	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		resp, err := http.Get("http://localhost:8083/health")
		if err == nil && resp.StatusCode == 200 {
			fmt.Println("✅ Framework is ready")
			if resp != nil {
				resp.Body.Close()
			}
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
		fmt.Print(".")
	}
	log.Fatal("❌ Framework not ready after 30 seconds")
}

func simulateSensorReadings(cycle int) {
	// Simulate varying pressure readings based on cycle
	pressureVariations := map[int][]float64{
		1: {5.2, 5.4, 5.3}, // Low pressure - pumps need to increase
		2: {5.7, 5.8, 5.6}, // Optimal pressure
		3: {6.1, 6.2, 5.9}, // High pressure - pumps need to decrease
	}

	readings := pressureVariations[cycle]
	sensors := []string{sensor1ID, sensor2ID, sensor3ID}
	zones := []string{"Zone A", "Zone B", "Zone C"}

	for i, sensorID := range sensors {
		pressure := readings[i]
		zone := zones[i]

		// Publish pressure reading to topic
		pubsubMsg := PubSubMessage{
			PublisherAgentID:   sensorID,
			PublisherAgentType: "sensor",
			EventName:          pressureReadingsTopic,
			PublicationType:    "metric",
			Payload: map[string]interface{}{
				"sensor_id":    sensorID,
				"zone":         zone,
				"pressure_bar": pressure,
				"timestamp":    time.Now().Format(time.RFC3339),
				"target_min":   5.5,
				"target_max":   6.0,
				"measurement":  "real-time",
				"quality":      "good",
			},
		}

		publishMessage(pubsubMsg)

		// Determine status
		status := "✅ OPTIMAL"
		if pressure < 5.5 {
			status = "⚠️  LOW"
		} else if pressure > 6.0 {
			status = "⚠️  HIGH"
		}

		fmt.Printf("   📡 %s (%s): %.1f bar %s\n", sensorID, zone, pressure, status)
	}

	// Coordinator receives aggregated data
	time.Sleep(500 * time.Millisecond)
	avgPressure := (readings[0] + readings[1] + readings[2]) / 3.0
	fmt.Printf("   📊 Average system pressure: %.2f bar (target: 5.5-6.0)\n", avgPressure)
}

func simulatePumpCoordination(cycle int) {
	// Pumps analyze pressure data and coordinate adjustments
	adjustments := map[int][]map[string]interface{}{
		1: { // Low pressure - increase output
			{"pump": pump1ID, "action": "INCREASE", "output_change": "+10%", "new_output": "75%"},
			{"pump": pump2ID, "action": "INCREASE", "output_change": "+8%", "new_output": "68%"},
			{"pump": pump3ID, "action": "INCREASE", "output_change": "+12%", "new_output": "72%"},
		},
		2: { // Optimal pressure - fine-tune
			{"pump": pump1ID, "action": "MAINTAIN", "output_change": "0%", "new_output": "75%"},
			{"pump": pump2ID, "action": "DECREASE", "output_change": "-3%", "new_output": "65%"},
			{"pump": pump3ID, "action": "MAINTAIN", "output_change": "0%", "new_output": "72%"},
		},
		3: { // High pressure - decrease output
			{"pump": pump1ID, "action": "DECREASE", "output_change": "-8%", "new_output": "67%"},
			{"pump": pump2ID, "action": "DECREASE", "output_change": "-10%", "new_output": "55%"},
			{"pump": pump3ID, "action": "DECREASE", "output_change": "-7%", "new_output": "65%"},
		},
	}

	pumpAdjustments := adjustments[cycle]

	// Pumps communicate with each other for coordination
	for i, adj := range pumpAdjustments {
		pumpID := adj["pump"].(string)
		action := adj["action"].(string)
		outputChange := adj["output_change"].(string)
		newOutput := adj["new_output"].(string)

		// Publish adjustment decision
		pubsubMsg := PubSubMessage{
			PublisherAgentID:   pumpID,
			PublisherAgentType: "pump",
			EventName:          pumpAdjustmentTopic,
			PublicationType:    "event",
			Payload: map[string]interface{}{
				"pump_id":           pumpID,
				"action":            action,
				"output_change":     outputChange,
				"new_output_level":  newOutput,
				"reason":            "pressure_optimization",
				"coordination_rank": i + 1,
				"timestamp":         time.Now().Format(time.RFC3339),
			},
		}

		publishMessage(pubsubMsg)

		// Visual feedback
		emoji := "➡️"
		if action == "INCREASE" {
			emoji = "⬆️"
		} else if action == "DECREASE" {
			emoji = "⬇️"
		}

		fmt.Printf("   %s %s: %s output %s → %s\n", emoji, pumpID, action, outputChange, newOutput)

		// Send coordination message to next pump
		if i < len(pumpAdjustments)-1 {
			nextPump := pumpAdjustments[i+1]["pump"].(string)
			coordMsg := Message{
				FromAgentID: pumpID,
				ToAgentID:   nextPump,
				MessageType: "coordination",
				Payload: map[string]interface{}{
					"message":      "adjustment_complete",
					"my_output":    newOutput,
					"your_turn":    true,
					"target_range": "5.5-6.0 bar",
				},
				Priority: 7,
			}
			sendDirectMessage(coordMsg)
		}
	}

	fmt.Println("   ✅ Pump coordination complete - adjustments applied")
}

func simulateZoneOptimization(cycle int) {
	// Zone coordinator analyzes overall system performance
	optimizationStatus := map[int]map[string]interface{}{
		1: {
			"status":            "OPTIMIZING",
			"efficiency":        "78%",
			"energy_usage":      "HIGH",
			"pressure_variance": 0.2,
			"recommendation":    "Pressure increased to meet demand",
		},
		2: {
			"status":            "OPTIMAL",
			"efficiency":        "94%",
			"energy_usage":      "NORMAL",
			"pressure_variance": 0.1,
			"recommendation":    "System balanced - maintaining current levels",
		},
		3: {
			"status":            "OPTIMIZING",
			"efficiency":        "89%",
			"energy_usage":      "LOW",
			"pressure_variance": 0.15,
			"recommendation":    "Pressure reduced to save energy",
		},
	}

	status := optimizationStatus[cycle]

	// Publish optimization status
	pubsubMsg := PubSubMessage{
		PublisherAgentID:   coordID,
		PublisherAgentType: "zone_coordinator",
		EventName:          optimizationTopic,
		PublicationType:    "event",
		Payload: map[string]interface{}{
			"coordinator_id":    coordID,
			"cycle":             cycle,
			"status":            status["status"],
			"system_efficiency": status["efficiency"],
			"energy_usage":      status["energy_usage"],
			"pressure_variance": status["pressure_variance"],
			"recommendation":    status["recommendation"],
			"pumps_active":      3,
			"sensors_active":    3,
			"timestamp":         time.Now().Format(time.RFC3339),
		},
	}

	publishMessage(pubsubMsg)

	// Send summary to control room
	summaryMsg := Message{
		FromAgentID: coordID,
		ToAgentID:   "CONTROL-ROOM",
		MessageType: "notification",
		Payload: map[string]interface{}{
			"subject":    "Pressure Optimization Update",
			"cycle":      cycle,
			"status":     status["status"],
			"efficiency": status["efficiency"],
			"energy":     status["energy_usage"],
			"action":     status["recommendation"],
			"priority":   "NORMAL",
			"timestamp":  time.Now().Format(time.RFC3339),
		},
		Priority: 5,
	}

	sendDirectMessage(summaryMsg)

	fmt.Printf("   📋 %s: System status = %s\n", coordID, status["status"])
	fmt.Printf("   ⚡ Efficiency: %s | Energy: %s\n", status["efficiency"], status["energy_usage"])
	fmt.Printf("   💡 %s\n", status["recommendation"])
}

func publishMessage(msg PubSubMessage) {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling pub/sub message: %v", err)
		return
	}

	resp, err := http.Post(baseURL+"/communications/publish", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error publishing message: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Publish failed (status %d): %s", resp.StatusCode, body)
	}
}

func sendDirectMessage(msg Message) {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling direct message: %v", err)
		return
	}

	resp, err := http.Post(baseURL+"/communications/messages", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error sending direct message: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Message send failed (status %d): %s", resp.StatusCode, body)
	}
}
