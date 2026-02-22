package chat

import (
	"github.com/flamingo/server/internal/base"
	"github.com/flamingo/server/internal/net"
	"go.uber.org/zap"
	"strconv"
	"sync"
)

type ChatServer struct {
	server        *net.TcpServer
	sessions      map[string]*ChatSession
	userSessions  map[int32][]*ChatSession
	mu            sync.RWMutex
	sessionIdGen  int
	logBinary     bool
}

type StoredUserInfo struct {
	UserId   int32
	Username string
	Password string
	Nickname string
}

func NewChatServer() *ChatServer {
	return &ChatServer{
		sessions:     make(map[string]*ChatSession),
		userSessions: make(map[int32][]*ChatSession),
		sessionIdGen: 1,
		logBinary:    false,
	}
}

func (cs *ChatServer) Init(ip string, port int, name string) error {
	addr, err := net.NewInetAddress(ip, port)
	if err != nil {
		base.GetLogger().Error("Create address error", zap.Error(err))
		return err
	}

	cs.server = net.NewTcpServer(addr, name, net.KReusePort)
	cs.server.SetConnectionCallback(cs.onConnection)
	cs.server.SetDisconnectionCallback(cs.onDisconnection)
	cs.server.SetMessageCallback(cs.onMessage)

	return cs.server.Start()
}

func (cs *ChatServer) Uninit() {
	if cs.server != nil {
		cs.server.Stop()
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	for _, session := range cs.sessions {
		session.Close()
	}
	cs.sessions = make(map[string]*ChatSession)
	cs.userSessions = make(map[int32][]*ChatSession)
}

func (cs *ChatServer) EnableLogPackageBinary(enable bool) {
	cs.logBinary = enable
}

func (cs *ChatServer) IsLogPackageBinaryEnabled() bool {
	return cs.logBinary
}

func (cs *ChatServer) GetSessions() []*ChatSession {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	sessions := make([]*ChatSession, 0, len(cs.sessions))
	for _, session := range cs.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func (cs *ChatServer) GetSessionByUserIdAndClientType(userId, clientType int32) *ChatSession {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	if sessions, ok := cs.userSessions[userId]; ok {
		for _, session := range sessions {
			if session.ClientType() == clientType {
				return session
			}
		}
	}
	return nil
}

func (cs *ChatServer) GetSessionsByUserId(userId int32) []*ChatSession {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	if sessions, ok := cs.userSessions[userId]; ok {
		// 返回副本
		result := make([]*ChatSession, len(sessions))
		copy(result, sessions)
		return result
	}
	return nil
}

func (cs *ChatServer) GetUserStatusByUserId(userId int32) int32 {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	if sessions, ok := cs.userSessions[userId]; ok && len(sessions) > 0 {
		return sessions[0].Status()
	}
	return 0
}

func (cs *ChatServer) GetUserClientTypeByUserId(userId int32) int32 {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	if sessions, ok := cs.userSessions[userId]; ok && len(sessions) > 0 {
		return sessions[0].ClientType()
	}
	return 0
}

func (cs *ChatServer) onConnection(conn *net.TcpConnection) {
	cs.mu.Lock()
	sessionId := "session_" + strconv.Itoa(cs.sessionIdGen)
	cs.sessionIdGen++
	cs.mu.Unlock()

	session := NewChatSession(conn, sessionId)

	cs.mu.Lock()
	cs.sessions[sessionId] = session
	cs.mu.Unlock()

	base.GetLogger().Info("New connection",
		zap.String("sessionId", sessionId),
		zap.String("remoteAddr", conn.PeerAddr().String()))
}

func (cs *ChatServer) onDisconnection(conn *net.TcpConnection) {
	cs.mu.RLock()
	var sessionId string
	for id, session := range cs.sessions {
		if session.Conn() == conn {
			sessionId = id
			break
		}
	}
	cs.mu.RUnlock()

	if sessionId != "" {
		cs.removeSession(sessionId)
	}
}

func (cs *ChatServer) onMessage(conn *net.TcpConnection, data []byte) {
	cs.mu.RLock()
	var session *ChatSession
	for _, s := range cs.sessions {
		if s.Conn() == conn {
			session = s
			break
		}
	}
	cs.mu.RUnlock()

	if session == nil {
		base.GetLogger().Warn("Message from unknown connection",
			zap.String("remoteAddr", conn.PeerAddr().String()))
		return
	}

	// 处理消息
	cs.handleMessage(session, data)
}

func (cs *ChatServer) handleMessage(session *ChatSession, data []byte) {
	// 这里实现具体的消息处理逻辑
	// 根据协议解析消息，处理业务逻辑
	base.GetLogger().Info("Received message",
		zap.String("sessionId", session.SessionId()),
		zap.Int("length", len(data)))

	// 示例：回显消息
	session.Send(data)
}

func (cs *ChatServer) removeSession(sessionId string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	session, ok := cs.sessions[sessionId]
	if !ok {
		return
	}

	// 从用户会话映射中移除
	if session.UserId() > 0 {
		if sessions, ok := cs.userSessions[session.UserId()]; ok {
			for i, s := range sessions {
				if s.SessionId() == sessionId {
					cs.userSessions[session.UserId()] = append(sessions[:i], sessions[i+1:]...)
					break
				}
			}
			// 如果用户没有会话了，从映射中删除
			if len(cs.userSessions[session.UserId()]) == 0 {
				delete(cs.userSessions, session.UserId())
			}
		}
	}

	// 从会话映射中移除
	delete(cs.sessions, sessionId)
	session.Close()

	base.GetLogger().Info("Session removed",
		zap.String("sessionId", sessionId),
		zap.Int32("userId", session.UserId()))
}

func (cs *ChatServer) AddUserSession(userId int32, session *ChatSession) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.userSessions[userId] = append(cs.userSessions[userId], session)
}

func (cs *ChatServer) BroadcastToUser(userId int32, message []byte) {
	sessions := cs.GetSessionsByUserId(userId)
	for _, session := range sessions {
		session.Send(message)
	}
}

func (cs *ChatServer) BroadcastToAll(message []byte) {
	cs.mu.RLock()
	sessions := make([]*ChatSession, 0, len(cs.sessions))
	for _, session := range cs.sessions {
		sessions = append(sessions, session)
	}
	cs.mu.RUnlock()

	for _, session := range sessions {
		session.Send(message)
	}
}
