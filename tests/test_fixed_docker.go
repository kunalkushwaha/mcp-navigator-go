package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kunalkushwaha/mcp-navigator-go/pkg/client"
	"github.com/kunalkushwaha/mcp-navigator-go/pkg/discovery"
	"github.com/kunalkushwaha/mcp-navigator-go/pkg/mcp"
)

func main() {
	// Test the fixed Docker transport from discovery service
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	disc := discovery.NewDiscovery(logger)

	// Get the Docker MCP transport (which should now use direct TCP)
	dockerTransport := disc.CreateDockerMCPTransport()

	// Create client
	clientConfig := client.ClientConfig{
		Name:    "test-client",
		Version: "1.0.0",
		Logger:  logger,
		Timeout: 30 * time.Second,
	}

	mcpClient := client.NewClient(dockerTransport, clientConfig)

	ctx := context.Background()

	// Connect
	fmt.Println("Connecting with fixed Docker transport...")
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

	// Try calling the search tool (this was failing before)
	fmt.Println("\nTrying to call 'search' tool with fixed Docker transport...")
	searchArgs := map[string]interface{}{
		"query":       "golang testing",
		"max_results": 2,
	}

	searchResult, err := mcpClient.CallTool(ctx, "search", searchArgs)
	if err != nil {
		fmt.Printf("Search tool failed: %v\n", err)
	} else {
		fmt.Printf("✅ Search tool SUCCESS with fixed Docker transport!\n")
		fmt.Printf("Result: %+v\n", searchResult)
	}
}
