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
	pump1ID   = "PUMP-001"
	pump2ID   = "PUMP-002"
	pump3ID   = "PUMP-003"
	coordID   = "COORD-NORTH"
	controlID = "CONTROL-ROOM"

	// Topics
	efficiencyTopic     = "zone.north.pump.efficiency"
	maintenanceTopic    = "zone.north.maintenance.alerts"
	workOrderTopic      = "zone.north.maintenance.workorders"
	diagnosticsTopic    = "zone.north.pump.diagnostics"
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
	fmt.Println("=== Starting Predictive Maintenance Scenario ===")

	// Load environment variables
	if err := godotenv.Load("../../.env"); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Wait for framework to be ready
	fmt.Println("Waiting for framework to be ready...")
	waitForFramework()

	// Run predictive maintenance demonstration
	fmt.Println("\n🔧 Starting Predictive Maintenance Demonstration...")
	fmt.Println("   (Monitoring 3 pumps over 4 weeks)")

	// Week 1: Normal operation
	fmt.Println("\n📅 === Week 1: Baseline Performance ===")
	simulateWeek1Baseline()
	time.Sleep(2 * time.Second)

	// Week 2: Early degradation signs
	fmt.Println("\n📅 === Week 2: Early Degradation Detected ===")
	simulateWeek2EarlyDegradation()
	time.Sleep(2 * time.Second)

	// Week 3: Declining performance
	fmt.Println("\n📅 === Week 3: Performance Decline ===")
	simulateWeek3DecliningPerformance()
	time.Sleep(2 * time.Second)

	// Week 4: Maintenance prediction & scheduling
	fmt.Println("\n📅 === Week 4: Maintenance Prediction & Scheduling ===")
	simulateWeek4MaintenancePrediction()

	// Final summary
	fmt.Println("\n=== Predictive Maintenance Scenario Complete ===")
	fmt.Println("\nSummary:")
	fmt.Println("✅ Monitored 3 pumps over 4-week period")
	fmt.Println("✅ Detected early degradation in PUMP-002 (efficiency drop from 92% → 78%)")
	fmt.Println("✅ Published maintenance alerts to zone coordinator")
	fmt.Println("✅ Generated predictive work orders before failure")
	fmt.Println("✅ Demonstrated proactive maintenance vs reactive repairs")
	fmt.Println("\nThis demonstrates AI-driven predictive maintenance for infrastructure!")
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

func simulateWeek1Baseline() {
	pumps := []struct {
		id         string
		efficiency float64
		vibration  float64
		temperature float64
	}{
		{pump1ID, 94.5, 0.8, 68.2},
		{pump2ID, 92.3, 1.2, 70.1},
		{pump3ID, 93.8, 0.9, 69.5},
	}

	fmt.Println("\n📊 Week 1 Performance Metrics:")
	for _, pump := range pumps {
		// Publish efficiency metrics
		pubsubMsg := PubSubMessage{
			PublisherAgentID:   pump.id,
			PublisherAgentType: "pump",
			EventName:          efficiencyTopic,
			PublicationType:    "metric",
			Payload: map[string]interface{}{
				"pump_id":             pump.id,
				"efficiency_percent":  pump.efficiency,
				"vibration_mm_s":      pump.vibration,
				"temperature_celsius": pump.temperature,
				"operating_hours":     2400,
				"week":                1,
				"status":              "NORMAL",
				"timestamp":           time.Now().Format(time.RFC3339),
			},
		}

		publishMessage(pubsubMsg)

		status := "✅ OPTIMAL"
		fmt.Printf("   %s: Efficiency %.1f%% | Vibration %.1f mm/s | Temp %.1f°C %s\n",
			pump.id, pump.efficiency, pump.vibration, pump.temperature, status)
	}

	// Coordinator baseline report
	coordMsg := Message{
		FromAgentID: coordID,
		ToAgentID:   controlID,
		MessageType: "report",
		Payload: map[string]interface{}{
			"subject":         "Week 1 Baseline Performance Report",
			"week":            1,
			"pumps_monitored": 3,
			"avg_efficiency":  93.5,
			"status":          "ALL_NORMAL",
			"alerts":          0,
			"recommendation":  "Continue normal monitoring schedule",
		},
		Priority: 5,
	}
	sendDirectMessage(coordMsg)

	fmt.Println("   📋 All pumps operating within normal parameters")
}

