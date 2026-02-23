package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"isac-cran-system/internal/handler"
	"isac-cran-system/internal/router"

	"github.com/gin-gonic/gin"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	irsHandler := handler.NewIRSHandler(nil)
	channelHandler := handler.NewChannelHandler(nil)
	algorithmHandler := handler.NewAlgorithmHandler(nil)
	sensorHandler := handler.NewSensorHandler(nil)
	systemHandler := handler.NewSystemHandler()

	return router.Setup(irsHandler, channelHandler, algorithmHandler, sensorHandler, systemHandler)
}

func TestHealthEndpoint(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["code"].(float64) != 0 {
		t.Errorf("Expected code 0, got %v", response["code"])
	}
}

func TestInfoEndpoint(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/info", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})
	if data["name"] != "ISAC-CRAN System" {
		t.Errorf("Expected name 'ISAC-CRAN System', got %v", data["name"])
	}
}

func TestAPIRoutesRegistered(t *testing.T) {
	router := setupTestRouter()

	routes := router.Routes()

	expectedRoutes := []string{
		"/api/v1/health",
		"/api/v1/info",
		"/api/v1/irs/config",
		"/api/v1/irs/status",
		"/api/v1/channel/data",
		"/api/v1/algorithm/beamforming",
		"/api/v1/algorithm/doa",
		"/api/v1/sensor/list",
	}

	routeMap := make(map[string]bool)
	for _, route := range routes {
		routeMap[route.Path] = true
	}

	for _, expected := range expectedRoutes {
		if !routeMap[expected] {
			t.Errorf("Expected route %s to be registered", expected)
		}
	}
}

func TestCORSMiddleware(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("OPTIONS", "/api/v1/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected CORS header to be set")
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	requestID := w.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Errorf("Expected X-Request-ID header to be set")
	}
}

func TestResponseFormat(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse response JSON: %v", err)
	}

	requiredFields := []string{"code", "message", "data"}
	for _, field := range requiredFields {
		if _, exists := response[field]; !exists {
			t.Errorf("Expected field '%s' in response", field)
		}
	}
}
