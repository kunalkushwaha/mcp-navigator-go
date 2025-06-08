package transport

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/kunalkushwaha/mcp-navigator-go/pkg/mcp"
)

// TCPTransport implements Transport for TCP connections
type TCPTransport struct {
	host      string
	port      int
	conn      net.Conn
	reader    *bufio.Reader
	writer    *bufio.Writer
	connected bool
	mu        sync.RWMutex
	timeout   time.Duration
}

// NewTCPTransport creates a new TCP transport
func NewTCPTransport(host string, port int) *TCPTransport {
	return &TCPTransport{
		host:    host,
		port:    port,
		timeout: 30 * time.Second,
	}
}

// Connect establishes TCP connection
func (t *TCPTransport) Connect(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.connected {
		return nil
	}

	address := fmt.Sprintf("%s:%d", t.host, t.port)

	dialer := &net.Dialer{
		Timeout: t.timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", address, err)
	}

	t.conn = conn
	t.reader = bufio.NewReader(conn)
	t.writer = bufio.NewWriter(conn)
	t.connected = true

	return nil
}

// Close closes the TCP connection
func (t *TCPTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.connected || t.conn == nil {
		return nil
	}

	err := t.conn.Close()
	t.connected = false
	t.conn = nil
	t.reader = nil
	t.writer = nil

	return err
}

// Send sends a message over TCP
func (t *TCPTransport) Send(message *mcp.Message) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if !t.connected {
		return fmt.Errorf("transport not connected")
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Write message with newline delimiter
	_, err = t.writer.Write(append(data, '\n'))
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = t.writer.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush message: %w", err)
	}

	return nil
}

// Receive receives a message from TCP
func (t *TCPTransport) Receive() (*mcp.Message, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if !t.connected {
		return nil, fmt.Errorf("transport not connected")
	}

	line, err := t.reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read message: %w", err)
	}

	var message mcp.Message
	if err := json.Unmarshal(line, &message); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return &message, nil
}

// GetReader returns the underlying reader
func (t *TCPTransport) GetReader() io.Reader {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.reader != nil {
		return t.reader
	}
	return nil
}

// GetWriter returns the underlying writer
func (t *TCPTransport) GetWriter() io.Writer {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.writer != nil {
		return t.writer
	}
	return nil
}

// IsConnected returns connection status
func (t *TCPTransport) IsConnected() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.connected
}

// SetTimeout sets the connection timeout
func (t *TCPTransport) SetTimeout(timeout time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.timeout = timeout
}