func simulateWeek2EarlyDegradation() {
	pumps := []struct {
		id         string
		efficiency float64
		vibration  float64
		temperature float64
		status     string
	}{
		{pump1ID, 94.2, 0.9, 68.5, "NORMAL"},
		{pump2ID, 88.7, 1.8, 72.3, "WATCH"},    // Efficiency drop detected
		{pump3ID, 93.5, 1.0, 69.8, "NORMAL"},
	}

	fmt.Println("\n📊 Week 2 Performance Metrics:")
	for _, pump := range pumps {
		// Publish efficiency metrics
		pubsubMsg := PubSubMessage{
			PublisherAgentID:   pump.id,
			PublisherAgentType: "pump",
			EventName:          efficiencyTopic,
			PublicationType:    "metric",
			Payload: map[string]interface{}{
				"pump_id":             pump.id,
				"efficiency_percent":  pump.efficiency,
				"vibration_mm_s":      pump.vibration,
				"temperature_celsius": pump.temperature,
				"operating_hours":     2568, // +168 hours (1 week)
				"week":                2,
				"status":              pump.status,
				"efficiency_delta":    pump.efficiency - 92.3, // vs baseline for PUMP-002
				"timestamp":           time.Now().Format(time.RFC3339),
			},
		}

		publishMessage(pubsubMsg)

		statusIcon := "✅ OPTIMAL"
		if pump.status == "WATCH" {
			statusIcon = "⚠️  WATCH"
		}
		fmt.Printf("   %s: Efficiency %.1f%% | Vibration %.1f mm/s | Temp %.1f°C %s\n",
			pump.id, pump.efficiency, pump.vibration, pump.temperature, statusIcon)
	}

	// PUMP-002 sends early degradation alert
	fmt.Println("\n🔔 Early Degradation Alert:")
	alertMsg := PubSubMessage{
		PublisherAgentID:   pump2ID,
		PublisherAgentType: "pump",
		EventName:          diagnosticsTopic,
		PublicationType:    "alert",
		Payload: map[string]interface{}{
			"pump_id":            pump2ID,
			"alert_type":         "EARLY_DEGRADATION",
			"severity":           "LOW",
			"efficiency_drop":    3.6, // 92.3% → 88.7%
			"vibration_increase": 0.6, // 1.2 → 1.8
			"temp_increase":      2.2, // 70.1 → 72.3
			"recommendation":     "Increase monitoring frequency, schedule inspection",
			"predicted_failure":  "4-6 weeks if trend continues",
			"timestamp":          time.Now().Format(time.RFC3339),
		},
	}
	publishMessage(alertMsg)

	fmt.Printf("   ⚠️  %s detected 3.6%% efficiency drop (92.3%% → 88.7%%)\n", pump2ID)
	fmt.Println("   📈 Vibration increased from 1.2 → 1.8 mm/s")
	fmt.Println("   🌡️  Temperature increased from 70.1 → 72.3°C")

	// Coordinator acknowledges alert
	coordResponse := Message{
		FromAgentID: coordID,
		ToAgentID:   pump2ID,
		MessageType: "acknowledgment",
		Payload: map[string]interface{}{
			"alert_received":   true,
			"action":           "MONITORING_INCREASED",
			"inspection_scheduled": "Week 3",
			"watchlist_added":  true,
		},
		Priority: 7,
	}
	sendDirectMessage(coordResponse)

	fmt.Println("   ✅ Coordinator: Monitoring frequency increased, inspection scheduled")
}

