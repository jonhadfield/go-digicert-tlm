package digicert

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

// TestPaginationIntegrationAcrossServices tests pagination consistency across all services
func TestPaginationIntegrationAcrossServices(t *testing.T) {
	client, _ := NewClient("test-key")
	_ = context.Background()

	tests := []struct {
		name        string
		serviceTest func(t *testing.T, client *Client, server *httptest.Server)
	}{
		{
			name:        "Certificates Service Pagination",
			serviceTest: testCertificatesPagination,
		},
		{
			name:        "Business Units Service Pagination",
			serviceTest: testBusinessUnitsPagination,
		},
		{
			name:        "Profiles Service Pagination",
			serviceTest: testProfilesPagination,
		},
		{
			name:        "Certificate Owners Service Pagination",
			serviceTest: testCertificateOwnersPagination,
		},
		{
			name:        "Enrollments Service Pagination",
			serviceTest: testEnrollmentsPagination,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Generic pagination handler that works for all services
				handlePaginationRequest(w, r, t)
			}))
			defer server.Close()

			client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")
			tt.serviceTest(t, client, server)
		})
	}
}

func handlePaginationRequest(w http.ResponseWriter, r *http.Request, t *testing.T) {
	q := r.URL.Query()

	// Parse pagination parameters
	offset := 0
	limit := 20
	var err error

	if offsetStr := q.Get("offset"); offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			t.Errorf("Invalid offset parameter: %s", offsetStr)
		}
	}

	if limitStr := q.Get("limit"); limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			t.Errorf("Invalid limit parameter: %s", limitStr)
		}
	}

	// Simulate total dataset size
	totalItems := 100

	// Calculate response items
	remainingItems := totalItems - offset
	if remainingItems < 0 {
		remainingItems = 0
	}
	if remainingItems > limit {
		remainingItems = limit
	}

	// Generate response based on endpoint
	switch {
	case r.URL.Path == "/mpki/api/v1/certificate-search":
		generateCertificateSearchResponse(w, offset, limit, totalItems, remainingItems)
	case r.URL.Path == "/mpki/api/v1/business-unit":
		generateBusinessUnitResponse(w, offset, limit, totalItems, remainingItems)
	case r.URL.Path == "/mpki/api/v1/profiles":
		generateProfileResponse(w, offset, limit, totalItems, remainingItems)
	case r.URL.Path == "/mpki/api/v1/certificate-owners":
		generateCertificateOwnerResponse(w, offset, limit, totalItems, remainingItems)
	case r.URL.Path == "/mpki/api/v1/enrollment-details":
		generateEnrollmentResponse(w, offset, limit, totalItems, remainingItems)
	default:
		t.Errorf("Unexpected endpoint: %s", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}
}

