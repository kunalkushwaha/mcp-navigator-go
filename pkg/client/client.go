// Package client provides a comprehensive Model Context Protocol (MCP) client implementation.
//
// The client supports multiple transport types (TCP, STDIO, WebSocket, Docker) and provides
// a simple API for connecting to MCP servers, discovering their capabilities, and executing tools.
//
// Basic usage:
//
//	transport := transport.NewTCPTransport("localhost", 8811)
//	config := client.ClientConfig{
//		Name:    "my-app",
//		Version: "1.0.0",
//	}
//	client := client.NewClient(transport, config)
//
//	ctx := context.Background()
//	if err := client.Connect(ctx); err != nil {
//		log.Fatal(err)
//	}
//	defer client.Disconnect()
//
//	if err := client.Initialize(ctx, mcp.ClientInfo{Name: "my-app", Version: "1.0.0"}); err != nil {
//		log.Fatal(err)
//	}
//
//	tools, err := client.ListTools(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// The client is thread-safe and can be used concurrently from multiple goroutines.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kunalkushwaha/mcp-navigator-go/pkg/mcp"
	"github.com/kunalkushwaha/mcp-navigator-go/pkg/transport"
)

// Client represents an MCP client
type Client struct {
	transport          transport.Transport
	serverInfo         *mcp.ServerInfo
	serverCapabilities *mcp.ServerCapabilities
	connected          bool
	initialized        bool
	mu                 sync.RWMutex
	requestID          int64
	logger             *log.Logger
	timeout            time.Duration
}

// ClientConfig holds configuration for the MCP client
type ClientConfig struct {
	Name    string
	Version string
	Logger  *log.Logger
	Timeout time.Duration
}

// NewClient creates a new MCP client with the given transport and configuration.
//
// The transport parameter specifies how to communicate with the MCP server (TCP, STDIO, etc.).
// The config parameter allows customization of client behavior including logging and timeouts.
//
// If config.Logger is nil, log.Default() will be used.
// If config.Timeout is 0, a default timeout of 30 seconds will be used.
//
// Example:
//
//	transport := transport.NewTCPTransport("localhost", 8811)
//	config := ClientConfig{
//		Name:    "my-app",
//		Version: "1.0.0",
//		Logger:  myLogger,
//		Timeout: 60 * time.Second,
//	}
//	client := NewClient(transport, config)
func NewClient(transport transport.Transport, config ClientConfig) *Client {
	if config.Logger == nil {
		config.Logger = log.Default()
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &Client{
		transport: transport,
		logger:    config.Logger,
		timeout:   config.Timeout,
	}
}

// Connect establishes connection to the MCP server.
//
// This method must be called before Initialize() or any other operations.
// It's safe to call Connect() multiple times - subsequent calls will be ignored
// if already connected.
//
// The context can be used to timeout or cancel the connection attempt.
//
// Returns an error if the connection fails.
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	c.logger.Println("Connecting to MCP server...")

	if err := c.transport.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect transport: %w", err)
	}

	c.connected = true
	c.logger.Println("Connected to MCP server")
	return nil
}

// Initialize performs the MCP protocol initialization handshake.
//
// This method must be called after Connect() and before any tool or resource operations.
// It exchanges capabilities with the server and establishes the protocol session.
//
// The clientInfo parameter identifies this client to the server and should contain
// a meaningful name and version.
//
// Returns an error if initialization fails or if the client is not connected.
func (c *Client) Initialize(ctx context.Context, clientInfo mcp.ClientInfo) error {
	if !c.IsConnected() {
		return ErrNotConnected
	}
	c.logger.Printf("Initializing MCP protocol with client: %s %s", clientInfo.Name, clientInfo.Version)

	c.logger.Printf("Creating initialize request...")
	// Create initialize request
	request := mcp.InitializeRequest{
		ProtocolVersion: mcp.Version,
		Capabilities: mcp.ClientCapabilities{
			Experimental: make(map[string]interface{}),
			Sampling:     &mcp.SamplingCapability{},
		},
		ClientInfo: clientInfo,
	}
	c.logger.Printf("Initialize request created successfully")
	// Send initialize request
	c.logger.Printf("Sending initialize request...")
	response, err := c.sendRequest(ctx, "initialize", request)
	if err != nil {
		return fmt.Errorf("initialize request failed: %w", err)
	}
	c.logger.Printf("Received initialize response")

	if response.Error != nil {
		return fmt.Errorf("initialize error: %s", response.Error.Message)
	}

	// Parse initialize response
	var initResponse mcp.InitializeResponse
	if err := parseResult(response.Result, &initResponse); err != nil {
		return fmt.Errorf("failed to parse initialize response: %w", err)
	}

	c.mu.Lock()
	c.serverInfo = &initResponse.ServerInfo
	c.serverCapabilities = &initResponse.Capabilities
	c.initialized = true
	c.mu.Unlock()

	c.logger.Printf("MCP protocol initialized. Server: %s %s",
		initResponse.ServerInfo.Name, initResponse.ServerInfo.Version)

	// Send initialized notification
	notification := mcp.NewNotification("notifications/initialized", nil)
	if err := c.transport.Send(notification); err != nil {
		return fmt.Errorf("failed to send initialized notification: %w", err)
	}

	return nil
}

