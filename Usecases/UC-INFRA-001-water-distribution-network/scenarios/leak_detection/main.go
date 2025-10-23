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
	// API Base URL
	baseURL = "http://localhost:8083/api/v1"

	// Agent IDs for the scenario
	sensorID = "SENSOR-001"
	pipeID   = "PIPE-001"
	valve1ID = "VALVE-001"
	valve2ID = "VALVE-002"
	coordID  = "COORD-NORTH"

	// Topics for pub/sub communication
	leakDetectedTopic  = "zone.north.leak.detected"
	leakConfirmedTopic = "zone.north.leak.confirmed"
)

// Message represents a message to be sent between agents
type Message struct {
	FromAgentID   string                 `json:"from_agent_id"`
	ToAgentID     string                 `json:"to_agent_id"`
	MessageType   string                 `json:"message_type"`
	Payload       map[string]interface{} `json:"payload"`
	Priority      int                    `json:"priority"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
}

// PubSubMessage represents a pub/sub message
type PubSubMessage struct {
	PublisherAgentID   string                 `json:"publisher_agent_id"`
	PublisherAgentType string                 `json:"publisher_agent_type,omitempty"`
	EventName          string                 `json:"event_name"`
	Payload            map[string]interface{} `json:"payload"`
	PublicationType    string                 `json:"publication_type,omitempty"`
	TTLSeconds         int                    `json:"ttl_seconds,omitempty"`
}

func main() {
	fmt.Println("=== Starting Leak Detection Scenario ===")

	// Load environment variables
	if err := godotenv.Load("../../.env"); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Wait for framework to be ready
	fmt.Println("Waiting for framework to be ready...")
	waitForFramework()

	// Step 1: Sensor detects anomaly
	fmt.Println("\n🔍 Step 1: Sensor detects pressure anomaly")
	simulateSensorAnomaly()

	time.Sleep(2 * time.Second)

	// Step 2: Pipe agent analyzes the situation
	fmt.Println("\n🔧 Step 2: Pipe agent analyzes leak probability")
	simulatePipeAnalysis()

	time.Sleep(2 * time.Second)

	// Step 3: Valve agents isolate the section
	fmt.Println("\n🚰 Step 3: Valve agents isolate pipe section")
	simulateValveIsolation()

	time.Sleep(2 * time.Second)

	// Step 4: Zone coordinator escalates to control room
	fmt.Println("\n📋 Step 4: Zone coordinator escalates incident")
	simulateZoneCoordinatorEscalation()

	time.Sleep(2 * time.Second)

	fmt.Println("\n=== Leak Detection Scenario Complete ===")
	fmt.Println("Summary:")
	fmt.Println("✅ Sensor detected pressure drop from 6.0 to 4.5 bar")
	fmt.Println("✅ Pipe agent confirmed 85% leak probability, 50 L/min loss")
	fmt.Println("✅ Isolation valves closed to contain leak")
	fmt.Println("✅ Incident escalated to control room for maintenance")
	fmt.Println("\nThis demonstrates multi-agent coordination for infrastructure monitoring!")
}

func waitForFramework() {
	for i := 0; i < 30; i++ {
		resp, err := http.Get("http://localhost:8083/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			fmt.Println("✅ Framework is ready")
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

func simulateSensorAnomaly() {
	// Publish leak detection alert to topic
	pubsubMsg := PubSubMessage{
		PublisherAgentID:   sensorID,
		PublisherAgentType: "sensor",
		EventName:          leakDetectedTopic,
		PublicationType:    "alert",
		Payload: map[string]interface{}{
			"sensor_id":         sensorID,
			"location":          "North Main Pipeline",
			"pressure_previous": 6.0,
			"pressure_current":  4.5,
			"pressure_drop_pct": -25.0,
			"timestamp":         time.Now().Format(time.RFC3339),
			"alert_level":       "HIGH",
			"message":           "Significant pressure drop detected - potential leak",
		},
	}

	publishMessage(pubsubMsg)

	// Send direct message to pipe agent
	directMsg := Message{
		FromAgentID: sensorID,
		ToAgentID:   pipeID,
		MessageType: "PRESSURE_ANOMALY_ALERT",
		Priority:    1,
		Payload: map[string]interface{}{
			"sensor_id":         sensorID,
			"pressure_drop":     1.5,
			"pressure_drop_pct": -25.0,
			"requires_analysis": true,
			"urgency":           "HIGH",
		},
	}

	sendDirectMessage(directMsg)

	fmt.Printf("   📡 %s: Pressure drop detected (6.0 → 4.5 bar, -25%%)\n", sensorID)
	fmt.Printf("   📤 Published alert to topic: %s\n", leakDetectedTopic)
	fmt.Printf("   📧 Sent direct alert to %s\n", pipeID)
}

func simulatePipeAnalysis() {
	// Pipe agent analyzes and confirms leak
	pubsubMsg := PubSubMessage{
		PublisherAgentID:   pipeID,
		PublisherAgentType: "pipe",
		EventName:          leakConfirmedTopic,
		PublicationType:    "event",
		Payload: map[string]interface{}{
			"pipe_id":            pipeID,
			"analysis_result":    "LEAK_CONFIRMED",
			"leak_probability":   85,
			"estimated_loss_lpm": 50,
			"location":           "North Main Pipeline, Section A",
			"severity":           "MODERATE",
			"isolation_required": true,
			"upstream_valve":     valve1ID,
			"downstream_valve":   valve2ID,
			"timestamp":          time.Now().Format(time.RFC3339),
		},
	}

	publishMessage(pubsubMsg)

	// Send isolation commands to valves
	valve1Cmd := Message{
		FromAgentID: pipeID,
		ToAgentID:   valve1ID,
		MessageType: "ISOLATION_COMMAND",
		Priority:    1,
		Payload: map[string]interface{}{
			"command":       "CLOSE",
			"reason":        "LEAK_ISOLATION",
			"urgency":       "HIGH",
			"leak_location": "downstream",
		},
	}

	valve2Cmd := Message{
		FromAgentID: pipeID,
		ToAgentID:   valve2ID,
		MessageType: "ISOLATION_COMMAND",
		Priority:    1,
		Payload: map[string]interface{}{
			"command":       "CLOSE",
			"reason":        "LEAK_ISOLATION",
			"urgency":       "HIGH",
			"leak_location": "upstream",
		},
	}

	sendDirectMessage(valve1Cmd)
	sendDirectMessage(valve2Cmd)

	fmt.Printf("   🔧 %s: Leak analysis complete\n", pipeID)
	fmt.Printf("   📊 Result: 85%% leak probability, estimated 50 L/min loss\n")
	fmt.Printf("   📤 Published confirmation to topic: %s\n", leakConfirmedTopic)
	fmt.Printf("   🚰 Sent CLOSE commands to %s and %s\n", valve1ID, valve2ID)
}

func simulateValveIsolation() {
	// Simulate valve responses
	valve1Response := Message{
		FromAgentID: valve1ID,
		ToAgentID:   pipeID,
		MessageType: "COMMAND_RESPONSE",
		Priority:    1,
		Payload: map[string]interface{}{
			"command_executed": "CLOSE",
			"status":           "SUCCESS",
			"position":         "CLOSED",
			"isolation_time":   time.Now().Format(time.RFC3339),
			"flow_stopped":     true,
		},
	}

	valve2Response := Message{
		FromAgentID: valve2ID,
		ToAgentID:   pipeID,
		MessageType: "COMMAND_RESPONSE",
		Priority:    1,
		Payload: map[string]interface{}{
			"command_executed": "CLOSE",
			"status":           "SUCCESS",
			"position":         "CLOSED",
			"isolation_time":   time.Now().Format(time.RFC3339),
			"flow_stopped":     true,
		},
	}

	sendDirectMessage(valve1Response)
	sendDirectMessage(valve2Response)

	fmt.Printf("   🚰 %s: Valve closed successfully\n", valve1ID)
	fmt.Printf("   🚰 %s: Valve closed successfully\n", valve2ID)
	fmt.Printf("   ✅ Pipe section isolated - leak contained\n")
}

func simulateZoneCoordinatorEscalation() {
	// Zone coordinator escalates to control room
	escalationMsg := Message{
		FromAgentID: coordID,
		ToAgentID:   "CONTROL-ROOM",
		MessageType: "INCIDENT_ESCALATION",
		Priority:    1,
		Payload: map[string]interface{}{
			"incident_type":        "WATER_LEAK",
			"severity":             "MODERATE",
			"location":             "North Main Pipeline, Section A",
			"affected_pipes":       []string{pipeID},
			"isolated_valves":      []string{valve1ID, valve2ID},
			"estimated_loss":       "50 L/min",
			"maintenance_required": true,
			"repair_priority":      "HIGH",
			"incident_time":        time.Now().Format(time.RFC3339),
			"response_time_sec":    120,
			"status":               "CONTAINED",
		},
	}

	sendDirectMessage(escalationMsg)

	// Also publish incident summary
	incidentSummary := PubSubMessage{
		PublisherAgentID:   coordID,
		PublisherAgentType: "zone_coordinator",
		EventName:          "incidents.water.leak.resolved",
		PublicationType:    "event",
		Payload: map[string]interface{}{
			"incident_id":     fmt.Sprintf("LEAK-%d", time.Now().Unix()),
			"status":          "CONTAINED",
			"response_time":   "2 minutes",
			"agents_involved": []string{sensorID, pipeID, valve1ID, valve2ID, coordID},
			"summary":         "Leak detected and isolated successfully via multi-agent coordination",
		},
	}

	publishMessage(incidentSummary)

	fmt.Printf("   📋 %s: Incident escalated to control room\n", coordID)
	fmt.Printf("   🚨 Maintenance dispatch requested for pipe repair\n")
	fmt.Printf("   📊 Response time: 2 minutes from detection to isolation\n")
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
