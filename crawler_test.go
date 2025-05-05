package webexcrawler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetRooms(t *testing.T) {
	// Mock response for the Webex API
	mockResponse := `{
		"items": [
			{"id": "1", "title": "Room 1"},
			{"id": "2", "title": "Room 2"},
			{"id": "3", "title": "Room 3"},
			{"id": "4", "title": "Room 4"},
			{"id": "5", "title": "Room 5"}
		]
	}`

	// Create a mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/rooms" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockResponse))
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	// Create a new crawler with a mock API key and baseURL
	crawler := &Crawler{
		ApiKey:  "mock-api-key",
		baseUrl: mockServer.URL,
		client:  *http.DefaultClient,
	}

	// Test the GetRooms function
	rooms, err := crawler.GetRooms(5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the number of rooms
	if len(rooms) != 5 {
		t.Errorf("expected 5 rooms, got %d", len(rooms))
	}
}
