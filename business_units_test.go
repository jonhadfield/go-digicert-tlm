package digicert

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBusinessUnitsService_Create(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("successful business unit creation", func(t *testing.T) {
		mockRequest := &BusinessUnitRequest{
			Name:        "Test Business Unit",
			Description: "A test business unit for integration testing",
			ParentID:    "parent-bu-123",
			Tags:        []string{"test", "development"},
			CustomAttributes: map[string]interface{}{
				"department": "IT",
				"cost_center": "CC-1234",
			},
		}

		createdAt := time.Now()
		mockResponse := &BusinessUnit{
			ID:               "bu-123",
			Name:             mockRequest.Name,
			Description:      mockRequest.Description,
			ParentID:         mockRequest.ParentID,
			AccountID:        "account-456",
			IsActive:         true,
			LicensedSeats:    100,
			UsedSeats:        25,
			AvailableSeats:   75,
			Tags:             mockRequest.Tags,
			CustomAttributes: mockRequest.CustomAttributes,
			CreatedAt:        &createdAt,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/mpki/api/v1/business-unit" {
				t.Errorf("Expected path /mpki/api/v1/business-unit, got %s", r.URL.Path)
			}
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST request, got %s", r.Method)
			}

			var reqBody BusinessUnitRequest
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			if reqBody.Name != mockRequest.Name {
				t.Errorf("Expected name %s, got %s", mockRequest.Name, reqBody.Name)
			}

			if reqBody.Description != mockRequest.Description {
				t.Errorf("Expected description %s, got %s", mockRequest.Description, reqBody.Description)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		result, resp, err := client.BusinessUnits.Create(ctx, mockRequest)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusCreated)
		}

		if result.ID != mockResponse.ID {
			t.Errorf("ID = %v, want %v", result.ID, mockResponse.ID)
		}

		if result.Name != mockResponse.Name {
			t.Errorf("Name = %v, want %v", result.Name, mockResponse.Name)
		}

		if result.AvailableSeats != mockResponse.AvailableSeats {
			t.Errorf("AvailableSeats = %v, want %v", result.AvailableSeats, mockResponse.AvailableSeats)
		}
	})

	t.Run("duplicate business unit name error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(APIError{
				Code:    "DUPLICATE_NAME",
				Message: "Business unit with this name already exists",
			})
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		mockRequest := &BusinessUnitRequest{
			Name: "Existing Business Unit",
		}

		_, _, err := client.BusinessUnits.Create(ctx, mockRequest)
		if err == nil {
			t.Fatal("Expected error for duplicate name")
		}

		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("Error type = %T, want *APIError", err)
		}

		if apiErr.Code != "DUPLICATE_NAME" {
			t.Errorf("Error Code = %v, want %v", apiErr.Code, "DUPLICATE_NAME")
		}
	})
}

func TestBusinessUnitsService_Get(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("successful business unit retrieval", func(t *testing.T) {
		buID := "bu-123"
		createdAt := time.Now().Add(-30 * 24 * time.Hour)
		updatedAt := time.Now().Add(-24 * time.Hour)

		mockBU := &BusinessUnit{
			ID:               buID,
			Name:             "Test Business Unit",
			Description:      "Test BU Description",
			ParentID:         "parent-bu-456",
			AccountID:        "account-789",
			IsActive:         true,
			LicensedSeats:    200,
			UsedSeats:        150,
			AvailableSeats:   50,
			Tags:             []string{"production", "web-services"},
			CustomAttributes: map[string]interface{}{"region": "us-west", "tier": "premium"},
			CreatedAt:        &createdAt,
			UpdatedAt:        &updatedAt,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/mpki/api/v1/business-unit/" + buID
			if r.URL.Path != expectedPath {
				t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
			}
			if r.Method != http.MethodGet {
				t.Errorf("Expected GET request, got %s", r.Method)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockBU)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		result, resp, err := client.BusinessUnits.Get(ctx, buID)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusOK)
		}

		if result.ID != mockBU.ID {
			t.Errorf("ID = %v, want %v", result.ID, mockBU.ID)
		}

		if result.Name != mockBU.Name {
			t.Errorf("Name = %v, want %v", result.Name, mockBU.Name)
		}

		if result.LicensedSeats != mockBU.LicensedSeats {
			t.Errorf("LicensedSeats = %v, want %v", result.LicensedSeats, mockBU.LicensedSeats)
		}

		if len(result.Tags) != len(mockBU.Tags) {
			t.Errorf("Tags length = %v, want %v", len(result.Tags), len(mockBU.Tags))
		}
	})

	t.Run("business unit not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(APIError{
				Code:    "BUSINESS_UNIT_NOT_FOUND",
				Message: "Business unit not found",
			})
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		_, _, err := client.BusinessUnits.Get(ctx, "nonexistent-bu")
		if err == nil {
			t.Fatal("Expected error for nonexistent business unit")
		}

		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("Error type = %T, want *APIError", err)
		}

		if apiErr.Code != "BUSINESS_UNIT_NOT_FOUND" {
			t.Errorf("Error Code = %v, want %v", apiErr.Code, "BUSINESS_UNIT_NOT_FOUND")
		}
	})
}

