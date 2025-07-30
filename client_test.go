package digicert

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		wantErr bool
	}{
		{
			name:    "valid API key",
			apiKey:  "test-api-key",
			wantErr: false,
		},
		{
			name:    "empty API key",
			apiKey:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() returned nil client")
			}
		})
	}
}

func TestClientOptions(t *testing.T) {
	t.Run("WithBaseURL", func(t *testing.T) {
		customURL := "https://custom.digicert.com"
		client, err := NewClient("test-key", WithBaseURL(customURL))
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		if client.BaseURL.String() != customURL {
			t.Errorf("BaseURL = %v, want %v", client.BaseURL.String(), customURL)
		}
	})

	t.Run("WithHTTPClient", func(t *testing.T) {
		customClient := &http.Client{}
		client, err := NewClient("test-key", WithHTTPClient(customClient))
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		if client.client != customClient {
			t.Error("HTTP client was not set correctly")
		}
	})

	t.Run("WithUserAgent", func(t *testing.T) {
		customUA := "test-app/1.0"
		client, err := NewClient("test-key", WithUserAgent(customUA))
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		if client.UserAgent != customUA {
			t.Errorf("UserAgent = %v, want %v", client.UserAgent, customUA)
		}
	})
}

func TestClient_NewRequest(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("GET request", func(t *testing.T) {
		req, err := client.NewRequest(ctx, http.MethodGet, "test", nil)
		if err != nil {
			t.Fatalf("NewRequest() error = %v", err)
		}

		if req.Method != http.MethodGet {
			t.Errorf("Method = %v, want %v", req.Method, http.MethodGet)
		}

		if req.Header.Get("X-API-Key") != "test-key" {
			t.Error("API key header not set correctly")
		}

		expectedURL := "https://one.digicert.com/mpki/api/v1/test"
		if req.URL.String() != expectedURL {
			t.Errorf("URL = %v, want %v", req.URL.String(), expectedURL)
		}
	})

	t.Run("POST request with body", func(t *testing.T) {
		body := map[string]string{"key": "value"}
		req, err := client.NewRequest(ctx, http.MethodPost, "test", body)
		if err != nil {
			t.Fatalf("NewRequest() error = %v", err)
		}

		if req.Header.Get("Content-Type") != "application/json" {
			t.Error("Content-Type header not set for POST request")
		}
	})
}

func TestClient_Do(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "success"})
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")
		req, _ := client.NewRequest(ctx, http.MethodGet, "test", nil)

		var result map[string]string
		resp, err := client.Do(ctx, req, &result)
		if err != nil {
			t.Fatalf("Do() error = %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusOK)
		}

		if result["status"] != "success" {
			t.Errorf("Response status = %v, want %v", result["status"], "success")
		}
	})

	t.Run("error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(APIError{
				Code:    "NOT_FOUND",
				Message: "Resource not found",
			})
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")
		req, _ := client.NewRequest(ctx, http.MethodGet, "test", nil)

		_, err := client.Do(ctx, req, nil)
		if err == nil {
			t.Fatal("Expected error for 404 response")
		}

		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("Error type = %T, want *APIError", err)
		}

		if apiErr.StatusCode != http.StatusNotFound {
			t.Errorf("Error StatusCode = %v, want %v", apiErr.StatusCode, http.StatusNotFound)
		}

		if apiErr.Code != "NOT_FOUND" {
			t.Errorf("Error Code = %v, want %v", apiErr.Code, "NOT_FOUND")
		}
	})
}

func TestErrorHelpers(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		checkFn  func(error) bool
		expected bool
	}{
		{
			name:     "IsNotFound with APIError 404",
			err:      &APIError{StatusCode: 404},
			checkFn:  IsNotFound,
			expected: true,
		},
		{
			name:     "IsNotFound with APIError 403",
			err:      &APIError{StatusCode: 403},
			checkFn:  IsNotFound,
			expected: false,
		},
		{
			name:     "IsUnauthorized with APIError 401",
			err:      &APIError{StatusCode: 401},
			checkFn:  IsUnauthorized,
			expected: true,
		},
		{
			name:     "IsForbidden with HTTPError 403",
			err:      &HTTPError{StatusCode: 403},
			checkFn:  IsForbidden,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.checkFn(tt.err); got != tt.expected {
				t.Errorf("%s() = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

func TestCertificatesService_GetCertificate(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("successful GetCertificate", func(t *testing.T) {
		mockCert := Certificate{
			ID:           "386e7879-df28-427b-b8ae-58205a8b87df",
			CommonName:   "helix-o2ukusm-dev.uk.pri.o2.com",
			Status:       "issued",
			SerialNumber: "520000AAEBDDE4D0EEBF6703DA00020000AAEB",
			Thumbprint:   "F28D04DDE45ADFBAEDF1B3A364227B17F89623FCC12489D858D14C5457A72547",
			ValidFrom:    "2023-12-06T04:42:39Z",
			ValidTo:      "2025-12-05T04:42:39Z",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/mpki/api/v1/certificate-by-id/386e7879-df28-427b-b8ae-58205a8b87df" {
				t.Errorf("Expected path /mpki/api/v1/certificate-by-id/386e7879-df28-427b-b8ae-58205a8b87df, got %s", r.URL.Path)
			}
			if r.Method != http.MethodGet {
				t.Errorf("Expected GET request, got %s", r.Method)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockCert)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		cert, resp, err := client.Certificates.GetCertificate(ctx, "386e7879-df28-427b-b8ae-58205a8b87df")
		if err != nil {
			t.Fatalf("GetCertificate() error = %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusOK)
		}

		if cert.ID != mockCert.ID {
			t.Errorf("Certificate ID = %v, want %v", cert.ID, mockCert.ID)
		}

		if cert.CommonName != mockCert.CommonName {
			t.Errorf("Certificate CommonName = %v, want %v", cert.CommonName, mockCert.CommonName)
		}

		if cert.Status != mockCert.Status {
			t.Errorf("Certificate Status = %v, want %v", cert.Status, mockCert.Status)
		}
	})

	t.Run("certificate not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(APIError{
				Code:    "NOT_FOUND",
				Message: "Certificate not found",
			})
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		_, _, err := client.Certificates.GetCertificate(ctx, "nonexistent-id")
		if err == nil {
			t.Fatal("Expected error for nonexistent certificate")
		}

		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("Error type = %T, want *APIError", err)
		}

		if apiErr.StatusCode != http.StatusNotFound {
			t.Errorf("Error StatusCode = %v, want %v", apiErr.StatusCode, http.StatusNotFound)
		}

		if apiErr.Code != "NOT_FOUND" {
			t.Errorf("Error Code = %v, want %v", apiErr.Code, "NOT_FOUND")
		}
	})
}