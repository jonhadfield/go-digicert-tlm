package digicert

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProfilesService_Get(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("successful profile retrieval", func(t *testing.T) {
		profileID := "profile-123"
		createdAt := time.Now().Add(-30 * 24 * time.Hour)
		updatedAt := time.Now().Add(-24 * time.Hour)

		mockProfile := &Profile{
			ID:               profileID,
			Name:             "TLS Server Certificate",
			Description:      "Standard TLS server certificate profile",
			Type:             "SERVER_CERTIFICATE",
			Status:           "active",
			EnrollmentMethod: "REST_API",
			KeyAlgorithm:     "RSA",
			KeySize:          2048,
			SignatureAlgorithm: "SHA256",
			Tags:             []string{"production", "web-server"},
			CreatedAt:        &createdAt,
			UpdatedAt:        &updatedAt,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/mpki/api/v1/profiles/" + profileID
			if r.URL.Path != expectedPath {
				t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
			}
			if r.Method != http.MethodGet {
				t.Errorf("Expected GET request, got %s", r.Method)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockProfile)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		result, resp, err := client.Profiles.Get(ctx, profileID)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusOK)
		}

		if result.ID != mockProfile.ID {
			t.Errorf("ID = %v, want %v", result.ID, mockProfile.ID)
		}

		if result.Name != mockProfile.Name {
			t.Errorf("Name = %v, want %v", result.Name, mockProfile.Name)
		}

		if result.Type != mockProfile.Type {
			t.Errorf("Type = %v, want %v", result.Type, mockProfile.Type)
		}

		if result.KeySize != mockProfile.KeySize {
			t.Errorf("KeySize = %v, want %v", result.KeySize, mockProfile.KeySize)
		}

		if len(result.Tags) != len(mockProfile.Tags) {
			t.Errorf("Tags length = %v, want %v", len(result.Tags), len(mockProfile.Tags))
		}
	})

	t.Run("profile not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(APIError{
				Code:    "PROFILE_NOT_FOUND",
				Message: "Profile not found",
			})
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		_, _, err := client.Profiles.Get(ctx, "nonexistent-profile")
		if err == nil {
			t.Fatal("Expected error for nonexistent profile")
		}

		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("Error type = %T, want *APIError", err)
		}

		if apiErr.Code != "PROFILE_NOT_FOUND" {
			t.Errorf("Error Code = %v, want %v", apiErr.Code, "PROFILE_NOT_FOUND")
		}
	})
}

