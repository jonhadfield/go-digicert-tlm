// Package digicert provides a client library for the DigiCert Trust Lifecycle Manager REST API.
//
// The library supports certificate management operations including issuance, renewal,
// revocation, and search, as well as management of business units, enrollments, and profiles.
package digicert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// https://one.nl.digicert.com/mpki/docs/openapi-public.json?7857

const (
	DefaultBaseURL = "https://one.digicert.com"
	APIVersion     = "v1"
	UserAgent      = "go-digicert/1.0"
)

type Client struct {
	client    *http.Client
	BaseURL   *url.URL
	UserAgent string
	apiKey    string

	// Services
	Certificates      *CertificatesService
	Orders            *OrdersService
	BusinessUnits     *BusinessUnitsService
	CertificateOwners *CertificateOwnersService
	Agents            *AgentsService
	Automation        *AutomationService
	AuditLog          *AuditLogService
	Enrollments       *EnrollmentsService
	Profiles          *ProfilesService
	CustomFields      *CustomFieldsService
	ACME              *ACMEService
}

type service struct {
	client *Client
}

type ClientOption func(*Client) error

func NewClient(apiKey string, opts ...ClientOption) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	baseURL, err := url.Parse(DefaultBaseURL)
	if err != nil {
		return nil, err
	}

	c := &Client{
		client:    &http.Client{Timeout: 30 * time.Second},
		BaseURL:   baseURL,
		UserAgent: UserAgent,
		apiKey:    apiKey,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	// Initialize services
	c.Certificates = &CertificatesService{client: c}
	c.Orders = &OrdersService{client: c}
	c.BusinessUnits = &BusinessUnitsService{client: c}
	c.CertificateOwners = &CertificateOwnersService{client: c}
	c.Agents = &AgentsService{client: c}
	c.Automation = &AutomationService{client: c}
	c.AuditLog = &AuditLogService{client: c}
	c.Enrollments = &EnrollmentsService{client: c}
	c.Profiles = &ProfilesService{client: c}
	c.CustomFields = &CustomFieldsService{client: c}
	c.ACME = &ACMEService{client: c}

	return c, nil
}

func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) error {
		if httpClient == nil {
			return fmt.Errorf("HTTP client cannot be nil")
		}
		c.client = httpClient
		return nil
	}
}

func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		u, err := url.Parse(baseURL)
		if err != nil {
			return fmt.Errorf("invalid base URL: %w", err)
		}
		c.BaseURL = u
		return nil
	}
}

func WithUserAgent(userAgent string) ClientOption {
	return func(c *Client) error {
		c.UserAgent = userAgent
		return nil
	}
}

func (c *Client) NewRequest(ctx context.Context, method, urlStr string, body interface{}) (*http.Request, error) {
	if !strings.HasSuffix(c.BaseURL.Path, "/") {
		c.BaseURL.Path += "/"
	}

	rel, err := url.Parse(fmt.Sprintf("mpki/api/%s/%s", APIVersion, strings.TrimPrefix(urlStr, "/")))
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)

	var buf io.ReadWriter
	if body != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("X-API-Key", c.apiKey)

	return req, nil
}

func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response := &Response{Response: resp}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	response.Body = data

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = c.checkError(resp, data)
		return response, err
	}

	if v != nil && len(data) > 0 {
		if err := json.Unmarshal(data, v); err != nil {
			return response, fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return response, nil
}

func (c *Client) checkError(resp *http.Response, data []byte) error {
	var apiError APIError
	if err := json.Unmarshal(data, &apiError); err != nil {
		return &HTTPError{
			StatusCode: resp.StatusCode,
			Message:    string(data),
		}
	}

	apiError.StatusCode = resp.StatusCode
	return &apiError
}

type Response struct {
	*http.Response
	Body []byte
}

type PaginationParams struct {
	Page     int `url:"page,omitempty"`
	PageSize int `url:"page_size,omitempty"`
}

type ListResponse struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
}
