# MCP Navigator (Go)

A comprehensive Model Context Protocol (MCP) client implementation in Go, providing both a powerful CLI and a library for integration into other applications.

## Features

- 🔍 **Server Discovery**: Automatically discover MCP servers on TCP ports and Docker containers
- 🔌 **Multiple Transports**: Support for TCP, STDIO, WebSocket, and Docker-based connections
- 💬 **Interactive CLI**: Full-featured command-line interface for server interaction
- 🛠️ **Complete MCP Protocol**: Full support for Tools, Resources, and Prompts
- 🐳 **Docker Support**: Direct support for Docker-based MCP servers
- 📚 **Library Integration**: Use as a library in your Go applications
- ⚡ **High Performance**: Written in Go for speed and efficiency

## Installation

### From Source

```bash
git clone https://github.com/kunalkushwaha/mcp-navigator-go.git
cd mcp-navigator-go
go mod download
go build -o mcp-navigator main.go
```

### Using Go Install

```bash
go install github.com/kunalkushwaha/mcp-navigator-go@latest
```

### As a Library

```bash
go get github.com/kunalkushwaha/mcp-navigator-go
```

## Quick Start

### 1. Discover Available MCP Servers

```bash
./mcp-navigator discover
```

This will show available servers including:
- TCP servers on common MCP ports (8810-8820)
- Docker-based MCP servers
- Standard Docker MCP configuration (alpine/socat)

### 2. Interactive Mode

```bash
./mcp-navigator interactive
```

This starts the interactive CLI where you can:
- `help` - Show available commands
- `discover` - Find MCP servers
- `connect <server-name-or-index>` - Connect to a server
- `list-tools` - List available tools on connected server
- `list-resources` - List available resources on connected server
- `call-tool <tool-name> [json-args]` - Execute a tool
- `status` - Show connection status
- `exit` - Exit the client

### 3. Direct Commands

Connect to a TCP server:
```bash
./mcp-navigator connect --tcp --host localhost --port 8811
```

Connect to Docker MCP server:
```bash
./mcp-navigator connect --docker
```

Execute a tool directly:
```bash
./mcp-navigator tool --name search --arguments '{"query": "golang"}' --docker
```

## Usage Examples

### Server Discovery

```bash
# Discover all servers
./mcp-navigator discover

# Scan specific host
./mcp-navigator discover --host 192.168.1.100

# Only scan TCP ports
./mcp-navigator discover --tcp-only

# Only check Docker containers
./mcp-navigator discover --docker-only

# Custom port range
./mcp-navigator discover --start-port 8000 --end-port 9000
```

### Connecting to Servers

```bash
# TCP connection
./mcp-navigator connect --tcp --host localhost --port 8811

# STDIO connection
./mcp-navigator connect --stdio --command "node" --args "server.js"

# Docker connection (uses alpine/socat bridge)
./mcp-navigator connect --docker

# With custom timeout
./mcp-navigator connect --tcp --host localhost --port 8811 --timeout 45s
```

### Tool Execution

```bash
# Execute tool with JSON arguments
./mcp-navigator tool --name search --arguments '{"query": "golang", "limit": 10}' --tcp

# Execute tool via Docker MCP server
./mcp-navigator tool --name docker --arguments '{"command": "ps -a"}' --docker

# Execute tool with no arguments
./mcp-navigator tool --name list-files --docker
```

## Docker MCP Server Support

The client automatically supports the standard Docker-based MCP server configuration used by Claude Desktop:

```json
{
  "mcpServers": {
    "MCP_DOCKER": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "alpine/socat",
        "STDIO",
        "TCP:host.docker.internal:8811"
      ]
    }
  }
}
```

This configuration allows MCP servers running in Docker containers to communicate with external TCP services.

## Interactive Mode Example

