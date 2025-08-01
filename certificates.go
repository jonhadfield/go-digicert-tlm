package digicert

import (
	"context"
	"fmt"
	"net/http"
)

type CertificatesService struct {
	client *Client
}

type Seat struct {
	SeatID string `json:"seat_id,omitempty"`
}

type SeatType struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type Account struct {
	ID string `json:"id,omitempty"`
}

type ICA struct {
	ID string `json:"id,omitempty"`
}

type Subject struct {
	CommonName        string   `json:"common_name,omitempty"`
	OrganizationName  string   `json:"organization_name,omitempty"`
	OrganizationUnits []string `json:"organization_units,omitempty"`
	Locality          string   `json:"locality,omitempty"`
	Country           string   `json:"country,omitempty"`
}

type Certificate struct {
	ID                 string                 `json:"id,omitempty"`
	Profile            ProfileReference       `json:"profile,omitempty"`
	Seat               *Seat                  `json:"seat,omitempty"`
	SeatType           *SeatType              `json:"seat_type,omitempty"`
	BusinessUnit       *BusinessUnit          `json:"business_unit,omitempty"`
	Account            *Account               `json:"account,omitempty"`
	Certificate        string                 `json:"certificate,omitempty"`
	ICA                *ICA                   `json:"ica,omitempty"`
	CommonName         string                 `json:"common_name,omitempty"`
	Status             string                 `json:"status,omitempty"`
	SerialNumber       string                 `json:"serial_number,omitempty"`
	Thumbprint         string                 `json:"thumbprint,omitempty"`
	ValidFrom          string                 `json:"valid_from,omitempty"`
	ValidTo            string                 `json:"valid_to,omitempty"`
	IssuingCAName      string                 `json:"issuing_ca_name,omitempty"`
	KeySize            string                 `json:"key_size,omitempty"`
	SignatureAlgorithm string                 `json:"signature_algorithm,omitempty"`
	Subject            *Subject               `json:"subject,omitempty"`
	CAVendor           string                 `json:"ca_vendor,omitempty"`
	Connector          string                 `json:"connector,omitempty"`
	Source             string                 `json:"source,omitempty"`
	ExpiresInDays      int                    `json:"expires_in_days,omitempty"`
	PQCVulnerable      bool                   `json:"pqc_vulnerable,omitempty"`
	ExtendedKeyUsage   string                 `json:"extended_key_usage,omitempty"`
	Escrow             bool                   `json:"escrow,omitempty"`
	Attributes         string                 `json:"attributes,omitempty"`
	CustomAttributes   map[string]interface{} `json:"custom_attributes,omitempty"`
}

type CertificateRequest struct {
	Profile          ProfileReference       `json:"profile"`
	Seat             *SeatReference         `json:"seat,omitempty"`
	CSR              string                 `json:"csr,omitempty"`
	Validity         *Validity              `json:"validity,omitempty"`
	DeliveryFormat   *DeliveryFormat        `json:"delivery_format,omitempty"`
	IncludeCAChain   bool                   `json:"include_ca_chain,omitempty"`
	Attributes       *CertificateAttributes `json:"attributes,omitempty"`
	Tags             []string               `json:"tags,omitempty"`
	CertOwnerIDs     []string               `json:"cert_owner_ids,omitempty"`
	CAAttributes     *CAAttributesWrapper   `json:"ca_attributes,omitempty"`
	CustomAttributes []CustomAttribute      `json:"custom_attributes,omitempty"`
}

type ProfileReference struct {
	ID string `json:"id"`
}

type SeatReference struct {
	SeatID string `json:"seat_id"`
}

type Validity struct {
	Years   int    `json:"years,omitempty"`
	Months  int    `json:"months,omitempty"`
	Days    int    `json:"days,omitempty"`
	EndDate string `json:"end_date,omitempty"`
}

type DeliveryFormat struct {
	Format string `json:"format,omitempty"`
}

type CertificateAttributes struct {
	CommonName         string           `json:"common_name,omitempty"`
	Organization       string           `json:"organization,omitempty"`
	OrganizationalUnit []string         `json:"organizational_unit,omitempty"`
	Country            string           `json:"country,omitempty"`
	State              string           `json:"state,omitempty"`
	Locality           string           `json:"locality,omitempty"`
	Email              string           `json:"email,omitempty"`
	SANs               *SubjectAltNames `json:"sans,omitempty"`
}

type SubjectAltNames struct {
	DNSNames    []string `json:"dns_names,omitempty"`
	IPAddresses []string `json:"ip_addresses,omitempty"`
	Emails      []string `json:"emails,omitempty"`
	URIs        []string `json:"uris,omitempty"`
	OtherNames  []string `json:"other_names,omitempty"`
}

type CAAttributesWrapper struct {
	Schema CAAttributes `json:"schema,omitempty"`
}

type CAAttributes map[string]interface{}

type CustomAttribute struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

type CertificateResponse struct {
	Certificate *Certificate `json:"certificate,omitempty"`
	RequestID   string       `json:"request_id,omitempty"`
	Chain       []string     `json:"chain,omitempty"`
	PrivateKey  string       `json:"private_key,omitempty"`
}

