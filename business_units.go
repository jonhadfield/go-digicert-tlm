package digicert

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type BusinessUnitsService struct {
	client *Client
}

type BusinessUnit struct {
	ID               string                 `json:"id,omitempty"`
	Name             string                 `json:"name,omitempty"`
	Description      string                 `json:"description,omitempty"`
	ParentID         string                 `json:"parent_id,omitempty"`
	AccountID        string                 `json:"account_id,omitempty"`
	IsActive         bool                   `json:"is_active,omitempty"`
	LicensedSeats    int                    `json:"licensed_seats,omitempty"`
	UsedSeats        int                    `json:"used_seats,omitempty"`
	AvailableSeats   int                    `json:"available_seats,omitempty"`
	Tags             []string               `json:"tags,omitempty"`
	CustomAttributes map[string]interface{} `json:"custom_attributes,omitempty"`
	CreatedAt        *time.Time             `json:"created_at,omitempty"`
	UpdatedAt        *time.Time             `json:"updated_at,omitempty"`
}

type BusinessUnitRequest struct {
	Name             string                 `json:"name"`
	Description      string                 `json:"description,omitempty"`
	ParentID         string                 `json:"parent_id,omitempty"`
	Tags             []string               `json:"tags,omitempty"`
	CustomAttributes map[string]interface{} `json:"custom_attributes,omitempty"`
}

type BusinessUnitAdmin struct {
	ID        string     `json:"id,omitempty"`
	Email     string     `json:"email,omitempty"`
	FirstName string     `json:"first_name,omitempty"`
	LastName  string     `json:"last_name,omitempty"`
	Role      string     `json:"role,omitempty"`
	IsActive  bool       `json:"is_active,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

type BusinessUnitAdminRequest struct {
	Email       string   `json:"email"`
	FirstName   string   `json:"first_name"`
	LastName    string   `json:"last_name"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions,omitempty"`
}

type LicensedSeats struct {
	TotalSeats     int                  `json:"total_seats"`
	UsedSeats      int                  `json:"used_seats"`
	AvailableSeats int                  `json:"available_seats"`
	SeatTypes      []SeatTypeAllocation `json:"seat_types,omitempty"`
}

type SeatTypeAllocation struct {
	Type      string `json:"type"`
	Total     int    `json:"total"`
	Used      int    `json:"used"`
	Available int    `json:"available"`
}

type BusinessUnitListOptions struct {
	PaginationParams
	Name      string `url:"name,omitempty"`
	ParentID  string `url:"parent_id,omitempty"`
	IsActive  *bool  `url:"is_active,omitempty"`
	SortBy    string `url:"sort_by,omitempty"`
	SortOrder string `url:"sort_order,omitempty"`
}

type BusinessUnitListResponse struct {
	ListResponse
	BusinessUnits []BusinessUnit `json:"business_units"`
}

// Create creates a new business unit
func (s *BusinessUnitsService) Create(ctx context.Context, req *BusinessUnitRequest) (*BusinessUnit, *Response, error) {
	u := "business-unit"

	httpReq, err := s.client.NewRequest(ctx, http.MethodPost, u, req)
	if err != nil {
		return nil, nil, err
	}

	var bu BusinessUnit
	resp, err := s.client.Do(ctx, httpReq, &bu)
	if err != nil {
		return nil, resp, err
	}

	return &bu, resp, nil
}

// Get retrieves a business unit by ID
func (s *BusinessUnitsService) Get(ctx context.Context, buID string) (*BusinessUnit, *Response, error) {
	u := fmt.Sprintf("business-unit/%s", buID)

	httpReq, err := s.client.NewRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var bu BusinessUnit
	resp, err := s.client.Do(ctx, httpReq, &bu)
	if err != nil {
		return nil, resp, err
	}

	return &bu, resp, nil
}

