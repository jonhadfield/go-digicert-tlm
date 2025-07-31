package digicert

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCertificateOwnersService_Create(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("successful certificate owner creation", func(t *testing.T) {
		mockRequest := &CertificateOwnerRequest{
			FirstName:   "John",
			LastName:    "Doe",
			Email:       "john.doe@example.com",
			PhoneNumber: "+1-555-123-4567",
			Department:  "IT Security",
			JobTitle:    "Senior Security Engineer",
			Company:     "Example Corp",
		}

		createdAt := time.Now()
		mockResponse := &CertificateOwner{
			ID:          "owner-123",
			FirstName:   mockRequest.FirstName,
			LastName:    mockRequest.LastName,
			Email:       mockRequest.Email,
			PhoneNumber: mockRequest.PhoneNumber,
			Department:  mockRequest.Department,
			JobTitle:    mockRequest.JobTitle,
			Company:     mockRequest.Company,
			IsActive:    true,
			CreatedAt:   &createdAt,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/mpki/api/v1/certificate-owners" {
				t.Errorf("Expected path /mpki/api/v1/certificate-owners, got %s", r.URL.Path)
			}
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST request, got %s", r.Method)
			}

			var reqBody CertificateOwnerRequest
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			if reqBody.Email != mockRequest.Email {
				t.Errorf("Expected email %s, got %s", mockRequest.Email, reqBody.Email)
			}

			if reqBody.Department != mockRequest.Department {
				t.Errorf("Expected department %s, got %s", mockRequest.Department, reqBody.Department)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		result, resp, err := client.CertificateOwners.Create(ctx, mockRequest)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusCreated)
		}

		if result.ID != mockResponse.ID {
			t.Errorf("ID = %v, want %v", result.ID, mockResponse.ID)
		}

		if result.Email != mockResponse.Email {
			t.Errorf("Email = %v, want %v", result.Email, mockResponse.Email)
		}

		if result.Department != mockResponse.Department {
			t.Errorf("Department = %v, want %v", result.Department, mockResponse.Department)
		}
	})

	t.Run("duplicate email error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(APIError{
				Code:    "DUPLICATE_EMAIL",
				Message: "Certificate owner with this email already exists",
			})
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		mockRequest := &CertificateOwnerRequest{
			FirstName: "Jane",
			LastName:  "Duplicate",
			Email:     "existing@example.com",
		}

		_, _, err := client.CertificateOwners.Create(ctx, mockRequest)
		if err == nil {
			t.Fatal("Expected error for duplicate email")
		}

		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("Error type = %T, want *APIError", err)
		}

		if apiErr.Code != "DUPLICATE_EMAIL" {
			t.Errorf("Error Code = %v, want %v", apiErr.Code, "DUPLICATE_EMAIL")
		}
	})
}

func TestCertificateOwnersService_Get(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("successful certificate owner retrieval", func(t *testing.T) {
		ownerID := "owner-123"
		createdAt := time.Now().Add(-60 * 24 * time.Hour)
		updatedAt := time.Now().Add(-10 * 24 * time.Hour)

		mockOwner := &CertificateOwner{
			ID:          ownerID,
			FirstName:   "Alice",
			LastName:    "Johnson",
			Email:       "alice.johnson@example.com",
			PhoneNumber: "+1-555-987-6543",
			Department:  "DevOps",
			JobTitle:    "DevOps Lead",
			Company:     "Example Corp",
			IsActive:    true,
			CreatedAt:   &createdAt,
			UpdatedAt:   &updatedAt,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/mpki/api/v1/certificate-owners/" + ownerID
			if r.URL.Path != expectedPath {
				t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
			}
			if r.Method != http.MethodGet {
				t.Errorf("Expected GET request, got %s", r.Method)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockOwner)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		result, resp, err := client.CertificateOwners.Get(ctx, ownerID)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusOK)
		}

		if result.ID != mockOwner.ID {
			t.Errorf("ID = %v, want %v", result.ID, mockOwner.ID)
		}

		if result.Email != mockOwner.Email {
			t.Errorf("Email = %v, want %v", result.Email, mockOwner.Email)
		}

		if result.JobTitle != mockOwner.JobTitle {
			t.Errorf("JobTitle = %v, want %v", result.JobTitle, mockOwner.JobTitle)
		}
	})

	t.Run("certificate owner not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(APIError{
				Code:    "CERTIFICATE_OWNER_NOT_FOUND",
				Message: "Certificate owner not found",
			})
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		_, _, err := client.CertificateOwners.Get(ctx, "nonexistent-owner")
		if err == nil {
			t.Fatal("Expected error for nonexistent certificate owner")
		}

		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("Error type = %T, want *APIError", err)
		}

		if apiErr.Code != "CERTIFICATE_OWNER_NOT_FOUND" {
			t.Errorf("Error Code = %v, want %v", apiErr.Code, "CERTIFICATE_OWNER_NOT_FOUND")
		}
	})
}