func TestBusinessUnitsService_Update(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("successful business unit update", func(t *testing.T) {
		buID := "bu-123"
		mockRequest := &BusinessUnitRequest{
			Name:        "Updated Business Unit",
			Description: "Updated description",
			Tags:        []string{"updated", "modified"},
			CustomAttributes: map[string]interface{}{
				"status": "active",
				"priority": "high",
			},
		}

		updatedAt := time.Now()
		mockResponse := &BusinessUnit{
			ID:               buID,
			Name:             mockRequest.Name,
			Description:      mockRequest.Description,
			IsActive:         true,
			LicensedSeats:    100,
			UsedSeats:        50,
			AvailableSeats:   50,
			Tags:             mockRequest.Tags,
			CustomAttributes: mockRequest.CustomAttributes,
			UpdatedAt:        &updatedAt,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/mpki/api/v1/business-unit/" + buID
			if r.URL.Path != expectedPath {
				t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
			}
			if r.Method != http.MethodPut {
				t.Errorf("Expected PUT request, got %s", r.Method)
			}

			var reqBody BusinessUnitRequest
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			if reqBody.Name != mockRequest.Name {
				t.Errorf("Expected name %s, got %s", mockRequest.Name, reqBody.Name)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		result, resp, err := client.BusinessUnits.Update(ctx, buID, mockRequest)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusOK)
		}

		if result.Name != mockResponse.Name {
			t.Errorf("Name = %v, want %v", result.Name, mockResponse.Name)
		}

		if result.Description != mockResponse.Description {
			t.Errorf("Description = %v, want %v", result.Description, mockResponse.Description)
		}
	})
}

func TestBusinessUnitsService_Delete(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("successful business unit deletion", func(t *testing.T) {
		buID := "bu-123"

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/mpki/api/v1/business-unit/" + buID
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

		resp, err := client.BusinessUnits.Delete(ctx, buID)
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusNoContent)
		}
	})

	t.Run("business unit has dependencies error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(APIError{
				Code:    "HAS_DEPENDENCIES",
				Message: "Cannot delete business unit with active certificates or sub-units",
			})
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		_, err := client.BusinessUnits.Delete(ctx, "bu-with-deps")
		if err == nil {
			t.Fatal("Expected error for business unit with dependencies")
		}

		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("Error type = %T, want *APIError", err)
		}

		if apiErr.Code != "HAS_DEPENDENCIES" {
			t.Errorf("Error Code = %v, want %v", apiErr.Code, "HAS_DEPENDENCIES")
		}
	})
}

