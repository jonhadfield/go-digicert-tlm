package digicert

import "fmt"

type APIError struct {
	StatusCode int      `json:"-"`
	Code       string   `json:"code"`
	Message    string   `json:"message"`
	Details    []string `json:"details,omitempty"`
	RequestID  string   `json:"request_id,omitempty"`
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("digicert: %s (code: %s, status: %d)", e.Message, e.Code, e.StatusCode)
	}
	return fmt.Sprintf("digicert: %s (status: %d)", e.Message, e.StatusCode)
}

type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("digicert: HTTP %d: %s", e.StatusCode, e.Message)
}

func IsNotFound(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 404
	}
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr.StatusCode == 404
	}
	return false
}

func IsUnauthorized(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 401
	}
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr.StatusCode == 401
	}
	return false
}

func IsForbidden(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 403
	}
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr.StatusCode == 403
	}
	return false
}