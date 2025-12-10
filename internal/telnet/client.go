package telnet

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

type Client struct {
	Host     string
	Port     string
	Password string
	conn     net.Conn
	reader   *bufio.Reader
	writer   *bufio.Writer
}

func NewClient(host, port, password string) *Client {
	return &Client{
		Host:     host,
		Port:     port,
		Password: password,
	}
}

func (c *Client) Connect() error {
	address := net.JoinHostPort(c.Host, c.Port)
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return err
	}
	c.conn = conn
	c.reader = bufio.NewReader(conn)
	c.writer = bufio.NewWriter(conn)

	// Authenticate
	return c.authenticate()
}

func (c *Client) authenticate() error {
	// Read initial prompt "Please enter password:"
	c.readUntil("password:")

	// Send password
	_, err := c.writer.WriteString(c.Password + "\r\n")
	if err != nil {
		return err
	}
	c.writer.Flush()

	// Wait for success
	_, err = c.readUntil("Logon successful")
	return err
}

func (c *Client) SendCommand(cmd string) (string, error) {
	if c.conn == nil {
		return "", fmt.Errorf("not connected")
	}

	_, err := c.writer.WriteString(cmd + "\r\n")
	if err != nil {
		return "", err
	}
	c.writer.Flush()

	// Wait a bit for the server to process strings
	// 7DTD can be slow.
	time.Sleep(300 * time.Millisecond)

	var output strings.Builder
	buffer := make([]byte, 4096)

	// Read Loop
	// We want to keep reading as long as there is data, or until a strict timeout.
	// Since we don't have a specific EOF/Prompt, we rely on a short inactivity timeout between chunks.

	// First read (blocking with timeout)
	c.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := c.conn.Read(buffer)
	if err != nil {
		// If timeout and no data, we might return empty (cmd had no output?)
		// But usually we get at least the log line.
		if strings.Contains(err.Error(), "timeout") {
			return "", nil
		}
		return "", err
	}
	output.Write(buffer[:n])

	// Subsequent reads: shorter timeout to drain buffer
	for {
		c.conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		n, err := c.conn.Read(buffer)
		if n > 0 {
			output.Write(buffer[:n])
		}
		if err != nil {
			// Timeout means we are done reading for now
			break
		}
	}
	// Reset deadline
	c.conn.SetReadDeadline(time.Time{})

	return output.String(), nil
}

func (c *Client) readUntil(substring string) (string, error) {
	var output strings.Builder
	buffer := make([]byte, 1024)
	for {
		n, err := c.conn.Read(buffer)
		if err != nil {
			return output.String(), err
		}
		chunk := string(buffer[:n])
		output.WriteString(chunk)
		if strings.Contains(output.String(), substring) {
			return output.String(), nil
		}
	}
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
