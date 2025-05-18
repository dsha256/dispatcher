package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dsha256/dispatcher/internal/dispatcher"
	"github.com/dsha256/dispatcher/internal/handler"
)

// setupTestServer creates a test server with the itinerary handler.
func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	// Create a test logger that discards output
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create a dispatcher service
	dispatcherService := dispatcher.New()

	// Create a handler with the dispatcher service
	h := handler.New(logger, dispatcherService)

	// Create a test server
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	server := httptest.NewServer(mux)

	// Add cleanup to ensure server is closed after test
	t.Cleanup(func() {
		server.Close()
	})

	return server
}

// sendRequest is a helper function to send HTTP requests in tests.
func sendRequest(t *testing.T, server *httptest.Server, method string, body interface{}) (*http.Response, map[string]interface{}) {
	t.Helper()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create request body
	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, method, server.URL+"/api/v1/dispatcher/itinerary", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	// Parse response body
	var respBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		resp.Body.Close()
		t.Fatalf("Failed to decode response body: %v", err)
	}

	return resp, respBody
}

//nolint:gocognit // Just needed.
func TestHandleItinerary(t *testing.T) {
	t.Parallel()

	// Setup test server
	server := setupTestServer(t)

	// Define test cases
	tests := []struct {
		requestBody    map[string]interface{}
		expectedBody   map[string][]string
		name           string
		method         string
		expectedStatus int
		expectedError  bool
	}{
		{
			name:   "Valid itinerary",
			method: http.MethodPost,
			requestBody: map[string]interface{}{
				"tickets": [][]string{{"LAX", "DXB"}, {"JFK", "LAX"}, {"SFO", "SJC"}, {"DXB", "SFO"}},
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string][]string{
				"linear_path": {"JFK", "LAX", "DXB", "SFO", "SJC"},
			},
			expectedError: false,
		},

		{
			name:   "Standard itinerary",
			method: http.MethodPost,
			requestBody: map[string]interface{}{
				"tickets": [][]string{{"LAX", "DXB"}, {"JFK", "LAX"}, {"SFO", "SJC"}, {"DXB", "SFO"}},
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string][]string{
				"linear_path": {"JFK", "LAX", "DXB", "SFO", "SJC"},
			},
			expectedError: false,
		},
		{
			name:   "Multiple possible paths",
			method: http.MethodPost,
			requestBody: map[string]interface{}{
				"tickets": [][]string{{"JFK", "SFO"}, {"JFK", "ATL"}, {"SFO", "ATL"}, {"ATL", "JFK"}},
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string][]string{
				"linear_path": {"JFK", "ATL", "JFK", "SFO", "ATL"},
			},
			expectedError: false,
		},
		{
			name:   "Single ticket",
			method: http.MethodPost,
			requestBody: map[string]interface{}{
				"tickets": [][]string{{"SFO", "JFK"}},
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string][]string{
				"linear_path": {"SFO", "JFK"},
			},
			expectedError: false,
		},
		{
			name:   "Empty tickets",
			method: http.MethodPost,
			requestBody: map[string]interface{}{
				"tickets": [][]string{},
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string][]string{
				"linear_path": {},
			},
			expectedError: false,
		},
		{
			name:   "Different starting point",
			method: http.MethodPost,
			requestBody: map[string]interface{}{
				"tickets": [][]string{{"SFO", "LAX"}, {"LAX", "JFK"}, {"JFK", "SFO"}},
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   nil,
			expectedError:  true,
		},
		{
			name:   "Cycle in itinerary",
			method: http.MethodPost,
			requestBody: map[string]interface{}{
				"tickets": [][]string{{"JFK", "SFO"}, {"SFO", "LAX"}, {"LAX", "JFK"}, {"JFK", "ATL"}},
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string][]string{
				"linear_path": {"JFK", "SFO", "LAX", "JFK", "ATL"},
			},
			expectedError: false,
		},
		{
			name:   "Multiple same destination",
			method: http.MethodPost,
			requestBody: map[string]interface{}{
				"tickets": [][]string{{"JFK", "SFO"}, {"JFK", "ATL"}, {"JFK", "SFO"}, {"SFO", "LAX"}, {"ATL", "LAX"}},
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   nil,
			expectedError:  true,
		},
		{
			name:   "Duplicate tickets turned into cycle",
			method: http.MethodPost,
			requestBody: map[string]interface{}{
				"tickets": [][]string{{"JFK", "SFO"}, {"JFK", "SFO"}, {"SFO", "LAX"}, {"LAX", "ATL"}},
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   nil,
			expectedError:  true,
		},
		{
			name:   "Longer complex itinerary",
			method: http.MethodPost,
			requestBody: map[string]interface{}{
				"tickets": [][]string{{"A", "B"}, {"B", "C"}, {"C", "D"}, {"D", "E"}, {"E", "F"}, {"F", "A"}, {"A", "G"}},
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string][]string{
				"linear_path": {"A", "B", "C", "D", "E", "F", "A", "G"},
			},
			expectedError: false,
		},
		{
			name:   "Multiple same destination error",
			method: http.MethodPost,
			requestBody: map[string]interface{}{
				"tickets": [][]string{{"JFK", "SFO"}, {"JFK", "SFO"}},
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   nil,
			expectedError:  true,
		},
		{
			name:   "Cycle in itinerary error",
			method: http.MethodPost,
			requestBody: map[string]interface{}{
				"tickets": [][]string{{"JFK", "SFO"}, {"SFO", "JFK"}, {"JFK", "SFO"}},
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   nil,
			expectedError:  true,
		},
		{
			name:           "Invalid method",
			method:         http.MethodGet,
			requestBody:    nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Send request and get response
			resp, respBody := sendRequest(t, server, tt.method, tt.requestBody)
			defer resp.Body.Close()

			// Check status code
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			// Check for error response.
			//nolint:nestif // Just needed.
			if tt.expectedError {
				if respBody["err"] == nil {
					t.Errorf("Expected error in response, got none")
				}
			} else {
				// Check response data
				data, ok := respBody["data"].(map[string]interface{})
				if !ok {
					t.Fatalf("Expected data field in response, got %v", respBody)
				}

				// Check linear_path
				linearPath, ok := data["linear_path"].([]interface{})
				if !ok {
					t.Fatalf("Expected linear_path field in data, got %v", data)
				}

				// Convert expected linear_path to []interface{} for comparison
				var expectedLinearPath []interface{}
				for _, v := range tt.expectedBody["linear_path"] {
					expectedLinearPath = append(expectedLinearPath, v)
				}

				// Compare linear_path
				if len(linearPath) != len(expectedLinearPath) {
					t.Errorf("Expected linear_path length %d, got %d", len(expectedLinearPath), len(linearPath))
				} else {
					for i, v := range linearPath {
						if v != expectedLinearPath[i] {
							t.Errorf("Expected linear_path[%d] = %v, got %v", i, expectedLinearPath[i], v)
						}
					}
				}
			}
		})
	}
}

// TestHandleItineraryEdgeCases tests additional edge cases for the itinerary handler.
func TestHandleItineraryEdgeCases(t *testing.T) {
	t.Parallel()

	// Setup test server
	server := setupTestServer(t)

	// Test case: Empty request body
	t.Run("Empty request body", func(t *testing.T) {
		t.Parallel()

		// Send request with nil body
		resp, respBody := sendRequest(t, server, http.MethodPost, nil)
		defer resp.Body.Close()

		// Check status code - should be bad request
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}

		// Check for error response
		if respBody["err"] == nil {
			t.Errorf("Expected error in response, got none")
		}
	})

	// Test case: Large itinerary
	t.Run("Large itinerary", func(t *testing.T) {
		t.Parallel()

		// Create a large itinerary with 100 tickets
		tickets := make([][]string, 100)
		for i := range tickets {
			tickets[i] = []string{fmt.Sprintf("CITY%d", i), fmt.Sprintf("CITY%d", i+1)}
		}
		tickets[99] = []string{fmt.Sprintf("CITY%d", 99), "FINAL"}

		// Create a request body.
		requestBody := map[string]interface{}{
			"tickets": tickets,
		}

		// Send request
		resp, respBody := sendRequest(t, server, http.MethodPost, requestBody)
		defer resp.Body.Close()

		// Check status code
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		// Check for error response - should not have an error
		if respBody["err"] != nil {
			t.Errorf("Expected no error in response, got %v", respBody["err"])
		}

		// Check response data
		data, ok := respBody["data"].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected data field in response, got %v", respBody)
		}

		// Check linear_path
		linearPath, ok := data["linear_path"].([]interface{})
		if !ok {
			t.Fatalf("Expected linear_path field in data, got %v", data)
		}

		// Check that the path has the expected length
		if len(linearPath) != 101 { // 100 tickets + 1 (the final destination)
			t.Errorf("Expected linear_path length %d, got %d", 101, len(linearPath))
		}
	})
}
