package digicert

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type ProfilesService struct {
	client *Client
}

type Profile struct {
	ID                     string                 `json:"id,omitempty"`
	Name                   string                 `json:"name,omitempty"`
	Description            string                 `json:"description,omitempty"`
	Type                   string                 `json:"type,omitempty"`
	Status                 string                 `json:"status,omitempty"`
	EnrollmentMethod       string                 `json:"enrollment_method,omitempty"`
	AuthenticationMethod   string                 `json:"authentication_method,omitempty"`
	KeyAlgorithm           string                 `json:"key_algorithm,omitempty"`
	KeySize                int                    `json:"key_size,omitempty"`
	SignatureAlgorithm     string                 `json:"signature_algorithm,omitempty"`
	Validity               ProfileValidity        `json:"validity,omitempty"`
	SubjectDNFields        []DNField              `json:"subject_dn_fields,omitempty"`
	SANFields              []SANField             `json:"san_fields,omitempty"`
	Extensions             []Extension            `json:"extensions,omitempty"`
	CustomFields           []CustomFieldDef       `json:"custom_fields,omitempty"`
	RequireApproval        bool                   `json:"require_approval,omitempty"`
	AutoRenew              bool                   `json:"auto_renew,omitempty"`
	AllowDuplicateCN       bool                   `json:"allow_duplicate_cn,omitempty"`
	Tags                   []string               `json:"tags,omitempty"`
	CreatedAt              *time.Time             `json:"created_at,omitempty"`
	UpdatedAt              *time.Time             `json:"updated_at,omitempty"`
}

type ProfileValidity struct {
	Type    string `json:"type,omitempty"`
	Years   int    `json:"years,omitempty"`
	Months  int    `json:"months,omitempty"`
	Days    int    `json:"days,omitempty"`
	MinDays int    `json:"min_days,omitempty"`
	MaxDays int    `json:"max_days,omitempty"`
}

type DNField struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
	Source   string `json:"source"`
	Value    string `json:"value,omitempty"`
}

type SANField struct {
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Source   string `json:"source"`
	Values   []string `json:"values,omitempty"`
}

type Extension struct {
	OID      string `json:"oid"`
	Critical bool   `json:"critical"`
	Value    string `json:"value"`
}

type CustomFieldDef struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Options  []string `json:"options,omitempty"`
}

type ProfileListOptions struct {
	PaginationParams
	Name             string `url:"name,omitempty"`
	Type             string `url:"type,omitempty"`
	Status           string `url:"status,omitempty"`
	EnrollmentMethod string `url:"enrollment_method,omitempty"`
	SortBy           string `url:"sort_by,omitempty"`
	SortOrder        string `url:"sort_order,omitempty"`
}

type ProfileListResponse struct {
	ListResponse
	Profiles []Profile `json:"profiles"`
}

type ProfileTemplateListResponse struct {
	Templates []ProfileTemplate `json:"templates"`
}

type ProfileTemplate struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Provider    string `json:"provider"`
}

// List lists certificate profiles
func (s *ProfilesService) List(ctx context.Context, opts *ProfileListOptions) (*ProfileListResponse, *Response, error) {
	u := "profiles"

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
		if opts.Type != "" {
			q.Add("type", opts.Type)
		}
		if opts.Status != "" {
			q.Add("status", opts.Status)
		}
		if opts.EnrollmentMethod != "" {
			q.Add("enrollment_method", opts.EnrollmentMethod)
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

	var result ProfileListResponse
	resp, err := s.client.Do(ctx, httpReq, &result)
	if err != nil {
		return nil, resp, err
	}

	return &result, resp, nil
}

// Get retrieves a certificate profile by ID
func (s *ProfilesService) Get(ctx context.Context, profileID string) (*Profile, *Response, error) {
	u := fmt.Sprintf("profiles/%s", profileID)

	httpReq, err := s.client.NewRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var profile Profile
	resp, err := s.client.Do(ctx, httpReq, &profile)
	if err != nil {
		return nil, resp, err
	}

	return &profile, resp, nil
}

// ListPublic lists publicly available certificate profiles
func (s *ProfilesService) ListPublic(ctx context.Context) (*ProfileListResponse, *Response, error) {
	u := "profiles/public"

	httpReq, err := s.client.NewRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var result ProfileListResponse
	resp, err := s.client.Do(ctx, httpReq, &result)
	if err != nil {
		return nil, resp, err
	}

	return &result, resp, nil
}

// ListTemplates lists available profile templates
func (s *ProfilesService) ListTemplates(ctx context.Context) (*ProfileTemplateListResponse, *Response, error) {
	u := "profiles/templates"

	httpReq, err := s.client.NewRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var result ProfileTemplateListResponse
	resp, err := s.client.Do(ctx, httpReq, &result)
	if err != nil {
		return nil, resp, err
	}

	return &result, resp, nil
}