package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Analyze_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("path = %s, want /v1/chat/completions", r.URL.Path)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-key" {
			t.Errorf("auth = %q, want %q", auth, "Bearer test-key")
		}

		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Model != "gpt-4o" {
			t.Errorf("model = %q, want %q", req.Model, "gpt-4o")
		}

		resp := chatResponse{
			Choices: []choice{
				{
					Message: responseMessage{
						Content: `{"type":"PLA","color_hex":"FF5733","brand":"Bambu","min_temp":190,"max_temp":220}`,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		APIKey:  "test-key",
		Model:   "gpt-4o",
		BaseURL: server.URL,
	}

	spool, err := client.Analyze(context.Background(), []byte("fake-image"), "image/jpeg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spool.Type != "PLA" {
		t.Errorf("type = %q, want %q", spool.Type, "PLA")
	}
	if spool.ColorHex != "FF5733" {
		t.Errorf("color_hex = %q, want %q", spool.ColorHex, "FF5733")
	}
	if spool.Brand != "Generic" {
		t.Errorf("brand = %q, want %q", spool.Brand, "Generic")
	}
	if spool.MinTemp != 190 {
		t.Errorf("min_temp = %d, want %d", spool.MinTemp, 190)
	}
	if spool.MaxTemp != 220 {
		t.Errorf("max_temp = %d, want %d", spool.MaxTemp, 220)
	}
	if spool.Protocol != "openspool" {
		t.Errorf("protocol = %q, want %q", spool.Protocol, "openspool")
	}
}

func TestClient_Analyze_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"invalid key"}`))
	}))
	defer server.Close()

	client := &Client{
		APIKey:  "bad-key",
		Model:   "gpt-4o",
		BaseURL: server.URL,
	}

	_, err := client.Analyze(context.Background(), []byte("fake-image"), "image/jpeg")
	if err == nil {
		t.Error("expected error for 401 response")
	}
}

func TestClient_Analyze_NoChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := chatResponse{Choices: []choice{}}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		APIKey:  "test-key",
		Model:   "gpt-4o",
		BaseURL: server.URL,
	}

	_, err := client.Analyze(context.Background(), []byte("fake-image"), "image/jpeg")
	if err == nil {
		t.Error("expected error for empty choices")
	}
}