// Disconnect closes the connection to the MCP server
func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	c.logger.Println("Disconnecting from MCP server...")

	err := c.transport.Close()
	c.connected = false
	c.initialized = false
	c.serverInfo = nil
	c.serverCapabilities = nil

	c.logger.Println("Disconnected from MCP server")
	return err
}

// IsConnected returns true if the client is connected
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// IsInitialized returns true if the MCP protocol has been initialized
func (c *Client) IsInitialized() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.initialized
}

// GetServerInfo returns information about the connected server
func (c *Client) GetServerInfo() *mcp.ServerInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.serverInfo == nil {
		return nil
	}
	info := *c.serverInfo
	return &info
}

// GetServerCapabilities returns the server's capabilities
func (c *Client) GetServerCapabilities() *mcp.ServerCapabilities {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.serverCapabilities == nil {
		return nil
	}
	caps := *c.serverCapabilities
	return &caps
}

// ListTools retrieves all available tools from the server
func (c *Client) ListTools(ctx context.Context) ([]mcp.Tool, error) {
	if !c.IsInitialized() {
		return nil, fmt.Errorf("client not initialized")
	}

	c.logger.Println("Listing available tools...")

	response, err := c.sendRequest(ctx, "tools/list", mcp.ListToolsRequest{})
	if err != nil {
		return nil, fmt.Errorf("list tools request failed: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("list tools error: %s", response.Error.Message)
	}

	var listResponse mcp.ListToolsResponse
	if err := parseResult(response.Result, &listResponse); err != nil {
		return nil, fmt.Errorf("failed to parse list tools response: %w", err)
	}

	c.logger.Printf("Found %d tools", len(listResponse.Tools))
	return listResponse.Tools, nil
}

// CallTool executes a tool on the server
func (c *Client) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*mcp.CallToolResponse, error) {
	if !c.IsInitialized() {
		return nil, fmt.Errorf("client not initialized")
	}

	// Check connection health before making the call
	if err := c.CheckConnection(); err != nil {
		return nil, fmt.Errorf("connection check failed: %w", err)
	}

	c.logger.Printf("Calling tool: %s", name)

	request := mcp.CallToolRequest{
		Name:      name,
		Arguments: arguments,
	}

	response, err := c.sendRequest(ctx, "tools/call", request)
	if err != nil {
		return nil, fmt.Errorf("call tool request failed: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("call tool error: %s", response.Error.Message)
	}

	var callResponse mcp.CallToolResponse
	if err := parseResult(response.Result, &callResponse); err != nil {
		return nil, fmt.Errorf("failed to parse call tool response: %w", err)
	}

	c.logger.Printf("Tool '%s' executed successfully", name)
	return &callResponse, nil
}

// ListResources retrieves all available resources from the server
func (c *Client) ListResources(ctx context.Context) ([]mcp.Resource, error) {
	if !c.IsInitialized() {
		return nil, fmt.Errorf("client not initialized")
	}

	c.logger.Println("Listing available resources...")

	response, err := c.sendRequest(ctx, "resources/list", mcp.ListResourcesRequest{})
	if err != nil {
		return nil, fmt.Errorf("list resources request failed: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("list resources error: %s", response.Error.Message)
	}

	var listResponse mcp.ListResourcesResponse
	if err := parseResult(response.Result, &listResponse); err != nil {
		return nil, fmt.Errorf("failed to parse list resources response: %w", err)
	}

	c.logger.Printf("Found %d resources", len(listResponse.Resources))
	return listResponse.Resources, nil
}

// ListPrompts retrieves all available prompts from the server
func (c *Client) ListPrompts(ctx context.Context) ([]mcp.Prompt, error) {
	if !c.IsInitialized() {
		return nil, fmt.Errorf("client not initialized")
	}

	c.logger.Println("Listing available prompts...")

	response, err := c.sendRequest(ctx, "prompts/list", mcp.ListPromptsRequest{})
	if err != nil {
		return nil, fmt.Errorf("list prompts request failed: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("list prompts error: %s", response.Error.Message)
	}

	var listResponse mcp.ListPromptsResponse
	if err := parseResult(response.Result, &listResponse); err != nil {
		return nil, fmt.Errorf("failed to parse list prompts response: %w", err)
	}

	c.logger.Printf("Found %d prompts", len(listResponse.Prompts))
	return listResponse.Prompts, nil
}

// GetPrompt retrieves a specific prompt from the server with optional arguments
func (c *Client) GetPrompt(ctx context.Context, name string, arguments map[string]interface{}) (*mcp.GetPromptResponse, error) {
	if !c.IsInitialized() {
		return nil, fmt.Errorf("client not initialized")
	}

	c.logger.Printf("Getting prompt: %s", name)

	request := mcp.GetPromptRequest{
		Name:      name,
		Arguments: arguments,
	}

	response, err := c.sendRequest(ctx, "prompts/get", request)
	if err != nil {
		return nil, fmt.Errorf("get prompt request failed: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("get prompt error: %s", response.Error.Message)
	}

	var promptResponse mcp.GetPromptResponse
	if err := parseResult(response.Result, &promptResponse); err != nil {
		return nil, fmt.Errorf("failed to parse get prompt response: %w", err)
	}

	c.logger.Printf("Retrieved prompt '%s' with %d messages", name, len(promptResponse.Messages))
	return &promptResponse, nil
}

// ReadResource retrieves the content of a specific resource from the server
func (c *Client) ReadResource(ctx context.Context, uri string) (*mcp.ReadResourceResponse, error) {
	if !c.IsInitialized() {
		return nil, fmt.Errorf("client not initialized")
	}

	c.logger.Printf("Reading resource: %s", uri)

	request := mcp.ReadResourceRequest{
		URI: uri,
	}

	response, err := c.sendRequest(ctx, "resources/read", request)
	if err != nil {
		return nil, fmt.Errorf("read resource request failed: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("read resource error: %s", response.Error.Message)
	}

	var resourceResponse mcp.ReadResourceResponse
	if err := parseResult(response.Result, &resourceResponse); err != nil {
		return nil, fmt.Errorf("failed to parse read resource response: %w", err)
	}

	c.logger.Printf("Read resource '%s' with %d content items", uri, len(resourceResponse.Contents))
	return &resourceResponse, nil
}

// sendRequest sends a request and waits for the response
func (c *Client) sendRequest(ctx context.Context, method string, params interface{}) (*mcp.Message, error) {
	requestID := atomic.AddInt64(&c.requestID, 1)

	request := mcp.NewRequest(requestID, method, params)

	// Check if transport is still connected before sending
	if !c.transport.IsConnected() {
		return nil, fmt.Errorf("transport disconnected")
	}

	if err := c.transport.Send(request); err != nil {
		// Mark client as disconnected if send fails
		c.mu.Lock()
		c.connected = false
		c.initialized = false
		c.mu.Unlock()
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Wait for response with timeout
	responseCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	for {
		select {
		case <-responseCtx.Done():
			c.logger.Printf("Request %d timed out", requestID)
			return nil, fmt.Errorf("request timeout")
		default:
			response, err := c.transport.Receive()
			if err != nil {
				// Mark client as disconnected if receive fails
				c.mu.Lock()
				c.connected = false
				c.initialized = false
				c.mu.Unlock()
				return nil, fmt.Errorf("failed to receive response: %w", err)
			}

			// Check if this is the response we're waiting for
			// Handle different ID types (JSON unmarshaling might convert int64 to float64)
			if c.isMatchingID(response.ID, requestID) {
				return response, nil
			}

			// Handle notifications or other messages
			c.handleMessage(response)
		}
	}
}

// handleMessage processes incoming messages (notifications, etc.)
func (c *Client) handleMessage(message *mcp.Message) {
	if message.Method != "" && message.ID == nil {
		// This is a notification
		c.logger.Printf("Received notification: %s", message.Method)
	}
}

// parseResult parses a response result into the target structure
func parseResult(result interface{}, target interface{}) error {
	if result == nil {
		return fmt.Errorf("result is nil")
	}

	// Convert result to JSON and back to properly unmarshal into target
	jsonData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	if err := json.Unmarshal(jsonData, target); err != nil {
		return fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return nil
}

// isMatchingID compares request IDs, handling JSON unmarshaling type conversions
func (c *Client) isMatchingID(responseID interface{}, requestID int64) bool {
	if responseID == nil {
		return false
	}

	switch id := responseID.(type) {
	case int64:
		return id == requestID
	case float64:
		return int64(id) == requestID
	case int:
		return int64(id) == requestID
	case string:
		// Try to parse string as int
		if parsedID, err := strconv.ParseInt(id, 10, 64); err == nil {
			return parsedID == requestID
		}
	}

	return false
}

// CheckConnection verifies the transport is still connected and updates client state
func (c *Client) CheckConnection() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.transport.IsConnected() {
		c.connected = false
		c.initialized = false
		return fmt.Errorf("transport disconnected")
	}

	return nil
}
