package transport

import (
	"context"
	"io"

	"github.com/kunalkushwaha/mcp-navigator-go/pkg/mcp"
)

// Transport represents a communication transport for MCP protocol
type Transport interface {
	// Connect establishes the connection
	Connect(ctx context.Context) error

	// Close closes the connection
	Close() error

	// Send sends a message to the server
	Send(message *mcp.Message) error

	// Receive receives a message from the server
	Receive() (*mcp.Message, error)

	// GetReader returns the underlying reader
	GetReader() io.Reader

	// GetWriter returns the underlying writer
	GetWriter() io.Writer

	// IsConnected returns true if the transport is connected
	IsConnected() bool
}
