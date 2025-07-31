package digicert

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCertificatesService_Issue(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("successful certificate issuance", func(t *testing.T) {
		mockRequest := &CertificateRequest{
			Profile: ProfileReference{ID: "profile-123"},
			CSR:     "-----BEGIN CERTIFICATE REQUEST-----\nMIICYjCCAUoCAQAwHTEbMBkGA1UEAwwSdGVzdC5leGFtcGxlLmNvbQowggEiMA0G\n-----END CERTIFICATE REQUEST-----",
			Validity: &Validity{
				Years: 1,
			},
		}

		mockResponse := &CertificateResponse{
			Certificate: &Certificate{
				ID:           "cert-123",
				CommonName:   "test.example.com",
				Status:       "issued",
				SerialNumber: "123456789",
			},
			RequestID: "req-123",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/mpki/api/v1/certificate" {
				t.Errorf("Expected path /mpki/api/v1/certificate, got %s", r.URL.Path)
			}
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST request, got %s", r.Method)
			}

			var reqBody CertificateRequest
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			if reqBody.Profile.ID != mockRequest.Profile.ID {
				t.Errorf("Expected profile ID %s, got %s", mockRequest.Profile.ID, reqBody.Profile.ID)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		result, resp, err := client.Certificates.Issue(ctx, mockRequest)
		if err != nil {
			t.Fatalf("Issue() error = %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusOK)
		}

		if result.Certificate.ID != mockResponse.Certificate.ID {
			t.Errorf("Certificate ID = %v, want %v", result.Certificate.ID, mockResponse.Certificate.ID)
		}
	})

	t.Run("invalid profile error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(APIError{
				Code:    "INVALID_PROFILE",
				Message: "Profile not found or inactive",
			})
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		mockRequest := &CertificateRequest{
			Profile: ProfileReference{ID: "invalid-profile"},
		}

		_, _, err := client.Certificates.Issue(ctx, mockRequest)
		if err == nil {
			t.Fatal("Expected error for invalid profile")
		}

		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("Error type = %T, want *APIError", err)
		}

		if apiErr.Code != "INVALID_PROFILE" {
			t.Errorf("Error Code = %v, want %v", apiErr.Code, "INVALID_PROFILE")
		}
	})
}

func TestCertificatesService_Search(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("successful search with pagination", func(t *testing.T) {
		mockResponse := &CertificateSearchResponse{
			ListResponse: ListResponse{
				Total:  100,
				Offset: 0,
				Limit:  0, // No limit sent when offset is 0
			},
			Items: []Certificate{
				{
					ID:           "cert-1",
					CommonName:   "example1.com",
					Status:       "issued",
					SerialNumber: "111111111",
					ValidFrom:    "2023-01-01T00:00:00Z",
					ValidTo:      "2024-01-01T00:00:00Z",
				},
				{
					ID:           "cert-2",
					CommonName:   "example2.com",
					Status:       "issued",
					SerialNumber: "222222222",
					ValidFrom:    "2023-02-01T00:00:00Z",
					ValidTo:      "2024-02-01T00:00:00Z",
				},
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/mpki/api/v1/certificate-search" {
				t.Errorf("Expected path /mpki/api/v1/certificate-search, got %s", r.URL.Path)
			}
			if r.Method != http.MethodGet {
				t.Errorf("Expected GET request, got %s", r.Method)
			}

			// Check query parameters
			q := r.URL.Query()
			if q.Get("common_name") != "example.com" {
				t.Errorf("Expected common_name=example.com, got %s", q.Get("common_name"))
			}
			if q.Get("status") != "issued" {
				t.Errorf("Expected status=issued, got %s", q.Get("status"))
			}
			// Neither offset nor limit should be present when offset is 0 (consistent pagination logic)
			if q.Has("offset") {
				t.Errorf("offset parameter should not be present when value is 0, but got %s", q.Get("offset"))
			}
			if q.Has("limit") {
				t.Errorf("limit parameter should not be present when offset is 0")
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		opts := &CertificateSearchOptions{
			PaginationParams: PaginationParams{
				Offset: 0, // Will not be added to query since it's 0
				Limit:  20,
			},
			CommonName: "example.com",
			Status:     "issued",
		}

		result, resp, err := client.Certificates.Search(ctx, opts)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusOK)
		}

		if result.Total != mockResponse.Total {
			t.Errorf("Total = %v, want %v", result.Total, mockResponse.Total)
		}

		if len(result.Items) != len(mockResponse.Items) {
			t.Errorf("Items count = %v, want %v", len(result.Items), len(mockResponse.Items))
		}

		if result.Items[0].CommonName != "example1.com" {
			t.Errorf("First certificate CommonName = %v, want %v", result.Items[0].CommonName, "example1.com")
		}
	})

	t.Run("search with multiple filters", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			
			// Verify all filter parameters
			if q.Get("common_name") != "test.com" {
				t.Errorf("Expected common_name=test.com, got %s", q.Get("common_name"))
			}
			if q.Get("serial_number") != "ABCD1234" {
				t.Errorf("Expected serial_number=ABCD1234, got %s", q.Get("serial_number"))
			}
			if q.Get("status") != "issued" {
				t.Errorf("Expected status=issued, got %s", q.Get("status"))
			}
			if q.Get("profile_id") != "profile-456" {
				t.Errorf("Expected profile_id=profile-456, got %s", q.Get("profile_id"))
			}

			// Check multiple tags
			tags := q["tags"]
			expectedTags := []string{"production", "web-server"}
			if len(tags) != len(expectedTags) {
				t.Errorf("Expected %d tags, got %d", len(expectedTags), len(tags))
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(&CertificateSearchResponse{
				ListResponse: ListResponse{Total: 1, Offset: 0, Limit: 20},
				Items:        []Certificate{},
			})
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		opts := &CertificateSearchOptions{
			CommonName:   "test.com",
			SerialNumber: "ABCD1234",
			Status:       "issued",
			ProfileID:    "profile-456",
			Tags:         []string{"production", "web-server"},
		}

		_, _, err := client.Certificates.Search(ctx, opts)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}
	})
}

