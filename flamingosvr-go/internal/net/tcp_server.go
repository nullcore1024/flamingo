package net

import (
	"github.com/flamingo/server/internal/base"
	"go.uber.org/zap"
	"net"
	"strconv"
	"sync"
)

type TcpServer struct {
	addr              *InetAddress
	listener          *net.TCPListener
	name              string
	connectionCallback ConnectionCallback
	disconnectionCallback DisconnectionCallback
	messageCallback    MessageCallback
	writeCompleteCallback WriteCompleteCallback
	connections       map[string]*TcpConnection
	mu                sync.Mutex
	nextConnId        int
	closed            bool
}

type Option int

const (
	KNoReusePort Option = iota
	KReusePort
)

func NewTcpServer(addr *InetAddress, name string, option Option) *TcpServer {
	return &TcpServer{
		addr:              addr,
		name:              name,
		connections:       make(map[string]*TcpConnection),
		nextConnId:        1,
		closed:            false,
	}
}

func (s *TcpServer) SetConnectionCallback(cb ConnectionCallback) {
	s.connectionCallback = cb
}

func (s *TcpServer) SetDisconnectionCallback(cb DisconnectionCallback) {
	s.disconnectionCallback = cb
}

func (s *TcpServer) SetMessageCallback(cb MessageCallback) {
	s.messageCallback = cb
}

func (s *TcpServer) SetWriteCompleteCallback(cb WriteCompleteCallback) {
	s.writeCompleteCallback = cb
}

func (s *TcpServer) Start() error {
	listener, err := net.ListenTCP("tcp", s.addr.Addr())
	if err != nil {
		base.GetLogger().Error("Listen error", zap.Error(err))
		return err
	}
	
	s.listener = listener
	s.closed = false
	
	base.GetLogger().Info("Server started", 
		zap.String("name", s.name),
		zap.String("address", s.addr.String()))
	
	// 启动接受连接的goroutine
	go s.acceptLoop()
	
	return nil
}

func (s *TcpServer) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.closed {
		return
	}
	
	s.closed = true
	
	// 关闭listener
	if s.listener != nil {
		s.listener.Close()
	}
	
	// 关闭所有连接
	for _, conn := range s.connections {
		conn.Close()
	}
	s.connections = make(map[string]*TcpConnection)
	
	base.GetLogger().Info("Server stopped", zap.String("name", s.name))
}

func (s *TcpServer) acceptLoop() {
	for {
		if s.closed {
			break
		}
		
		conn, err := s.listener.AcceptTCP()
		if err != nil {
			if !s.closed {
				base.GetLogger().Error("Accept error", zap.Error(err))
			}
			continue
		}
		
		// 创建新的TcpConnection
		connId := s.name + "#" + strconv.Itoa(s.nextConnId)
		s.nextConnId++
		
		tcpConn := NewTcpConnection(conn, connId)
	tcpConn.SetConnectionCallback(s.connectionCallback)
	tcpConn.SetDisconnectionCallback(s.disconnectionCallback)
	tcpConn.SetMessageCallback(s.messageCallback)
	tcpConn.SetWriteCompleteCallback(s.writeCompleteCallback)
		
		// 添加到连接管理
		s.mu.Lock()
		s.connections[connId] = tcpConn
		s.mu.Unlock()
		
		// 启动读取goroutine
		go tcpConn.Read()
		
		// 调用连接回调
		if s.connectionCallback != nil {
			s.connectionCallback(tcpConn)
		}
	}
}

func (s *TcpServer) RemoveConnection(connId string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if _, ok := s.connections[connId]; ok {
		delete(s.connections, connId)
		base.GetLogger().Info("Connection removed", zap.String("connId", connId))
	}
}

func (s *TcpServer) GetConnection(connId string) *TcpConnection {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	return s.connections[connId]
}

func (s *TcpServer) GetConnections() map[string]*TcpConnection {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// 返回副本
	connections := make(map[string]*TcpConnection)
	for k, v := range s.connections {
		connections[k] = v
	}
	return connections
}

func (s *TcpServer) ConnectionCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	return len(s.connections)
}
