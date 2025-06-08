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
	"github.com/kunalkushwaha/mcp-navigator-go/pkg/transport"
)

// Example demonstrates how to use the MCP client as a library
func main() {
	fmt.Println("MCP Client Library Example")
	fmt.Println("==========================")

	// 1. Create a client with TCP transport
	if err := exampleTCPClient(); err != nil {
		log.Printf("TCP client example failed: %v", err)
	}

	fmt.Println()

	// 2. Use server discovery
	if err := exampleServerDiscovery(); err != nil {
		log.Printf("Server discovery example failed: %v", err)
	}
}

// exampleTCPClient demonstrates basic client usage with TCP transport
func exampleTCPClient() error {
	fmt.Println("1. TCP Client Example")
	fmt.Println("---------------------")

	// Create TCP transport
	tcpTransport := transport.NewTCPTransport("localhost", 8811)

	// Configure client
	config := client.ClientConfig{
		Name:    "example-library-app",
		Version: "1.0.0",
		Logger:  log.Default(),
		Timeout: 30 * time.Second,
	}

	// Create client
	mcpClient := client.NewClient(tcpTransport, config)

	// Connect with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("📡 Connecting to MCP server...")
	if err := mcpClient.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer func() {
		fmt.Println("🔌 Disconnecting...")
		mcpClient.Disconnect()
	}()

	fmt.Println("✅ Connected successfully!")

	// Initialize protocol
	clientInfo := mcp.ClientInfo{
		Name:    "example-library-app",
		Version: "1.0.0",
	}

	fmt.Println("🚀 Initializing MCP protocol...")
	if err := mcpClient.Initialize(ctx, clientInfo); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	fmt.Println("✅ Protocol initialized!")

	// Get server information
	if serverInfo := mcpClient.GetServerInfo(); serverInfo != nil {
		fmt.Printf("🖥️  Server: %s v%s\n", serverInfo.Name, serverInfo.Version)
	}

	// List available tools
	fmt.Println("📋 Listing available tools...")
	tools, err := mcpClient.ListTools(ctx)
	if err != nil {
		return fmt.Errorf("failed to list tools: %w", err)
	}

	fmt.Printf("📝 Found %d tools:\n", len(tools))
	for i, tool := range tools {
		fmt.Printf("   %d. %s - %s\n", i+1, tool.Name, tool.Description)
	}

	// Call a tool if available
	if len(tools) > 0 {
		toolName := tools[0].Name
		fmt.Printf("🔧 Calling tool: %s\n", toolName)

		// Example arguments (adjust based on your server's tools)
		arguments := map[string]interface{}{
			"query": "example search from library",
		}

		result, err := mcpClient.CallTool(ctx, toolName, arguments)
		if err != nil {
			return fmt.Errorf("failed to call tool: %w", err)
		}

		fmt.Printf("✅ Tool result received (%d content items)\n", len(result.Content))
		if len(result.Content) > 0 {
			fmt.Printf("   First item: %+v\n", result.Content[0])
		}
	}

	return nil
}

// exampleServerDiscovery demonstrates server discovery functionality
func exampleServerDiscovery() error {
	fmt.Println("2. Server Discovery Example")
	fmt.Println("---------------------------")

	// Note: This would require importing the discovery package
	// For now, just demonstrate the concept
	fmt.Println("📡 Server discovery functionality available")
	fmt.Println("   - Scan TCP ports for MCP servers")
	fmt.Println("   - Discover Docker-based MCP servers")
	fmt.Println("   - Test server connectivity")
	fmt.Println("   - See pkg/discovery for full API")

	return nil
}
