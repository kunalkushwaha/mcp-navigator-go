// Package discovery provides automatic MCP server discovery capabilities.
//
// This package can discover MCP servers through multiple methods:
//   - TCP port scanning for servers listening on common MCP ports
//   - Docker container inspection for MCP-related containers
//   - Direct connection testing to validate discovered servers
//
// Basic usage:
//
//	disco := discovery.NewDiscovery(logger)
//	servers := disco.DiscoverAll(ctx, "localhost")
//
//	for _, server := range servers {
//		if disco.TestConnection(ctx, server) {
//			// Use server.Transport to create MCP client
//			client := client.NewClient(server.Transport, config)
//			// ... connect and use client
//		}
//	}
//
// The discovery system is particularly useful for applications that need to
// automatically find and connect to available MCP servers without manual configuration.
package discovery

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/kunalkushwaha/mcp-navigator-go/pkg/transport"
)

// ServerInfo represents information about a discovered server.
//
// This structure contains all the information needed to connect to a discovered
// MCP server, including the pre-configured transport for immediate use.
type ServerInfo struct {
	Name        string              // Human-readable name for the server
	Type        string              // "tcp", "docker", "process"
	Address     string              // Server address (hostname, container ID, etc.)
	Port        int                 // Port number (0 for non-TCP transports)
	Transport   transport.Transport // Ready-to-use transport for this server
	Description string              // Detailed description of the server
}

// Discovery handles MCP server discovery
type Discovery struct {
	logger  *log.Logger
	timeout time.Duration
}

// NewDiscovery creates a new server discovery instance.
//
// If logger is nil, log.Default() will be used.
// The discovery instance uses a default timeout of 5 seconds for connection tests.
//
// Example:
//
//	disco := NewDiscovery(log.Default())
//	disco.SetTimeout(10 * time.Second) // Optional: custom timeout
func NewDiscovery(logger *log.Logger) *Discovery {
	if logger == nil {
		logger = log.Default()
	}
	return &Discovery{
		logger:  logger,
		timeout: 5 * time.Second,
	}
}

// DiscoverTCPServers scans for MCP servers on TCP ports
func (d *Discovery) DiscoverTCPServers(ctx context.Context, host string, ports []int) []ServerInfo {
	d.logger.Printf("Scanning for MCP servers on %s, ports: %v", host, ports)

	var servers []ServerInfo

	for _, port := range ports {
		if d.isPortOpen(host, port) {
			server := ServerInfo{
				Name:        fmt.Sprintf("TCP Server %s:%d", host, port),
				Type:        "tcp",
				Address:     host,
				Port:        port,
				Transport:   transport.NewTCPTransport(host, port),
				Description: fmt.Sprintf("MCP server on TCP %s:%d", host, port),
			}
			servers = append(servers, server)
			d.logger.Printf("Found TCP server: %s:%d", host, port)
		}
	}

	d.logger.Printf("TCP discovery complete. Found %d servers", len(servers))
	return servers
}

// DiscoverDockerServers scans for MCP servers in Docker containers
func (d *Discovery) DiscoverDockerServers(ctx context.Context) []ServerInfo {
	d.logger.Println("Scanning for MCP servers in Docker containers...")

	var servers []ServerInfo

	// Check if Docker is available
	if !d.isDockerAvailable() {
		d.logger.Println("Docker not available, skipping Docker discovery")
		return servers
	}

	// Look for containers with MCP-related labels or names
	containers := d.getDockerContainers()

	for _, container := range containers {
		if d.isMCPContainer(container) {
			server := ServerInfo{
				Name:        fmt.Sprintf("Docker Container %s", container.Name),
				Type:        "docker",
				Address:     container.ID,
				Port:        0,
				Transport:   d.createDockerTransport(container),
				Description: fmt.Sprintf("MCP server in Docker container %s", container.Name),
			}
			servers = append(servers, server)
			d.logger.Printf("Found Docker MCP server: %s", container.Name)
		}
	}

	d.logger.Printf("Docker discovery complete. Found %d servers", len(servers))
	return servers
}

// CreateDockerMCPTransport creates a transport for the Docker MCP configuration
func (d *Discovery) CreateDockerMCPTransport() transport.Transport {
	d.logger.Println("Creating Docker MCP transport with direct TCP connection")

	// Instead of using alpine/socat proxy (which fails during tool calls),
	// create a direct TCP connection to localhost:8811
	// This assumes the MCP server is accessible on the host at localhost:8811
	return transport.NewTCPTransport("localhost", 8811)
}