func TestCertificateOwnersService_Update(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("successful certificate owner update", func(t *testing.T) {
		ownerID := "owner-123"
		mockRequest := &CertificateOwnerRequest{
			FirstName:   "Alice",
			LastName:    "Johnson-Miller",
			Email:       "alice.johnson-miller@example.com",
			PhoneNumber: "+1-555-987-1234",
			Department:  "Site Reliability Engineering",
			JobTitle:    "Senior SRE",
			Company:     "Example Corp",
		}

		updatedAt := time.Now()
		mockResponse := &CertificateOwner{
			ID:          ownerID,
			FirstName:   mockRequest.FirstName,
			LastName:    mockRequest.LastName,
			Email:       mockRequest.Email,
			PhoneNumber: mockRequest.PhoneNumber,
			Department:  mockRequest.Department,
			JobTitle:    mockRequest.JobTitle,
			Company:     mockRequest.Company,
			IsActive:    true,
			UpdatedAt:   &updatedAt,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/mpki/api/v1/certificate-owners/" + ownerID
			if r.URL.Path != expectedPath {
				t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
			}
			if r.Method != http.MethodPut {
				t.Errorf("Expected PUT request, got %s", r.Method)
			}

			var reqBody CertificateOwnerRequest
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			if reqBody.LastName != mockRequest.LastName {
				t.Errorf("Expected last name %s, got %s", mockRequest.LastName, reqBody.LastName)
			}

			if reqBody.JobTitle != mockRequest.JobTitle {
				t.Errorf("Expected job title %s, got %s", mockRequest.JobTitle, reqBody.JobTitle)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		result, resp, err := client.CertificateOwners.Update(ctx, ownerID, mockRequest)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusOK)
		}

		if result.LastName != mockResponse.LastName {
			t.Errorf("LastName = %v, want %v", result.LastName, mockResponse.LastName)
		}

		if result.JobTitle != mockResponse.JobTitle {
			t.Errorf("JobTitle = %v, want %v", result.JobTitle, mockResponse.JobTitle)
		}

		if result.Department != mockResponse.Department {
			t.Errorf("Department = %v, want %v", result.Department, mockResponse.Department)
		}
	})
}

func TestCertificateOwnersService_Delete(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("successful certificate owner deletion", func(t *testing.T) {
		ownerID := "owner-123"

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/mpki/api/v1/certificate-owners/" + ownerID
			if r.URL.Path != expectedPath {
				t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
			}
			if r.Method != http.MethodDelete {
				t.Errorf("Expected DELETE request, got %s", r.Method)
			}

			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		resp, err := client.CertificateOwners.Delete(ctx, ownerID)
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusNoContent)
		}
	})

	t.Run("certificate owner has active certificates error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(APIError{
				Code:    "HAS_ACTIVE_CERTIFICATES",
				Message: "Cannot delete certificate owner with active certificates",
			})
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		_, err := client.CertificateOwners.Delete(ctx, "owner-with-certs")
		if err == nil {
			t.Fatal("Expected error for owner with active certificates")
		}

		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("Error type = %T, want *APIError", err)
		}

		if apiErr.Code != "HAS_ACTIVE_CERTIFICATES" {
			t.Errorf("Error Code = %v, want %v", apiErr.Code, "HAS_ACTIVE_CERTIFICATES")
		}
	})
}

