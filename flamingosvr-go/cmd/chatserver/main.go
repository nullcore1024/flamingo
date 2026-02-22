package main

import (
	"flag"
	"github.com/flamingo/server/internal/base"
	"github.com/flamingo/server/internal/chat"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 解析命令行参数
	ip := flag.String("ip", "0.0.0.0", "Server IP address")
	port := flag.Int("port", 8000, "Server port")
	name := flag.String("name", "ChatServer", "Server name")
	flag.Parse()

	// 初始化日志
	logger := base.GetLogger()
	logger.Info("Starting ChatServer",
		zap.String("ip", *ip),
		zap.Int("port", *port),
		zap.String("name", *name))

	// 创建并初始化ChatServer
	chatServer := chat.NewChatServer()
	err := chatServer.Init(*ip, *port, *name)
	if err != nil {
		logger.Error("Failed to initialize ChatServer", zap.Error(err))
		os.Exit(1)
	}

	// 启用二进制日志
	chatServer.EnableLogPackageBinary(true)

	// 等待信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// 关闭服务器
	logger.Info("Shutting down ChatServer")
	chatServer.Uninit()
	logger.Info("ChatServer stopped")
}
