package digicert

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestEnrollmentsService_Create(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("successful enrollment creation", func(t *testing.T) {
		mockRequest := &EnrollmentRequest{
			Profile: ProfileReference{
				ID: "profile-123",
			},
			Email:      "user@example.com",
			CommonName: "test.example.com",
		}

		mockResponse := &EnrollmentResponse{
			EnrollmentID:   "enrollment-123",
			EnrollmentCode: "CODE-123",
			Status:         "pending",
			Message:        "Enrollment created successfully",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/mpki/api/v1/enrollment" {
				t.Errorf("Expected path /mpki/api/v1/enrollment, got %s", r.URL.Path)
			}
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST request, got %s", r.Method)
			}

			var reqBody EnrollmentRequest
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

		result, resp, err := client.Enrollments.Create(ctx, mockRequest)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusCreated)
		}

		if result.EnrollmentID != "enrollment-123" {
			t.Errorf("EnrollmentID = %v, want %v", result.EnrollmentID, "enrollment-123")
		}
	})
}

func TestEnrollmentsService_Get(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("successful enrollment retrieval", func(t *testing.T) {
		enrollmentCode := "CODE-123"
		createdAt := time.Now()

		mockEnrollment := &Enrollment{
			ID:             "enrollment-123",
			EnrollmentCode: enrollmentCode,
			Status:         "pending",
			ProfileID:      "profile-123",
			CommonName:     "test.example.com",
			Email:          "user@example.com",
			CreatedAt:      &createdAt,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/mpki/api/v1/enrollment/" + enrollmentCode
			if r.URL.Path != expectedPath {
				t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
			}
			if r.Method != http.MethodGet {
				t.Errorf("Expected GET request, got %s", r.Method)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockEnrollment)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		result, resp, err := client.Enrollments.Get(ctx, enrollmentCode)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusOK)
		}

		if result.ID != mockEnrollment.ID {
			t.Errorf("ID = %v, want %v", result.ID, mockEnrollment.ID)
		}
	})
}

func TestEnrollmentsService_ListDetails(t *testing.T) {
	client, _ := NewClient("test-key")
	ctx := context.Background()

	t.Run("successful enrollment details listing", func(t *testing.T) {
		createdAt := time.Now()
		mockResponse := &EnrollmentDetailsResponse{
			ListResponse: ListResponse{
				Total:  25,
				Offset: 10,
				Limit:  10,
			},
			Enrollments: []Enrollment{
				{
					ID:             "enrollment-1",
					EnrollmentCode: "CODE-001",
					Status:         "pending",
					ProfileID:      "profile-123",
					CommonName:     "test1.example.com",
					Email:          "user1@example.com",
					CreatedAt:      &createdAt,
				},
				{
					ID:             "enrollment-2",
					EnrollmentCode: "CODE-002",
					Status:         "completed",
					ProfileID:      "profile-456",
					CommonName:     "test2.example.com",
					Email:          "user2@example.com",
					CreatedAt:      &createdAt,
				},
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/mpki/api/v1/enrollment-details" {
				t.Errorf("Expected path /mpki/api/v1/enrollment-details, got %s", r.URL.Path)
			}
			if r.Method != http.MethodGet {
				t.Errorf("Expected GET request, got %s", r.Method)
			}

			q := r.URL.Query()
			if q.Get("status") != "pending" {
				t.Errorf("Expected status=pending, got %s", q.Get("status"))
			}
			if q.Get("offset") != "10" {
				t.Errorf("Expected offset=10, got %s", q.Get("offset"))
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

		opts := &EnrollmentDetailsOptions{
			PaginationParams: PaginationParams{
				Offset: 10,
				Limit:  10,
			},
			Status: "pending",
		}

		result, resp, err := client.Enrollments.ListDetails(ctx, opts)
		if err != nil {
			t.Fatalf("ListDetails() error = %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %v, want %v", resp.StatusCode, http.StatusOK)
		}

		if result.Total != 25 {
			t.Errorf("Total = %v, want %v", result.Total, 25)
		}

		if len(result.Enrollments) != 2 {
			t.Errorf("Enrollments count = %v, want %v", len(result.Enrollments), 2)
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

			mockResponse := &EnrollmentDetailsResponse{
				ListResponse: ListResponse{
					Total:  5,
					Offset: 0,
					Limit:  0,
				},
				Enrollments: make([]Enrollment, 5),
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

		opts := &EnrollmentDetailsOptions{
			PaginationParams: PaginationParams{
				Offset: 0, // Should not be added to query
				Limit:  0, // Should not be added to query
			},
		}

		_, _, err := client.Enrollments.ListDetails(ctx, opts)
		if err != nil {
			t.Fatalf("ListDetails() error = %v", err)
		}
	})
}