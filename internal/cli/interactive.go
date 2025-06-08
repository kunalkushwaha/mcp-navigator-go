package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kunalkushwaha/mcp-navigator-go/pkg/client"
	"github.com/kunalkushwaha/mcp-navigator-go/pkg/discovery"
	"github.com/kunalkushwaha/mcp-navigator-go/pkg/mcp"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// interactiveCmd represents the interactive command
var interactiveCmd = &cobra.Command{
	Use:     "interactive",
	Short:   "Start interactive mode",
	Aliases: []string{"i", "shell"},
	Long: `Start interactive mode for the MCP client.

Interactive mode provides a command-line interface for discovering, connecting to,
and interacting with MCP servers. Available commands:

  help                    - Show available commands
  discover                - Discover available MCP servers
  connect <name|index>    - Connect to a server by name or index
  disconnect              - Disconnect from current server
  list-tools              - List tools available on current server
  list-resources          - List resources available on current server
  call-tool <name> [args] - Execute a tool with optional JSON arguments
  status                  - Show connection status
  exit/quit               - Exit interactive mode

Examples:
  > discover
  > connect 1
  > list-tools
  > call-tool search {"query": "golang"}
  > exit`,
	Run: runInteractive,
}

func init() {
	rootCmd.AddCommand(interactiveCmd)
}

type InteractiveSession struct {
	logger           *log.Logger
	discoveryService *discovery.Discovery
	availableServers []discovery.ServerInfo
	currentClient    *client.Client
	currentServer    string
	reader           *bufio.Reader

	// Colors for output
	promptColor  *color.Color
	successColor *color.Color
	errorColor   *color.Color
	infoColor    *color.Color
}

func runInteractive(cmd *cobra.Command, args []string) {
	session := &InteractiveSession{
		logger:           log.New(os.Stdout, "", 0),
		discoveryService: discovery.NewDiscovery(nil),
		reader:           bufio.NewReader(os.Stdin),
		promptColor:      color.New(color.FgCyan, color.Bold),
		successColor:     color.New(color.FgGreen),
		errorColor:       color.New(color.FgRed),
		infoColor:        color.New(color.FgBlue),
	}

	if verbose {
		session.logger = log.New(os.Stdout, "[INTERACTIVE] ", log.LstdFlags)
	}

	session.start()
}

