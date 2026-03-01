package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tlanfer/SpoolToTag/openspool"
)

type mockAnalyzer struct {
	result openspool.SpoolData
	err    error
}

func (m *mockAnalyzer) Analyze(_ context.Context, _ []byte, _ string) (openspool.SpoolData, error) {
	return m.result, m.err
}

func TestAnalyze_Success(t *testing.T) {
	spool, _ := openspool.New("PLA", "#FF5733", "eSun", 190, 220)
	mock := &mockAnalyzer{result: spool}
	h := New(mock)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("image", "photo.jpg")
	part.Write([]byte("fake image data"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/analyze", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var result openspool.SpoolData
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result.Type != "PLA" {
		t.Errorf("type = %q, want %q", result.Type, "PLA")
	}
	if result.Protocol != "openspool" {
		t.Errorf("protocol = %q, want %q", result.Protocol, "openspool")
	}
}

func TestAnalyze_MissingImage(t *testing.T) {
	mock := &mockAnalyzer{}
	h := New(mock)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/analyze", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestAnalyze_AnalyzerError(t *testing.T) {
	mock := &mockAnalyzer{err: fmt.Errorf("API error")}
	h := New(mock)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("image", "photo.jpg")
	part.Write([]byte("fake image data"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/analyze", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestStaticFiles(t *testing.T) {
	mock := &mockAnalyzer{}
	h := New(mock)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct == "" {
		t.Error("expected Content-Type header")
	}
}
