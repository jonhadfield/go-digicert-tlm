package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jonhadfield/go-digicert"
)

func main() {
	// Get API key from environment variable
	apiKey := os.Getenv("DIGICERT_API_KEY")
	if apiKey == "" {
		log.Fatal("DIGICERT_API_KEY environment variable is required")
	}

	// Create a new client
	client, err := digicert.NewClient(apiKey)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Example 1: List certificate profiles
	fmt.Println("=== Listing Certificate Profiles ===")
	profileOpts := &digicert.ProfileListOptions{}
	profileOpts.Limit = 10
	profiles, _, err := client.Profiles.List(ctx, profileOpts)
	if err != nil {
		log.Printf("Error listing profiles: %v", err)
	} else {
		for _, profile := range profiles.Profiles {
			fmt.Printf("Profile: %s (ID: %s, Type: %s)\n", profile.Name, profile.ID, profile.Type)
		}
	}

	// Example 2: Search for certificates
	fmt.Println("\n=== Searching for Certificates ===")
	certOpts := &digicert.CertificateSearchOptions{
		Status: "active",
	}
	certOpts.Limit = 5
	certs, _, err := client.Certificates.Search(ctx, certOpts)
	if err != nil {
		log.Printf("Error searching certificates: %v", err)
	} else {
		fmt.Printf("Found %d certificates (showing first %d)\n", certs.Total, len(certs.Items))
		for _, cert := range certs.Items {
			fmt.Printf("- %s (Serial: %s, Expires: %v)\n", 
				cert.CommonName, cert.SerialNumber, cert.ValidTo)
		}
	}

	// Example 3: Issue a new certificate (requires valid profile ID and CSR)
	/*
	fmt.Println("\n=== Issuing a New Certificate ===")
	certReq := &digicert.CertificateRequest{
		Profile: digicert.ProfileReference{
			ID: "your-profile-id-here",
		},
		CSR: "-----BEGIN CERTIFICATE REQUEST-----\n...\n-----END CERTIFICATE REQUEST-----",
		Validity: &digicert.Validity{
			Years: 1,
		},
		Attributes: &digicert.CertificateAttributes{
			CommonName:   "example.com",
			Organization: "Example Corp",
			Country:      "US",
			State:        "California",
			Locality:     "San Francisco",
		},
		Tags: []string{"production", "web-server"},
	}

	newCert, _, err := client.Certificates.Issue(ctx, certReq)
	if err != nil {
		log.Printf("Error issuing certificate: %v", err)
	} else {
		if newCert.RequestID != "" {
			fmt.Printf("Certificate request submitted. Request ID: %s\n", newCert.RequestID)
			// For Microsoft CA certificates, use Pickup method with request ID
		} else if newCert.Certificate != nil {
			fmt.Printf("Certificate issued: %s\n", newCert.Certificate.SerialNumber)
		}
	}
	*/

	// Example 4: List business units
	fmt.Println("\n=== Listing Business Units ===")
	buOpts := &digicert.BusinessUnitListOptions{}
	buOpts.Limit = 5
	businessUnits, _, err := client.BusinessUnits.List(ctx, buOpts)
	if err != nil {
		log.Printf("Error listing business units: %v", err)
	} else {
		for _, bu := range businessUnits.BusinessUnits {
			fmt.Printf("Business Unit: %s (ID: %s, Seats: %d/%d)\n", 
				bu.Name, bu.ID, bu.UsedSeats, bu.LicensedSeats)
		}
	}

	// Example 5: Create an enrollment
	/*
	fmt.Println("\n=== Creating an Enrollment ===")
	enrollReq := &digicert.EnrollmentRequest{
		Profile: digicert.ProfileReference{
			ID: "your-profile-id-here",
		},
		Email:       "user@example.com",
		CommonName:  "user@example.com",
		Validity: &digicert.Validity{
			Years: 1,
		},
		Tags: []string{"user-cert"},
	}

	enrollment, _, err := client.Enrollments.Create(ctx, enrollReq)
	if err != nil {
		log.Printf("Error creating enrollment: %v", err)
	} else {
		fmt.Printf("Enrollment created. Code: %s, ID: %s\n", 
			enrollment.EnrollmentCode, enrollment.EnrollmentID)
	}
	*/

	// Example 6: Error handling
	fmt.Println("\n=== Error Handling Example ===")
	_, _, err = client.Certificates.Get(ctx, "invalid-serial-number")
	if err != nil {
		if digicert.IsNotFound(err) {
			fmt.Println("Certificate not found (404)")
		} else if digicert.IsUnauthorized(err) {
			fmt.Println("Unauthorized - check your API key (401)")
		} else if apiErr, ok := err.(*digicert.APIError); ok {
			fmt.Printf("API Error: %s (Code: %s)\n", apiErr.Message, apiErr.Code)
		} else {
			fmt.Printf("Other error: %v\n", err)
		}
	}
}