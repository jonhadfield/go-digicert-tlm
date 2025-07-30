package digicert

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type EnrollmentsService struct {
	client *Client
}

type Enrollment struct {
	ID                 string                 `json:"id,omitempty"`
	EnrollmentCode     string                 `json:"enrollment_code,omitempty"`
	Status             string                 `json:"status,omitempty"`
	ProfileID          string                 `json:"profile_id,omitempty"`
	ProfileName        string                 `json:"profile_name,omitempty"`
	SeatID             string                 `json:"seat_id,omitempty"`
	CertificateID      string                 `json:"certificate_id,omitempty"`
	CommonName         string                 `json:"common_name,omitempty"`
	Email              string                 `json:"email,omitempty"`
	PhoneNumber        string                 `json:"phone_number,omitempty"`
	ExpirationDate     *time.Time             `json:"expiration_date,omitempty"`
	CreatedAt          *time.Time             `json:"created_at,omitempty"`
	UpdatedAt          *time.Time             `json:"updated_at,omitempty"`
	Tags               []string               `json:"tags,omitempty"`
	CustomAttributes   map[string]interface{} `json:"custom_attributes,omitempty"`
}

type EnrollmentRequest struct {
	Profile          ProfileReference        `json:"profile"`
	Seat             *SeatReference          `json:"seat,omitempty"`
	Validity         *Validity               `json:"validity,omitempty"`
	Email            string                  `json:"email,omitempty"`
	PhoneNumber      string                  `json:"phone_number,omitempty"`
	CommonName       string                  `json:"common_name,omitempty"`
	Attributes       *EnrollmentAttributes   `json:"attributes,omitempty"`
	Tags             []string                `json:"tags,omitempty"`
	CustomAttributes []CustomAttribute       `json:"custom_attributes,omitempty"`
	NotificationEmails []string              `json:"notification_emails,omitempty"`
}

type EnrollmentAttributes struct {
	CommonName         string           `json:"common_name,omitempty"`
	Organization       string           `json:"organization,omitempty"`
	OrganizationalUnit []string         `json:"organizational_unit,omitempty"`
	Country            string           `json:"country,omitempty"`
	State              string           `json:"state,omitempty"`
	Locality           string           `json:"locality,omitempty"`
	Email              string           `json:"email,omitempty"`
	SANs               *SubjectAltNames `json:"sans,omitempty"`
}

type EnrollmentResponse struct {
	EnrollmentID   string `json:"enrollment_id,omitempty"`
	EnrollmentCode string `json:"enrollment_code,omitempty"`
	Status         string `json:"status,omitempty"`
	Message        string `json:"message,omitempty"`
}

type EnrollmentStatusResponse struct {
	Status         string     `json:"status"`
	CertificateID  string     `json:"certificate_id,omitempty"`
	Message        string     `json:"message,omitempty"`
	LastUpdated    *time.Time `json:"last_updated,omitempty"`
}

type RedeemEnrollmentRequest struct {
	EnrollmentCode string `json:"enrollment_code"`
	CSR            string `json:"csr"`
}

type ManualEnrollmentRequest struct {
	Profile          ProfileReference        `json:"profile"`
	Seat             *SeatReference          `json:"seat,omitempty"`
	CSR              string                  `json:"csr"`
	Validity         *Validity               `json:"validity,omitempty"`
	DeliveryFormat   *DeliveryFormat         `json:"delivery_format,omitempty"`
	IncludeCAChain   bool                    `json:"include_ca_chain,omitempty"`
	Attributes       *CertificateAttributes  `json:"attributes,omitempty"`
	Tags             []string                `json:"tags,omitempty"`
	CertOwnerIDs     []string                `json:"cert_owner_ids,omitempty"`
	CustomAttributes []CustomAttribute       `json:"custom_attributes,omitempty"`
	ApproverEmail    string                  `json:"approver_email,omitempty"`
	Comments         string                  `json:"comments,omitempty"`
}

type EnrollmentDetailsOptions struct {
	PaginationParams
	Status    string `url:"status,omitempty"`
	ProfileID string `url:"profile_id,omitempty"`
	SortBy    string `url:"sort_by,omitempty"`
	SortOrder string `url:"sort_order,omitempty"`
}

type EnrollmentDetailsResponse struct {
	ListResponse
	Enrollments []Enrollment `json:"enrollments"`
}

// Create creates a new enrollment
func (s *EnrollmentsService) Create(ctx context.Context, req *EnrollmentRequest) (*EnrollmentResponse, *Response, error) {
	u := "enrollment"

	httpReq, err := s.client.NewRequest(ctx, http.MethodPost, u, req)
	if err != nil {
		return nil, nil, err
	}

	var enrollment EnrollmentResponse
	resp, err := s.client.Do(ctx, httpReq, &enrollment)
	if err != nil {
		return nil, resp, err
	}

	return &enrollment, resp, nil
}

