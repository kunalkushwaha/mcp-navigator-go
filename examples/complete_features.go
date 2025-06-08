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

// Example demonstrating complete MCP protocol support including Prompts and Resources
func main() {
	fmt.Println("MCP Client - Complete Features Example")
	fmt.Println("=====================================")

	// Create client with builder pattern
	mcpClient := client.NewClientBuilder().
		WithTCPTransport("localhost", 8811).
		WithName("complete-features-example").
		WithVersion("1.0.0").
		WithTimeout(30 * time.Second).
		Build()

	ctx := context.Background()

	// Connect and initialize
	if err := mcpClient.Connect(ctx); err != nil {
		log.Printf("Failed to connect: %v", err)
		return
	}
	defer mcpClient.Disconnect()

	clientInfo := mcp.ClientInfo{
		Name:    "complete-features-example",
		Version: "1.0.0",
	}

	if err := mcpClient.Initialize(ctx, clientInfo); err != nil {
		log.Printf("Failed to initialize: %v", err)
		return
	}

	// Demonstrate all MCP features
	demonstrateTools(ctx, mcpClient)
	fmt.Println()
	demonstrateResources(ctx, mcpClient)
	fmt.Println()
	demonstratePrompts(ctx, mcpClient)
}

// demonstrateTools shows tool listing and execution
func demonstrateTools(ctx context.Context, client *client.Client) {
	fmt.Println("1. Tools Support")
	fmt.Println("---------------")

	// List available tools
	tools, err := client.ListTools(ctx)
	if err != nil {
		log.Printf("Failed to list tools: %v", err)
		return
	}

	fmt.Printf("Found %d tools:\n", len(tools))
	for i, tool := range tools {
		fmt.Printf("  %d. %s - %s\n", i+1, tool.Name, tool.Description)
	}

	// Execute a tool if available
	if len(tools) > 0 {
		toolName := tools[0].Name
		fmt.Printf("\nExecuting tool '%s'...\n", toolName)

		result, err := client.CallTool(ctx, toolName, map[string]interface{}{})
		if err != nil {
			log.Printf("Failed to execute tool: %v", err)
		} else {
			fmt.Printf("Tool execution result: %+v\n", result)
		}
	}
}

// demonstrateResources shows resource listing and reading
func demonstrateResources(ctx context.Context, client *client.Client) {
	fmt.Println("2. Resources Support")
	fmt.Println("-------------------")

	// List available resources
	resources, err := client.ListResources(ctx)
	if err != nil {
		log.Printf("Failed to list resources: %v", err)
		return
	}

	fmt.Printf("Found %d resources:\n", len(resources))
	for i, resource := range resources {
		fmt.Printf("  %d. %s (%s) - %s\n", i+1, resource.Name, resource.URI, resource.Description)
	}

	// Read a resource if available
	if len(resources) > 0 {
		resourceURI := resources[0].URI
		fmt.Printf("\nReading resource '%s'...\n", resourceURI)

		content, err := client.ReadResource(ctx, resourceURI)
		if err != nil {
			log.Printf("Failed to read resource: %v", err)
		} else {
			fmt.Printf("Resource content (%d items):\n", len(content.Contents))
			for i, item := range content.Contents {
				fmt.Printf("  Item %d: Type=%s, Length=%d bytes\n", i+1, item.Type, len(item.Text))
			}
		}
	}
}

// demonstratePrompts shows prompt listing and execution
func demonstratePrompts(ctx context.Context, client *client.Client) {
	fmt.Println("3. Prompts Support")
	fmt.Println("-----------------")

	// List available prompts
	prompts, err := client.ListPrompts(ctx)
	if err != nil {
		log.Printf("Failed to list prompts: %v", err)
		return
	}

	fmt.Printf("Found %d prompts:\n", len(prompts))
	for i, prompt := range prompts {
		fmt.Printf("  %d. %s - %s\n", i+1, prompt.Name, prompt.Description)
		if len(prompt.Arguments) > 0 {
			fmt.Printf("     Arguments: ")
			for j, arg := range prompt.Arguments {
				if j > 0 {
					fmt.Print(", ")
				}
				fmt.Printf("%s", arg.Name)
				if arg.Required {
					fmt.Print("*")
				}
			}
			fmt.Println()
		}
	}

	// Get a prompt if available
	if len(prompts) > 0 {
		promptName := prompts[0].Name
		fmt.Printf("\nGetting prompt '%s'...\n", promptName)

		// Prepare arguments if needed
		arguments := make(map[string]interface{})
		for _, arg := range prompts[0].Arguments {
			if arg.Required {
				// Provide sample values for required arguments
				arguments[arg.Name] = fmt.Sprintf("sample_%s", arg.Name)
			}
		}

		promptResult, err := client.GetPrompt(ctx, promptName, arguments)
		if err != nil {
			log.Printf("Failed to get prompt: %v", err)
		} else {
			fmt.Printf("Prompt result (%d messages):\n", len(promptResult.Messages))
			for i, message := range promptResult.Messages {
				fmt.Printf("  Message %d: Role=%s, Content=%s\n", i+1, message.Role, message.Content.Text)
			}
		}
	}
}
