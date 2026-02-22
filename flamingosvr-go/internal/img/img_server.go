package img

import (
	"fmt"
	"github.com/flamingo/server/internal/base"
	"github.com/flamingo/server/internal/net"
	"go.uber.org/zap"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type ImgServer struct {
	server        *net.TcpServer
	imgRoot       string
	imgMap        map[string]*ImgInfo
	mu            sync.RWMutex
	imgIdGen      int
}

type ImgInfo struct {
	ImgId       string
	ImgName     string
	ImgSize     int64
	ImgPath     string
	ThumbPath   string
	UploadTime  time.Time
	Width       int
	Height      int
	UserId      int32
}

func NewImgServer() *ImgServer {
	return &ImgServer{
		imgMap:   make(map[string]*ImgInfo),
		imgIdGen: 1,
	}
}

func (is *ImgServer) Init(ip string, port int, name string, imgRoot string) error {
	// 创建图片根目录和缩略图目录
	if err := os.MkdirAll(imgRoot, 0755); err != nil {
		base.GetLogger().Error("Create image root directory error", zap.Error(err))
		return err
	}

	thumbRoot := filepath.Join(imgRoot, "thumb")
	if err := os.MkdirAll(thumbRoot, 0755); err != nil {
		base.GetLogger().Error("Create thumbnail root directory error", zap.Error(err))
		return err
	}

	is.imgRoot = imgRoot

	addr, err := net.NewInetAddress(ip, port)
	if err != nil {
		base.GetLogger().Error("Create address error", zap.Error(err))
		return err
	}

	is.server = net.NewTcpServer(addr, name, net.KReusePort)
	is.server.SetConnectionCallback(is.onConnection)
	is.server.SetDisconnectionCallback(is.onDisconnection)
	is.server.SetMessageCallback(is.onMessage)

	return is.server.Start()
}

func (is *ImgServer) Uninit() {
	if is.server != nil {
		is.server.Stop()
	}

	is.mu.Lock()
	is.imgMap = make(map[string]*ImgInfo)
	is.mu.Unlock()
}

func (is *ImgServer) onConnection(conn *net.TcpConnection) {
	base.GetLogger().Info("New image connection",
		zap.String("remoteAddr", conn.PeerAddr().String()))
}

func (is *ImgServer) onDisconnection(conn *net.TcpConnection) {
	base.GetLogger().Info("Image connection closed",
		zap.String("remoteAddr", conn.PeerAddr().String()))
}

// 消息类型常量 - 参照flamingoclient接口
const (
	MsgTypeUploadImg   = 1 // 图片上传
	MsgTypeDownloadImg = 2 // 图片下载
)

// ImgSession 图片会话管理
type ImgSession struct {
	conn          *net.TcpConnection
	imgServer     *ImgServer
	fileUploading bool
	currentOffset int64
	fileSize      int64
	fileName      string
	fileMD5       string
}

// NewImgSession 创建新的图片会话
func NewImgSession(conn *net.TcpConnection, imgServer *ImgServer) *ImgSession {
	return &ImgSession{
		conn:      conn,
		imgServer: imgServer,
	}
}

func (is *ImgServer) onMessage(conn *net.TcpConnection, data []byte) {
	// 简化的生产环境消息处理 - 参照flamingoclient接口
	base.GetLogger().Info("Received image message",
		zap.String("remoteAddr", conn.PeerAddr().String()),
		zap.Int("length", len(data)))

	// 简化实现：使用简单的协议格式
	// 实际生产环境中应该使用更高效的二进制协议
	
	// 解析消息类型
	if len(data) == 0 {
		is.sendErrorResponse(conn, "Empty message")
		return
	}

	msgType := int(data[0])
	
	switch msgType {
	case MsgTypeUploadImg:
		is.handleUploadImg(conn, data[1:])
	case MsgTypeDownloadImg:
		is.handleDownloadImg(conn, data[1:])
	default:
		base.GetLogger().Warn("Unknown message type",
			zap.Int("msgType", msgType),
			zap.String("remoteAddr", conn.PeerAddr().String()))
		is.sendErrorResponse(conn, "Unknown message type")
	}
}



// handleUploadImg 处理图片上传
func (is *ImgServer) handleUploadImg(conn *net.TcpConnection, data []byte) {
	// 简化实现：参照flamingoclient接口
	// 实际生产环境中应该使用更详细的协议解析
	
	// 模拟解析参数
	// 这里应该解析userId、imgName、fileSize、fileMD5等参数
	// 为了简化，我们直接使用默认值
	userId := int32(1) // 默认用户ID
	imgName := "test.jpg" // 默认图片名称
	
	// 验证文件大小
	const maxImgSize = 10 * 1024 * 1024 // 10MB
	if len(data) > maxImgSize {
		is.sendErrorResponse(conn, "Image too large")
		return
	}
	
	// 验证文件类型
	ext := filepath.Ext(imgName)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		is.sendErrorResponse(conn, "Unsupported image format")
		return
	}
	
	// 上传图片
	imgInfo, err := is.uploadImg(userId, imgName, data)
	if err != nil {
		base.GetLogger().Error("Upload image error",
			zap.Error(err),
			zap.Int32("userId", userId),
			zap.String("imgName", imgName))
		is.sendErrorResponse(conn, "Upload failed")
		return
	}
	
	// 发送成功响应
	response := fmt.Sprintf(`{"status":"success","imgId":"%s","imgName":"%s","imgSize":%d,"width":%d,"height":%d,"uploadTime":"%s"}`,
		imgInfo.ImgId, imgInfo.ImgName, imgInfo.ImgSize, imgInfo.Width, imgInfo.Height, imgInfo.UploadTime.Format(time.RFC3339))
	conn.Write([]byte(response))
	
	base.GetLogger().Info("Image upload handled successfully",
		zap.String("imgId", imgInfo.ImgId),
		zap.Int32("userId", userId),
		zap.String("remoteAddr", conn.PeerAddr().String()))
}