// Update updates a business unit
func (s *BusinessUnitsService) Update(ctx context.Context, buID string, req *BusinessUnitRequest) (*BusinessUnit, *Response, error) {
	u := fmt.Sprintf("business-unit/%s", buID)

	httpReq, err := s.client.NewRequest(ctx, http.MethodPut, u, req)
	if err != nil {
		return nil, nil, err
	}

	var bu BusinessUnit
	resp, err := s.client.Do(ctx, httpReq, &bu)
	if err != nil {
		return nil, resp, err
	}

	return &bu, resp, nil
}

// Delete deletes a business unit
func (s *BusinessUnitsService) Delete(ctx context.Context, buID string) (*Response, error) {
	u := fmt.Sprintf("business-unit/%s", buID)

	httpReq, err := s.client.NewRequest(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(ctx, httpReq, nil)
	return resp, err
}

// List lists business units
func (s *BusinessUnitsService) List(ctx context.Context, opts *BusinessUnitListOptions) (*BusinessUnitListResponse, *Response, error) {
	u := "business-unit"

	httpReq, err := s.client.NewRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	// Add query parameters
	if opts != nil {
		q := httpReq.URL.Query()
		if opts.Name != "" {
			q.Add("name", opts.Name)
		}
		if opts.ParentID != "" {
			q.Add("parent_id", opts.ParentID)
		}
		if opts.IsActive != nil {
			q.Add("is_active", fmt.Sprintf("%t", *opts.IsActive))
		}
		if opts.Offset > 0 && opts.Limit > 0 {
			q.Add("offset", fmt.Sprintf("%d", opts.Offset))
			q.Add("limit", fmt.Sprintf("%d", opts.Limit))
		}
		if opts.SortBy != "" {
			q.Add("sort_by", opts.SortBy)
		}
		if opts.SortOrder != "" {
			q.Add("sort_order", opts.SortOrder)
		}
		httpReq.URL.RawQuery = q.Encode()
	}

	var result BusinessUnitListResponse
	resp, err := s.client.Do(ctx, httpReq, &result)
	if err != nil {
		return nil, resp, err
	}

	return &result, resp, nil
}

// GetLicensedSeats retrieves licensed seat information for a business unit
func (s *BusinessUnitsService) GetLicensedSeats(ctx context.Context, buID string) (*LicensedSeats, *Response, error) {
	u := fmt.Sprintf("business-unit/%s/licensed-seats", buID)

	httpReq, err := s.client.NewRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var seats LicensedSeats
	resp, err := s.client.Do(ctx, httpReq, &seats)
	if err != nil {
		return nil, resp, err
	}

	return &seats, resp, nil
}

// AddAdmin adds an administrator to a business unit
func (s *BusinessUnitsService) AddAdmin(ctx context.Context, buID string, req *BusinessUnitAdminRequest) (*BusinessUnitAdmin, *Response, error) {
	u := fmt.Sprintf("business-unit/%s/admin", buID)

	httpReq, err := s.client.NewRequest(ctx, http.MethodPost, u, req)
	if err != nil {
		return nil, nil, err
	}

	var admin BusinessUnitAdmin
	resp, err := s.client.Do(ctx, httpReq, &admin)
	if err != nil {
		return nil, resp, err
	}

	return &admin, resp, nil
}

// RemoveAdmin removes an administrator from a business unit
func (s *BusinessUnitsService) RemoveAdmin(ctx context.Context, buID, adminID string) (*Response, error) {
	u := fmt.Sprintf("business-unit/%s/admin/%s", buID, adminID)

	httpReq, err := s.client.NewRequest(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(ctx, httpReq, nil)
	return resp, err
}

// ListAdmins lists administrators for a business unit
func (s *BusinessUnitsService) ListAdmins(ctx context.Context, buID string) ([]BusinessUnitAdmin, *Response, error) {
	u := fmt.Sprintf("business-unit/%s/admin", buID)

	httpReq, err := s.client.NewRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var admins []BusinessUnitAdmin
	resp, err := s.client.Do(ctx, httpReq, &admins)
	if err != nil {
		return nil, resp, err
	}

	return admins, resp, nil
}
