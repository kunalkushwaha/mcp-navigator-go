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
	// Create Docker transport using STDIO with docker exec
	// This mimics how the discovery service creates Docker transports
	command := "docker"
	args := []string{"run", "--rm", "-i", "alpine/socat", "socat", "STDIO", "TCP-CONNECT:host.docker.internal:8811"}
	dockerTransport := transport.NewStdioTransport(command, args)

	// Create client
	clientConfig := client.ClientConfig{
		Name:    "test-client",
		Version: "1.0.0",
		Logger:  log.New(os.Stdout, "[TEST] ", log.LstdFlags),
		Timeout: 30 * time.Second,
	}

	mcpClient := client.NewClient(dockerTransport, clientConfig)

	ctx := context.Background()

	// Connect
	fmt.Println("Connecting...")
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

	// Try calling the docker tool first (which should be safer)
	fmt.Println("\nTrying to call 'docker' tool...")
	dockerArgs := map[string]interface{}{
		"args": []string{"--version"},
	}

	dockerResult, err := mcpClient.CallTool(ctx, "docker", dockerArgs)
	if err != nil {
		fmt.Printf("Docker tool failed: %v\n", err)
	} else {
		fmt.Printf("Docker tool success: %+v\n", dockerResult)
	}

	// Now try the search tool
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
