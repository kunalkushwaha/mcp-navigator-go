package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kunalkushwaha/mcp-navigator-go/pkg/client"
	"github.com/kunalkushwaha/mcp-navigator-go/pkg/mcp"
	"github.com/kunalkushwaha/mcp-navigator-go/pkg/transport"
)

func main() {
	// Create TCP transport directly (instead of Docker proxy)
	tcpTransport := transport.NewTCPTransport("localhost", 8811)

	// Create client
	clientConfig := client.ClientConfig{
		Name:    "test-client",
		Version: "1.0.0",
		Logger:  log.New(os.Stdout, "[TEST] ", log.LstdFlags),
		Timeout: 30 * time.Second,
	}

	mcpClient := client.NewClient(tcpTransport, clientConfig)

	ctx := context.Background()

	// Connect
	fmt.Println("Connecting directly to TCP...")
	if err := mcpClient.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer mcpClient.Disconnect()

	// Initialize
	fmt.Println("Initializing...")
	clientInfo := mcp.ClientInfo{
		Name:    "test-client",
		Version: "1.0.0",
	}

	if err := mcpClient.Initialize(ctx, clientInfo); err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}

	// List tools
	fmt.Println("Listing tools...")
	tools, err := mcpClient.ListTools(ctx)
	if err != nil {
		log.Fatalf("Failed to list tools: %v", err)
	}

	fmt.Printf("Found %d tools:\n", len(tools))
	for i, tool := range tools {
		fmt.Printf("  %d. %s: %s\n", i+1, tool.Name, tool.Description)
	}

	// Try calling the search tool
	fmt.Println("\nTrying to call 'search' tool...")
	searchArgs := map[string]interface{}{
		"query":       "hello world",
		"max_results": 3,
	}

	searchResult, err := mcpClient.CallTool(ctx, "search", searchArgs)
	if err != nil {
		fmt.Printf("Search tool failed: %v\n", err)
	} else {
		fmt.Printf("Search tool success: %+v\n", searchResult)
	}
}
