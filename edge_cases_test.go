package digicert

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestClientErrorHandling tests various HTTP error scenarios
func TestClientErrorHandling(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("network timeout", func(t *testing.T) {
		// Use a client with very short timeout
		client.client.Timeout = 1 * time.Millisecond

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate slow response
			time.Sleep(10 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		_, _, err := client.Certificates.Search(ctx, nil)
		if err == nil {
			t.Fatal("Expected timeout error")
		}

		if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "deadline") {
			t.Errorf("Expected timeout error, got: %v", err)
		}
	})

	t.Run("malformed JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"invalid": json}`)) // Invalid JSON
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		_, _, err := client.Certificates.Search(ctx, nil)
		if err == nil {
			t.Fatal("Expected JSON decode error")
		}

		if !strings.Contains(err.Error(), "decode") && !strings.Contains(err.Error(), "unmarshal") {
			t.Errorf("Expected JSON decode error, got: %v", err)
		}
	})

	t.Run("empty response body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			// Empty body
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		result, resp, err := client.Certificates.Search(ctx, nil)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusOK)
		}

		if result == nil {
			t.Error("Expected non-nil result, even for empty response")
		}
		
		// For empty response, expect zero values
		if result.Total != 0 || result.Offset != 0 || result.Limit != 0 {
			t.Errorf("Expected zero values for empty response, got Total=%d, Offset=%d, Limit=%d", 
				result.Total, result.Offset, result.Limit)
		}
		
		if result.Items != nil {
			t.Errorf("Expected nil Items for empty response, got %v", result.Items)
		}
	})

	t.Run("HTTP 500 internal server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(APIError{
				Code:    "INTERNAL_ERROR",
				Message: "Internal server error occurred",
			})
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		_, _, err := client.Certificates.Search(ctx, nil)
		if err == nil {
			t.Fatal("Expected server error")
		}

		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("Error type = %T, want *APIError", err)
		}

		if apiErr.StatusCode != http.StatusInternalServerError {
			t.Errorf("StatusCode = %v, want %v", apiErr.StatusCode, http.StatusInternalServerError)
		}

		if apiErr.Code != "INTERNAL_ERROR" {
			t.Errorf("Error Code = %v, want %v", apiErr.Code, "INTERNAL_ERROR")
		}
	})

	t.Run("HTTP 429 rate limiting", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "30")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(APIError{
				Code:    "RATE_LIMIT_EXCEEDED",
				Message: "Rate limit exceeded. Please try again later.",
			})
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		_, _, err := client.BusinessUnits.List(ctx, nil)
		if err == nil {
			t.Fatal("Expected rate limit error")
		}

		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("Error type = %T, want *APIError", err)
		}

		if apiErr.StatusCode != http.StatusTooManyRequests {
			t.Errorf("StatusCode = %v, want %v", apiErr.StatusCode, http.StatusTooManyRequests)
		}

		if apiErr.Code != "RATE_LIMIT_EXCEEDED" {
			t.Errorf("Error Code = %v, want %v", apiErr.Code, "RATE_LIMIT_EXCEEDED")
		}
	})

	t.Run("HTTP 503 service unavailable", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("<html><body>Service Unavailable</body></html>"))
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		_, _, err := client.Profiles.List(ctx, nil)
		if err == nil {
			t.Fatal("Expected service unavailable error")
		}

		httpErr, ok := err.(*HTTPError)
		if !ok {
			t.Fatalf("Error type = %T, want *HTTPError", err)
		}

		if httpErr.StatusCode != http.StatusServiceUnavailable {
			t.Errorf("StatusCode = %v, want %v", httpErr.StatusCode, http.StatusServiceUnavailable)
		}
	})
}

