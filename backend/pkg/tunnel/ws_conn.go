package tunnel

import (
	"io"
	"net"
	"time"

	"github.com/gorilla/websocket"
)

// WSConn adapts a gorilla/websocket.Conn into a standard net.Conn.
// This allows multiplexers like Yamux to transparently operate on top of
// standard WebSocket frames, bypassing strict corporate proxies that drop non-WS traffic.
type WSConn struct {
	ws     *websocket.Conn
	reader io.Reader
}

// Ensure WSConn implements net.Conn
var _ net.Conn = (*WSConn)(nil)

// NewWSConn wraps a websocket connection into a net.Conn stream.
func NewWSConn(ws *websocket.Conn) *WSConn {
	return &WSConn{ws: ws}
}

func (c *WSConn) Read(p []byte) (int, error) {
	for {
		if c.reader == nil {
			msgType, r, err := c.ws.NextReader()
			if err != nil {
				return 0, err // Connection closed or error
			}
			// Only tunnel traffic over binary frames to prevent encoding overhead
			if msgType != websocket.BinaryMessage {
				continue
			}
			c.reader = r
		}

		n, err := c.reader.Read(p)
		if err == io.EOF {
			c.reader = nil // Reached end of current WebSocket message frame
			if n > 0 {
				return n, nil // Return what we read before moving to next message
			}
			continue
		}
		return n, err
	}
}

func (c *WSConn) Write(p []byte) (int, error) {
	err := c.ws.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (c *WSConn) Close() error {
	return c.ws.Close()
}

func (c *WSConn) LocalAddr() net.Addr {
	return c.ws.LocalAddr()
}

func (c *WSConn) RemoteAddr() net.Addr {
	return c.ws.RemoteAddr()
}

func (c *WSConn) SetDeadline(t time.Time) error {
	if err := c.SetReadDeadline(t); err != nil {
		return err
	}
	return c.SetWriteDeadline(t)
}

func (c *WSConn) SetReadDeadline(t time.Time) error {
	return c.ws.SetReadDeadline(t)
}

func (c *WSConn) SetWriteDeadline(t time.Time) error {
	return c.ws.SetWriteDeadline(t)
}