// DiscoverCommonPorts scans commonly used MCP ports
func (d *Discovery) DiscoverCommonPorts(ctx context.Context, host string) []ServerInfo {
	commonPorts := []int{
		8811, // Common MCP port
		8080, // HTTP alternative
		3000, // Development server
		4000, // Development server
		5000, // Development server
		8000, // HTTP alternative
		8888, // Alternative
		9000, // Alternative
	}

	return d.DiscoverTCPServers(ctx, host, commonPorts)
}

// DiscoverAll performs comprehensive server discovery
func (d *Discovery) DiscoverAll(ctx context.Context, host string) []ServerInfo {
	d.logger.Println("Starting comprehensive MCP server discovery...")

	var allServers []ServerInfo

	// Discover TCP servers on common ports
	tcpServers := d.DiscoverCommonPorts(ctx, host)
	allServers = append(allServers, tcpServers...)

	// Discover Docker servers
	dockerServers := d.DiscoverDockerServers(ctx)
	allServers = append(allServers, dockerServers...)
	// Add the standard Docker MCP configuration
	dockerMCP := ServerInfo{
		Name:        "Docker MCP (Direct TCP)",
		Type:        "docker",
		Address:     "localhost",
		Port:        8811,
		Transport:   d.CreateDockerMCPTransport(),
		Description: "Standard Docker MCP server using direct TCP connection to localhost:8811",
	}
	allServers = append(allServers, dockerMCP)

	d.logger.Printf("Discovery complete. Found %d total servers", len(allServers))
	return allServers
}

// isPortOpen checks if a TCP port is open
func (d *Discovery) isPortOpen(host string, port int) bool {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, d.timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// isDockerAvailable checks if Docker is available
func (d *Discovery) isDockerAvailable() bool {
	cmd := exec.Command("docker", "version")
	err := cmd.Run()
	return err == nil
}

// DockerContainer represents a Docker container
type DockerContainer struct {
	ID    string
	Name  string
	Image string
	Ports []string
}

// getDockerContainers lists running Docker containers
func (d *Discovery) getDockerContainers() []DockerContainer {
	cmd := exec.Command("docker", "ps", "--format", "{{.ID}}\t{{.Names}}\t{{.Image}}\t{{.Ports}}")
	output, err := cmd.Output()
	if err != nil {
		d.logger.Printf("Failed to list Docker containers: %v", err)
		return nil
	}

	var containers []DockerContainer
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) >= 4 {
			container := DockerContainer{
				ID:    parts[0],
				Name:  parts[1],
				Image: parts[2],
				Ports: strings.Split(parts[3], ","),
			}
			containers = append(containers, container)
		}
	}

	return containers
}

// isMCPContainer checks if a container is likely an MCP server
func (d *Discovery) isMCPContainer(container DockerContainer) bool {
	// Check for MCP-related keywords in name or image
	mcpKeywords := []string{"mcp", "model-context-protocol", "context"}

	name := strings.ToLower(container.Name)
	image := strings.ToLower(container.Image)

	for _, keyword := range mcpKeywords {
		if strings.Contains(name, keyword) || strings.Contains(image, keyword) {
			return true
		}
	}

	// Check for specific ports that might indicate MCP servers
	for _, portStr := range container.Ports {
		if strings.Contains(portStr, "8811") || strings.Contains(portStr, "3000") {
			return true
		}
	}

	return false
}

// createDockerTransport creates a transport for a Docker container
func (d *Discovery) createDockerTransport(container DockerContainer) transport.Transport {
	// For now, create a generic Docker exec transport
	// This could be enhanced to detect the specific transport needed
	command := "docker"
	args := []string{"exec", "-i", container.ID, "sh"}

	return transport.NewStdioTransport(command, args)
}

// SetTimeout sets the connection timeout for discovery
func (d *Discovery) SetTimeout(timeout time.Duration) {
	d.timeout = timeout
}

// ScanPortRange scans a range of ports for MCP servers
func (d *Discovery) ScanPortRange(ctx context.Context, host string, startPort, endPort int) []ServerInfo {
	d.logger.Printf("Scanning port range %d-%d on %s", startPort, endPort, host)

	var ports []int
	for port := startPort; port <= endPort; port++ {
		ports = append(ports, port)
	}

	return d.DiscoverTCPServers(ctx, host, ports)
}

// TestConnection tests if a discovered server is actually an MCP server
func (d *Discovery) TestConnection(ctx context.Context, server ServerInfo) bool {
	d.logger.Printf("Testing connection to %s", server.Name)

	testCtx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	err := server.Transport.Connect(testCtx)
	if err != nil {
		d.logger.Printf("Failed to connect to %s: %v", server.Name, err)
		return false
	}

	defer server.Transport.Close()

	d.logger.Printf("Successfully connected to %s", server.Name)
	return true
}
