package file

import (
	"fmt"
	"github.com/flamingo/server/internal/base"
	"github.com/flamingo/server/internal/net"
	"go.uber.org/zap"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type FileServer struct {
	server        *net.TcpServer
	fileRoot      string
	fileMap       map[string]*FileInfo
	mu            sync.RWMutex
	fileIdGen     int
}

type FileInfo struct {
	FileId      string
	FileName    string
	FileSize    int64
	FilePath    string
	UploadTime  time.Time
	ContentType string
	UserId      int32
}

func NewFileServer() *FileServer {
	return &FileServer{
		fileMap:   make(map[string]*FileInfo),
		fileIdGen: 1,
	}
}

func (fs *FileServer) Init(ip string, port int, name string, fileRoot string) error {
	// 创建文件根目录
	if err := os.MkdirAll(fileRoot, 0755); err != nil {
		base.GetLogger().Error("Create file root directory error", zap.Error(err))
		return err
	}

	fs.fileRoot = fileRoot

	addr, err := net.NewInetAddress(ip, port)
	if err != nil {
		base.GetLogger().Error("Create address error", zap.Error(err))
		return err
	}

	fs.server = net.NewTcpServer(addr, name, net.KReusePort)
	fs.server.SetConnectionCallback(fs.onConnection)
	fs.server.SetDisconnectionCallback(fs.onDisconnection)
	fs.server.SetMessageCallback(fs.onMessage)

	return fs.server.Start()
}

func (fs *FileServer) Uninit() {
	if fs.server != nil {
		fs.server.Stop()
	}

	fs.mu.Lock()
	fs.fileMap = make(map[string]*FileInfo)
	fs.mu.Unlock()
}

func (fs *FileServer) onConnection(conn *net.TcpConnection) {
	base.GetLogger().Info("New file connection",
		zap.String("remoteAddr", conn.PeerAddr().String()))
}

func (fs *FileServer) onDisconnection(conn *net.TcpConnection) {
	base.GetLogger().Info("File connection closed",
		zap.String("remoteAddr", conn.PeerAddr().String()))
}

func (fs *FileServer) onMessage(conn *net.TcpConnection, data []byte) {
	// 这里实现文件消息的处理逻辑
	// 根据协议解析消息，处理文件上传、下载等请求
	base.GetLogger().Info("Received file message",
		zap.String("remoteAddr", conn.PeerAddr().String()),
		zap.Int("length", len(data)))
}

func (fs *FileServer) generateFileId() string {
	fs.mu.Lock()
	fileId := "file_" + strconv.Itoa(fs.fileIdGen)
	fs.fileIdGen++
	fs.mu.Unlock()
	return fileId
}

func (fs *FileServer) uploadFile(userId int32, fileName string, content []byte, contentType string) (*FileInfo, error) {
	fileId := fs.generateFileId()

	// 创建存储目录（按日期分目录）
	dateDir := time.Now().Format("2006/01/02")
	storeDir := filepath.Join(fs.fileRoot, dateDir)
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		base.GetLogger().Error("Create store directory error", zap.Error(err))
		return nil, err
	}

	// 生成文件路径
	fileExt := filepath.Ext(fileName)
	storePath := filepath.Join(storeDir, fileId+fileExt)

	// 写入文件
	if err := os.WriteFile(storePath, content, 0644); err != nil {
		base.GetLogger().Error("Write file error", zap.Error(err))
		return nil, err
	}

	// 创建文件信息
	fileInfo := &FileInfo{
		FileId:      fileId,
		FileName:    fileName,
		FileSize:    int64(len(content)),
		FilePath:    storePath,
		UploadTime:  time.Now(),
		ContentType: contentType,
		UserId:      userId,
	}

	// 添加到文件映射
	fs.mu.Lock()
	fs.fileMap[fileId] = fileInfo
	fs.mu.Unlock()

	base.GetLogger().Info("File uploaded successfully",
		zap.String("fileId", fileId),
		zap.String("fileName", fileName),
		zap.Int64("fileSize", fileInfo.FileSize),
		zap.Int32("userId", userId))

	return fileInfo, nil
}