func TestProfilesService_List(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("successful profiles listing with filters", func(t *testing.T) {
		mockResponse := &ProfileListResponse{
			ListResponse: ListResponse{
				Total:  15,
				Offset: 5,
				Limit:  10,
			},
			Profiles: []Profile{
				{
					ID:               "profile-1",
					Name:             "Web Server SSL",
					Description:      "Standard web server SSL certificate",
					Type:             "SERVER_CERTIFICATE",
					Status:           "active",
					EnrollmentMethod: "REST_API",
					KeyAlgorithm:     "RSA",
					KeySize:          2048,
					SignatureAlgorithm:    "SHA256",
					Tags:             []string{"web-server", "production"},
				},
				{
					ID:               "profile-2",
					Name:             "Client Authentication",
					Description:      "Client authentication certificate profile",
					Type:             "CLIENT_CERTIFICATE",
					Status:           "active",
					EnrollmentMethod: "MANUAL",
					KeyAlgorithm:     "RSA",
					KeySize:          4096,
					SignatureAlgorithm:    "SHA256",
					Tags:             []string{"client-auth", "vpn"},
				},
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/mpki/api/v1/profiles" {
				t.Errorf("Expected path /mpki/api/v1/profile, got %s", r.URL.Path)
			}
			if r.Method != http.MethodGet {
				t.Errorf("Expected GET request, got %s", r.Method)
			}

			q := r.URL.Query()
			if q.Get("type") != "SERVER_CERTIFICATE" {
				t.Errorf("Expected type=SERVER_CERTIFICATE, got %s", q.Get("type"))
			}
			if q.Get("status") != "active" {
				t.Errorf("Expected status=active, got %s", q.Get("status"))
			}
			if q.Get("enrollment_method") != "REST_API" {
				t.Errorf("Expected enrollment_method=REST_API, got %s", q.Get("enrollment_method"))
			}
			if q.Get("offset") != "5" {
				t.Errorf("Expected offset=5, got %s", q.Get("offset"))
			}
			if q.Get("limit") != "10" {
				t.Errorf("Expected limit=10, got %s", q.Get("limit"))
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		opts := &ProfileListOptions{
			PaginationParams: PaginationParams{
				Offset: 5,
				Limit:  10,
			},
			Type:             "SERVER_CERTIFICATE",
			Status:           "active",
			EnrollmentMethod: "REST_API",
		}

		result, resp, err := client.Profiles.List(ctx, opts)
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

		if len(result.Profiles) != len(mockResponse.Profiles) {
			t.Errorf("Profiles count = %v, want %v", len(result.Profiles), len(mockResponse.Profiles))
		}

		if result.Profiles[0].Type != "SERVER_CERTIFICATE" {
			t.Errorf("First profile type = %v, want %v", result.Profiles[0].Type, "SERVER_CERTIFICATE")
		}

		if result.Profiles[1].EnrollmentMethod != "MANUAL" {
			t.Errorf("Second profile enrollment method = %v, want %v", result.Profiles[1].EnrollmentMethod, "MANUAL")
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

			mockResponse := &ProfileListResponse{
				ListResponse: ListResponse{
					Total:  8,
					Offset: 0,
					Limit:  0,
				},
				Profiles: make([]Profile, 8),
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		opts := &ProfileListOptions{
			PaginationParams: PaginationParams{
				Offset: 0, // Should not be added to query
				Limit:  0, // Should not be added to query
			},
			Status: "active",
		}

		_, _, err := client.Profiles.List(ctx, opts)
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
	})
}

func TestProfilesService_ListPublic(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("successful public profiles listing", func(t *testing.T) {
		mockResponse := &ProfileListResponse{
			ListResponse: ListResponse{
				Total:  3,
				Offset: 0,
				Limit:  0,
			},
			Profiles: []Profile{
				{
					ID:               "public-profile-1",
					Name:             "Public SSL Certificate",
					Description:      "Publicly trusted SSL certificate",
					Type:             "SERVER_CERTIFICATE",
					Status:           "active",
					EnrollmentMethod: "REST_API",
					KeyAlgorithm:     "RSA",
					KeySize:          2048,
					SignatureAlgorithm:    "SHA256",
					Tags:             []string{"public", "trusted"},
				},
				{
					ID:               "public-profile-2", 
					Name:             "Public Code Signing",
					Description:      "Publicly trusted code signing certificate",
					Type:             "CODE_SIGNING",
					Status:           "active",
					EnrollmentMethod: "MANUAL",
					KeyAlgorithm:     "RSA",
					KeySize:          4096,
					SignatureAlgorithm:    "SHA256",
					Tags:             []string{"public", "code-signing"},
				},
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/mpki/api/v1/profiles/public" {
				t.Errorf("Expected path /mpki/api/v1/profiles/public, got %s", r.URL.Path)
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

		result, resp, err := client.Profiles.ListPublic(ctx)
		if err != nil {
			t.Fatalf("ListPublic() error = %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusOK)
		}

		if result.Total != mockResponse.Total {
			t.Errorf("Total = %v, want %v", result.Total, mockResponse.Total)
		}

		if len(result.Profiles) != len(mockResponse.Profiles) {
			t.Errorf("Profiles count = %v, want %v", len(result.Profiles), len(mockResponse.Profiles))
		}

		// Verify public profiles have expected characteristics
		for _, profile := range result.Profiles {
			if profile.Status != "active" {
				t.Errorf("Public profile should be active, got %s", profile.Status)
			}
		}

		if result.Profiles[0].Name != "Public SSL Certificate" {
			t.Errorf("First profile name = %v, want %v", result.Profiles[0].Name, "Public SSL Certificate")
		}

		if result.Profiles[1].Type != "CODE_SIGNING" {
			t.Errorf("Second profile type = %v, want %v", result.Profiles[1].Type, "CODE_SIGNING")
		}
	})
}

func TestProfilesService_ListTemplates(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("successful profile templates listing", func(t *testing.T) {
		mockResponse := &ProfileTemplateListResponse{
			Templates: []ProfileTemplate{
				{
					ID:          "template-1",
					Name:        "Standard Web Server",
					Description: "Template for standard web server certificates",
					Type:        "SERVER_CERTIFICATE",
					Provider:    "DigiCert",
				},
				{
					ID:          "template-2",
					Name:        "Client Authentication",
					Description: "Template for client authentication certificates",
					Type:        "CLIENT_CERTIFICATE",
					Provider:    "DigiCert",
				},
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/mpki/api/v1/profiles/templates" {
				t.Errorf("Expected path /mpki/api/v1/profiles/templates, got %s", r.URL.Path)
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

		result, resp, err := client.Profiles.ListTemplates(ctx)
		if err != nil {
			t.Fatalf("ListTemplates() error = %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusOK)
		}

		if len(result.Templates) != len(mockResponse.Templates) {
			t.Errorf("Templates count = %v, want %v", len(result.Templates), len(mockResponse.Templates))
		}

		if result.Templates[0].Name != "Standard Web Server" {
			t.Errorf("First template name = %v, want %v", result.Templates[0].Name, "Standard Web Server")
		}

		if result.Templates[0].Provider != "DigiCert" {
			t.Errorf("First template provider = %v, want %v", result.Templates[0].Provider, "DigiCert")
		}

		if result.Templates[1].Type != "CLIENT_CERTIFICATE" {
			t.Errorf("Second template type = %v, want %v", result.Templates[1].Type, "CLIENT_CERTIFICATE")
		}
	})
}


// TestProfileListOptionsCombinations tests various filter combinations
func TestProfileListOptionsCombinations(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	tests := []struct {
		name     string
		options  *ProfileListOptions
		expected func(q map[string][]string) bool
	}{
		{
			name: "type and status filters",
			options: &ProfileListOptions{
				Type:   "SERVER_CERTIFICATE",
				Status: "active",
			},
			expected: func(q map[string][]string) bool {
				return q["type"][0] == "SERVER_CERTIFICATE" && q["status"][0] == "active"
			},
		},
		{
			name: "enrollment method filter",
			options: &ProfileListOptions{
				EnrollmentMethod: "MANUAL",
			},
			expected: func(q map[string][]string) bool {
				return q["enrollment_method"][0] == "MANUAL"
			},
		},
		{
			name: "sorting options",
			options: &ProfileListOptions{
				SortBy:    "name",
				SortOrder: "desc",
			},
			expected: func(q map[string][]string) bool {
				return q["sort_by"][0] == "name" && q["sort_order"][0] == "desc"
			},
		},
		{
			name: "pagination with filters",
			options: &ProfileListOptions{
				PaginationParams: PaginationParams{
					Offset: 25,
					Limit:  50,
				},
				Type:   "CLIENT_CERTIFICATE",
				Status: "draft",
			},
			expected: func(q map[string][]string) bool {
				return q["offset"][0] == "25" && q["limit"][0] == "50" &&
					q["type"][0] == "CLIENT_CERTIFICATE" && q["status"][0] == "draft"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				queryParams := r.URL.Query()
				if !tt.expected(queryParams) {
					t.Errorf("Query parameters don't match expected values: %v", queryParams)
				}

				mockResponse := &ProfileListResponse{
					ListResponse: ListResponse{Total: 1, Offset: 0, Limit: 10},
					Profiles:     []Profile{},
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(mockResponse)
			}))
			defer server.Close()

			client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

			_, _, err := client.Profiles.List(ctx, tt.options)
			if err != nil {
				t.Fatalf("List() error = %v", err)
			}
		})
	}
}

// TestProfileResponseValidation tests profile response structure validation
func TestProfileResponseValidation(t *testing.T) {
	// Test data based on typical profile response structure
	profileJSON := `{
		"id": "profile-abc123",
		"name": "TLS Server Certificate Profile",
		"description": "Standard TLS server certificate for web servers",
		"type": "SERVER_CERTIFICATE",
		"status": "active",
		"enrollment_method": "REST_API",
		"key_algorithm": "RSA",
		"key_size": 2048,
		"signature_algorithm": "SHA256",
		"tags": ["production", "web-server", "ssl"],
		"created_at": "2023-01-15T10:30:00Z",
		"updated_at": "2023-12-01T14:45:00Z"
	}`

	var profile Profile
	err := json.Unmarshal([]byte(profileJSON), &profile)
	if err != nil {
		t.Fatalf("Failed to unmarshal profile JSON: %v", err)
	}

	// Verify key fields are correctly parsed
	if profile.ID != "profile-abc123" {
		t.Errorf("ID = %v, want %v", profile.ID, "profile-abc123")
	}

	if profile.Name != "TLS Server Certificate Profile" {
		t.Errorf("Name = %v, want %v", profile.Name, "TLS Server Certificate Profile")
	}

	if profile.Type != "SERVER_CERTIFICATE" {
		t.Errorf("Type = %v, want %v", profile.Type, "SERVER_CERTIFICATE")
	}

	if profile.Status != "active" {
		t.Errorf("Status = %v, want %v", profile.Status, "active")
	}

	if profile.EnrollmentMethod != "REST_API" {
		t.Errorf("EnrollmentMethod = %v, want %v", profile.EnrollmentMethod, "REST_API")
	}

	if profile.KeySize != 2048 {
		t.Errorf("KeySize = %v, want %v", profile.KeySize, 2048)
	}

	if profile.KeyAlgorithm != "RSA" {
		t.Errorf("KeyAlgorithm = %v, want %v", profile.KeyAlgorithm, "RSA")
	}

	if profile.SignatureAlgorithm != "SHA256" {
		t.Errorf("SignatureAlgorithm = %v, want %v", profile.SignatureAlgorithm, "SHA256")
	}

	// Verify arrays
	expectedTags := []string{"production", "web-server", "ssl"}
	if len(profile.Tags) != len(expectedTags) {
		t.Errorf("Tags length = %v, want %v", len(profile.Tags), len(expectedTags))
	}

	for i, tag := range expectedTags {
		if i < len(profile.Tags) && profile.Tags[i] != tag {
			t.Errorf("Tags[%d] = %v, want %v", i, profile.Tags[i], tag)
		}
	}

	// Verify arrays
	if len(profile.Tags) != len(expectedTags) {
		t.Errorf("Tags length = %v, want %v", len(profile.Tags), len(expectedTags))
	}

	// Verify time parsing
	if profile.CreatedAt == nil {
		t.Error("CreatedAt should not be nil")
	}

	if profile.UpdatedAt == nil {
		t.Error("UpdatedAt should not be nil")
	}
}