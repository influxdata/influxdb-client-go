package influxclient

import "fmt"

// ServerError holds InfluxDB server error info
// ServerError represents an error returned from an InfluxDB API server.
type ServerError struct {
	// Code holds the Influx error code, or empty if the code is unknown.
	Code string `json:"code"`
	// Message holds the error message.
	Message string `json:"message"`
	// StatusCode holds the HTTP response status code.
	StatusCode int `json:"-"`
}

// NewServerError returns new with just a message
func NewServerError(message string) *ServerError {
	return &ServerError{Message: message}
}

// Error implements Error interface
func (e *ServerError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	} else {
		return e.Message
	}
}