// Get retrieves an enrollment by enrollment code
func (s *EnrollmentsService) Get(ctx context.Context, enrollmentCode string) (*Enrollment, *Response, error) {
	u := fmt.Sprintf("enrollment/%s", enrollmentCode)

	httpReq, err := s.client.NewRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var enrollment Enrollment
	resp, err := s.client.Do(ctx, httpReq, &enrollment)
	if err != nil {
		return nil, resp, err
	}

	return &enrollment, resp, nil
}

// GetStatus retrieves the status of an enrollment
func (s *EnrollmentsService) GetStatus(ctx context.Context, enrollmentID string) (*EnrollmentStatusResponse, *Response, error) {
	u := fmt.Sprintf("enrollment/%s/status", enrollmentID)

	httpReq, err := s.client.NewRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var status EnrollmentStatusResponse
	resp, err := s.client.Do(ctx, httpReq, &status)
	if err != nil {
		return nil, resp, err
	}

	return &status, resp, nil
}

// Redeem redeems an enrollment code
func (s *EnrollmentsService) Redeem(ctx context.Context, req *RedeemEnrollmentRequest) (*CertificateResponse, *Response, error) {
	u := "enrollment/redeem"

	httpReq, err := s.client.NewRequest(ctx, http.MethodPost, u, req)
	if err != nil {
		return nil, nil, err
	}

	var cert CertificateResponse
	resp, err := s.client.Do(ctx, httpReq, &cert)
	if err != nil {
		return nil, resp, err
	}

	return &cert, resp, nil
}

// CreateManualEnrollment creates a manual enrollment (requires approval)
func (s *EnrollmentsService) CreateManualEnrollment(ctx context.Context, req *ManualEnrollmentRequest) (*EnrollmentResponse, *Response, error) {
	u := "manual-enrollment"

	httpReq, err := s.client.NewRequest(ctx, http.MethodPost, u, req)
	if err != nil {
		return nil, nil, err
	}

	var enrollment EnrollmentResponse
	resp, err := s.client.Do(ctx, httpReq, &enrollment)
	if err != nil {
		return nil, resp, err
	}

	return &enrollment, resp, nil
}

// RenewManualEnrollment renews a certificate through manual enrollment
func (s *EnrollmentsService) RenewManualEnrollment(ctx context.Context, certificateID string, req *ManualEnrollmentRequest) (*EnrollmentResponse, *Response, error) {
	u := fmt.Sprintf("manual-enrollment/renew/%s", certificateID)

	httpReq, err := s.client.NewRequest(ctx, http.MethodPost, u, req)
	if err != nil {
		return nil, nil, err
	}

	var enrollment EnrollmentResponse
	resp, err := s.client.Do(ctx, httpReq, &enrollment)
	if err != nil {
		return nil, resp, err
	}

	return &enrollment, resp, nil
}

// ListDetails lists enrollment details
func (s *EnrollmentsService) ListDetails(ctx context.Context, opts *EnrollmentDetailsOptions) (*EnrollmentDetailsResponse, *Response, error) {
	u := "enrollment-details"

	httpReq, err := s.client.NewRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	// Add query parameters
	if opts != nil {
		q := httpReq.URL.Query()
		if opts.Status != "" {
			q.Add("status", opts.Status)
		}
		if opts.ProfileID != "" {
			q.Add("profile_id", opts.ProfileID)
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

	var result EnrollmentDetailsResponse
	resp, err := s.client.Do(ctx, httpReq, &result)
	if err != nil {
		return nil, resp, err
	}

	return &result, resp, nil
}

// GetDetails retrieves enrollment details by ID
func (s *EnrollmentsService) GetDetails(ctx context.Context, enrollmentID string) (*Enrollment, *Response, error) {
	u := fmt.Sprintf("enrollment-details/%s", enrollmentID)

	httpReq, err := s.client.NewRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var enrollment Enrollment
	resp, err := s.client.Do(ctx, httpReq, &enrollment)
	if err != nil {
		return nil, resp, err
	}

	return &enrollment, resp, nil
}

// GetByCertificate retrieves enrollment information for a certificate
func (s *EnrollmentsService) GetByCertificate(ctx context.Context, certificateID string) (*Enrollment, *Response, error) {
	u := fmt.Sprintf("enrollment/certificate/%s", certificateID)

	httpReq, err := s.client.NewRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var enrollment Enrollment
	resp, err := s.client.Do(ctx, httpReq, &enrollment)
	if err != nil {
		return nil, resp, err
	}

	return &enrollment, resp, nil
}