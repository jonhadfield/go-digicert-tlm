/*
Package digicert provides a Go client library for the DigiCert Trust Lifecycle Manager REST API.

# Basic Usage

Create a client with your API key:

	client, err := digicert.NewClient("your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	// Issue a certificate
	cert, _, err := client.Certificates.Issue(ctx, &digicert.CertificateRequest{
		Profile: digicert.ProfileReference{ID: "profile-id"},
		CSR: "-----BEGIN CERTIFICATE REQUEST-----...",
	})

# Authentication

The library uses API key authentication via the X-API-Key header:

	client, err := digicert.NewClient("your-api-key")

# Error Handling

The library provides typed errors for better error handling:

	cert, _, err := client.Certificates.Get(ctx, "serial-number")
	if err != nil {
		if digicert.IsNotFound(err) {
			// Handle 404 error
		} else if apiErr, ok := err.(*digicert.APIError); ok {
			// Handle API error with code and message
			log.Printf("API Error: %s (Code: %s)", apiErr.Message, apiErr.Code)
		}
	}

# Services

The client provides access to various API services:

  - Certificates: Issue, search, get, revoke, and renew certificates
  - Enrollments: Create and manage certificate enrollments
  - BusinessUnits: Manage organizational units and seat allocations
  - CertificateOwners: Manage certificate ownership
  - Profiles: List and retrieve certificate profiles
  - Agents: Certificate discovery agents (placeholder)
  - Automation: Certificate lifecycle automation (placeholder)
  - AuditLog: Audit log retrieval (placeholder)
  - CustomFields: Custom field management (placeholder)
  - ACME: ACME protocol operations (placeholder)

# Configuration

The client can be configured with various options:

	// Use custom HTTP client
	httpClient := &http.Client{Timeout: 60 * time.Second}
	client, err := digicert.NewClient("api-key", 
		digicert.WithHTTPClient(httpClient))

	// Use custom base URL
	client, err := digicert.NewClient("api-key",
		digicert.WithBaseURL("https://your-instance.digicert.com"))

	// Set custom user agent
	client, err := digicert.NewClient("api-key",
		digicert.WithUserAgent("my-app/1.0"))

For more information, see https://docs.digicert.com/trust-lifecycle-manager-api/
*/
package digicert