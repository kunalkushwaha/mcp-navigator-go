package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"sync"
	"time"

	"github.com/kunalkushwaha/mcp-navigator-go/pkg/mcp"

	"github.com/gorilla/websocket"
)

// WebSocketTransport implements Transport for WebSocket connections
type WebSocketTransport struct {
	url       string
	conn      *websocket.Conn
	connected bool
	mu        sync.RWMutex
	timeout   time.Duration
	readChan  chan []byte
	writeChan chan []byte
	stopChan  chan struct{}
	errorChan chan error
}

// NewWebSocketTransport creates a new WebSocket transport
func NewWebSocketTransport(wsURL string) *WebSocketTransport {
	return &WebSocketTransport{
		url:       wsURL,
		timeout:   30 * time.Second,
		readChan:  make(chan []byte, 100),
		writeChan: make(chan []byte, 100),
		stopChan:  make(chan struct{}),
		errorChan: make(chan error, 10),
	}
}

// Connect establishes WebSocket connection
func (w *WebSocketTransport) Connect(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.connected {
		return nil
	}

	// Parse and validate URL
	u, err := url.Parse(w.url)
	if err != nil {
		return fmt.Errorf("invalid WebSocket URL '%s': %w", w.url, err)
	}

	// Create dialer with timeout
	dialer := websocket.Dialer{
		HandshakeTimeout: w.timeout,
	}

	// Connect to WebSocket
	conn, _, err := dialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket %s: %w", w.url, err)
	}

	w.conn = conn
	w.connected = true

	// Start goroutines for reading and writing
	go w.readLoop()
	go w.writeLoop()

	return nil
}

// Close closes the WebSocket connection
func (w *WebSocketTransport) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.connected || w.conn == nil {
		return nil
	}

	// Signal stop
	close(w.stopChan)

	// Close WebSocket connection
	w.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	err := w.conn.Close()

	w.connected = false
	w.conn = nil

	return err
}

// Send sends a message over WebSocket
func (w *WebSocketTransport) Send(message *mcp.Message) error {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if !w.connected {
		return fmt.Errorf("transport not connected")
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	select {
	case w.writeChan <- data:
		return nil
	case <-time.After(w.timeout):
		return fmt.Errorf("timeout sending message")
	}
}

// Receive receives a message from WebSocket
func (w *WebSocketTransport) Receive() (*mcp.Message, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if !w.connected {
		return nil, fmt.Errorf("transport not connected")
	}

	select {
	case data := <-w.readChan:
		var message mcp.Message
		if err := json.Unmarshal(data, &message); err != nil {
			return nil, fmt.Errorf("failed to unmarshal message: %w", err)
		}
		return &message, nil
	case err := <-w.errorChan:
		return nil, err
	case <-time.After(w.timeout):
		return nil, fmt.Errorf("timeout receiving message")
	}
}

// GetReader returns nil for WebSocket (not applicable)
func (w *WebSocketTransport) GetReader() io.Reader {
	return nil
}

// GetWriter returns nil for WebSocket (not applicable)
func (w *WebSocketTransport) GetWriter() io.Writer {
	return nil
}

// IsConnected returns connection status
func (w *WebSocketTransport) IsConnected() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.connected
}

// SetTimeout sets the connection timeout
func (w *WebSocketTransport) SetTimeout(timeout time.Duration) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.timeout = timeout
}

// readLoop handles reading messages from WebSocket
func (w *WebSocketTransport) readLoop() {
	defer func() {
		if r := recover(); r != nil {
			w.errorChan <- fmt.Errorf("read loop panic: %v", r)
		}
	}()

	for {
		select {
		case <-w.stopChan:
			return
		default:
			_, message, err := w.conn.ReadMessage()
			if err != nil {
				w.errorChan <- fmt.Errorf("failed to read WebSocket message: %w", err)
				return
			}
			w.readChan <- message
		}
	}
}

// writeLoop handles writing messages to WebSocket
func (w *WebSocketTransport) writeLoop() {
	defer func() {
		if r := recover(); r != nil {
			w.errorChan <- fmt.Errorf("write loop panic: %v", r)
		}
	}()

	for {
		select {
		case <-w.stopChan:
			return
		case data := <-w.writeChan:
			if err := w.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				w.errorChan <- fmt.Errorf("failed to write WebSocket message: %w", err)
				return
			}
		}
	}
}

// GetURL returns the WebSocket URL
func (w *WebSocketTransport) GetURL() string {
	return w.url
}
