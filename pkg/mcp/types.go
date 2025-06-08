package mcp

import (
	"fmt"
)

// MCP Protocol Version
const Version = "2024-11-05"

// Message Types
type MessageType string

const (
	MessageTypeRequest      MessageType = "request"
	MessageTypeResponse     MessageType = "response"
	MessageTypeNotification MessageType = "notification"
)

// JSON-RPC 2.0 Base Message
type Message struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method,omitempty"`
	Params  interface{} `json:"params,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// Error Information
type ErrorInfo struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Client Capabilities
type ClientCapabilities struct {
	Experimental map[string]interface{} `json:"experimental,omitempty"`
	Sampling     *SamplingCapability    `json:"sampling,omitempty"`
}

type SamplingCapability struct{}

// Server Capabilities
type ServerCapabilities struct {
	Experimental map[string]interface{} `json:"experimental,omitempty"`
	Logging      *LoggingCapability     `json:"logging,omitempty"`
	Prompts      *PromptsCapability     `json:"prompts,omitempty"`
	Resources    *ResourcesCapability   `json:"resources,omitempty"`
	Tools        *ToolsCapability       `json:"tools,omitempty"`
}

type LoggingCapability struct{}
type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}
type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// Initialize Request/Response
type InitializeRequest struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      ClientInfo         `json:"clientInfo"`
}

type InitializeResponse struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Tool Definitions
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type ListToolsRequest struct{}

type ListToolsResponse struct {
	Tools []Tool `json:"tools"`
}

// Call Tool Request/Response
type CallToolRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type CallToolResponse struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

// Content Types
type Content struct {
	Type        string                 `json:"type"`
	Text        string                 `json:"text,omitempty"`
	Data        string                 `json:"data,omitempty"`
	MimeType    string                 `json:"mimeType,omitempty"`
	Name        string                 `json:"name,omitempty"`
	URI         string                 `json:"uri,omitempty"`
	Annotations map[string]interface{} `json:"annotations,omitempty"`
}

// Resource Definitions
type Resource struct {
	URI         string                 `json:"uri"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	MimeType    string                 `json:"mimeType,omitempty"`
	Annotations map[string]interface{} `json:"annotations,omitempty"`
}

type ListResourcesRequest struct{}

type ListResourcesResponse struct {
	Resources []Resource `json:"resources"`
}

// Prompt Definitions
type Prompt struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
}

type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

type ListPromptsRequest struct{}

type ListPromptsResponse struct {
	Prompts []Prompt `json:"prompts"`
}

// GetPrompt request/response types
type GetPromptRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type PromptMessage struct {
	Role    string  `json:"role"`
	Content Content `json:"content"`
}

type GetPromptResponse struct {
	Description string          `json:"description,omitempty"`
	Messages    []PromptMessage `json:"messages"`
}

// ReadResource request/response types
type ReadResourceRequest struct {
	URI string `json:"uri"`
}

type ReadResourceResponse struct {
	Contents []Content `json:"contents"`
}

// Utility functions for creating messages
func NewRequest(id interface{}, method string, params interface{}) *Message {
	return &Message{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}
}

func NewResponse(id interface{}, result interface{}) *Message {
	return &Message{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

func NewErrorResponse(id interface{}, code int, message string, data interface{}) *Message {
	return &Message{
		JSONRPC: "2.0",
		ID:      id,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

func NewNotification(method string, params interface{}) *Message {
	return &Message{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
}

// Error codes based on JSON-RPC 2.0 and MCP specification
const (
	ErrorCodeParseError     = -32700
	ErrorCodeInvalidRequest = -32600
	ErrorCodeMethodNotFound = -32601
	ErrorCodeInvalidParams  = -32602
	ErrorCodeInternalError  = -32603
)

// MCP-specific error codes
const (
	ErrorCodeInvalidMessage  = -32001
	ErrorCodeInvalidTool     = -32002
	ErrorCodeInvalidResource = -32003
)

func (e *ErrorInfo) Error() string {
	return fmt.Sprintf("JSON-RPC Error %d: %s", e.Code, e.Message)
}