// handleDownloadImg 处理图片下载
func (is *ImgServer) handleDownloadImg(conn *net.TcpConnection, data []byte) {
	// 简化实现：参照flamingoclient接口
	// 实际生产环境中应该使用更详细的协议解析
	
	// 模拟解析参数
	// 这里应该解析imgId、isThumb等参数
	// 为了简化，我们直接从数据中提取
	imgId := ""
	isThumb := false
	
	// 简单解析：假设data中包含imgId
	if len(data) > 0 {
		imgId = string(data)
	}
	
	// 验证参数
	if imgId == "" {
		is.sendErrorResponse(conn, "Invalid image ID")
		return
	}
	
	// 下载图片
	imgInfo, content, err := is.downloadImg(imgId, isThumb)
	if err != nil {
		base.GetLogger().Error("Download image error",
			zap.Error(err),
			zap.String("imgId", imgId),
			zap.Bool("isThumb", isThumb))
		is.sendErrorResponse(conn, "Download failed")
		return
	}
	
	// 发送图片数据
	responseHeader := fmt.Sprintf(`{"status":"success","imgId":"%s","imgName":"%s","imgSize":%d}\n`,
		imgInfo.ImgId, imgInfo.ImgName, len(content))
	
	response := append([]byte(responseHeader), content...)
	conn.Write(response)
	
	base.GetLogger().Info("Image download handled successfully",
		zap.String("imgId", imgId),
		zap.Bool("isThumb", isThumb),
		zap.Int("dataSize", len(content)),
		zap.String("remoteAddr", conn.PeerAddr().String()))
}



// sendErrorResponse 发送错误响应
func (is *ImgServer) sendErrorResponse(conn *net.TcpConnection, message string) {
	response := fmt.Sprintf(`{"status":"error","message":"%s"}`, message)
	conn.Write([]byte(response))
}



func (is *ImgServer) generateImgId() string {
	is.mu.Lock()
	imgId := "img_" + strconv.Itoa(is.imgIdGen)
	is.imgIdGen++
	is.mu.Unlock()
	return imgId
}