type CertificateSearchOptions struct {
	PaginationParams
	CommonName   string   `url:"common_name,omitempty"`
	SerialNumber string   `url:"serial_number,omitempty"`
	Status       string   `url:"status,omitempty"`
	ProfileID    string   `url:"profile_id,omitempty"`
	Tags         []string `url:"tags,omitempty"`
	SortBy       string   `url:"sort_by,omitempty"`
	SortOrder    string   `url:"sort_order,omitempty"`
}

type CertificateSearchResponse struct {
	ListResponse
	Items []Certificate `json:"items"`
}

type RevokeRequest struct {
	Reason  string `json:"reason"`
	Comment string `json:"comment,omitempty"`
}

type RenewRequest struct {
	CSR              string                 `json:"csr,omitempty"`
	Validity         *Validity              `json:"validity,omitempty"`
	DeliveryFormat   *DeliveryFormat        `json:"delivery_format,omitempty"`
	IncludeCAChain   bool                   `json:"include_ca_chain,omitempty"`
	Attributes       *CertificateAttributes `json:"attributes,omitempty"`
	Tags             []string               `json:"tags,omitempty"`
	CustomAttributes []CustomAttribute      `json:"custom_attributes,omitempty"`
}

type AdditionalFormatsResponse struct {
	Formats map[string]string `json:"formats"`
}

// Issue creates a new certificate
func (s *CertificatesService) Issue(ctx context.Context, req *CertificateRequest) (*CertificateResponse, *Response, error) {
	u := "certificate"

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

// Get retrieves a certificate by serial number
func (s *CertificatesService) Get(ctx context.Context, serialNumber string) (*Certificate, *Response, error) {
	u := fmt.Sprintf("certificate/%s", serialNumber)

	httpReq, err := s.client.NewRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var cert Certificate
	resp, err := s.client.Do(ctx, httpReq, &cert)
	if err != nil {
		return nil, resp, err
	}

	return &cert, resp, nil
}

// GetCertificate retrieves a certificate by ID
func (s *CertificatesService) GetCertificate(ctx context.Context, certificateID string) (*Certificate, *Response, error) {
	u := fmt.Sprintf("certificate-by-id/%s", certificateID)

	httpReq, err := s.client.NewRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var cert Certificate
	resp, err := s.client.Do(ctx, httpReq, &cert)
	if err != nil {
		return nil, resp, err
	}

	return &cert, resp, nil
}

// Search searches for certificates
func (s *CertificatesService) Search(ctx context.Context, opts *CertificateSearchOptions) (*CertificateSearchResponse, *Response, error) {
	u := "certificate-search"

	httpReq, err := s.client.NewRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	// Add query parameters
	if opts != nil {
		q := httpReq.URL.Query()
		if opts.CommonName != "" {
			q.Add("common_name", opts.CommonName)
		}
		if opts.SerialNumber != "" {
			q.Add("serial_number", opts.SerialNumber)
		}
		if opts.Status != "" {
			q.Add("status", opts.Status)
		}
		if opts.ProfileID != "" {
			q.Add("profile_id", opts.ProfileID)
		}
		for _, tag := range opts.Tags {
			q.Add("tags", tag)
		}
		if opts.Offset > 0 {
			q.Add("offset", fmt.Sprintf("%d", opts.Offset))
		}
		if opts.Limit > 0 {
			q.Add("limit", fmt.Sprintf("%d", opts.Limit))
		}
		httpReq.URL.RawQuery = q.Encode()
	}

	fmt.Println(httpReq.URL.String())
	var result CertificateSearchResponse
	resp, err := s.client.Do(ctx, httpReq, &result)
	if err != nil {
		return nil, resp, err
	}

	return &result, resp, nil
}

// Revoke revokes a certificate
func (s *CertificatesService) Revoke(ctx context.Context, serialNumber string, req *RevokeRequest) (*Response, error) {
	u := fmt.Sprintf("certificate/%s/revoke", serialNumber)

	httpReq, err := s.client.NewRequest(ctx, http.MethodPut, u, req)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(ctx, httpReq, nil)
	return resp, err
}

// Unrevoke unrevokes a certificate
func (s *CertificatesService) Unrevoke(ctx context.Context, serialNumber string) (*Response, error) {
	u := fmt.Sprintf("certificate/%s/revoke", serialNumber)

	httpReq, err := s.client.NewRequest(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(ctx, httpReq, nil)
	return resp, err
}

// Renew renews a certificate
func (s *CertificatesService) Renew(ctx context.Context, serialNumber string, req *RenewRequest) (*CertificateResponse, *Response, error) {
	u := fmt.Sprintf("certificate/%s/renew", serialNumber)

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

// GetAdditionalFormats retrieves additional certificate formats
func (s *CertificatesService) GetAdditionalFormats(ctx context.Context, serialNumber string) (*AdditionalFormatsResponse, *Response, error) {
	u := fmt.Sprintf("certificate/%s/additional-formats", serialNumber)

	httpReq, err := s.client.NewRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var formats AdditionalFormatsResponse
	resp, err := s.client.Do(ctx, httpReq, &formats)
	if err != nil {
		return nil, resp, err
	}

	return &formats, resp, nil
}

// Pickup retrieves a certificate by request ID (for Microsoft CA certificates)
func (s *CertificatesService) Pickup(ctx context.Context, requestID string) (*CertificateResponse, *Response, error) {
	u := fmt.Sprintf("certificate-pickup/%s", requestID)

	httpReq, err := s.client.NewRequest(ctx, http.MethodPost, u, nil)
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
