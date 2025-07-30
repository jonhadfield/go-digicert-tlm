# go-digicert

A Go client library for the DigiCert® Trust Lifecycle Manager REST API.

## Installation

```bash
go get github.com/jonhadfield/go-digicert
```

## Usage

```go
package main

import (
    "context"
    "log"
    "github.com/jonhadfield/go-digicert"
)

func main() {
    // Create client
    client, err := digicert.NewClient("your-api-key")
    if err != nil {
        log.Fatal(err)
    }

    // List certificates
    ctx := context.Background()
    certs, _, err := client.Certificates.Search(ctx, &digicert.CertificateSearchOptions{
        Status: "active",
    })
    if err != nil {
        log.Fatal(err)
    }

    for _, cert := range certs.Certificates {
        log.Printf("Certificate: %s", cert.CommonName)
    }
}
```

## Features

### Implemented Services

- **Certificates**: Issue, search, get, revoke, renew certificates
- **Enrollments**: Create and manage certificate enrollments
- **Business Units**: Manage organizational units and seat allocations
- **Certificate Owners**: Manage certificate ownership
- **Profiles**: List and retrieve certificate profiles

### Core Features

- Full REST API coverage
- Context support for cancellation
- Comprehensive error handling
- Pagination support
- Custom HTTP client support

## Authentication

The library uses API key authentication. You can obtain an API key from your DigiCert Trust Lifecycle Manager account.

```go
client, err := digicert.NewClient("your-api-key")
```

## Error Handling

The library provides typed errors for better error handling:

```go
_, _, err := client.Certificates.Get(ctx, "serial-number")
if err != nil {
    if digicert.IsNotFound(err) {
        // Handle 404
    } else if apiErr, ok := err.(*digicert.APIError); ok {
        log.Printf("API Error: %s (Code: %s)", apiErr.Message, apiErr.Code)
    }
}
```

## Examples

See the [examples](examples/) directory for more detailed usage examples.

## Configuration Options

```go
// Use custom base URL
client, err := digicert.NewClient("api-key",
    digicert.WithBaseURL("https://your-digicert-instance.com"))

// Use custom HTTP client
httpClient := &http.Client{Timeout: 60 * time.Second}
client, err := digicert.NewClient("api-key",
    digicert.WithHTTPClient(httpClient))

// Set custom user agent
client, err := digicert.NewClient("api-key",
    digicert.WithUserAgent("my-app/1.0"))
```

## API Documentation

For complete API documentation, see the [DigiCert Trust Lifecycle Manager API docs](https://docs.digicert.com/trust-lifecycle-manager-api/).

## License

This library is released under the MIT License. See [LICENSE](LICENSE) file for details.