func (s *InteractiveSession) start() {
	s.successColor.Println("🚀 MCP Client Interactive Mode")
	s.infoColor.Println("Type 'help' for available commands.")

	// Auto-discover servers on startup
	s.discoverServers()
	// Main command loop
	for {
		s.promptColor.Print("\nmcp-client> ")
		input, err := s.reader.ReadString('\n')
		if err != nil {
			// Handle EOF (Ctrl+D) and other input errors gracefully
			if err == io.EOF {
				s.infoColor.Println("\n👋 Exiting interactive mode...")
				s.exit()
				return
			}
			s.errorColor.Printf("Error reading input: %v\n", err)
			s.infoColor.Println("Exiting due to input error...")
			s.exit()
			return
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		parts := strings.Fields(input)
		command := parts[0]
		args := parts[1:]

		switch command {
		case "help", "h":
			s.showHelp()
		case "discover", "d":
			s.discoverServers()
		case "connect", "c":
			s.connectToServer(args)
		case "disconnect", "dc":
			s.disconnectFromServer()
		case "list-tools", "lt":
			s.listTools()
		case "list-resources", "lr":
			s.listResources()
		case "call-tool", "ct":
			s.callTool(args)
		case "status", "s":
			s.showStatus()
		case "exit", "quit", "q":
			s.exit()
			return
		default:
			s.errorColor.Printf("Unknown command: %s. Type 'help' for available commands.\n", command)
		}
	}
}

func (s *InteractiveSession) showHelp() {
	fmt.Println("\n📋 Available Commands:")
	fmt.Println("  help              - Show this help message")
	fmt.Println("  discover          - Discover available MCP servers")
	fmt.Println("  connect <n>       - Connect to a server by name or index")
	fmt.Println("  disconnect        - Disconnect from current server")
	fmt.Println("  list-tools        - List tools available on current server")
	fmt.Println("  list-resources    - List resources available on current server")
	fmt.Println("  call-tool <n> [args] - Call a tool with optional JSON arguments")
	fmt.Println("  status            - Show connection status")
	fmt.Println("  exit/quit         - Exit the client")
}

func (s *InteractiveSession) discoverServers() {
	s.infoColor.Println("🔍 Discovering MCP servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.availableServers = s.discoveryService.DiscoverAll(ctx, "localhost")

	if len(s.availableServers) == 0 {
		s.errorColor.Println("❌ No MCP servers discovered")
		return
	}

	s.successColor.Printf("✅ Found %d server(s):\n", len(s.availableServers))
	for i, server := range s.availableServers {
		fmt.Printf("  %d. %s (%s)\n", i+1, server.Name, server.Type)
		if server.Address != "" && server.Port > 0 {
			fmt.Printf("     Address: %s:%d\n", server.Address, server.Port)
		}
	}
}

func (s *InteractiveSession) connectToServer(args []string) {
	if len(args) == 0 {
		s.errorColor.Println("❌ Please specify a server name or index")
		return
	}

	if len(s.availableServers) == 0 {
		s.errorColor.Println("❌ No servers available. Run 'discover' first.")
		return
	}

	// Parse server selection
	var selectedServer discovery.ServerInfo
	var found bool

	// Try to parse as index
	if index, err := strconv.Atoi(args[0]); err == nil {
		if index > 0 && index <= len(s.availableServers) {
			selectedServer = s.availableServers[index-1]
			found = true
		}
	} else {
		// Try to find by name
		for _, server := range s.availableServers {
			if strings.Contains(strings.ToLower(server.Name), strings.ToLower(args[0])) {
				selectedServer = server
				found = true
				break
			}
		}
	}

	if !found {
		s.errorColor.Println("❌ Server not found")
		return
	}

	// Disconnect from current server if any
	if s.currentClient != nil {
		s.disconnectFromServer()
	}

	s.infoColor.Printf("🔌 Connecting to %s...\n", selectedServer.Name)

	// Create client
	clientConfig := client.ClientConfig{
		Name:    "mcp-client-go",
		Version: "1.0.0",
		Logger:  s.logger,
		Timeout: 30 * time.Second,
	}

	s.currentClient = client.NewClient(selectedServer.Transport, clientConfig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect
	if err := s.currentClient.Connect(ctx); err != nil {
		s.errorColor.Printf("❌ Failed to connect: %v\n", err)
		s.currentClient = nil
		return
	}

	// Initialize MCP protocol
	clientInfo := mcp.ClientInfo{
		Name:    "mcp-client-go",
		Version: "1.0.0",
	}

	if err := s.currentClient.Initialize(ctx, clientInfo); err != nil {
		s.errorColor.Printf("❌ Failed to initialize MCP protocol: %v\n", err)
		s.currentClient.Disconnect()
		s.currentClient = nil
		return
	}

	s.currentServer = selectedServer.Name
	s.successColor.Printf("✅ Connected to %s\n", selectedServer.Name)

	// Show server info
	if serverInfo := s.currentClient.GetServerInfo(); serverInfo != nil {
		s.infoColor.Printf("🚀 Server: %s %s\n", serverInfo.Name, serverInfo.Version)
	}
}

func (s *InteractiveSession) disconnectFromServer() {
	if s.currentClient == nil {
		s.errorColor.Println("❌ No active connection")
		return
	}

	if err := s.currentClient.Disconnect(); err != nil {
		s.errorColor.Printf("❌ Error during disconnect: %v\n", err)
	} else {
		s.successColor.Printf("🔌 Disconnected from %s\n", s.currentServer)
	}

	s.currentClient = nil
	s.currentServer = ""
}

func (s *InteractiveSession) listTools() {
	if s.currentClient == nil {
		s.errorColor.Println("❌ No active connection. Use 'connect' first.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tools, err := s.currentClient.ListTools(ctx)
	if err != nil {
		s.errorColor.Printf("❌ Failed to list tools: %v\n", err)
		return
	}

	if len(tools) == 0 {
		s.infoColor.Println("📝 No tools available")
		return
	}

	s.successColor.Printf("📝 Available tools (%d):\n", len(tools))
	for i, tool := range tools {
		fmt.Printf("  %d. %s\n", i+1, tool.Name)
		if tool.Description != "" {
			fmt.Printf("     Description: %s\n", tool.Description)
		}
	}
}

func (s *InteractiveSession) listResources() {
	if s.currentClient == nil {
		s.errorColor.Println("❌ No active connection. Use 'connect' first.")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resources, err := s.currentClient.ListResources(ctx)
	if err != nil {
		s.errorColor.Printf("❌ Failed to list resources: %v\n", err)
		return
	}

	if len(resources) == 0 {
		s.infoColor.Println("📂 No resources available")
		return
	}

	s.successColor.Printf("📄 Available resources (%d):\n", len(resources))
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

func (s *InteractiveSession) callTool(args []string) {
	if s.currentClient == nil {
		s.errorColor.Println("❌ No active connection. Use 'connect' first.")
		return
	}

	if len(args) == 0 {
		s.errorColor.Println("❌ Please specify a tool name")
		return
	}

	toolName := args[0]
	var arguments map[string]interface{}

	// Parse JSON arguments if provided
	if len(args) > 1 {
		argsStr := strings.Join(args[1:], " ")
		if err := json.Unmarshal([]byte(argsStr), &arguments); err != nil {
			s.errorColor.Printf("❌ Invalid JSON arguments: %v\n", err)
			return
		}
	}

	s.infoColor.Printf("🔧 Calling tool: %s\n", toolName)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := s.currentClient.CallTool(ctx, toolName, arguments)
	if err != nil {
		s.errorColor.Printf("❌ Tool execution failed: %v\n", err)
		return
	}

	s.successColor.Println("📤 Tool result:")
	if result.IsError {
		s.errorColor.Println("⚠️  Tool returned an error:")
	}

	for _, content := range result.Content {
		switch content.Type {
		case "text":
			fmt.Println(content.Text)
		default:
			fmt.Printf("Content type: %s\n", content.Type)
			if content.Data != "" {
				fmt.Printf("Data: %s\n", content.Data)
			}
		}
	}
}

func (s *InteractiveSession) showStatus() {
	fmt.Println("\n📊 Status:")
	fmt.Printf("  Available servers: %d\n", len(s.availableServers))

	if s.currentClient != nil {
		fmt.Printf("  Current connection: %s ✅\n", s.currentServer)
		if serverInfo := s.currentClient.GetServerInfo(); serverInfo != nil {
			fmt.Printf("  Server info: %s %s\n", serverInfo.Name, serverInfo.Version)
		}
	} else {
		fmt.Println("  Current connection: None ❌")
	}
}

func (s *InteractiveSession) exit() {
	s.infoColor.Println("\n👋 Shutting down MCP client...")

	if s.currentClient != nil {
		s.disconnectFromServer()
	}

	s.successColor.Println("✅ Goodbye!")
}
