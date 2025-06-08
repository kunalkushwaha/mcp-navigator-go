package tests

import (
	"testing"

	"github.com/kunalkushwaha/mcp-navigator-go/pkg/client"
	"github.com/kunalkushwaha/mcp-navigator-go/pkg/transport"
)

func TestClientBuilder(t *testing.T) {
	t.Run("Create TCP client builder", func(t *testing.T) {
		builder := client.NewClientBuilder().
			WithTCPTransport("localhost", 8811)

		if builder == nil {
			t.Error("Expected non-nil builder")
		}

		// Test that we can build a client
		client := builder.Build()
		if client == nil {
			t.Error("Expected non-nil client")
		}
	})

	t.Run("Create STDIO client builder", func(t *testing.T) {
		builder := client.NewClientBuilder().
			WithSTDIOTransport("echo", []string{"hello"})

		if builder == nil {
			t.Error("Expected non-nil builder")
		}

		client := builder.Build()
		if client == nil {
			t.Error("Expected non-nil client")
		}
	})

	t.Run("Create WebSocket client builder", func(t *testing.T) {
		builder := client.NewClientBuilder().
			WithWebSocketTransport("ws://localhost:8811")

		if builder == nil {
			t.Error("Expected non-nil builder")
		}

		client := builder.Build()
		if client == nil {
			t.Error("Expected non-nil client")
		}
	})
}

func TestClientConfig(t *testing.T) {
	t.Run("Client with custom config", func(t *testing.T) {
		transport := transport.NewTCPTransport("localhost", 8811)
		config := client.ClientConfig{
			Name:    "test-client",
			Version: "1.0.0",
		}

		c := client.NewClient(transport, config)
		if c == nil {
			t.Error("Expected non-nil client")
		}
	})
}