func TestCertificateOwnersService_List(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("successful certificate owners listing with filters", func(t *testing.T) {
		createdAt := time.Now().Add(-90 * 24 * time.Hour)
		mockResponse := &CertificateOwnerListResponse{
			ListResponse: ListResponse{
				Total:  75,
				Offset: 30,
				Limit:  20,
			},
			Owners: []CertificateOwner{
				{
					ID:          "owner-1",
					FirstName:   "Bob",
					LastName:    "Smith",
					Email:       "bob.smith@example.com",
					PhoneNumber: "+1-555-111-2222",
					Department:  "Security",
					JobTitle:    "Security Analyst",
					Company:     "Example Corp",
					IsActive:    true,
					CreatedAt:   &createdAt,
				},
				{
					ID:          "owner-2",
					FirstName:   "Carol",
					LastName:    "Davis",
					Email:       "carol.davis@example.com",
					PhoneNumber: "+1-555-333-4444",
					Department:  "Security",
					JobTitle:    "Security Manager",
					Company:     "Example Corp",
					IsActive:    true,
					CreatedAt:   &createdAt,
				},
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/mpki/api/v1/certificate-owners" {
				t.Errorf("Expected path /mpki/api/v1/certificate-owners, got %s", r.URL.Path)
			}
			if r.Method != http.MethodGet {
				t.Errorf("Expected GET request, got %s", r.Method)
			}

			q := r.URL.Query()
			if q.Get("first_name") != "Bob" {
				t.Errorf("Expected first_name=Bob, got %s", q.Get("first_name"))
			}
			if q.Get("is_active") != "true" {
				t.Errorf("Expected is_active=true, got %s", q.Get("is_active"))
			}
			if q.Get("offset") != "30" {
				t.Errorf("Expected offset=30, got %s", q.Get("offset"))
			}
			if q.Get("limit") != "20" {
				t.Errorf("Expected limit=20, got %s", q.Get("limit"))
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		isActive := true
		opts := &CertificateOwnerListOptions{
			PaginationParams: PaginationParams{
				Offset: 30,
				Limit:  20,
			},
			FirstName: "Bob",
			IsActive:  &isActive,
		}

		result, resp, err := client.CertificateOwners.List(ctx, opts)
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusOK)
		}

		if result.Total != mockResponse.Total {
			t.Errorf("Total = %v, want %v", result.Total, mockResponse.Total)
		}

		if result.Offset != mockResponse.Offset {
			t.Errorf("Offset = %v, want %v", result.Offset, mockResponse.Offset)
		}

		if len(result.Owners) != len(mockResponse.Owners) {
			t.Errorf("Owners count = %v, want %v", len(result.Owners), len(mockResponse.Owners))
		}

		if result.Owners[0].Department != "Security" {
			t.Errorf("First owner department = %v, want %v", result.Owners[0].Department, "Security")
		}

		if result.Owners[1].JobTitle != "Security Manager" {
			t.Errorf("Second owner job title = %v, want %v", result.Owners[1].JobTitle, "Security Manager")
		}
	})

	t.Run("list without pagination parameters", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()

			// Verify offset and limit are not present when they are 0
			if q.Has("offset") {
				t.Errorf("offset parameter should not be present when value is 0")
			}
			if q.Has("limit") {
				t.Errorf("limit parameter should not be present when value is 0")
			}

			mockResponse := &CertificateOwnerListResponse{
				ListResponse: ListResponse{
					Total:  12,
					Offset: 0,
					Limit:  0,
				},
				Owners: make([]CertificateOwner, 12),
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		opts := &CertificateOwnerListOptions{
			PaginationParams: PaginationParams{
				Offset: 0, // Should not be added to query
				Limit:  0, // Should not be added to query
			},
			FirstName: "Engineering",
		}

		_, _, err := client.CertificateOwners.List(ctx, opts)
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
	})
}