package net

import (
	"github.com/flamingo/server/internal/base"
	"go.uber.org/zap"
	"net"
	"sync"
	"time"
)

type ConnectionCallback func(conn *TcpConnection)
type DisconnectionCallback func(conn *TcpConnection)
type MessageCallback func(conn *TcpConnection, data []byte)
type WriteCompleteCallback func(conn *TcpConnection)

type TcpConnection struct {
	conn              *net.TCPConn
	connId            string
	localAddr         *InetAddress
	peerAddr          *InetAddress
	buffer            *ByteBuffer
	connectionCallback ConnectionCallback
	disconnectionCallback DisconnectionCallback
	messageCallback    MessageCallback
	writeCompleteCallback WriteCompleteCallback
	closed            bool
	mu                sync.Mutex
}

func NewTcpConnection(conn *net.TCPConn, connId string) *TcpConnection {
	localAddr := NewInetAddressFromAddr(conn.LocalAddr().(*net.TCPAddr))
	peerAddr := NewInetAddressFromAddr(conn.RemoteAddr().(*net.TCPAddr))
	
	return &TcpConnection{
		conn:              conn,
		connId:            connId,
		localAddr:         localAddr,
		peerAddr:          peerAddr,
		buffer:            NewByteBuffer(),
		closed:            false,
	}
}

func (c *TcpConnection) SetConnectionCallback(cb ConnectionCallback) {
	c.connectionCallback = cb
}

func (c *TcpConnection) SetDisconnectionCallback(cb DisconnectionCallback) {
	c.disconnectionCallback = cb
}

func (c *TcpConnection) SetMessageCallback(cb MessageCallback) {
	c.messageCallback = cb
}

func (c *TcpConnection) SetWriteCompleteCallback(cb WriteCompleteCallback) {
	c.writeCompleteCallback = cb
}

func (c *TcpConnection) ConnId() string {
	return c.connId
}

func (c *TcpConnection) LocalAddr() *InetAddress {
	return c.localAddr
}

func (c *TcpConnection) PeerAddr() *InetAddress {
	return c.peerAddr
}

func (c *TcpConnection) IsClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}

func (c *TcpConnection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.closed {
		return nil
	}
	
	c.closed = true
	return c.conn.Close()
}

func (c *TcpConnection) Read() {
	buf := make([]byte, 4096)
	
	for {
		n, err := c.conn.Read(buf)
		if err != nil {
			base.GetLogger().Error("Read error", zap.Error(err))
			c.Close()
			if c.disconnectionCallback != nil {
				c.disconnectionCallback(c)
			}
			return
		}
		
		if n > 0 {
			data := make([]byte, n)
			copy(data, buf[:n])
			if c.messageCallback != nil {
				c.messageCallback(c, data)
			}
		}
	}
}

func (c *TcpConnection) Write(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.closed {
		return nil
	}
	
	n, err := c.conn.Write(data)
	if err != nil {
		return err
	}
	
	if n == len(data) && c.writeCompleteCallback != nil {
		c.writeCompleteCallback(c)
	}
	
	return nil
}

func (c *TcpConnection) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *TcpConnection) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

func (c *TcpConnection) SetKeepAlive(keepalive bool) error {
	return c.conn.SetKeepAlive(keepalive)
}

func (c *TcpConnection) SetKeepAlivePeriod(d time.Duration) error {
	return c.conn.SetKeepAlivePeriod(d)
}