func generateCertificateSearchResponse(w http.ResponseWriter, offset, limit, total, count int) {
	items := make([]Certificate, count)
	for i := 0; i < count; i++ {
		items[i] = Certificate{
			ID:         fmt.Sprintf("cert-%d", offset+i+1),
			CommonName: fmt.Sprintf("cert%d.example.com", offset+i+1),
			Status:     "issued",
			SerialNumber: fmt.Sprintf("SN%06d", offset+i+1),
		}
	}

	response := &CertificateSearchResponse{
		ListResponse: ListResponse{
			Total:  total,
			Offset: offset,
			Limit:  limit,
		},
		Items: items,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func generateBusinessUnitResponse(w http.ResponseWriter, offset, limit, total, count int) {
	items := make([]BusinessUnit, count)
	for i := 0; i < count; i++ {
		items[i] = BusinessUnit{
			ID:             fmt.Sprintf("bu-%d", offset+i+1),
			Name:           fmt.Sprintf("Business Unit %d", offset+i+1),
			Description:    fmt.Sprintf("Description for BU %d", offset+i+1),
			IsActive:       true,
			LicensedSeats:  100,
			UsedSeats:      50,
			AvailableSeats: 50,
		}
	}

	response := &BusinessUnitListResponse{
		ListResponse: ListResponse{
			Total:  total,
			Offset: offset,
			Limit:  limit,
		},
		BusinessUnits: items,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func generateProfileResponse(w http.ResponseWriter, offset, limit, total, count int) {
	items := make([]Profile, count)
	for i := 0; i < count; i++ {
		items[i] = Profile{
			ID:               fmt.Sprintf("profile-%d", offset+i+1),
			Name:             fmt.Sprintf("Profile %d", offset+i+1),
			Description:      fmt.Sprintf("Description for Profile %d", offset+i+1),
			Type:             "SERVER_CERTIFICATE",
			Status:           "active",
			EnrollmentMethod: "REST_API",
			KeySize:          2048,
		}
	}

	response := &ProfileListResponse{
		ListResponse: ListResponse{
			Total:  total,
			Offset: offset,
			Limit:  limit,
		},
		Profiles: items,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func generateCertificateOwnerResponse(w http.ResponseWriter, offset, limit, total, count int) {
	items := make([]CertificateOwner, count)
	for i := 0; i < count; i++ {
		items[i] = CertificateOwner{
			ID:           fmt.Sprintf("owner-%d", offset+i+1),
			FirstName:    fmt.Sprintf("Owner%d", offset+i+1),
			LastName:     "LastName",
			Email:        fmt.Sprintf("owner%d@example.com", offset+i+1),
			Department:   "IT",
			Company:      "Infrastructure Corp",
			IsActive:     true,
		}
	}

	response := &CertificateOwnerListResponse{
		ListResponse: ListResponse{
			Total:  total,
			Offset: offset,
			Limit:  limit,
		},
		Owners: items,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func generateEnrollmentResponse(w http.ResponseWriter, offset, limit, total, count int) {
	items := make([]Enrollment, count)
	for i := 0; i < count; i++ {
		items[i] = Enrollment{
			ID:             fmt.Sprintf("enrollment-%d", offset+i+1),
			EnrollmentCode: fmt.Sprintf("CODE-%04d-%04d", offset+i+1, offset+i+1),
			Status:         "pending",
			ProfileID:      "profile-123",
			CommonName:     fmt.Sprintf("enrollment%d.example.com", offset+i+1),
			Email:          fmt.Sprintf("enrollment%d@example.com", offset+i+1),
		}
	}

	response := &EnrollmentDetailsResponse{
		ListResponse: ListResponse{
			Total:  total,
			Offset: offset,
			Limit:  limit,
		},
		Enrollments: items,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func testCertificatesPagination(t *testing.T, client *Client, server *httptest.Server) {
	ctx := context.Background()

	// Test first page
	opts := &CertificateSearchOptions{
		PaginationParams: PaginationParams{
			Offset: 0,
			Limit:  20,
		},
	}

	result, _, err := client.Certificates.Search(ctx, opts)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if result.Total != 100 {
		t.Errorf("Total = %v, want %v", result.Total, 100)
	}

	if result.Offset != 0 {
		t.Errorf("Offset = %v, want %v", result.Offset, 0)
	}

	if len(result.Items) != 20 {
		t.Errorf("Items count = %v, want %v", len(result.Items), 20)
	}

	if result.Items[0].ID != "cert-1" {
		t.Errorf("First item ID = %v, want %v", result.Items[0].ID, "cert-1")
	}

	// Test middle page
	opts.Offset = 40
	result, _, err = client.Certificates.Search(ctx, opts)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if result.Offset != 40 {
		t.Errorf("Offset = %v, want %v", result.Offset, 40)
	}

	if result.Items[0].ID != "cert-41" {
		t.Errorf("First item ID = %v, want %v", result.Items[0].ID, "cert-41")
	}

	// Test last page
	opts.Offset = 90
	result, _, err = client.Certificates.Search(ctx, opts)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(result.Items) != 10 {
		t.Errorf("Last page items count = %v, want %v", len(result.Items), 10)
	}

	if result.Items[0].ID != "cert-91" {
		t.Errorf("Last page first item ID = %v, want %v", result.Items[0].ID, "cert-91")
	}
}

func testBusinessUnitsPagination(t *testing.T, client *Client, server *httptest.Server) {
	ctx := context.Background()

	opts := &BusinessUnitListOptions{
		PaginationParams: PaginationParams{
			Offset: 10,
			Limit:  15,
		},
	}

	result, _, err := client.BusinessUnits.List(ctx, opts)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if result.Total != 100 {
		t.Errorf("Total = %v, want %v", result.Total, 100)
	}

	if result.Offset != 10 {
		t.Errorf("Offset = %v, want %v", result.Offset, 10)
	}

	if len(result.BusinessUnits) != 15 {
		t.Errorf("BusinessUnits count = %v, want %v", len(result.BusinessUnits), 15)
	}

	if result.BusinessUnits[0].ID != "bu-11" {
		t.Errorf("First item ID = %v, want %v", result.BusinessUnits[0].ID, "bu-11")
	}
}

func testProfilesPagination(t *testing.T, client *Client, server *httptest.Server) {
	ctx := context.Background()

	opts := &ProfileListOptions{
		PaginationParams: PaginationParams{
			Offset: 25,
			Limit:  30,
		},
	}

	result, _, err := client.Profiles.List(ctx, opts)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if result.Total != 100 {
		t.Errorf("Total = %v, want %v", result.Total, 100)
	}

	if result.Offset != 25 {
		t.Errorf("Offset = %v, want %v", result.Offset, 25)
	}

	if len(result.Profiles) != 30 {
		t.Errorf("Profiles count = %v, want %v", len(result.Profiles), 30)
	}

	if result.Profiles[0].ID != "profile-26" {
		t.Errorf("First item ID = %v, want %v", result.Profiles[0].ID, "profile-26")
	}
}

func testCertificateOwnersPagination(t *testing.T, client *Client, server *httptest.Server) {
	ctx := context.Background()

	opts := &CertificateOwnerListOptions{
		PaginationParams: PaginationParams{
			Offset: 50,
			Limit:  25,
		},
	}

	result, _, err := client.CertificateOwners.List(ctx, opts)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if result.Total != 100 {
		t.Errorf("Total = %v, want %v", result.Total, 100)
	}

	if result.Offset != 50 {
		t.Errorf("Offset = %v, want %v", result.Offset, 50)
	}

	if len(result.Owners) != 25 {
		t.Errorf("Owners count = %v, want %v", len(result.Owners), 25)
	}

	if result.Owners[0].ID != "owner-51" {
		t.Errorf("First item ID = %v, want %v", result.Owners[0].ID, "owner-51")
	}
}

func testEnrollmentsPagination(t *testing.T, client *Client, server *httptest.Server) {
	ctx := context.Background()

	opts := &EnrollmentDetailsOptions{
		PaginationParams: PaginationParams{
			Offset: 75,
			Limit:  10,
		},
	}

	result, _, err := client.Enrollments.ListDetails(ctx, opts)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if result.Total != 100 {
		t.Errorf("Total = %v, want %v", result.Total, 100)
	}

	if result.Offset != 75 {
		t.Errorf("Offset = %v, want %v", result.Offset, 75)
	}

	if len(result.Enrollments) != 10 {
		t.Errorf("Enrollments count = %v, want %v", len(result.Enrollments), 10)
	}

	if result.Enrollments[0].ID != "enrollment-76" {
		t.Errorf("First item ID = %v, want %v", result.Enrollments[0].ID, "enrollment-76")
	}
}

// TestPaginationParameterConsistency tests that pagination parameters are handled consistently
func TestPaginationParameterConsistency(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	// Test cases for pagination parameter validation
	testCases := []struct {
		name   string
		offset int
		limit  int
		expectInURL bool
	}{
		{"zero values", 0, 0, false},
		{"negative values", -1, -5, false},
		{"positive offset only", 10, 0, false}, // Only offset > 0, limit = 0
		{"positive limit only", 0, 20, false},  // Only limit > 0, offset = 0
		{"both positive", 30, 40, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				q := r.URL.Query()

				if tc.expectInURL {
					if tc.offset > 0 && q.Get("offset") != fmt.Sprintf("%d", tc.offset) {
						t.Errorf("Expected offset=%d in URL, got %s", tc.offset, q.Get("offset"))
					}
					if tc.limit > 0 && q.Get("limit") != fmt.Sprintf("%d", tc.limit) {
						t.Errorf("Expected limit=%d in URL, got %s", tc.limit, q.Get("limit"))
					}
				} else {
					if q.Has("offset") {
						t.Errorf("offset should not be present in URL for case: %s", tc.name)
					}
					if q.Has("limit") {
						t.Errorf("limit should not be present in URL for case: %s", tc.name)
					}
				}

				// Return minimal response
				response := &CertificateSearchResponse{
					ListResponse: ListResponse{Total: 1, Offset: 0, Limit: 0},
					Items:        []Certificate{},
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

			opts := &CertificateSearchOptions{
				PaginationParams: PaginationParams{
					Offset: tc.offset,
					Limit:  tc.limit,
				},
			}

			_, _, err := client.Certificates.Search(ctx, opts)
			if err != nil {
				t.Fatalf("Search() error = %v", err)
			}
		})
	}
}

// TestPaginationBoundaryConditions tests edge cases in pagination
func TestPaginationBoundaryConditions(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("offset beyond total results", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate offset beyond available data
			response := &CertificateSearchResponse{
				ListResponse: ListResponse{
					Total:  50,   // Total items available
					Offset: 100,  // Offset beyond total
					Limit:  20,
				},
				Items: []Certificate{}, // No items returned
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		opts := &CertificateSearchOptions{
			PaginationParams: PaginationParams{
				Offset: 100,
				Limit:  20,
			},
		}

		result, _, err := client.Certificates.Search(ctx, opts)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		if len(result.Items) != 0 {
			t.Errorf("Expected 0 items for offset beyond total, got %d", len(result.Items))
		}

		if result.Total != 50 {
			t.Errorf("Total = %v, want %v", result.Total, 50)
		}
	})

	t.Run("very large limit", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			// Since offset=0 and our pagination logic requires both > 0, no limit should be sent
			if q.Has("limit") {
				t.Errorf("limit parameter should not be present when offset is 0")
			}

			// Server might cap the actual returned items
			items := make([]Certificate, 1000) // Server caps at 1000
			for i := range items {
				items[i] = Certificate{ID: fmt.Sprintf("cert-%d", i+1)}
			}

			response := &CertificateSearchResponse{
				ListResponse: ListResponse{
					Total:  100000,
					Offset: 0,
					Limit:  0, // No limit sent since offset=0
				},
				Items: items, // Actual returned items (capped)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		opts := &CertificateSearchOptions{
			PaginationParams: PaginationParams{
				Offset: 0,
				Limit:  999999,
			},
		}

		result, _, err := client.Certificates.Search(ctx, opts)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		// Server implementation might cap results
		if len(result.Items) > 10000 {
			t.Errorf("Server should cap large responses, got %d items", len(result.Items))
		}

		if result.Limit != 0 {
			t.Errorf("Limit = %v, want %v", result.Limit, 0)
		}
	})
}

// TestServiceSpecificPaginationFeatures tests unique pagination features per service
func TestServiceSpecificPaginationFeatures(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("business units with active filter and pagination", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()

			// Verify both pagination and filtering parameters
			if q.Get("is_active") != "true" {
				t.Errorf("Expected is_active=true, got %s", q.Get("is_active"))
			}
			if q.Get("offset") != "15" {
				t.Errorf("Expected offset=15, got %s", q.Get("offset"))
			}
			if q.Get("limit") != "25" {
				t.Errorf("Expected limit=25, got %s", q.Get("limit"))
			}

			response := &BusinessUnitListResponse{
				ListResponse: ListResponse{Total: 60, Offset: 15, Limit: 25},
				BusinessUnits: make([]BusinessUnit, 25),
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		isActive := true
		opts := &BusinessUnitListOptions{
			PaginationParams: PaginationParams{Offset: 15, Limit: 25},
			IsActive:         &isActive,
		}

		_, _, err := client.BusinessUnits.List(ctx, opts)
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
	})

	t.Run("profiles with type filter and sorting", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()

			// Verify pagination, filtering, and sorting parameters
			if q.Get("type") != "SERVER_CERTIFICATE" {
				t.Errorf("Expected type=SERVER_CERTIFICATE, got %s", q.Get("type"))
			}
			if q.Get("sort_by") != "name" {
				t.Errorf("Expected sort_by=name, got %s", q.Get("sort_by"))
			}
			if q.Get("sort_order") != "asc" {
				t.Errorf("Expected sort_order=asc, got %s", q.Get("sort_order"))
			}
			if q.Get("offset") != "5" {
				t.Errorf("Expected offset=5, got %s", q.Get("offset"))
			}

			response := &ProfileListResponse{
				ListResponse: ListResponse{Total: 30, Offset: 5, Limit: 10},
				Profiles:     make([]Profile, 10),
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		opts := &ProfileListOptions{
			PaginationParams: PaginationParams{Offset: 5, Limit: 10},
			Type:             "SERVER_CERTIFICATE",
			SortBy:           "name",
			SortOrder:        "asc",
		}

		_, _, err := client.Profiles.List(ctx, opts)
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
	})
}