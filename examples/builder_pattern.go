//go:build example
// +build example

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kunalkushwaha/mcp-navigator-go/pkg/client"
	"github.com/kunalkushwaha/mcp-navigator-go/pkg/mcp"
)

// Example showing different ways to create and use the MCP client as a library
func main() {
	fmt.Println("MCP Client Library - Builder Pattern Examples")
	fmt.Println("===========================================")

	// Example 1: Simple TCP client
	simpleExample()

	fmt.Println()

	// Example 2: Builder pattern with customization
	builderExample()

	fmt.Println()

	// Example 3: Service wrapper pattern
	serviceExample()
}

// simpleExample shows the simplest way to create a client
func simpleExample() {
	fmt.Println("1. Simple TCP Client")
	fmt.Println("-------------------")

	// One-liner client creation
	mcpClient := client.NewTCPClient("localhost", 8811)

	ctx := context.Background()

	fmt.Println("📡 Connecting...")
	if err := mcpClient.Connect(ctx); err != nil {
		fmt.Printf("❌ Connection failed: %v\n", err)
		return
	}
	defer mcpClient.Disconnect()

	clientInfo := mcp.ClientInfo{Name: "simple-app", Version: "1.0.0"}
	if err := mcpClient.Initialize(ctx, clientInfo); err != nil {
		fmt.Printf("❌ Initialization failed: %v\n", err)
		return
	}

	fmt.Println("✅ Connected and initialized!")

	// Quick tool listing
	tools, err := mcpClient.ListTools(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to list tools: %v\n", err)
		return
	}

	fmt.Printf("📝 Found %d tools\n", len(tools))
}

// builderExample shows the fluent builder pattern
func builderExample() {
	fmt.Println("2. Builder Pattern with Customization")
	fmt.Println("-------------------------------------")

	// Fluent builder pattern
	mcpClient := client.NewClientBuilder().
		WithTCPTransport("localhost", 8811).
		WithName("builder-app").
		WithVersion("2.0.0").
		WithTimeout(45 * time.Second).
		WithLogger(log.Default()).
		Build()

	ctx := context.Background()

	fmt.Println("📡 Connecting with custom configuration...")
	if err := mcpClient.Connect(ctx); err != nil {
		fmt.Printf("❌ Connection failed: %v\n", err)
		return
	}
	defer mcpClient.Disconnect()

	clientInfo := mcp.ClientInfo{Name: "builder-app", Version: "2.0.0"}
	if err := mcpClient.Initialize(ctx, clientInfo); err != nil {
		fmt.Printf("❌ Initialization failed: %v\n", err)
		return
	}

	fmt.Println("✅ Connected with custom settings!")

	// Show server info
	if serverInfo := mcpClient.GetServerInfo(); serverInfo != nil {
		fmt.Printf("🖥️  Connected to: %s v%s\n", serverInfo.Name, serverInfo.Version)
	}
}

// serviceExample shows a service wrapper pattern
func serviceExample() {
	fmt.Println("3. Service Wrapper Pattern")
	fmt.Println("--------------------------")

	service := NewMCPService()

	fmt.Println("📡 Service connecting...")
	if err := service.Connect("localhost", 8811); err != nil {
		fmt.Printf("❌ Service connection failed: %v\n", err)
		return
	}
	defer service.Disconnect()

	fmt.Println("✅ Service connected!")

	// Use the service
	tools, err := service.ListTools()
	if err != nil {
		fmt.Printf("❌ Failed to list tools: %v\n", err)
		return
	}

	fmt.Printf("📝 Service found %d tools\n", len(tools))

	// Execute a tool through the service
	if len(tools) > 0 {
		fmt.Printf("🔧 Executing tool: %s\n", tools[0].Name)
		result, err := service.ExecuteTool(tools[0].Name, map[string]interface{}{
			"query": "test from service",
		})
		if err != nil {
			fmt.Printf("❌ Tool execution failed: %v\n", err)
		} else {
			fmt.Printf("✅ Tool executed successfully (%d content items)\n", len(result.Content))
		}
	}
}

// MCPService demonstrates a service wrapper pattern
type MCPService struct {
	client    *client.Client
	connected bool
}

// NewMCPService creates a new MCP service wrapper
func NewMCPService() *MCPService {
	return &MCPService{}
}

// Connect establishes connection to an MCP server
func (s *MCPService) Connect(host string, port int) error {
	s.client = client.NewClientBuilder().
		WithTCPTransport(host, port).
		WithName("mcp-service").
		WithVersion("1.0.0").
		Build()

	ctx := context.Background()

	if err := s.client.Connect(ctx); err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	clientInfo := mcp.ClientInfo{Name: "mcp-service", Version: "1.0.0"}
	if err := s.client.Initialize(ctx, clientInfo); err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}

	s.connected = true
	return nil
}

// Disconnect closes the connection
func (s *MCPService) Disconnect() error {
	if s.client != nil {
		s.connected = false
		return s.client.Disconnect()
	}
	return nil
}

// ListTools returns available tools
func (s *MCPService) ListTools() ([]mcp.Tool, error) {
	if !s.connected {
		return nil, fmt.Errorf("service not connected")
	}

	ctx := context.Background()
	return s.client.ListTools(ctx)
}

// ExecuteTool executes a tool with the given arguments
func (s *MCPService) ExecuteTool(name string, args map[string]interface{}) (*mcp.CallToolResponse, error) {
	if !s.connected {
		return nil, fmt.Errorf("service not connected")
	}

	ctx := context.Background()
	return s.client.CallTool(ctx, name, args)
}

// IsConnected returns the connection status
func (s *MCPService) IsConnected() bool {
	return s.connected && s.client != nil && s.client.IsConnected()
}
