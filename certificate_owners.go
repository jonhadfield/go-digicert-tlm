package digicert

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type CertificateOwnersService struct {
	client *Client
}

type CertificateOwner struct {
	ID               string     `json:"id,omitempty"`
	Email            string     `json:"email,omitempty"`
	FirstName        string     `json:"first_name,omitempty"`
	LastName         string     `json:"last_name,omitempty"`
	PhoneNumber      string     `json:"phone_number,omitempty"`
	JobTitle         string     `json:"job_title,omitempty"`
	Company          string     `json:"company,omitempty"`
	Department       string     `json:"department,omitempty"`
	IsActive         bool       `json:"is_active,omitempty"`
	CreatedAt        *time.Time `json:"created_at,omitempty"`
	UpdatedAt        *time.Time `json:"updated_at,omitempty"`
}

type CertificateOwnerRequest struct {
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number,omitempty"`
	JobTitle    string `json:"job_title,omitempty"`
	Company     string `json:"company,omitempty"`
	Department  string `json:"department,omitempty"`
}

type CertificateOwnerListOptions struct {
	PaginationParams
	Email     string `url:"email,omitempty"`
	FirstName string `url:"first_name,omitempty"`
	LastName  string `url:"last_name,omitempty"`
	IsActive  *bool  `url:"is_active,omitempty"`
	SortBy    string `url:"sort_by,omitempty"`
	SortOrder string `url:"sort_order,omitempty"`
}

type CertificateOwnerListResponse struct {
	ListResponse
	Owners []CertificateOwner `json:"certificate_owners"`
}

// Create creates a new certificate owner
func (s *CertificateOwnersService) Create(ctx context.Context, req *CertificateOwnerRequest) (*CertificateOwner, *Response, error) {
	u := "certificate-owners"

	httpReq, err := s.client.NewRequest(ctx, http.MethodPost, u, req)
	if err != nil {
		return nil, nil, err
	}

	var owner CertificateOwner
	resp, err := s.client.Do(ctx, httpReq, &owner)
	if err != nil {
		return nil, resp, err
	}

	return &owner, resp, nil
}

// Get retrieves a certificate owner by ID
func (s *CertificateOwnersService) Get(ctx context.Context, ownerID string) (*CertificateOwner, *Response, error) {
	u := fmt.Sprintf("certificate-owners/%s", ownerID)

	httpReq, err := s.client.NewRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var owner CertificateOwner
	resp, err := s.client.Do(ctx, httpReq, &owner)
	if err != nil {
		return nil, resp, err
	}

	return &owner, resp, nil
}

// Update updates a certificate owner
func (s *CertificateOwnersService) Update(ctx context.Context, ownerID string, req *CertificateOwnerRequest) (*CertificateOwner, *Response, error) {
	u := fmt.Sprintf("certificate-owners/%s", ownerID)

	httpReq, err := s.client.NewRequest(ctx, http.MethodPut, u, req)
	if err != nil {
		return nil, nil, err
	}

	var owner CertificateOwner
	resp, err := s.client.Do(ctx, httpReq, &owner)
	if err != nil {
		return nil, resp, err
	}

	return &owner, resp, nil
}

// Delete deletes a certificate owner
func (s *CertificateOwnersService) Delete(ctx context.Context, ownerID string) (*Response, error) {
	u := fmt.Sprintf("certificate-owners/%s", ownerID)

	httpReq, err := s.client.NewRequest(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(ctx, httpReq, nil)
	return resp, err
}

// List lists certificate owners
func (s *CertificateOwnersService) List(ctx context.Context, opts *CertificateOwnerListOptions) (*CertificateOwnerListResponse, *Response, error) {
	u := "certificate-owners"

	httpReq, err := s.client.NewRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	// Add query parameters
	if opts != nil {
		q := httpReq.URL.Query()
		if opts.Email != "" {
			q.Add("email", opts.Email)
		}
		if opts.FirstName != "" {
			q.Add("first_name", opts.FirstName)
		}
		if opts.LastName != "" {
			q.Add("last_name", opts.LastName)
		}
		if opts.IsActive != nil {
			q.Add("is_active", fmt.Sprintf("%t", *opts.IsActive))
		}
		if opts.Page > 0 {
			q.Add("page", fmt.Sprintf("%d", opts.Page))
		}
		if opts.PageSize > 0 {
			q.Add("page_size", fmt.Sprintf("%d", opts.PageSize))
		}
		if opts.SortBy != "" {
			q.Add("sort_by", opts.SortBy)
		}
		if opts.SortOrder != "" {
			q.Add("sort_order", opts.SortOrder)
		}
		httpReq.URL.RawQuery = q.Encode()
	}

	var result CertificateOwnerListResponse
	resp, err := s.client.Do(ctx, httpReq, &result)
	if err != nil {
		return nil, resp, err
	}

	return &result, resp, nil
}

// AssignToCertificate assigns owners to a certificate
func (s *CertificateOwnersService) AssignToCertificate(ctx context.Context, certificateID string, ownerIDs []string) (*Response, error) {
	u := fmt.Sprintf("certificate-owners/certificate/%s", certificateID)

	req := struct {
		OwnerIDs []string `json:"owner_ids"`
	}{
		OwnerIDs: ownerIDs,
	}

	httpReq, err := s.client.NewRequest(ctx, http.MethodPut, u, req)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(ctx, httpReq, nil)
	return resp, err
}

// RemoveFromCertificate removes owners from a certificate
func (s *CertificateOwnersService) RemoveFromCertificate(ctx context.Context, certificateID string) (*Response, error) {
	u := fmt.Sprintf("certificate-owners/certificate/%s", certificateID)

	httpReq, err := s.client.NewRequest(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(ctx, httpReq, nil)
	return resp, err
}