// TestPaginationEdgeCases tests edge cases in pagination handling
func TestPaginationEdgeCases(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("negative offset and limit values", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()

			// Negative values should not be added to query
			if q.Has("offset") {
				t.Errorf("offset parameter should not be present for negative value")
			}
			if q.Has("limit") {
				t.Errorf("limit parameter should not be present for negative value")
			}

			mockResponse := &CertificateSearchResponse{
				ListResponse: ListResponse{Total: 0, Offset: 0, Limit: 0},
				Items:        []Certificate{},
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		opts := &CertificateSearchOptions{
			PaginationParams: PaginationParams{
				Offset: -10, // Negative value should not be added
				Limit:  -5,  // Negative value should not be added
			},
		}

		_, _, err := client.Certificates.Search(ctx, opts)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}
	})

	t.Run("very large pagination values", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()

			if q.Get("offset") != "999999" {
				t.Errorf("Expected offset=999999, got %s", q.Get("offset"))
			}
			if q.Get("limit") != "10000" {
				t.Errorf("Expected limit=10000, got %s", q.Get("limit"))
			}

			// Server might return an error for unreasonable pagination
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(APIError{
				Code:    "INVALID_PAGINATION",
				Message: "Offset exceeds maximum allowed value",
			})
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		opts := &CertificateSearchOptions{
			PaginationParams: PaginationParams{
				Offset: 999999,
				Limit:  10000,
			},
		}

		_, _, err := client.Certificates.Search(ctx, opts)
		if err == nil {
			t.Fatal("Expected error for large pagination values")
		}

		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("Error type = %T, want *APIError", err)
		}

		if apiErr.Code != "INVALID_PAGINATION" {
			t.Errorf("Error Code = %v, want %v", apiErr.Code, "INVALID_PAGINATION")
		}
	})

	t.Run("mismatched pagination response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Server returns inconsistent pagination data
			mockResponse := &CertificateSearchResponse{
				ListResponse: ListResponse{
					Total:  100,
					Offset: 20,
					Limit:  10,
				},
				Items: make([]Certificate, 15), // More items than limit suggests
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		opts := &CertificateSearchOptions{
			PaginationParams: PaginationParams{
				Offset: 20,
				Limit:  10,
			},
		}

		result, _, err := client.Certificates.Search(ctx, opts)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		// Verify we still get the response even with inconsistent data
		if result.Limit != 10 {
			t.Errorf("Limit = %v, want %v", result.Limit, 10)
		}

		if len(result.Items) != 15 {
			t.Errorf("Items count = %v, want %v", len(result.Items), 15)
		}
	})
}

// TestInputValidationEdgeCases tests edge cases in input validation
func TestInputValidationEdgeCases(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("empty string parameters", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()

			// Empty strings should not be added to query parameters
			if q.Has("common_name") {
				t.Errorf("common_name parameter should not be present for empty value")
			}
			if q.Has("serial_number") {
				t.Errorf("serial_number parameter should not be present for empty value")
			}
			if q.Has("status") {
				t.Errorf("status parameter should not be present for empty value")
			}

			mockResponse := &CertificateSearchResponse{
				ListResponse: ListResponse{Total: 10, Offset: 0, Limit: 0},
				Items:        make([]Certificate, 10),
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		opts := &CertificateSearchOptions{
			CommonName:   "", // Empty string should not be added
			SerialNumber: "", // Empty string should not be added
			Status:       "", // Empty string should not be added
			ProfileID:    "profile-123", // Non-empty should be added
		}

		_, _, err := client.Certificates.Search(ctx, opts)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}
	})

	t.Run("special characters in search parameters", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()

			// URL encoding should handle special characters
			expectedCN := "test@domain.com & 'special' chars"
			if q.Get("common_name") != expectedCN {
				t.Errorf("Expected common_name=%s, got %s", expectedCN, q.Get("common_name"))
			}

			mockResponse := &CertificateSearchResponse{
				ListResponse: ListResponse{Total: 1, Offset: 0, Limit: 0},
				Items:        []Certificate{{CommonName: expectedCN}},
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		opts := &CertificateSearchOptions{
			CommonName: "test@domain.com & 'special' chars",
		}

		result, _, err := client.Certificates.Search(ctx, opts)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		if result.Items[0].CommonName != opts.CommonName {
			t.Errorf("CommonName = %v, want %v", result.Items[0].CommonName, opts.CommonName)
		}
	})

	t.Run("unicode characters in parameters", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()

			expectedName := "測試證書" // Chinese characters
			if q.Get("common_name") != expectedName {
				t.Errorf("Expected common_name=%s, got %s", expectedName, q.Get("common_name"))
			}

			mockResponse := &CertificateSearchResponse{
				ListResponse: ListResponse{Total: 1, Offset: 0, Limit: 0},
				Items:        []Certificate{{CommonName: expectedName}},
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		opts := &CertificateSearchOptions{
			CommonName: "測試證書",
		}

		result, _, err := client.Certificates.Search(ctx, opts)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		if result.Items[0].CommonName != opts.CommonName {
			t.Errorf("CommonName = %v, want %v", result.Items[0].CommonName, opts.CommonName)
		}
	})
}

