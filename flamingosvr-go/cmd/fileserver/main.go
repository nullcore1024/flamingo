package main

import (
	"flag"
	"github.com/flamingo/server/internal/base"
	"github.com/flamingo/server/internal/file"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 解析命令行参数
	ip := flag.String("ip", "0.0.0.0", "Server IP address")
	port := flag.Int("port", 8001, "Server port")
	name := flag.String("name", "FileServer", "Server name")
	fileRoot := flag.String("fileRoot", "./files", "File root directory")
	flag.Parse()

	// 初始化日志
	logger := base.GetLogger()
	logger.Info("Starting FileServer",
		zap.String("ip", *ip),
		zap.Int("port", *port),
		zap.String("name", *name),
		zap.String("fileRoot", *fileRoot))

	// 创建并初始化FileServer
	fileServer := file.NewFileServer()
	err := fileServer.Init(*ip, *port, *name, *fileRoot)
	if err != nil {
		logger.Error("Failed to initialize FileServer", zap.Error(err))
		os.Exit(1)
	}

	// 等待信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// 关闭服务器
	logger.Info("Shutting down FileServer")
	fileServer.Uninit()
	logger.Info("FileServer stopped")
}
