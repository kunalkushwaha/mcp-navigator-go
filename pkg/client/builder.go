package client

import (
	"log"
	"time"

	"github.com/kunalkushwaha/mcp-navigator-go/pkg/transport"
)

// ClientBuilder provides a fluent interface for building MCP clients
type ClientBuilder struct {
	transport transport.Transport
	config    ClientConfig
}

// NewClientBuilder creates a new client builder
func NewClientBuilder() *ClientBuilder {
	return &ClientBuilder{
		config: ClientConfig{
			Name:    "mcp-client",
			Version: "1.0.0",
			Timeout: 30 * time.Second,
		},
	}
}

// WithTCPTransport configures the client to use TCP transport
func (b *ClientBuilder) WithTCPTransport(host string, port int) *ClientBuilder {
	b.transport = transport.NewTCPTransport(host, port)
	return b
}

// WithSTDIOTransport configures the client to use STDIO transport
func (b *ClientBuilder) WithSTDIOTransport(command string, args []string) *ClientBuilder {
	b.transport = transport.NewStdioTransport(command, args)
	return b
}

// WithWebSocketTransport configures the client to use WebSocket transport
func (b *ClientBuilder) WithWebSocketTransport(url string) *ClientBuilder {
	b.transport = transport.NewWebSocketTransport(url)
	return b
}

// WithTransport sets a custom transport
func (b *ClientBuilder) WithTransport(transport transport.Transport) *ClientBuilder {
	b.transport = transport
	return b
}

// WithName sets the client name
func (b *ClientBuilder) WithName(name string) *ClientBuilder {
	b.config.Name = name
	return b
}

// WithVersion sets the client version
func (b *ClientBuilder) WithVersion(version string) *ClientBuilder {
	b.config.Version = version
	return b
}

// WithLogger sets the logger
func (b *ClientBuilder) WithLogger(logger *log.Logger) *ClientBuilder {
	b.config.Logger = logger
	return b
}

// WithTimeout sets the operation timeout
func (b *ClientBuilder) WithTimeout(timeout time.Duration) *ClientBuilder {
	b.config.Timeout = timeout
	return b
}

// Build creates the MCP client
func (b *ClientBuilder) Build() *Client {
	if b.transport == nil {
		// Default to TCP localhost:8811
		b.transport = transport.NewTCPTransport("localhost", 8811)
	}

	return NewClient(b.transport, b.config)
}

// Convenience functions for common configurations

// NewTCPClient creates a client with TCP transport using builder pattern
func NewTCPClient(host string, port int) *Client {
	return NewClientBuilder().
		WithTCPTransport(host, port).
		Build()
}

// NewSTDIOClient creates a client with STDIO transport using builder pattern
func NewSTDIOClient(command string, args []string) *Client {
	return NewClientBuilder().
		WithSTDIOTransport(command, args).
		Build()
}

// NewWebSocketClient creates a client with WebSocket transport using builder pattern
func NewWebSocketClient(url string) *Client {
	return NewClientBuilder().
		WithWebSocketTransport(url).
		Build()
}
