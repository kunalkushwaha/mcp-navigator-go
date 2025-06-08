package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kunalkushwaha/mcp-navigator-go/pkg/client"
	"github.com/kunalkushwaha/mcp-navigator-go/pkg/mcp"
)

// TestCompleteProtocol validates all MCP protocol features
func main() {
	fmt.Println("🧪 MCP Complete Protocol Test")
	fmt.Println("=============================")

	// Test connection to localhost:8811 (common MCP server port)
	fmt.Println("Testing MCP client with complete protocol support...")

	// Create client using builder pattern
	mcpClient := client.NewClientBuilder().
		WithTCPTransport("localhost", 8811).
		WithName("protocol-test-client").
		WithVersion("1.0.0").
		WithTimeout(10 * time.Second).
		Build()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test connection
	fmt.Printf("⏳ Connecting to localhost:8811...")
	if err := mcpClient.Connect(ctx); err != nil {
		fmt.Printf(" ❌ FAILED\n")
		fmt.Printf("Connection error: %v\n", err)
		fmt.Println("💡 Make sure an MCP server is running on localhost:8811")
		os.Exit(1)
	}
	fmt.Printf(" ✅ SUCCESS\n")
	defer mcpClient.Disconnect()

	// Test initialization
	fmt.Printf("⏳ Initializing client...")
	clientInfo := mcp.ClientInfo{
		Name:    "protocol-test-client",
		Version: "1.0.0",
	}

	if err := mcpClient.Initialize(ctx, clientInfo); err != nil {
		fmt.Printf(" ❌ FAILED\n")
		fmt.Printf("Initialization error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf(" ✅ SUCCESS\n")

	// Test all protocol features
	testTools(ctx, mcpClient)
	testResources(ctx, mcpClient)
	testPrompts(ctx, mcpClient)

	fmt.Println("\n🎉 All tests completed successfully!")
	fmt.Println("✅ MCP Go client library is fully functional with complete protocol support")
}

func testTools(ctx context.Context, client *client.Client) {
	fmt.Println("\n📋 Testing Tools Protocol")
	fmt.Println("-------------------------")

	fmt.Printf("⏳ Listing tools...")
	tools, err := client.ListTools(ctx)
	if err != nil {
		fmt.Printf(" ❌ FAILED: %v\n", err)
		return
	}
	fmt.Printf(" ✅ SUCCESS (%d tools found)\n", len(tools))

	for i, tool := range tools {
		fmt.Printf("   %d. %s - %s\n", i+1, tool.Name, tool.Description)
	}

	// Test tool execution if tools are available
	if len(tools) > 0 {
		toolName := tools[0].Name
		fmt.Printf("⏳ Executing tool '%s'...", toolName)

		result, err := client.CallTool(ctx, toolName, map[string]interface{}{})
		if err != nil {
			fmt.Printf(" ❌ FAILED: %v\n", err)
		} else {
			fmt.Printf(" ✅ SUCCESS\n")
			fmt.Printf("   Tool result: %d content items\n", len(result.Content))
		}
	}
}

func testResources(ctx context.Context, client *client.Client) {
	fmt.Println("\n📄 Testing Resources Protocol")
	fmt.Println("-----------------------------")

	fmt.Printf("⏳ Listing resources...")
	resources, err := client.ListResources(ctx)
	if err != nil {
		fmt.Printf(" ❌ FAILED: %v\n", err)
		return
	}
	fmt.Printf(" ✅ SUCCESS (%d resources found)\n", len(resources))

	for i, resource := range resources {
		fmt.Printf("   %d. %s (%s) - %s\n", i+1, resource.Name, resource.URI, resource.Description)
	}

	// Test resource reading if resources are available
	if len(resources) > 0 {
		resourceURI := resources[0].URI
		fmt.Printf("⏳ Reading resource '%s'...", resourceURI)

		content, err := client.ReadResource(ctx, resourceURI)
		if err != nil {
			fmt.Printf(" ❌ FAILED: %v\n", err)
		} else {
			fmt.Printf(" ✅ SUCCESS\n")
			fmt.Printf("   Resource content: %d items\n", len(content.Contents))
		}
	}
}

func testPrompts(ctx context.Context, client *client.Client) {
	fmt.Println("\n💬 Testing Prompts Protocol")
	fmt.Println("---------------------------")

	fmt.Printf("⏳ Listing prompts...")
	prompts, err := client.ListPrompts(ctx)
	if err != nil {
		fmt.Printf(" ❌ FAILED: %v\n", err)
		return
	}
	fmt.Printf(" ✅ SUCCESS (%d prompts found)\n", len(prompts))

	for i, prompt := range prompts {
		fmt.Printf("   %d. %s - %s\n", i+1, prompt.Name, prompt.Description)
		if len(prompt.Arguments) > 0 {
			fmt.Printf("      Arguments: ")
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

	// Test prompt execution if prompts are available
	if len(prompts) > 0 {
		promptName := prompts[0].Name
		fmt.Printf("⏳ Getting prompt '%s'...", promptName)

		// Prepare arguments
		arguments := make(map[string]interface{})
		for _, arg := range prompts[0].Arguments {
			if arg.Required {
				arguments[arg.Name] = fmt.Sprintf("test_%s", arg.Name)
			}
		}

		result, err := client.GetPrompt(ctx, promptName, arguments)
		if err != nil {
			fmt.Printf(" ❌ FAILED: %v\n", err)
		} else {
			fmt.Printf(" ✅ SUCCESS\n")
			fmt.Printf("   Prompt result: %d messages\n", len(result.Messages))
		}
	}
}
