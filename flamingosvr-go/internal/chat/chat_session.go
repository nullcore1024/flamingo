package chat

import (
	"github.com/flamingo/server/internal/net"
	"sync"
)

const (
	CLIENT_TYPE_UNKNOWN = iota
	CLIENT_TYPE_PC
	CLIENT_TYPE_ANDROID
	CLIENT_TYPE_IOS
	CLIENT_TYPE_MAC
)

type ChatSession struct {
	conn      *net.TcpConnection
	sessionId string
	userId    int32
	clientType int32
	username  string
	nickname  string
	status    int32
	mu        sync.Mutex
}

func NewChatSession(conn *net.TcpConnection, sessionId string) *ChatSession {
	return &ChatSession{
		conn:      conn,
		sessionId: sessionId,
		userId:    0,
		clientType: CLIENT_TYPE_UNKNOWN,
		username:  "",
		nickname:  "",
		status:    0,
	}
}

func (s *ChatSession) SessionId() string {
	return s.sessionId
}

func (s *ChatSession) UserId() int32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.userId
}

func (s *ChatSession) SetUserId(userId int32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.userId = userId
}

func (s *ChatSession) ClientType() int32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.clientType
}

func (s *ChatSession) SetClientType(clientType int32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clientType = clientType
}

func (s *ChatSession) Username() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.username
}

func (s *ChatSession) SetUsername(username string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.username = username
}

func (s *ChatSession) Nickname() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.nickname
}

func (s *ChatSession) SetNickname(nickname string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nickname = nickname
}

func (s *ChatSession) Status() int32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status
}

func (s *ChatSession) SetStatus(status int32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = status
}

func (s *ChatSession) Conn() *net.TcpConnection {
	return s.conn
}

func (s *ChatSession) Send(data []byte) error {
	if s.conn != nil {
		return s.conn.Write(data)
	}
	return nil
}

func (s *ChatSession) Close() {
	if s.conn != nil {
		s.conn.Close()
	}
}

func (s *ChatSession) IsValid() bool {
	if s.conn == nil {
		return false
	}
	return !s.conn.IsClosed()
}