func TestBusinessUnitsService_List(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("successful business units listing with filters", func(t *testing.T) {
		mockResponse := &BusinessUnitListResponse{
			ListResponse: ListResponse{
				Total:  50,
				Offset: 20,
				Limit:  10,
			},
			BusinessUnits: []BusinessUnit{
				{
					ID:             "bu-1",
					Name:           "Engineering",
					Description:    "Engineering Department",
					IsActive:       true,
					LicensedSeats:  100,
					UsedSeats:      75,
					AvailableSeats: 25,
					Tags:           []string{"engineering", "development"},
				},
				{
					ID:             "bu-2",
					Name:           "Marketing",
					Description:    "Marketing Department",
					IsActive:       true,
					LicensedSeats:  50,
					UsedSeats:      30,
					AvailableSeats: 20,
					Tags:           []string{"marketing", "sales"},
				},
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/mpki/api/v1/business-unit" {
				t.Errorf("Expected path /mpki/api/v1/business-unit, got %s", r.URL.Path)
			}
			if r.Method != http.MethodGet {
				t.Errorf("Expected GET request, got %s", r.Method)
			}

			q := r.URL.Query()
			if q.Get("name") != "Engineering" {
				t.Errorf("Expected name=Engineering, got %s", q.Get("name"))
			}
			if q.Get("is_active") != "true" {
				t.Errorf("Expected is_active=true, got %s", q.Get("is_active"))
			}
			if q.Get("offset") != "20" {
				t.Errorf("Expected offset=20, got %s", q.Get("offset"))
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

		isActive := true
		opts := &BusinessUnitListOptions{
			PaginationParams: PaginationParams{
				Offset: 20,
				Limit:  10,
			},
			Name:     "Engineering",
			IsActive: &isActive,
		}

		result, resp, err := client.BusinessUnits.List(ctx, opts)
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusOK)
		}

		if result.Total != mockResponse.Total {
			t.Errorf("Total = %v, want %v", result.Total, mockResponse.Total)
		}

		if len(result.BusinessUnits) != len(mockResponse.BusinessUnits) {
			t.Errorf("BusinessUnits count = %v, want %v", len(result.BusinessUnits), len(mockResponse.BusinessUnits))
		}

		if result.BusinessUnits[0].Name != "Engineering" {
			t.Errorf("First BU name = %v, want %v", result.BusinessUnits[0].Name, "Engineering")
		}
	})

	t.Run("list with pagination parameters disabled", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()

			// Verify offset and limit are not present when they are 0
			if q.Has("offset") {
				t.Errorf("offset parameter should not be present when value is 0")
			}
			if q.Has("limit") {
				t.Errorf("limit parameter should not be present when value is 0")
			}

			mockResponse := &BusinessUnitListResponse{
				ListResponse: ListResponse{
					Total:  3,
					Offset: 0,
					Limit:  0,
				},
				BusinessUnits: make([]BusinessUnit, 3),
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		opts := &BusinessUnitListOptions{
			PaginationParams: PaginationParams{
				Offset: 0, // Should not be added to query
				Limit:  0, // Should not be added to query
			},
		}

		_, _, err := client.BusinessUnits.List(ctx, opts)
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
	})
}