// TestConcurrentRequests tests concurrent API calls
func TestConcurrentRequests(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("concurrent certificate searches", func(t *testing.T) {
		requestCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			// Simulate some processing time
			time.Sleep(10 * time.Millisecond)

			mockResponse := &CertificateSearchResponse{
				ListResponse: ListResponse{Total: 1, Offset: 0, Limit: 0},
				Items:        []Certificate{{ID: "cert-concurrent"}},
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		// Launch multiple concurrent requests
		const numRequests = 10
		results := make(chan error, numRequests)

		for i := 0; i < numRequests; i++ {
			go func() {
				_, _, err := client.Certificates.Search(ctx, &CertificateSearchOptions{
					Status: "active",
				})
				results <- err
			}()
		}

		// Wait for all requests to complete
		for i := 0; i < numRequests; i++ {
			if err := <-results; err != nil {
				t.Errorf("Concurrent request %d failed: %v", i, err)
			}
		}

		if requestCount != numRequests {
			t.Errorf("Expected %d requests, got %d", numRequests, requestCount)
		}
	})
}

// TestContextCancellation tests context cancellation handling
func TestContextCancellation(t *testing.T) {
	client, _ := NewClient("test-key")

	t.Run("cancelled context", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate slow response
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		// Create context that cancels immediately
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, _, err := client.Certificates.Search(ctx, nil)
		if err == nil {
			t.Fatal("Expected error for cancelled context")
		}

		if !strings.Contains(err.Error(), "context canceled") {
			t.Errorf("Expected context canceled error, got: %v", err)
		}
	})

	t.Run("timeout context", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate slow response
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		// Create context with very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		_, _, err := client.Certificates.Search(ctx, nil)
		if err == nil {
			t.Fatal("Expected timeout error")
		}

		if !strings.Contains(err.Error(), "deadline exceeded") &&
			!strings.Contains(err.Error(), "context deadline exceeded") {
			t.Errorf("Expected deadline exceeded error, got: %v", err)
		}
	})
}

// TestClientConfiguration tests client configuration edge cases
func TestClientConfiguration(t *testing.T) {
	t.Run("invalid base URL", func(t *testing.T) {
		_, err := NewClient("test-key", WithBaseURL("://invalid-url"))
		if err == nil {
			t.Fatal("Expected error for invalid base URL")
		}

		if !strings.Contains(err.Error(), "invalid base URL") {
			t.Errorf("Expected base URL error, got: %v", err)
		}
	})

	t.Run("nil HTTP client", func(t *testing.T) {
		_, err := NewClient("test-key", WithHTTPClient(nil))
		if err == nil {
			t.Fatal("Expected error for nil HTTP client")
		}

		if !strings.Contains(err.Error(), "cannot be nil") {
			t.Errorf("Expected nil client error, got: %v", err)
		}
	})

	t.Run("empty user agent", func(t *testing.T) {
		client, err := NewClient("test-key", WithUserAgent(""))
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}

		if client.UserAgent != "" {
			t.Errorf("UserAgent = %v, want empty string", client.UserAgent)
		}
	})

	t.Run("multiple client options", func(t *testing.T) {
		customClient := &http.Client{Timeout: 5 * time.Second}
		customURL := "https://custom.digicert.com"
		customUA := "custom-agent/2.0"

		client, err := NewClient("test-key",
			WithHTTPClient(customClient),
			WithBaseURL(customURL),
			WithUserAgent(customUA))

		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}

		if client.client != customClient {
			t.Error("HTTP client not set correctly")
		}

		if client.BaseURL.String() != customURL {
			t.Errorf("BaseURL = %v, want %v", client.BaseURL.String(), customURL)
		}

		if client.UserAgent != customUA {
			t.Errorf("UserAgent = %v, want %v", client.UserAgent, customUA)
		}
	})
}

// TestResponseSizeHandling tests handling of very large responses
func TestResponseSizeHandling(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("large response handling", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a large response with many certificates
			items := make([]Certificate, 1000)
			for i := range items {
				items[i] = Certificate{
					ID:         string(rune('A' + i%26)) + string(rune('a' + (i/26)%26)) + string(rune('0' + (i/676)%10)),
					CommonName: "large-response-test-" + string(rune('0' + i%10)) + ".example.com",
					Status:     "active",
					SerialNumber: string(rune('1'+i%9)) + string(rune('0'+i%10)) + string(rune('A'+(i/10)%26)),
				}
			}

			mockResponse := &CertificateSearchResponse{
				ListResponse: ListResponse{Total: 1000, Offset: 0, Limit: 1000},
				Items:        items,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		result, resp, err := client.Certificates.Search(ctx, &CertificateSearchOptions{
			PaginationParams: PaginationParams{Limit: 1000},
		})

		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusOK)
		}

		if len(result.Items) != 1000 {
			t.Errorf("Items count = %v, want %v", len(result.Items), 1000)
		}

		// Verify first and last items
		if result.Items[0].CommonName != "large-response-test-0.example.com" {
			t.Errorf("First item CommonName = %v", result.Items[0].CommonName)
		}

		if result.Items[999].CommonName != "large-response-test-9.example.com" {
			t.Errorf("Last item CommonName = %v", result.Items[999].CommonName)
		}
	})
}