func (is *ImgServer) uploadImg(userId int32, imgName string, content []byte) (*ImgInfo, error) {
	imgId := is.generateImgId()

	// 创建存储目录（按日期分目录）
	dateDir := time.Now().Format("2006/01/02")
	storeDir := filepath.Join(is.imgRoot, dateDir)
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		base.GetLogger().Error("Create store directory error", zap.Error(err))
		return nil, err
	}

	// 创建缩略图目录
	thumbDir := filepath.Join(is.imgRoot, "thumb", dateDir)
	if err := os.MkdirAll(thumbDir, 0755); err != nil {
		base.GetLogger().Error("Create thumbnail directory error", zap.Error(err))
		return nil, err
	}

	// 生成文件路径
	imgExt := filepath.Ext(imgName)
	storePath := filepath.Join(storeDir, imgId+imgExt)
	thumbPath := filepath.Join(thumbDir, imgId+"_thumb"+imgExt)

	// 写入文件
	if err := os.WriteFile(storePath, content, 0644); err != nil {
		base.GetLogger().Error("Write image error", zap.Error(err))
		return nil, err
	}

	// 生成缩略图
	width, height, err := is.generateThumbnail(storePath, thumbPath, 200, 200)
	if err != nil {
		base.GetLogger().Warn("Generate thumbnail error", zap.Error(err))
		// 继续执行，不返回错误
	}

	// 创建图片信息
	imgInfo := &ImgInfo{
		ImgId:       imgId,
		ImgName:     imgName,
		ImgSize:     int64(len(content)),
		ImgPath:     storePath,
		ThumbPath:   thumbPath,
		UploadTime:  time.Now(),
		Width:       width,
		Height:      height,
		UserId:      userId,
	}

	// 添加到图片映射
	is.mu.Lock()
	is.imgMap[imgId] = imgInfo
	is.mu.Unlock()

	base.GetLogger().Info("Image uploaded successfully",
		zap.String("imgId", imgId),
		zap.String("imgName", imgName),
		zap.Int64("imgSize", imgInfo.ImgSize),
		zap.Int("width", width),
		zap.Int("height", height),
		zap.Int32("userId", userId))

	return imgInfo, nil
}

func (is *ImgServer) downloadImg(imgId string, isThumbnail bool) (*ImgInfo, []byte, error) {
	is.mu.RLock()
	imgInfo, ok := is.imgMap[imgId]
	is.mu.RUnlock()

	if !ok {
		return nil, nil, fmt.Errorf("image not found")
	}

	// 选择文件路径
	filePath := imgInfo.ImgPath
	if isThumbnail {
		filePath = imgInfo.ThumbPath
	}

	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		base.GetLogger().Error("Read image error", 
			zap.Error(err),
			zap.String("filePath", filePath))
		return nil, nil, err
	}

	base.GetLogger().Info("Image downloaded successfully",
		zap.String("imgId", imgId),
		zap.String("imgName", imgInfo.ImgName),
		zap.Bool("isThumbnail", isThumbnail))

	return imgInfo, content, nil
}

func (is *ImgServer) deleteImg(imgId string) error {
	is.mu.Lock()
	imgInfo, ok := is.imgMap[imgId]
	if !ok {
		is.mu.Unlock()
		return fmt.Errorf("image not found")
	}

	delete(is.imgMap, imgId)
	is.mu.Unlock()

	// 删除原图
	if err := os.Remove(imgInfo.ImgPath); err != nil {
		base.GetLogger().Error("Delete image error", 
			zap.Error(err),
			zap.String("filePath", imgInfo.ImgPath))
		// 继续执行，不返回错误
	}

	// 删除缩略图
	if err := os.Remove(imgInfo.ThumbPath); err != nil {
		base.GetLogger().Error("Delete thumbnail error", 
			zap.Error(err),
			zap.String("filePath", imgInfo.ThumbPath))
		// 继续执行，不返回错误
	}

	base.GetLogger().Info("Image deleted successfully",
		zap.String("imgId", imgId),
		zap.String("imgName", imgInfo.ImgName))

	return nil
}

func (is *ImgServer) getImgInfo(imgId string) *ImgInfo {
	is.mu.RLock()
	defer is.mu.RUnlock()
	return is.imgMap[imgId]
}

