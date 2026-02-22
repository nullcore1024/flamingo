package main

import (
	"flag"
	"github.com/flamingo/server/internal/base"
	"github.com/flamingo/server/internal/img"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 解析命令行参数
	ip := flag.String("ip", "0.0.0.0", "Server IP address")
	port := flag.Int("port", 8002, "Server port")
	name := flag.String("name", "ImgServer", "Server name")
	imgRoot := flag.String("imgRoot", "./images", "Image root directory")
	flag.Parse()

	// 初始化日志
	logger := base.GetLogger()
	logger.Info("Starting ImgServer",
		zap.String("ip", *ip),
		zap.Int("port", *port),
		zap.String("name", *name),
		zap.String("imgRoot", *imgRoot))

	// 创建并初始化ImgServer
	imgServer := img.NewImgServer()
	err := imgServer.Init(*ip, *port, *name, *imgRoot)
	if err != nil {
		logger.Error("Failed to initialize ImgServer", zap.Error(err))
		os.Exit(1)
	}

	// 等待信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// 关闭服务器
	logger.Info("Shutting down ImgServer")
	imgServer.Uninit()
	logger.Info("ImgServer stopped")
}
