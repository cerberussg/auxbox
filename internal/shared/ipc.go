package shared

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
)

type Transport interface {
	Send(cmd Command) (*Response, error)
	Listen(handler func(Command) Response) error
	Close() error
}

type UnixSocketTransport struct {
	socketPath string
	listener   net.Listener
	conn       net.Conn
}

func NewUnixSocketTransport() *UnixSocketTransport {
	// Put socket in user's config directory or fallback to /tmp
	socketPath := "/tmp/auxbox.sock"

	if runtimeDir := os.Getenv("XDG_RUNTIME_DIR"); runtimeDir != "" {
		socketPath = filepath.Join(runtimeDir, "auxbox.sock")
	} else if homeDir, err := os.UserHomeDir(); err == nil {
		configDir := filepath.Join(homeDir, ".config", "auxbox")
		os.MkdirAll(configDir, 0755) // Create directory if it doesn't exist
		socketPath = filepath.Join(configDir, "auxbox.sock")
	}

	return &UnixSocketTransport{
		socketPath: socketPath,
	}
}

func (u *UnixSocketTransport) Send(cmd Command) (*Response, error) {
	conn, err := net.Dial("unix", u.socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to daemon: %w (is auxbox daemon running?)", err)
	}
	defer conn.Close()

	data, err := cmd.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize command: %w", err)
	}

	_, err = fmt.Fprintf(conn, "%s\n", data)
	if err != nil {
		return nil, fmt.Errorf("failed to send command: %w", err)
	}

	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		return nil, fmt.Errorf("failed to read response: %w", scanner.Err())
	}

	resp, err := ResponseFromJSON(scanner.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &resp, nil
}

func (u *UnixSocketTransport) Listen(handler func(Command) Response) error {
	os.Remove(u.socketPath)

	listener, err := net.Listen("unix", u.socketPath)
	if err != nil {
		return fmt.Errorf("failed to create socket listener: %w", err)
	}
	u.listener = listener

	os.Chmod(u.socketPath, 0600)

	fmt.Printf("auxbox daemon listening on %s\n", u.socketPath)

	for {
		conn, err := listener.Accept()
		if err != nil {
			if u.listener == nil {
				return nil
			}
			return fmt.Errorf("failed to accept connection: %w", err)
		}

		go u.handleConnection(conn, handler)
	}
}

func (u *UnixSocketTransport) handleConnection(conn net.Conn, handler func(Command) Response) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		cmd, err := CommandFromJSON(scanner.Bytes())
		if err != nil {
			resp := NewErrorResponse(fmt.Sprintf("invalid command: %v", err))
			respData, _ := resp.ToJSON()
			fmt.Fprintf(conn, "%s\n", respData)
			continue
		}

		resp := handler(cmd)

		respData, err := resp.ToJSON()
		if err != nil {
			fallback := NewErrorResponse("failed to serialize response")
			fallbackData, _ := fallback.ToJSON()
			fmt.Fprintf(conn, "%s\n", fallbackData)
			continue
		}

		fmt.Fprintf(conn, "%s\n", respData)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Connection error: %v\n", err)
	}
}

func (u *UnixSocketTransport) Close() error {
	if u.listener != nil {
		u.listener.Close()
		u.listener = nil
	}
	if u.conn != nil {
		u.conn.Close()
		u.conn = nil
	}
	os.Remove(u.socketPath)
	return nil
}

func (u *UnixSocketTransport) IsRunning() bool {
	conn, err := net.Dial("unix", u.socketPath)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func (u *UnixSocketTransport) GetSocketPath() string {
	return u.socketPath
}