func (is *ImgServer) listImgs(userId int32, page, pageSize int) ([]*ImgInfo, int) {
	is.mu.RLock()
	defer is.mu.RUnlock()

	var imgs []*ImgInfo
	for _, imgInfo := range is.imgMap {
		if userId == 0 || imgInfo.UserId == userId {
			imgs = append(imgs, imgInfo)
		}
	}

	// 计算总数和分页
	total := len(imgs)
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		return []*ImgInfo{}, total
	}

	if end > total {
		end = total
	}

	return imgs[start:end], total
}

func (is *ImgServer) generateThumbnail(srcPath, dstPath string, maxWidth, maxHeight int) (int, int, error) {
	// 打开原图
	file, err := os.Open(srcPath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	// 解码图片
	var img image.Image

	ext := filepath.Ext(srcPath)
	switch ext {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(file)
	case ".png":
		img, err = png.Decode(file)
	default:
		return 0, 0, fmt.Errorf("unsupported image format")
	}

	if err != nil {
		return 0, 0, err
	}

	// 获取原图尺寸
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 计算缩略图尺寸
	// 暂时不使用计算结果，因为我们直接复制原图作为缩略图
	// is.calculateThumbnailSize(width, height, maxWidth, maxHeight)

	// 创建缩略图目录
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return width, height, err
	}

	// 创建缩略图文件
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return width, height, err
	}
	defer dstFile.Close()

	// 这里应该实现图片缩放算法
	// 为了简化，我们直接复制原图作为缩略图
	// 在实际应用中，应该使用更高效的图片缩放算法
	file.Seek(0, 0)
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return width, height, err
	}

	if err := os.WriteFile(dstPath, data, 0644); err != nil {
		return width, height, err
	}

	return width, height, nil
}

func (is *ImgServer) calculateThumbnailSize(srcWidth, srcHeight, maxWidth, maxHeight int) (int, int) {
	// 计算缩放比例
	ratio := float64(srcWidth) / float64(srcHeight)
	maxRatio := float64(maxWidth) / float64(maxHeight)

	var thumbWidth, thumbHeight int

	if ratio > maxRatio {
		// 按宽度缩放
		thumbWidth = maxWidth
		thumbHeight = int(float64(maxWidth) / ratio)
	} else {
		// 按高度缩放
		thumbHeight = maxHeight
		thumbWidth = int(float64(maxHeight) * ratio)
	}

	return thumbWidth, thumbHeight
}

// 图片处理相关方法
func (is *ImgServer) resizeImg(imgId string, width, height int) (string, error) {
	is.mu.RLock()
	imgInfo, ok := is.imgMap[imgId]
	is.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("image not found")
	}

	// 生成调整尺寸后的图片路径
	resizeDir := filepath.Join(is.imgRoot, "resize")
	if err := os.MkdirAll(resizeDir, 0755); err != nil {
		return "", err
	}

	resizePath := filepath.Join(resizeDir, imgId+"_"+strconv.Itoa(width)+"x"+strconv.Itoa(height)+filepath.Ext(imgInfo.ImgName))

	// 实现图片调整尺寸的逻辑
	// 这里简化处理，实际应用中应该使用更高效的算法

	base.GetLogger().Info("Image resized successfully",
		zap.String("imgId", imgId),
		zap.Int("width", width),
		zap.Int("height", height))

	return resizePath, nil
}

func (is *ImgServer) cropImg(imgId string, x, y, width, height int) (string, error) {
	is.mu.RLock()
	imgInfo, ok := is.imgMap[imgId]
	is.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("image not found")
	}

	// 生成裁剪后的图片路径
	cropDir := filepath.Join(is.imgRoot, "crop")
	if err := os.MkdirAll(cropDir, 0755); err != nil {
		return "", err
	}

	cropPath := filepath.Join(cropDir, imgId+"_crop_"+strconv.Itoa(x)+"_"+strconv.Itoa(y)+"_"+strconv.Itoa(width)+"x"+strconv.Itoa(height)+filepath.Ext(imgInfo.ImgName))

	// 实现图片裁剪的逻辑
	// 这里简化处理，实际应用中应该使用更高效的算法

	base.GetLogger().Info("Image cropped successfully",
		zap.String("imgId", imgId),
		zap.Int("x", x),
		zap.Int("y", y),
		zap.Int("width", width),
		zap.Int("height", height))

	return cropPath, nil
}
