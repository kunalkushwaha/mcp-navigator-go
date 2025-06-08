package tests

import (
	"testing"

	"github.com/kunalkushwaha/mcp-navigator-go/pkg/discovery"
)

func TestDiscovery(t *testing.T) {
	t.Run("Create discovery instance", func(t *testing.T) {
		disco := discovery.NewDiscovery(nil)

		if disco == nil {
			t.Error("Expected non-nil discovery instance")
		}
	})

	t.Run("Create server info", func(t *testing.T) {
		server := discovery.ServerInfo{
			Name:        "test-server",
			Type:        "tcp",
			Address:     "localhost",
			Port:        8811,
			Description: "Test MCP server",
		}

		if server.Name != "test-server" {
			t.Errorf("Expected server name 'test-server', got %s", server.Name)
		}

		if server.Type != "tcp" {
			t.Errorf("Expected server type 'tcp', got %s", server.Type)
		}

		if server.Port != 8811 {
			t.Errorf("Expected server port 8811, got %d", server.Port)
		}
	})
}