func TestCertificatesService_Revoke(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("successful certificate revocation", func(t *testing.T) {
		serialNumber := "123456789ABCDEF"
		revokeReq := &RevokeRequest{
			Reason:  "keyCompromise",
			Comment: "Private key was compromised",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/mpki/api/v1/certificate/" + serialNumber + "/revoke"
			if r.URL.Path != expectedPath {
				t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
			}
			if r.Method != http.MethodPut {
				t.Errorf("Expected PUT request, got %s", r.Method)
			}

			var reqBody RevokeRequest
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			if reqBody.Reason != revokeReq.Reason {
				t.Errorf("Expected reason %s, got %s", revokeReq.Reason, reqBody.Reason)
			}

			if reqBody.Comment != revokeReq.Comment {
				t.Errorf("Expected comment %s, got %s", revokeReq.Comment, reqBody.Comment)
			}

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		resp, err := client.Certificates.Revoke(ctx, serialNumber, revokeReq)
		if err != nil {
			t.Fatalf("Revoke() error = %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusOK)
		}
	})
}

func TestCertificatesService_Renew(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("successful certificate renewal", func(t *testing.T) {
		serialNumber := "123456789ABCDEF"
		renewReq := &RenewRequest{
			CSR: "-----BEGIN CERTIFICATE REQUEST-----\nMIICYjCCAUoCAQAwHTEbMBkGA1UEAwwSdGVzdC5leGFtcGxlLmNvbQowggEiMA0G\n-----END CERTIFICATE REQUEST-----",
			Validity: &Validity{
				Years: 2,
			},
		}

		mockResponse := &CertificateResponse{
			Certificate: &Certificate{
				ID:           "cert-renewed-123",
				CommonName:   "test.example.com",
				Status:       "issued",
				SerialNumber: "987654321FEDCBA",
			},
			RequestID: "renew-req-123",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/mpki/api/v1/certificate/" + serialNumber + "/renew"
			if r.URL.Path != expectedPath {
				t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
			}
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST request, got %s", r.Method)
			}

			var reqBody RenewRequest
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			if reqBody.Validity.Years != renewReq.Validity.Years {
				t.Errorf("Expected validity years %d, got %d", renewReq.Validity.Years, reqBody.Validity.Years)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		result, resp, err := client.Certificates.Renew(ctx, serialNumber, renewReq)
		if err != nil {
			t.Fatalf("Renew() error = %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusOK)
		}

		if result.Certificate.SerialNumber != mockResponse.Certificate.SerialNumber {
			t.Errorf("New certificate serial = %v, want %v", result.Certificate.SerialNumber, mockResponse.Certificate.SerialNumber)
		}
	})

	t.Run("renewal outside renewal period", func(t *testing.T) {
		serialNumber := "123456789ABCDEF"
		renewReq := &RenewRequest{
			Validity: &Validity{Years: 1},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(APIError{
				Code:    "RENEWAL_NOT_ALLOWED",
				Message: "Certificate is not within renewal period",
			})
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		_, _, err := client.Certificates.Renew(ctx, serialNumber, renewReq)
		if err == nil {
			t.Fatal("Expected error for renewal outside period")
		}

		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("Error type = %T, want *APIError", err)
		}

		if apiErr.Code != "RENEWAL_NOT_ALLOWED" {
			t.Errorf("Error Code = %v, want %v", apiErr.Code, "RENEWAL_NOT_ALLOWED")
		}
	})
}

func TestCertificatesService_Pickup(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("successful certificate pickup", func(t *testing.T) {
		requestID := "pickup-req-123"
		
		mockResponse := &CertificateResponse{
			Certificate: &Certificate{
				ID:           "cert-pickup-123",
				CommonName:   "pickup.example.com",
				Status:       "issued",
				SerialNumber: "PICKUP123456789",
			},
			RequestID: requestID,
			Chain:     []string{"-----BEGIN CERTIFICATE-----\nIntermediateCert\n-----END CERTIFICATE-----"},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/mpki/api/v1/certificate-pickup/" + requestID
			if r.URL.Path != expectedPath {
				t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
			}
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST request, got %s", r.Method)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		result, resp, err := client.Certificates.Pickup(ctx, requestID)
		if err != nil {
			t.Fatalf("Pickup() error = %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusOK)
		}

		if result.Certificate.ID != mockResponse.Certificate.ID {
			t.Errorf("Certificate ID = %v, want %v", result.Certificate.ID, mockResponse.Certificate.ID)
		}

		if len(result.Chain) != len(mockResponse.Chain) {
			t.Errorf("Chain length = %v, want %v", len(result.Chain), len(mockResponse.Chain))
		}
	})

	t.Run("pickup request not found", func(t *testing.T) {
		requestID := "nonexistent-req"

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(APIError{
				Code:    "REQUEST_NOT_FOUND",
				Message: "Certificate request not found",
			})
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		_, _, err := client.Certificates.Pickup(ctx, requestID)
		if err == nil {
			t.Fatal("Expected error for nonexistent request")
		}

		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("Error type = %T, want *APIError", err)
		}

		if apiErr.Code != "REQUEST_NOT_FOUND" {
			t.Errorf("Error Code = %v, want %v", apiErr.Code, "REQUEST_NOT_FOUND")
		}
	})
}

func TestCertificatesService_GetAdditionalFormats(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("successful additional formats retrieval", func(t *testing.T) {
		serialNumber := "123456789ABCDEF"
		
		mockResponse := &AdditionalFormatsResponse{
			Formats: map[string]string{
				"pkcs12": "base64encodedpkcs12data",
				"pem":    "-----BEGIN CERTIFICATE-----\nCertificateData\n-----END CERTIFICATE-----",
				"der":    "base64encodedderdata",
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/mpki/api/v1/certificate/" + serialNumber + "/additional-formats"
			if r.URL.Path != expectedPath {
				t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
			}
			if r.Method != http.MethodGet {
				t.Errorf("Expected GET request, got %s", r.Method)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		result, resp, err := client.Certificates.GetAdditionalFormats(ctx, serialNumber)
		if err != nil {
			t.Fatalf("GetAdditionalFormats() error = %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusOK)
		}

		if len(result.Formats) != len(mockResponse.Formats) {
			t.Errorf("Formats count = %v, want %v", len(result.Formats), len(mockResponse.Formats))
		}

		if result.Formats["pkcs12"] != mockResponse.Formats["pkcs12"] {
			t.Errorf("PKCS12 format mismatch")
		}
	})
}

func TestCertificateRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		request *CertificateRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request with CSR",
			request: &CertificateRequest{
				Profile: ProfileReference{ID: "profile-123"},
				CSR:     "-----BEGIN CERTIFICATE REQUEST-----\nvalidcsr\n-----END CERTIFICATE REQUEST-----",
				Validity: &Validity{Years: 1},
			},
			wantErr: false,
		},
		{
			name: "valid request with attributes",
			request: &CertificateRequest{
				Profile: ProfileReference{ID: "profile-123"},
				Attributes: &CertificateAttributes{
					CommonName: "test.example.com",
					Country:    "US",
					State:      "CA",
				},
				Validity: &Validity{Years: 1},
			},
			wantErr: false,
		},
		{
			name: "request with custom attributes",
			request: &CertificateRequest{
				Profile:  ProfileReference{ID: "profile-123"},
				CSR:      "-----BEGIN CERTIFICATE REQUEST-----\nvalidcsr\n-----END CERTIFICATE REQUEST-----",
				Validity: &Validity{Years: 1},
				CustomAttributes: []CustomAttribute{
					{ID: "attr1", Value: "value1"},
					{ID: "attr2", Value: "value2"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling/unmarshaling
			data, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			var unmarshaledReq CertificateRequest
			if err := json.Unmarshal(data, &unmarshaledReq); err != nil {
				t.Fatalf("Failed to unmarshal request: %v", err)
			}

			if unmarshaledReq.Profile.ID != tt.request.Profile.ID {
				t.Errorf("Profile ID mismatch after JSON round-trip")
			}
		})
	}
}

func TestCertificateSearchPagination(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("pagination parameters", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			
			if q.Get("offset") != "50" {
				t.Errorf("Expected offset=50, got %s", q.Get("offset"))
			}
			if q.Get("limit") != "50" {
				t.Errorf("Expected limit=50, got %s", q.Get("limit"))
			}

			mockResponse := &CertificateSearchResponse{
				ListResponse: ListResponse{
					Total:  150,
					Offset: 50,
					Limit:  50,
				},
				Items: make([]Certificate, 50), // Simulate 50 certificates
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		opts := &CertificateSearchOptions{
			PaginationParams: PaginationParams{
				Offset: 50,
				Limit:  50,
			},
		}

		result, _, err := client.Certificates.Search(ctx, opts)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		if result.Total != 150 {
			t.Errorf("Total = %v, want %v", result.Total, 150)
		}

		if result.Offset != 50 {
			t.Errorf("Offset = %v, want %v", result.Offset, 50)
		}

		if len(result.Items) != 50 {
			t.Errorf("Items count = %v, want %v", len(result.Items), 50)
		}
	})

	t.Run("offset and limit not added when zero", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			
			// Verify offset and limit are not in query params when they are 0
			if q.Has("offset") {
				t.Errorf("offset parameter should not be present when value is 0")
			}
			if q.Has("limit") {
				t.Errorf("limit parameter should not be present when value is 0")
			}

			mockResponse := &CertificateSearchResponse{
				ListResponse: ListResponse{
					Total:  10,
					Offset: 0,
					Limit:  0,
				},
				Items: make([]Certificate, 10),
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		opts := &CertificateSearchOptions{
			PaginationParams: PaginationParams{
				Offset: 0, // Should not be added to query
				Limit:  0, // Should not be added to query
			},
			CommonName: "test.example.com",
		}

		_, _, err := client.Certificates.Search(ctx, opts)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}
	})

	t.Run("offset and limit added when greater than zero", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			
			// Verify both offset and limit are present when > 0
			if q.Get("offset") != "10" {
				t.Errorf("Expected offset=10, got %s", q.Get("offset"))
			}
			if q.Get("limit") != "25" {
				t.Errorf("Expected limit=25, got %s", q.Get("limit"))
			}

			mockResponse := &CertificateSearchResponse{
				ListResponse: ListResponse{
					Total:  100,
					Offset: 10,
					Limit:  25,
				},
				Items: make([]Certificate, 25),
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		opts := &CertificateSearchOptions{
			PaginationParams: PaginationParams{
				Offset: 10, // Should be added to query
				Limit:  25, // Should be added to query
			},
		}

		result, _, err := client.Certificates.Search(ctx, opts)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		if result.Offset != 10 {
			t.Errorf("Offset = %v, want %v", result.Offset, 10)
		}

		if result.Limit != 25 {
			t.Errorf("Limit = %v, want %v", result.Limit, 25)
		}
	})
}

// TestCertificateResponseFormat tests that certificate responses match expected API format
func TestCertificateResponseFormat(t *testing.T) {
	// Test data based on the provided certificate JSON response
	certificateJSON := `{
		"id": "386e7879-df28-427b-b8ae-58205a8b87df",
		"profile": {},
		"seat": {
			"seat_id": "helix-o2ukusm-dev.uk.pri.o2.com_25Jun16_FiJNMw"
		},
		"seat_type": {
			"id": "2688787a-983a-45c7-8a5d-4261721cbf38",
			"name": "DISCOVERY_SEAT"
		},
		"business_unit": {
			"id": "276e9968-b34a-477c-b2d2-e63cdb77f09e",
			"name": "Default Business Unit"
		},
		"account": {
			"id": "fce887a9-58a7-4c1c-a47c-a25c54d0de83"
		},
		"certificate": "MIIJUTCCCDmgAwIBAgITUgAAquvd5NDuv2cD2gACAACq6zANBgkqhkiG9w0BAQsFADBoMRMwEQ...",
		"ica": {
			"id": "EXTERNAL_CA"
		},
		"common_name": "helix-o2ukusm-dev.uk.pri.o2.com",
		"status": "issued",
		"serial_number": "520000AAEBDDE4D0EEBF6703DA00020000AAEB",
		"thumbprint": "F28D04DDE45ADFBAEDF1B3A364227B17F89623FCC12489D858D14C5457A72547",
		"valid_from": "2023-12-06T04:42:39Z",
		"valid_to": "2025-12-05T04:42:39Z",
		"issuing_ca_name": "O2-EntCA-01",
		"key_size": "RSA_2048",
		"signature_algorithm": "SHA256withRSA",
		"subject": {
			"common_name": "helix-o2ukusm-dev.uk.pri.o2.com",
			"organization_name": "TELEFONICA UK LIMITED",
			"organization_units": ["Operations"],
			"locality": "Slough",
			"country": "GB"
		},
		"ca_vendor": "Unknown",
		"connector": "common-services proxy nonprod",
		"source": "KEY_VAULT_IMPORT",
		"expires_in_days": 128,
		"pqc_vulnerable": true,
		"extended_key_usage": "Server Authentication, Client Authentication",
		"escrow": false,
		"custom_attributes": {}
	}`

	var cert Certificate
	err := json.Unmarshal([]byte(certificateJSON), &cert)
	if err != nil {
		t.Fatalf("Failed to unmarshal certificate JSON: %v", err)
	}

	// Verify key fields are correctly parsed
	if cert.ID != "386e7879-df28-427b-b8ae-58205a8b87df" {
		t.Errorf("ID = %v, want %v", cert.ID, "386e7879-df28-427b-b8ae-58205a8b87df")
	}

	if cert.CommonName != "helix-o2ukusm-dev.uk.pri.o2.com" {
		t.Errorf("CommonName = %v, want %v", cert.CommonName, "helix-o2ukusm-dev.uk.pri.o2.com")
	}

	if cert.Status != "issued" {
		t.Errorf("Status = %v, want %v", cert.Status, "issued")
	}

	if cert.SerialNumber != "520000AAEBDDE4D0EEBF6703DA00020000AAEB" {
		t.Errorf("SerialNumber = %v, want %v", cert.SerialNumber, "520000AAEBDDE4D0EEBF6703DA00020000AAEB")
	}

	if cert.ValidFrom != "2023-12-06T04:42:39Z" {
		t.Errorf("ValidFrom = %v, want %v", cert.ValidFrom, "2023-12-06T04:42:39Z")
	}

	if cert.ValidTo != "2025-12-05T04:42:39Z" {
		t.Errorf("ValidTo = %v, want %v", cert.ValidTo, "2025-12-05T04:42:39Z")
	}

	// Verify nested structures
	if cert.Seat == nil || cert.Seat.SeatID != "helix-o2ukusm-dev.uk.pri.o2.com_25Jun16_FiJNMw" {
		t.Errorf("Seat.SeatID not correctly parsed")
	}

	if cert.SeatType == nil || cert.SeatType.Name != "DISCOVERY_SEAT" {
		t.Errorf("SeatType.Name not correctly parsed")
	}

	if cert.BusinessUnit == nil || cert.BusinessUnit.Name != "Default Business Unit" {
		t.Errorf("BusinessUnit.Name not correctly parsed")
	}

	if cert.Subject == nil || cert.Subject.OrganizationName != "TELEFONICA UK LIMITED" {
		t.Errorf("Subject.OrganizationName not correctly parsed")
	}

	if cert.Subject.Country != "GB" {
		t.Errorf("Subject.Country = %v, want %v", cert.Subject.Country, "GB")
	}

	if len(cert.Subject.OrganizationUnits) != 1 || cert.Subject.OrganizationUnits[0] != "Operations" {
		t.Errorf("Subject.OrganizationUnits not correctly parsed")
	}

	// Verify boolean and numeric fields
	if !cert.PQCVulnerable {
		t.Errorf("PQCVulnerable = %v, want %v", cert.PQCVulnerable, true)
	}

	if cert.Escrow {
		t.Errorf("Escrow = %v, want %v", cert.Escrow, false)
	}

	if cert.ExpiresInDays != 128 {
		t.Errorf("ExpiresInDays = %v, want %v", cert.ExpiresInDays, 128)
	}
}