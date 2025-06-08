package cli

import (
	"context"
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
	connectHost    string
	connectPort    int
	connectCommand string
	connectArgs    []string
	connectType    string
	connectTimeout time.Duration
)

// connectCmd represents the connect command
var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to an MCP server",
	Long: `Connect to an MCP server and list its available tools.

This command can connect to MCP servers using different transport methods:
- TCP: Direct TCP connection to a server
- STDIO: Execute a command and communicate via stdin/stdout  
- Docker: Use Docker to run alpine/socat bridge to TCP server

Examples:
  mcp-client connect --tcp --host localhost --port 8811
  mcp-client connect --stdio --command node --args server.js
  mcp-client connect --docker  # Uses standard Docker MCP configuration
  mcp-client connect --type tcp --host 192.168.1.100 --port 8811`,
	Run: runConnect,
}

func init() {
	rootCmd.AddCommand(connectCmd)

	// Connection flags
	connectCmd.Flags().StringVar(&connectType, "type", "tcp", "Connection type: tcp, stdio, or docker")
	connectCmd.Flags().BoolP("tcp", "t", false, "Use TCP transport")
	connectCmd.Flags().BoolP("stdio", "s", false, "Use STDIO transport")
	connectCmd.Flags().BoolP("docker", "d", false, "Use Docker transport (alpine/socat)")

	connectCmd.Flags().StringVar(&connectHost, "host", "localhost", "TCP host to connect to")
	connectCmd.Flags().IntVar(&connectPort, "port", 8811, "TCP port to connect to")
	connectCmd.Flags().StringVar(&connectCommand, "command", "", "Command to execute for STDIO transport")
	connectCmd.Flags().StringSliceVar(&connectArgs, "args", []string{}, "Arguments for the command")
	connectCmd.Flags().DurationVar(&connectTimeout, "timeout", 30*time.Second, "Connection timeout")
}

func runConnect(cmd *cobra.Command, args []string) {
	logger := log.New(os.Stdout, "", 0)
	if verbose {
		logger = log.New(os.Stdout, "[MCP] ", log.LstdFlags)
	}

	// Determine transport type from flags
	tcpFlag, _ := cmd.Flags().GetBool("tcp")
	stdioFlag, _ := cmd.Flags().GetBool("stdio")
	dockerFlag, _ := cmd.Flags().GetBool("docker")

	transportType := connectType
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
	var err error

	switch transportType {
	case "tcp":
		fmt.Printf("   Host: %s:%d\n", connectHost, connectPort)
		mcpTransport = transport.NewTCPTransport(connectHost, connectPort)

	case "stdio":
		if connectCommand == "" {
			fmt.Println("❌ STDIO transport requires --command flag")
			os.Exit(1)
		}
		fmt.Printf("   Command: %s %s\n", connectCommand, strings.Join(connectArgs, " "))
		mcpTransport = transport.NewStdioTransport(connectCommand, connectArgs)

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
		Timeout: connectTimeout,
	}

	mcpClient := client.NewClient(mcpTransport, clientConfig)

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	// Connect to server
	if err := mcpClient.Connect(ctx); err != nil {
		fmt.Printf("❌ Failed to connect: %v\n", err)
		os.Exit(1)
	}

	// Initialize MCP protocol
	clientInfo := mcp.ClientInfo{
		Name:    "mcp-client-go",
		Version: "1.0.0",
	}

	if err := mcpClient.Initialize(ctx, clientInfo); err != nil {
		fmt.Printf("❌ Failed to initialize MCP protocol: %v\n", err)
		mcpClient.Disconnect()
		os.Exit(1)
	}

	fmt.Println("✅ Connected and initialized MCP protocol")

	// Get server info
	if serverInfo := mcpClient.GetServerInfo(); serverInfo != nil {
		fmt.Printf("🚀 Server: %s %s\n", serverInfo.Name, serverInfo.Version)
	}

	// List available tools
	fmt.Println("\n📋 Listing available tools...")
	tools, err := mcpClient.ListTools(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to list tools: %v\n", err)
	} else {
		if len(tools) == 0 {
			fmt.Println("   No tools available")
		} else {
			fmt.Printf("📝 Available tools (%d):\n", len(tools))
			for i, tool := range tools {
				fmt.Printf("  %d. %s\n", i+1, tool.Name)
				if tool.Description != "" {
					fmt.Printf("     Description: %s\n", tool.Description)
				}
			}
		}
	}

	// List available resources
	fmt.Println("\n📂 Listing available resources...")
	resources, err := mcpClient.ListResources(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to list resources: %v\n", err)
	} else {
		if len(resources) == 0 {
			fmt.Println("   No resources available")
		} else {
			fmt.Printf("📄 Available resources (%d):\n", len(resources))
			for i, resource := range resources {
				fmt.Printf("  %d. %s\n", i+1, resource.Name)
				if resource.Description != "" {
					fmt.Printf("     Description: %s\n", resource.Description)
				}
				if resource.URI != "" {
					fmt.Printf("     URI: %s\n", resource.URI)
				}
			}
		}
	}

	// Disconnect
	fmt.Println("\n🔌 Disconnecting...")
	if err := mcpClient.Disconnect(); err != nil {
		fmt.Printf("❌ Error during disconnect: %v\n", err)
	} else {
		fmt.Println("✅ Disconnected successfully")
	}
}
