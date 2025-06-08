package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kunalkushwaha/mcp-navigator-go/pkg/discovery"

	"github.com/spf13/cobra"
)

var (
	discoveryHost      string
	discoveryStartPort int
	discoveryEndPort   int
	discoveryTimeout   time.Duration
	includeTCP         bool
	includeDocker      bool
)

// discoverCmd represents the discover command
var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover available MCP servers",
	Long: `Discover MCP servers running on the local network.

This command scans for MCP servers using multiple discovery methods:
- TCP ports scanning for servers listening on common MCP ports
- Docker container inspection for MCP-related containers
- Standard Docker MCP configuration (alpine/socat)

Examples:
  mcp-client discover                    # Discover all servers
  mcp-client discover --host 192.168.1.1  # Scan specific host
  mcp-client discover --tcp-only         # Only scan TCP ports
  mcp-client discover --docker-only      # Only check Docker containers`,
	Run: runDiscover,
}

func init() {
	rootCmd.AddCommand(discoverCmd)

	// Discovery flags
	discoverCmd.Flags().StringVar(&discoveryHost, "host", "localhost", "Host to scan for TCP servers")
	discoverCmd.Flags().IntVar(&discoveryStartPort, "start-port", 8810, "Start port for range scanning")
	discoverCmd.Flags().IntVar(&discoveryEndPort, "end-port", 8820, "End port for range scanning")
	discoverCmd.Flags().DurationVar(&discoveryTimeout, "timeout", 5*time.Second, "Connection timeout for discovery")
	discoverCmd.Flags().BoolVar(&includeTCP, "tcp-only", false, "Only scan TCP ports")
	discoverCmd.Flags().BoolVar(&includeDocker, "docker-only", false, "Only check Docker containers")
}

func runDiscover(cmd *cobra.Command, args []string) {
	logger := log.New(os.Stdout, "", 0)
	if verbose {
		logger = log.New(os.Stdout, "[DISCOVERY] ", log.LstdFlags)
	}

	discoveryService := discovery.NewDiscovery(logger)
	discoveryService.SetTimeout(discoveryTimeout)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("🔍 Discovering MCP servers...")

	var servers []discovery.ServerInfo

	// Determine what to discover based on flags
	if includeTCP && includeDocker {
		fmt.Println("❌ Cannot use --tcp-only and --docker-only together")
		os.Exit(1)
	}

	if includeDocker {
		// Only Docker discovery
		servers = discoveryService.DiscoverDockerServers(ctx)

		// Add standard Docker MCP config
		dockerMCP := discovery.ServerInfo{
			Name:        "Docker MCP (alpine/socat)",
			Type:        "docker",
			Address:     "host.docker.internal",
			Port:        8811,
			Transport:   discoveryService.CreateDockerMCPTransport(),
			Description: "Standard Docker MCP server using alpine/socat",
		}
		servers = append(servers, dockerMCP)
	} else if includeTCP {
		// Only TCP discovery
		servers = discoveryService.ScanPortRange(ctx, discoveryHost, discoveryStartPort, discoveryEndPort)
	} else {
		// Comprehensive discovery
		servers = discoveryService.DiscoverAll(ctx, discoveryHost)
	}

	// Display results
	if len(servers) == 0 {
		fmt.Println("❌ No MCP servers discovered")
		return
	}

	fmt.Printf("✅ Found %d MCP server(s):\n\n", len(servers))

	for i, server := range servers {
		fmt.Printf("%d. %s (%s)\n", i+1, server.Name, server.Type)

		if server.Address != "" {
			if server.Port > 0 {
				fmt.Printf("   Address: %s:%d\n", server.Address, server.Port)
			} else {
				fmt.Printf("   Address: %s\n", server.Address)
			}
		}

		if server.Description != "" {
			fmt.Printf("   Description: %s\n", server.Description)
		}

		fmt.Println()
	}

	fmt.Printf("Discovery completed in %v\n", discoveryTimeout)
}
