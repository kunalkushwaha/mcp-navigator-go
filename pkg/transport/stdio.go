package transport

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"github.com/kunalkushwaha/mcp-navigator-go/pkg/mcp"
)

// StdioTransport implements Transport for STDIO-based connections (processes)
type StdioTransport struct {
	command   string
	args      []string
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    io.ReadCloser
	stderr    io.ReadCloser
	reader    *bufio.Reader
	writer    *bufio.Writer
	connected bool
	mu        sync.RWMutex
}

// NewStdioTransport creates a new STDIO transport
func NewStdioTransport(command string, args []string) *StdioTransport {
	return &StdioTransport{
		command: command,
		args:    args,
	}
}

// Connect starts the process and establishes STDIO connection
func (s *StdioTransport) Connect(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.connected {
		return nil
	}

	s.cmd = exec.CommandContext(ctx, s.command, s.args...)

	stdin, err := s.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	s.stdin = stdin

	stdout, err := s.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	s.stdout = stdout

	stderr, err := s.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	s.stderr = stderr

	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command '%s %v': %w", s.command, s.args, err)
	}

	s.reader = bufio.NewReader(s.stdout)
	s.writer = bufio.NewWriter(s.stdin)
	s.connected = true

	return nil
}

// Close closes the STDIO connection and terminates the process
func (s *StdioTransport) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.connected {
		return nil
	}

	var errs []error

	if s.stdin != nil {
		if err := s.stdin.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if s.stdout != nil {
		if err := s.stdout.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if s.stderr != nil {
		if err := s.stderr.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if s.cmd != nil && s.cmd.Process != nil {
		if err := s.cmd.Process.Kill(); err != nil {
			errs = append(errs, err)
		}
		s.cmd.Wait() // Wait for process to exit
	}

	s.connected = false
	s.cmd = nil
	s.stdin = nil
	s.stdout = nil
	s.stderr = nil
	s.reader = nil
	s.writer = nil

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}

	return nil
}

// Send sends a message via STDIO
func (s *StdioTransport) Send(message *mcp.Message) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.connected {
		return fmt.Errorf("transport not connected")
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Write message with newline delimiter
	_, err = s.writer.Write(append(data, '\n'))
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return s.writer.Flush()
}

// Receive receives a message from STDIO
func (s *StdioTransport) Receive() (*mcp.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.connected {
		return nil, fmt.Errorf("transport not connected")
	}

	line, err := s.reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read message: %w", err)
	}

	var message mcp.Message
	if err := json.Unmarshal(line, &message); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return &message, nil
}

// GetReader returns the stdout reader
func (s *StdioTransport) GetReader() io.Reader {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.reader
}

// GetWriter returns the stdin writer
func (s *StdioTransport) GetWriter() io.Writer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.writer
}

// IsConnected returns connection status
func (s *StdioTransport) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connected
}

// GetCommand returns the command and args being executed
func (s *StdioTransport) GetCommand() (string, []string) {
	return s.command, s.args
}
