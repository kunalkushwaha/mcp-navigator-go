package client

import "errors"

// Library-friendly error types for better error handling in third-party applications

var (
	// ErrNotConnected indicates the client is not connected to a server
	ErrNotConnected = errors.New("client not connected")

	// ErrNotInitialized indicates the MCP protocol has not been initialized
	ErrNotInitialized = errors.New("client not initialized")

	// ErrConnectionClosed indicates the connection was closed unexpectedly
	ErrConnectionClosed = errors.New("connection closed")

	// ErrTimeout indicates an operation timed out
	ErrTimeout = errors.New("operation timed out")

	// ErrInvalidResponse indicates the server returned an invalid response
	ErrInvalidResponse = errors.New("invalid server response")
)

// MCPError represents an error from the MCP server
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e *MCPError) Error() string {
	return e.Message
}

// IsErrorCode checks if an error is an MCPError with a specific code
func IsErrorCode(err error, code int) bool {
	if mcpErr, ok := err.(*MCPError); ok {
		return mcpErr.Code == code
	}
	return false
}

// TransportError represents a transport-level error
type TransportError struct {
	Type    string // "tcp", "stdio", "websocket", etc.
	Message string
	Cause   error
}

func (e *TransportError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *TransportError) Unwrap() error {
	return e.Cause
}

// NewTransportError creates a new transport error
func NewTransportError(transportType, message string, cause error) *TransportError {
	return &TransportError{
		Type:    transportType,
		Message: message,
		Cause:   cause,
	}
}
