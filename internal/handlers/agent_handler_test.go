package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agent"
	"github.com/aosanya/CodeValdCortex/internal/runtime"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() (*gin.Engine, *runtime.Manager) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	config := runtime.ManagerConfig{
		MaxAgents:           10,
		HealthCheckInterval: 30 * time.Second,
		ShutdownTimeout:     5 * time.Second,
		EnableMetrics:       true,
	}

	manager := runtime.NewManager(logger, config)
	handler := NewAgentHandler(manager, logger) // Register routes
	v1 := router.Group("/api/v1")
	{
		agents := v1.Group("/agents")
		{
			agents.POST("", handler.CreateAgent)
			agents.GET("", handler.ListAgents)
			agents.GET("/:id", handler.GetAgent)
			agents.POST("/:id/start", handler.StartAgent)
			agents.POST("/:id/stop", handler.StopAgent)
			agents.POST("/:id/tasks", handler.SubmitTask)
		}
		v1.GET("/metrics", handler.GetMetrics)
	}

	return router, manager
}

func TestCreateAgent(t *testing.T) {
	router, manager := setupTestRouter()
	defer manager.Shutdown()

	reqBody := CreateAgentRequest{
		Name: "test-agent",
		Type: "worker",
		Config: agent.Config{
			MaxConcurrentTasks: 5,
			TaskQueueSize:      50,
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response["id"])
	assert.Equal(t, "test-agent", response["name"])
	assert.Equal(t, "worker", response["type"])
	assert.Equal(t, "created", response["state"])
}

func TestCreateAgentInvalidRequest(t *testing.T) {
	router, manager := setupTestRouter()
	defer manager.Shutdown()

	// Missing required fields
	reqBody := CreateAgentRequest{
		Name: "", // Empty name
		Type: "worker",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListAgents(t *testing.T) {
	router, manager := setupTestRouter()
	defer manager.Shutdown()

	// Create a few agents
	config := agent.Config{MaxConcurrentTasks: 5, TaskQueueSize: 50}
	manager.CreateAgent("agent1", "worker", config)
	manager.CreateAgent("agent2", "coordinator", config)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []AgentResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response, 2)
	assert.Equal(t, "agent1", response[0].Name)
	assert.Equal(t, "agent2", response[1].Name)
}

func TestGetAgent(t *testing.T) {
	router, manager := setupTestRouter()
	defer manager.Shutdown()

	config := agent.Config{MaxConcurrentTasks: 5, TaskQueueSize: 50}
	ag, err := manager.CreateAgent("test-agent", "worker", config)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents/"+ag.ID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, ag.ID, response["id"])
	assert.Equal(t, "test-agent", response["name"])
}

func TestGetAgentNotFound(t *testing.T) {
	router, manager := setupTestRouter()
	defer manager.Shutdown()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents/nonexistent-id", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestStartAgent(t *testing.T) {
	router, manager := setupTestRouter()
	defer manager.Shutdown()

	config := agent.Config{MaxConcurrentTasks: 5, TaskQueueSize: 50}
	ag, err := manager.CreateAgent("test-agent", "worker", config)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/"+ag.ID+"/start", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "running", response["state"])
}

func TestStopAgent(t *testing.T) {
	router, manager := setupTestRouter()
	defer manager.Shutdown()

	config := agent.Config{MaxConcurrentTasks: 5, TaskQueueSize: 50}
	ag, err := manager.CreateAgent("test-agent", "worker", config)
	require.NoError(t, err)

	// Start the agent first
	err = manager.StartAgent(ag.ID)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/"+ag.ID+"/stop", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "stopped", response["state"])
}

func TestSubmitTask(t *testing.T) {
	router, manager := setupTestRouter()
	defer manager.Shutdown()

	config := agent.Config{MaxConcurrentTasks: 5, TaskQueueSize: 50}
	ag, err := manager.CreateAgent("test-agent", "worker", config)
	require.NoError(t, err)

	// Start the agent
	err = manager.StartAgent(ag.ID)
	require.NoError(t, err)

	taskReq := SubmitTaskRequest{
		Type:     "process",
		Payload:  map[string]interface{}{"data": "test"},
		Priority: 1,
	}

	body, _ := json.Marshal(taskReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/"+ag.ID+"/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response["task_id"])
	assert.Equal(t, ag.ID, response["agent_id"])
	assert.Equal(t, "task submitted successfully", response["message"])
}

func TestSubmitTaskInvalidAgent(t *testing.T) {
	router, manager := setupTestRouter()
	defer manager.Shutdown()

	taskReq := SubmitTaskRequest{
		Type:     "process",
		Payload:  map[string]interface{}{"data": "test"},
		Priority: 1,
	}

	body, _ := json.Marshal(taskReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/nonexistent/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetMetrics(t *testing.T) {
	router, manager := setupTestRouter()
	defer manager.Shutdown()

	// Create and start some agents
	config := agent.Config{MaxConcurrentTasks: 5, TaskQueueSize: 50}
	ag, _ := manager.CreateAgent("test-agent", "worker", config)
	manager.StartAgent(ag.ID)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response["total_agents_created"])
	assert.NotNil(t, response["total_agents_stopped"])
	assert.NotNil(t, response["total_tasks_executed"])
	assert.NotNil(t, response["total_tasks_failed"])
	assert.NotNil(t, response["current_active_agents"])
	assert.NotNil(t, response["current_running_tasks"])
}