func simulateWeek3DecliningPerformance() {
	pumps := []struct {
		id         string
		efficiency float64
		vibration  float64
		temperature float64
		status     string
	}{
		{pump1ID, 93.9, 1.0, 68.8, "NORMAL"},
		{pump2ID, 82.1, 2.5, 75.8, "DEGRADED"}, // Significant decline
		{pump3ID, 93.2, 1.1, 70.0, "NORMAL"},
	}

	fmt.Println("\n📊 Week 3 Performance Metrics:")
	for _, pump := range pumps {
		// Publish efficiency metrics
		pubsubMsg := PubSubMessage{
			PublisherAgentID:   pump.id,
			PublisherAgentType: "pump",
			EventName:          efficiencyTopic,
			PublicationType:    "metric",
			Payload: map[string]interface{}{
				"pump_id":             pump.id,
				"efficiency_percent":  pump.efficiency,
				"vibration_mm_s":      pump.vibration,
				"temperature_celsius": pump.temperature,
				"operating_hours":     2736, // +168 hours (1 week)
				"week":                3,
				"status":              pump.status,
				"efficiency_delta":    pump.efficiency - 92.3, // vs baseline for PUMP-002
				"timestamp":           time.Now().Format(time.RFC3339),
			},
		}

		publishMessage(pubsubMsg)

		statusIcon := "✅ OPTIMAL"
		if pump.status == "DEGRADED" {
			statusIcon = "🔴 DEGRADED"
		}
		fmt.Printf("   %s: Efficiency %.1f%% | Vibration %.1f mm/s | Temp %.1f°C %s\n",
			pump.id, pump.efficiency, pump.vibration, pump.temperature, statusIcon)
	}

	// PUMP-002 sends degradation alert
	fmt.Println("\n🚨 Performance Degradation Alert:")
	alertMsg := PubSubMessage{
		PublisherAgentID:   pump2ID,
		PublisherAgentType: "pump",
		EventName:          maintenanceTopic,
		PublicationType:    "alert",
		Payload: map[string]interface{}{
			"pump_id":            pump2ID,
			"alert_type":         "PERFORMANCE_DEGRADATION",
			"severity":           "MEDIUM",
			"efficiency_drop":    10.2, // 92.3% → 82.1%
			"vibration_increase": 1.3,  // 1.2 → 2.5
			"temp_increase":      5.7,  // 70.1 → 75.8
			"degradation_rate":   "3.3% per week",
			"recommendation":     "Schedule maintenance within 1 week",
			"predicted_failure":  "2-3 weeks",
			"root_cause_likely":  "Bearing wear or impeller damage",
			"timestamp":          time.Now().Format(time.RFC3339),
		},
	}
	publishMessage(alertMsg)

	fmt.Printf("   🚨 %s efficiency critically low: 82.1%% (baseline 92.3%%)\n", pump2ID)
	fmt.Println("   📉 Degradation rate: 3.3% per week")
	fmt.Println("   ⚠️  Predicted failure: 2-3 weeks if not addressed")
	fmt.Println("   🔍 Likely cause: Bearing wear or impeller damage")

	// Coordinator escalates to maintenance team
	coordEscalation := Message{
		FromAgentID: coordID,
		ToAgentID:   controlID,
		MessageType: "escalation",
		Payload: map[string]interface{}{
			"subject":         "URGENT: PUMP-002 Maintenance Required",
			"pump_id":         pump2ID,
			"severity":        "MEDIUM",
			"efficiency":      82.1,
			"efficiency_drop": 10.2,
			"recommendation":  "Schedule maintenance within 1 week to prevent failure",
			"estimated_downtime": "4-6 hours",
			"action_required": "MAINTENANCE_SCHEDULING",
		},
		Priority: 9,
	}
	sendDirectMessage(coordEscalation)

	fmt.Println("   📤 Coordinator escalated to control room for urgent maintenance")
}

