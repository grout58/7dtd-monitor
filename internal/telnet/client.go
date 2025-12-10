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

	// Read response. This is tricky with Telnet as there is no specific EOF.
	// We rely on the fact that the server usually sends a newline at the end of the block?
	// Or we wait for a bit. For 7DTD, commands usually return immediate output.
	// A robust solution reads until a known prompt, but 7DTD telnet is raw.
	// We'll read buffered available data after a short sleep to allow server to process.
	// *Note*: In a real prod client we'd be smarter, but this is simple for now.

	time.Sleep(100 * time.Millisecond)

	var output strings.Builder
	buffer := make([]byte, 4096)

	// Read at least once
	n, err := c.conn.Read(buffer)
	if err != nil {
		return "", err
	}
	output.Write(buffer[:n])

	// Attempt to read more if available
	// This is blocking, but we set a very short Read deadline just to check buffer
	c.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	for {
		n, err := c.conn.Read(buffer)
		if n > 0 {
			output.Write(buffer[:n])
		}
		if err != nil {
			break // Timeout or EOF
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
