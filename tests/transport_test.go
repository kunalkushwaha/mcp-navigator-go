package tests

import (
	"testing"

	"github.com/kunalkushwaha/mcp-navigator-go/pkg/transport"
)

func TestTransportTypes(t *testing.T) {
	t.Run("TCP transport creation", func(t *testing.T) {
		tcpTransport := transport.NewTCPTransport("localhost", 8811)

		if tcpTransport == nil {
			t.Fatal("Expected non-nil TCP transport")
		}

		// Test that it implements the Transport interface
		var _ transport.Transport = tcpTransport
	})

	t.Run("STDIO transport creation", func(t *testing.T) {
		stdioTransport := transport.NewStdioTransport("echo", []string{"hello"})

		if stdioTransport == nil {
			t.Fatal("Expected non-nil STDIO transport")
		}

		// Test that it implements the Transport interface
		var _ transport.Transport = stdioTransport
	})

	t.Run("WebSocket transport creation", func(t *testing.T) {
		wsTransport := transport.NewWebSocketTransport("ws://localhost:8811/mcp")

		if wsTransport == nil {
			t.Fatal("Expected non-nil WebSocket transport")
		}

		// Test that it implements the Transport interface
		var _ transport.Transport = wsTransport
	})
}

func TestTransportInterface(t *testing.T) {
	// Test that all transport types implement the expected interface methods
	transports := []struct {
		name      string
		transport transport.Transport
	}{
		{"TCP", transport.NewTCPTransport("localhost", 8811)},
		{"STDIO", transport.NewStdioTransport("echo", []string{"test"})},
		{"WebSocket", transport.NewWebSocketTransport("ws://localhost:8811/mcp")},
	}

	for _, tt := range transports {
		t.Run(tt.name+" implements Transport", func(t *testing.T) {
			// Test that the transport has the expected methods
			// (This is compile-time verification that they implement the interface)
			var _ transport.Transport = tt.transport

			// Test that Close method exists and can be called safely
			// (Even though we're not connected)
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s transport Close() panicked: %v", tt.name, r)
				}
			}()

			err := tt.transport.Close()
			// Close might return an error if not connected, that's fine
			_ = err
		})
	}
}

func TestTransportCreationEdgeCases(t *testing.T) {
	t.Run("TCP with empty host", func(t *testing.T) {
		tcpTransport := transport.NewTCPTransport("", 8811)
		if tcpTransport == nil {
			t.Error("TCP transport should handle empty host gracefully")
		}
	})

	t.Run("TCP with invalid port", func(t *testing.T) {
		tcpTransport := transport.NewTCPTransport("localhost", 0)
		if tcpTransport == nil {
			t.Error("TCP transport should handle invalid port gracefully")
		}
	})

	t.Run("STDIO with empty command", func(t *testing.T) {
		stdioTransport := transport.NewStdioTransport("", []string{})
		if stdioTransport == nil {
			t.Error("STDIO transport should handle empty command gracefully")
		}
	})

	t.Run("WebSocket with invalid URL", func(t *testing.T) {
		wsTransport := transport.NewWebSocketTransport("invalid-url")
		if wsTransport == nil {
			t.Error("WebSocket transport should handle invalid URL gracefully")
		}
	})
}
