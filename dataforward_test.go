package mapon

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	maponv1 "github.com/way-platform/mapon-go/proto/gen/go/wayplatform/connect/mapon/v1"
)

func TestSaveDataForward(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/data_forward/save.json" {
			t.Errorf("expected /data_forward/save.json, got %s", r.URL.Path)
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

	req := &maponv1.SaveDataForwardRequest{}
	req.SetUrl("https://example.com/webhook")
	req.SetPacks([]int32{1, 3, 5})
	resp, err := client.SaveDataForward(context.Background(), req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp.GetEndpointId() != 12345 {
		t.Errorf("expected endpoint ID 12345, got %d", resp.GetEndpointId())
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

	req := &maponv1.DeleteDataForwardRequest{}
	req.SetEndpointId(12345)
	_, err = client.DeleteDataForward(context.Background(), req)
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

	resp, err := client.ListDataForwards(context.Background(), &maponv1.ListDataForwardsRequest{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	endpoints := resp.GetEndpoints()
	if len(endpoints) != 2 {
		t.Errorf("expected 2 endpoints, got %d", len(endpoints))
	}
	if endpoints[0].GetId() != 12345 {
		t.Errorf("expected first endpoint ID 12345, got %d", endpoints[0].GetId())
	}
	if endpoints[0].GetUrl() != "https://example.com/webhook1" {
		t.Errorf("expected first endpoint URL https://example.com/webhook1, got %s", endpoints[0].GetUrl())
	}
	if len(endpoints[0].GetPacks()) != 3 {
		t.Errorf("expected 3 packs for first endpoint, got %d", len(endpoints[0].GetPacks()))
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

	req := &maponv1.SaveDataForwardRequest{}
	req.SetUrl("")
	req.SetPacks([]int32{1})
	_, err = client.SaveDataForward(context.Background(), req)
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

	req := &maponv1.DeleteDataForwardRequest{}
	req.SetEndpointId(99999)
	_, err = client.DeleteDataForward(context.Background(), req)
	if err == nil {
		t.Error("expected error, got nil")
	}
}