func simulateWeek4MaintenancePrediction() {
	pumps := []struct {
		id         string
		efficiency float64
		vibration  float64
		temperature float64
		status     string
	}{
		{pump1ID, 93.6, 1.1, 69.0, "NORMAL"},
		{pump2ID, 78.4, 3.2, 78.5, "CRITICAL"}, // Critical degradation
		{pump3ID, 92.9, 1.2, 70.2, "NORMAL"},
	}

	fmt.Println("\n📊 Week 4 Performance Metrics:")
	for _, pump := range pumps {
		// Publish efficiency metrics
		pubsubMsg := PubSubMessage{
			PublisherAgentID:   pump.id,
			PublisherAgentType: "pump",
			EventName:          efficiencyTopic,
			PublicationType:    "metric",
			Payload: map[string]interface{}{
				"pump_id":             pump.id,
				"efficiency_percent":  pump.efficiency,
				"vibration_mm_s":      pump.vibration,
				"temperature_celsius": pump.temperature,
				"operating_hours":     2904, // +168 hours (1 week)
				"week":                4,
				"status":              pump.status,
				"efficiency_delta":    pump.efficiency - 92.3,
				"timestamp":           time.Now().Format(time.RFC3339),
			},
		}

		publishMessage(pubsubMsg)

		statusIcon := "✅ OPTIMAL"
		if pump.status == "CRITICAL" {
			statusIcon = "🔴 CRITICAL"
		}
		fmt.Printf("   %s: Efficiency %.1f%% | Vibration %.1f mm/s | Temp %.1f°C %s\n",
			pump.id, pump.efficiency, pump.vibration, pump.temperature, statusIcon)
	}

	// PUMP-002 sends critical alert
	fmt.Println("\n🚨 CRITICAL: Imminent Failure Prediction:")
	alertMsg := PubSubMessage{
		PublisherAgentID:   pump2ID,
		PublisherAgentType: "pump",
		EventName:          maintenanceTopic,
		PublicationType:    "alert",
		Payload: map[string]interface{}{
			"pump_id":            pump2ID,
			"alert_type":         "IMMINENT_FAILURE",
			"severity":           "CRITICAL",
			"efficiency_drop":    13.9, // 92.3% → 78.4%
			"vibration_increase": 2.0,  // 1.2 → 3.2
			"temp_increase":      8.4,  // 70.1 → 78.5
			"degradation_rate":   "3.7% per week",
			"recommendation":     "IMMEDIATE maintenance required - take offline",
			"predicted_failure":  "3-7 days",
			"failure_mode":       "Bearing seizure or impeller failure likely",
			"timestamp":          time.Now().Format(time.RFC3339),
		},
	}
	publishMessage(alertMsg)

	fmt.Printf("   🔴 %s CRITICAL: Efficiency 78.4%% (13.9%% drop from baseline)\n", pump2ID)
	fmt.Println("   ⚠️  Imminent failure predicted: 3-7 days")
	fmt.Println("   🛑 Recommendation: Take pump offline immediately")

	// Coordinator generates work order
	fmt.Println("\n📝 Generating Predictive Maintenance Work Order:")
	workOrderMsg := PubSubMessage{
		PublisherAgentID:   coordID,
		PublisherAgentType: "zone_coordinator",
		EventName:          workOrderTopic,
		PublicationType:    "event",
		Payload: map[string]interface{}{
			"work_order_id":   "WO-2025-1023-001",
			"pump_id":         pump2ID,
			"priority":        "CRITICAL",
			"type":            "PREDICTIVE_MAINTENANCE",
			"scheduled_date":  time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
			"estimated_hours": 6,
			"tasks": []string{
				"Inspect and replace bearings",
				"Inspect impeller for damage",
				"Check motor alignment",
				"Replace seals and gaskets",
				"Full system calibration",
			},
			"parts_required": []string{
				"Bearing set (SKF 6308)",
				"Mechanical seal",
				"Impeller (if damaged)",
			},
			"downtime_window":     "02:00-08:00",
			"backup_pump":         "PUMP-001 (capacity boost)",
			"cost_savings":        "Prevented catastrophic failure - estimated $45,000 savings",
			"timestamp":           time.Now().Format(time.RFC3339),
		},
	}
	publishMessage(workOrderMsg)

	fmt.Println("   📋 Work Order: WO-2025-1023-001")
	fmt.Println("   📅 Scheduled: Tomorrow, 02:00-08:00 (6 hours)")
	fmt.Println("   🔧 Tasks: Bearing replacement, impeller inspection, alignment")
	fmt.Println("   💰 Cost Savings: $45,000 (prevented catastrophic failure)")
	fmt.Println("   ♻️  Backup: PUMP-001 will boost capacity during maintenance")

	// Send final report to control room
	finalReport := Message{
		FromAgentID: coordID,
		ToAgentID:   controlID,
		MessageType: "report",
		Payload: map[string]interface{}{
			"subject":           "Predictive Maintenance Success - PUMP-002",
			"pump_id":           pump2ID,
			"detection_week":    2,
			"intervention_week": 4,
			"efficiency_drop":   13.9,
			"failure_prevented": true,
			"work_order":        "WO-2025-1023-001",
			"cost_savings":      45000,
			"downtime_planned":  6,
			"downtime_avoided":  48, // Would have been 2 days emergency repair
			"success_metrics": map[string]interface{}{
				"early_detection":     "2 weeks advance notice",
				"cost_avoidance":      "$45,000",
				"downtime_reduction":  "87.5% (6h vs 48h)",
				"service_continuity":  "100% (backup pump)",
			},
		},
		Priority: 8,
	}
	sendDirectMessage(finalReport)

	fmt.Println("\n✅ Predictive Maintenance Report Sent to Control Room")
	fmt.Println("   ⏱️  Early Detection: 2 weeks advance notice")
	fmt.Println("   💵 Cost Avoidance: $45,000")
	fmt.Println("   📉 Downtime Reduction: 87.5% (6h vs 48h emergency repair)")
	fmt.Println("   ✅ Service Continuity: 100% maintained")
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