func (fs *FileServer) downloadFile(fileId string) (*FileInfo, []byte, error) {
	fs.mu.RLock()
	fileInfo, ok := fs.fileMap[fileId]
	fs.mu.RUnlock()

	if !ok {
		return nil, nil, fmt.Errorf("file not found")
	}

	// 读取文件内容
	content, err := os.ReadFile(fileInfo.FilePath)
	if err != nil {
		base.GetLogger().Error("Read file error", 
			zap.Error(err),
			zap.String("filePath", fileInfo.FilePath))
		return nil, nil, err
	}

	base.GetLogger().Info("File downloaded successfully",
		zap.String("fileId", fileId),
		zap.String("fileName", fileInfo.FileName),
		zap.Int64("fileSize", fileInfo.FileSize))

	return fileInfo, content, nil
}

func (fs *FileServer) deleteFile(fileId string) error {
	fs.mu.Lock()
	fileInfo, ok := fs.fileMap[fileId]
	if !ok {
		fs.mu.Unlock()
		return fmt.Errorf("file not found")
	}

	delete(fs.fileMap, fileId)
	fs.mu.Unlock()

	// 删除文件
	if err := os.Remove(fileInfo.FilePath); err != nil {
		base.GetLogger().Error("Delete file error", 
			zap.Error(err),
			zap.String("filePath", fileInfo.FilePath))
		return err
	}

	base.GetLogger().Info("File deleted successfully",
		zap.String("fileId", fileId),
		zap.String("fileName", fileInfo.FileName))

	return nil
}

func (fs *FileServer) getFileInfo(fileId string) *FileInfo {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.fileMap[fileId]
}

func (fs *FileServer) listFiles(userId int32, page, pageSize int) ([]*FileInfo, int) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var files []*FileInfo
	for _, fileInfo := range fs.fileMap {
		if userId == 0 || fileInfo.UserId == userId {
			files = append(files, fileInfo)
		}
	}

	// 计算总数和分页
	total := len(files)
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		return []*FileInfo{}, total
	}

	if end > total {
		end = total
	}

	return files[start:end], total
}

// 大文件上传相关方法
func (fs *FileServer) initLargeFileUpload(userId int32, fileName string, fileSize int64, contentType string) (string, error) {
	fileId := fs.generateFileId()

	// 创建存储目录
	dateDir := time.Now().Format("2006/01/02")
	storeDir := filepath.Join(fs.fileRoot, dateDir)
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		base.GetLogger().Error("Create store directory error", zap.Error(err))
		return "", err
	}

	// 生成文件路径
	fileExt := filepath.Ext(fileName)
	storePath := filepath.Join(storeDir, fileId+fileExt)

	// 创建空文件
	file, err := os.Create(storePath)
	if err != nil {
		base.GetLogger().Error("Create file error", zap.Error(err))
		return "", err
	}
	file.Close()

	// 创建文件信息
	fileInfo := &FileInfo{
		FileId:      fileId,
		FileName:    fileName,
		FileSize:    fileSize,
		FilePath:    storePath,
		UploadTime:  time.Now(),
		ContentType: contentType,
		UserId:      userId,
	}

	// 添加到文件映射
	fs.mu.Lock()
	fs.fileMap[fileId] = fileInfo
	fs.mu.Unlock()

	return fileId, nil
}

func (fs *FileServer) uploadFileChunk(fileId string, offset int64, chunk []byte) error {
	fs.mu.RLock()
	fileInfo, ok := fs.fileMap[fileId]
	fs.mu.RUnlock()

	if !ok {
		return fmt.Errorf("file not found")
	}

	// 打开文件
	file, err := os.OpenFile(fileInfo.FilePath, os.O_WRONLY, 0644)
	if err != nil {
		base.GetLogger().Error("Open file error", zap.Error(err))
		return err
	}
	defer file.Close()

	// 定位到偏移量
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		base.GetLogger().Error("Seek file error", zap.Error(err))
		return err
	}

	// 写入数据
	if _, err := file.Write(chunk); err != nil {
		base.GetLogger().Error("Write chunk error", zap.Error(err))
		return err
	}

	return nil
}

func (fs *FileServer) completeLargeFileUpload(fileId string) error {
	fs.mu.RLock()
	fileInfo, ok := fs.fileMap[fileId]
	fs.mu.RUnlock()

	if !ok {
		return fmt.Errorf("file not found")
	}

	// 验证文件大小
	statInfo, err := os.Stat(fileInfo.FilePath)
	if err != nil {
		base.GetLogger().Error("Stat file error", zap.Error(err))
		return err
	}

	base.GetLogger().Info("Large file upload completed",
		zap.String("fileId", fileId),
		zap.Int64("fileSize", statInfo.Size()))

	return nil
}