```bash
$ ./mcp-navigator interactive

🚀 MCP Navigator Interactive Mode
Type 'help' for available commands.
🔍 Discovering MCP servers...
✅ Found 3 server(s):
  1. TCP Server localhost:8811 (tcp)
     Address: localhost:8811
  2. Docker Container mcp-server (docker)
  3. Docker MCP (alpine/socat) (docker)
     Address: host.docker.internal:8811

mcp-client> help

📋 Available Commands:
  help              - Show this help message
  discover          - Discover available MCP servers
  connect <n>       - Connect to a server by name or index
  disconnect        - Disconnect from current server
  list-tools        - List tools available on current server
  list-resources    - List resources available on current server
  call-tool <n> [args] - Call a tool with optional JSON arguments
  status            - Show connection status
  exit/quit         - Exit the client

mcp-client> connect 1
🔌 Connecting to TCP Server localhost:8811...
✅ Connected to TCP Server localhost:8811
🚀 Server: mcp-server 1.0.0

mcp-client> list-tools
📝 Available tools (3):
  1. search
     Description: Search for information using DuckDuckGo
  2. fetch_content
     Description: Fetch and parse content from a webpage URL
  3. docker
     Description: Execute Docker commands

mcp-client> call-tool search {"query": "Model Context Protocol"}
🔧 Calling tool: search
📤 Tool result:
The Model Context Protocol (MCP) is an open standard that enables secure connections between AI assistants and data sources...

mcp-client> status

📊 Status:
  Available servers: 3
  Current connection: TCP Server localhost:8811 ✅
  Server info: mcp-server 1.0.0

mcp-client> exit

👋 Shutting down MCP client...
✅ Goodbye!
```

## Configuration

### Environment Variables

- `MCP_CLIENT_TIMEOUT`: Default timeout for operations (default: 30s)
- `MCP_CLIENT_HOST`: Default host for TCP connections (default: localhost)
- `MCP_CLIENT_PORT`: Default port for TCP connections (default: 8811)
- `MCP_CLIENT_VERBOSE`: Enable verbose logging (default: false)

### Config File

Create `~/.mcp-client.yaml`:

```yaml
# Default connection settings
host: localhost
port: 8811
timeout: 30s

# Discovery settings
discovery:
  startPort: 8810
  endPort: 8820
  timeout: 5s

# Logging
verbose: false
```

## Command Reference

### Global Flags

- `--config`: Config file path
- `--verbose, -v`: Enable verbose output

### Commands

#### `discover`
Discover available MCP servers

**Flags:**
- `--host`: Host to scan (default: localhost)
- `--start-port`: Start port for scanning (default: 8810)
- `--end-port`: End port for scanning (default: 8820)
- `--timeout`: Discovery timeout (default: 5s)
- `--tcp-only`: Only scan TCP ports
- `--docker-only`: Only check Docker containers

#### `connect`
Connect to an MCP server and show available tools/resources

**Flags:**
- `--type`: Connection type (tcp, stdio, docker)
- `--tcp, -t`: Use TCP transport
- `--stdio, -s`: Use STDIO transport
- `--docker, -d`: Use Docker transport
- `--host`: TCP host (default: localhost)
- `--port`: TCP port (default: 8811)
- `--command`: Command for STDIO transport
- `--args`: Arguments for STDIO command
- `--timeout`: Connection timeout (default: 30s)

#### `tool`
Execute a specific tool on an MCP server

**Flags:**
- `--name`: Tool name (required)
- `--arguments`: JSON arguments for the tool (default: "{}")
- All connection flags from `connect` command

#### `interactive`
Start interactive mode

**Aliases:** `i`, `shell`

## Development

### Project Structure

```
├── main.go                 # Entry point
├── internal/
│   └── cli/               # CLI commands
│       ├── root.go        # Root command and configuration
│       ├── discover.go    # Server discovery command
│       ├── connect.go     # Connection command
│       ├── tool.go        # Tool execution command
│       └── interactive.go # Interactive mode
├── pkg/
│   ├── client/           # MCP client implementation
│   ├── discovery/        # Server discovery logic
│   ├── mcp/             # MCP protocol types and utilities
│   └── transport/       # Transport implementations (TCP, STDIO, WebSocket)
├── go.mod
└── README.md
```

### Building

```bash
go build -o mcp-client main.go
```

### Testing

```bash
go test ./...
```

### Running in Development

```bash
go run main.go interactive
```

## Requirements

- Go 1.21 or later
- Docker (for Docker-based MCP servers)
- MCP Server running on TCP port or Docker

## Related

- [Model Context Protocol](https://modelcontextprotocol.io/)
- [MCP Specification](https://spec.modelcontextprotocol.io/)
- [TypeScript MCP SDK](https://github.com/modelcontextprotocol/typescript-sdk)

## License

MIT License - see LICENSE file for details.
