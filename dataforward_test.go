package mapon

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSaveDataForward(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/data_forward/insert_update.json" {
			t.Errorf("expected /data_forward/insert_update.json, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id": int64(12345),
			},
		}); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewClient(context.Background(), WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	client.baseURL = server.URL

	endpointID, err := client.SaveDataForward(context.Background(), &SaveDataForwardRequest{
		URL:   "https://example.com/webhook",
		Packs: []int32{1, 3, 5},
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if endpointID != 12345 {
		t.Errorf("expected endpoint ID 12345, got %d", endpointID)
	}
}

func TestDeleteDataForward(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/data_forward/delete.json" {
			t.Errorf("expected /data_forward/delete.json, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{}); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewClient(context.Background(), WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	client.baseURL = server.URL

	err = client.DeleteDataForward(context.Background(), &DeleteDataForwardRequest{
		EndpointID: 12345,
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestListDataForwards(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/data_forward/list.json" {
			t.Errorf("expected /data_forward/list.json, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"endpoints": []map[string]interface{}{
					{
						"id":    int64(12345),
						"url":   "https://example.com/webhook1",
						"packs": []int32{1, 3, 5},
					},
					{
						"id":    int64(12346),
						"url":   "https://example.com/webhook2",
						"packs": []int32{26, 55},
					},
				},
			},
		}); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewClient(context.Background(), WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	client.baseURL = server.URL

	resp, err := client.ListDataForwards(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(resp.Endpoints) != 2 {
		t.Errorf("expected 2 endpoints, got %d", len(resp.Endpoints))
	}
	if resp.Endpoints[0].ID != 12345 {
		t.Errorf("expected first endpoint ID 12345, got %d", resp.Endpoints[0].ID)
	}
	if resp.Endpoints[0].URL != "https://example.com/webhook1" {
		t.Errorf("expected first endpoint URL https://example.com/webhook1, got %s", resp.Endpoints[0].URL)
	}
	if len(resp.Endpoints[0].Packs) != 3 {
		t.Errorf("expected 3 packs for first endpoint, got %d", len(resp.Endpoints[0].Packs))
	}
}

func TestSaveDataForwardError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"code": 400,
				"msg":  "Invalid URL",
			},
		}); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewClient(context.Background(), WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	client.baseURL = server.URL

	_, err = client.SaveDataForward(context.Background(), &SaveDataForwardRequest{
		URL:   "",
		Packs: []int32{1},
	})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestDeleteDataForwardHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("Not found"))
	}))
	defer server.Close()

	client, err := NewClient(context.Background(), WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	client.baseURL = server.URL

	err = client.DeleteDataForward(context.Background(), &DeleteDataForwardRequest{
		EndpointID: 99999,
	})
	if err == nil {
		t.Error("expected error, got nil")
	}
}
