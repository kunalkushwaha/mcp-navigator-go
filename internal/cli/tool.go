package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/kunalkushwaha/mcp-navigator-go/pkg/client"
	"github.com/kunalkushwaha/mcp-navigator-go/pkg/mcp"
	"github.com/kunalkushwaha/mcp-navigator-go/pkg/transport"

	"github.com/spf13/cobra"
)

var (
	toolHost      string
	toolPort      int
	toolCommand   string
	toolArgs      []string
	toolType      string
	toolTimeout   time.Duration
	toolName      string
	toolArguments string
)

// toolCmd represents the tool command
var toolCmd = &cobra.Command{
	Use:   "tool",
	Short: "Execute a tool on an MCP server",
	Long: `Execute a specific tool on an MCP server.

This command connects to an MCP server, lists available tools, and executes
the specified tool with the provided arguments.

Examples:
  mcp-client tool --name search --args '{"query": "golang"}' --tcp --host localhost --port 8811
  mcp-client tool --name docker --args '{"command": "ps"}' --docker
  mcp-client tool --name fetch_content --args '{"url": "https://example.com"}' --type tcp`,
	Run: runTool,
}

func init() {
	rootCmd.AddCommand(toolCmd)

	// Connection flags (same as connect command)
	toolCmd.Flags().StringVar(&toolType, "type", "tcp", "Connection type: tcp, stdio, or docker")
	toolCmd.Flags().BoolP("tcp", "t", false, "Use TCP transport")
	toolCmd.Flags().BoolP("stdio", "s", false, "Use STDIO transport")
	toolCmd.Flags().BoolP("docker", "d", false, "Use Docker transport (alpine/socat)")

	toolCmd.Flags().StringVar(&toolHost, "host", "localhost", "TCP host to connect to")
	toolCmd.Flags().IntVar(&toolPort, "port", 8811, "TCP port to connect to")
	toolCmd.Flags().StringVar(&toolCommand, "command", "", "Command to execute for STDIO transport")
	toolCmd.Flags().StringSliceVar(&toolArgs, "args", []string{}, "Arguments for the command")
	toolCmd.Flags().DurationVar(&toolTimeout, "timeout", 30*time.Second, "Connection timeout")

	// Tool-specific flags
	toolCmd.Flags().StringVar(&toolName, "name", "", "Name of the tool to execute (required)")
	toolCmd.Flags().StringVar(&toolArguments, "arguments", "{}", "JSON arguments for the tool")

	// Mark required flags
	toolCmd.MarkFlagRequired("name")
}

func runTool(cmd *cobra.Command, args []string) {
	logger := log.New(os.Stdout, "", 0)
	if verbose {
		logger = log.New(os.Stdout, "[TOOL] ", log.LstdFlags)
	}

	// Determine transport type from flags
	tcpFlag, _ := cmd.Flags().GetBool("tcp")
	stdioFlag, _ := cmd.Flags().GetBool("stdio")
	dockerFlag, _ := cmd.Flags().GetBool("docker")

	transportType := toolType
	if tcpFlag {
		transportType = "tcp"
	} else if stdioFlag {
		transportType = "stdio"
	} else if dockerFlag {
		transportType = "docker"
	}

	fmt.Printf("🔌 Connecting to MCP server using %s transport...\n", transportType)

	// Create transport based on type
	var mcpTransport transport.Transport

	switch transportType {
	case "tcp":
		fmt.Printf("   Host: %s:%d\n", toolHost, toolPort)
		mcpTransport = transport.NewTCPTransport(toolHost, toolPort)

	case "stdio":
		if toolCommand == "" {
			fmt.Println("❌ STDIO transport requires --command flag")
			os.Exit(1)
		}
		fmt.Printf("   Command: %s %s\n", toolCommand, strings.Join(toolArgs, " "))
		mcpTransport = transport.NewStdioTransport(toolCommand, toolArgs)

	case "docker":
		fmt.Println("   Using Docker alpine/socat -> host.docker.internal:8811")
		dockerCommand := "docker"
		dockerArgs := []string{
			"run", "-i", "--rm", "alpine/socat",
			"STDIO", "TCP:host.docker.internal:8811",
		}
		mcpTransport = transport.NewStdioTransport(dockerCommand, dockerArgs)

	default:
		fmt.Printf("❌ Unsupported transport type: %s\n", transportType)
		os.Exit(1)
	}

	// Create client
	clientConfig := client.ClientConfig{
		Name:    "mcp-client-go",
		Version: "1.0.0",
		Logger:  logger,
		Timeout: toolTimeout,
	}

	mcpClient := client.NewClient(mcpTransport, clientConfig)

	ctx, cancel := context.WithTimeout(context.Background(), toolTimeout)
	defer cancel()

	// Connect to server
	if err := mcpClient.Connect(ctx); err != nil {
		fmt.Printf("❌ Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer mcpClient.Disconnect()

	// Initialize MCP protocol
	clientInfo := mcp.ClientInfo{
		Name:    "mcp-client-go",
		Version: "1.0.0",
	}

	if err := mcpClient.Initialize(ctx, clientInfo); err != nil {
		fmt.Printf("❌ Failed to initialize MCP protocol: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Connected and initialized MCP protocol")

	// Parse tool arguments
	var arguments map[string]interface{}
	if toolArguments != "" {
		if err := json.Unmarshal([]byte(toolArguments), &arguments); err != nil {
			fmt.Printf("❌ Invalid JSON arguments: %v\n", err)
			os.Exit(1)
		}
	}

	// Execute tool
	fmt.Printf("🔧 Executing tool: %s\n", toolName)
	if len(arguments) > 0 {
		fmt.Printf("📝 Arguments: %s\n", toolArguments)
	}

	result, err := mcpClient.CallTool(ctx, toolName, arguments)
	if err != nil {
		fmt.Printf("❌ Tool execution failed: %v\n", err)
		os.Exit(1)
	}

	// Display result
	fmt.Println("\n📤 Tool result:")
	if result.IsError {
		fmt.Println("⚠️  Tool returned an error:")
	}

	for i, content := range result.Content {
		if len(result.Content) > 1 {
			fmt.Printf("\n--- Content %d ---\n", i+1)
		}

		switch content.Type {
		case "text":
			fmt.Println(content.Text)
		case "image":
			fmt.Printf("Image content: %s\n", content.Data)
			if content.MimeType != "" {
				fmt.Printf("MIME type: %s\n", content.MimeType)
			}
		default:
			fmt.Printf("Content type: %s\n", content.Type)
			if content.Data != "" {
				fmt.Printf("Data: %s\n", content.Data)
			}
			if content.Text != "" {
				fmt.Printf("Text: %s\n", content.Text)
			}
		}
	}

	fmt.Println("\n✅ Tool execution completed")
}