func TestBusinessUnitsService_GetLicensedSeats(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("successful licensed seats retrieval", func(t *testing.T) {
		buID := "bu-123"
		mockSeats := &LicensedSeats{
			TotalSeats:     1000,
			UsedSeats:      650,
			AvailableSeats: 350,
			SeatTypes: []SeatTypeAllocation{
				{
					Type:      "DISCOVERY_SEAT",
					Total:     500,
					Used:      300,
					Available: 200,
				},
				{
					Type:      "MANAGEMENT_SEAT",
					Total:     500,
					Used:      350,
					Available: 150,
				},
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/mpki/api/v1/business-unit/" + buID + "/licensed-seats"
			if r.URL.Path != expectedPath {
				t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
			}
			if r.Method != http.MethodGet {
				t.Errorf("Expected GET request, got %s", r.Method)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockSeats)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		result, resp, err := client.BusinessUnits.GetLicensedSeats(ctx, buID)
		if err != nil {
			t.Fatalf("GetLicensedSeats() error = %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusOK)
		}

		if result.TotalSeats != mockSeats.TotalSeats {
			t.Errorf("TotalSeats = %v, want %v", result.TotalSeats, mockSeats.TotalSeats)
		}

		if result.UsedSeats != mockSeats.UsedSeats {
			t.Errorf("UsedSeats = %v, want %v", result.UsedSeats, mockSeats.UsedSeats)
		}

		if len(result.SeatTypes) != len(mockSeats.SeatTypes) {
			t.Errorf("SeatTypes count = %v, want %v", len(result.SeatTypes), len(mockSeats.SeatTypes))
		}

		if result.SeatTypes[0].Type != "DISCOVERY_SEAT" {
			t.Errorf("First seat type = %v, want %v", result.SeatTypes[0].Type, "DISCOVERY_SEAT")
		}
	})
}

func TestBusinessUnitsService_AdminManagement(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("add admin successfully", func(t *testing.T) {
		buID := "bu-123"
		mockRequest := &BusinessUnitAdminRequest{
			Email:       "admin@example.com",
			FirstName:   "Jane",
			LastName:    "Admin",
			Role:        "Administrator",
			Permissions: []string{"read", "write", "admin"},
		}

		createdAt := time.Now()
		mockResponse := &BusinessUnitAdmin{
			ID:        "admin-456",
			Email:     mockRequest.Email,
			FirstName: mockRequest.FirstName,
			LastName:  mockRequest.LastName,
			Role:      mockRequest.Role,
			IsActive:  true,
			CreatedAt: &createdAt,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/mpki/api/v1/business-unit/" + buID + "/admin"
			if r.URL.Path != expectedPath {
				t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
			}
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST request, got %s", r.Method)
			}

			var reqBody BusinessUnitAdminRequest
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			if reqBody.Email != mockRequest.Email {
				t.Errorf("Expected email %s, got %s", mockRequest.Email, reqBody.Email)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		result, resp, err := client.BusinessUnits.AddAdmin(ctx, buID, mockRequest)
		if err != nil {
			t.Fatalf("AddAdmin() error = %v", err)
		}

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusCreated)
		}

		if result.Email != mockResponse.Email {
			t.Errorf("Email = %v, want %v", result.Email, mockResponse.Email)
		}
	})

	t.Run("remove admin successfully", func(t *testing.T) {
		buID := "bu-123"
		adminID := "admin-456"

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/mpki/api/v1/business-unit/" + buID + "/admin/" + adminID
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

		resp, err := client.BusinessUnits.RemoveAdmin(ctx, buID, adminID)
		if err != nil {
			t.Fatalf("RemoveAdmin() error = %v", err)
		}

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusNoContent)
		}
	})

	t.Run("list admins successfully", func(t *testing.T) {
		buID := "bu-123"
		createdAt := time.Now()
		mockAdmins := []BusinessUnitAdmin{
			{
				ID:        "admin-1",
				Email:     "admin1@example.com",
				FirstName: "John",
				LastName:  "Admin",
				Role:      "Administrator",
				IsActive:  true,
				CreatedAt: &createdAt,
			},
			{
				ID:        "admin-2",
				Email:     "admin2@example.com",
				FirstName: "Jane",
				LastName:  "Manager",
				Role:      "Manager",
				IsActive:  true,
				CreatedAt: &createdAt,
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/mpki/api/v1/business-unit/" + buID + "/admin"
			if r.URL.Path != expectedPath {
				t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
			}
			if r.Method != http.MethodGet {
				t.Errorf("Expected GET request, got %s", r.Method)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockAdmins)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		result, resp, err := client.BusinessUnits.ListAdmins(ctx, buID)
		if err != nil {
			t.Fatalf("ListAdmins() error = %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusOK)
		}

		if len(result) != len(mockAdmins) {
			t.Errorf("Admins count = %v, want %v", len(result), len(mockAdmins))
		}

		if result[0].Role != "Administrator" {
			t.Errorf("First admin role = %v, want %v", result[0].Role, "Administrator")
		}
	})
}

// TestBusinessUnitRequestValidation tests various business unit request configurations  
func TestBusinessUnitRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		request *BusinessUnitRequest
		wantErr bool
	}{
		{
			name: "valid business unit with all fields",
			request: &BusinessUnitRequest{
				Name:        "Complete Business Unit",
				Description: "A complete business unit with all fields",
				ParentID:    "parent-123",
				Tags:        []string{"production", "critical"},
				CustomAttributes: map[string]interface{}{
					"region":      "us-west-1",
					"cost_center": "CC-2024-001",
					"priority":    1,
				},
			},
			wantErr: false,
		},
		{
			name: "minimal business unit",
			request: &BusinessUnitRequest{
				Name: "Minimal BU",
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

			var unmarshaledReq BusinessUnitRequest
			if err := json.Unmarshal(data, &unmarshaledReq); err != nil {
				t.Fatalf("Failed to unmarshal request: %v", err)
			}

			if unmarshaledReq.Name != tt.request.Name {
				t.Errorf("Name mismatch after JSON round-trip")
			}

			if unmarshaledReq.Description != tt.request.Description {
				t.Errorf("Description mismatch after JSON round-trip")
			}
		})
	}
}