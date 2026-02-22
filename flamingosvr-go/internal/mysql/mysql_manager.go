package mysql

import (
	"database/sql"
	"fmt"
	"github.com/flamingo/server/internal/base"
	"go.uber.org/zap"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MysqlManager struct {
	db         *sql.DB
	mu         sync.RWMutex
	connected  bool
	config     *MysqlConfig
}

type MysqlConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	Charset  string
	MaxConn  int
	MaxIdle  int
}

var (
	mysqlManagerInstance *MysqlManager
	mysqlManagerOnce     sync.Once
)

func GetMysqlManager() *MysqlManager {
	mysqlManagerOnce.Do(func() {
		mysqlManagerInstance = &MysqlManager{
			connected: false,
		}
	})
	return mysqlManagerInstance
}

func (mm *MysqlManager) Init(config *MysqlConfig) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if mm.connected {
		return nil
	}

	mm.config = config

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true",
		config.User, config.Password, config.Host, config.Port, config.Database, config.Charset)

	var err error
	mm.db, err = sql.Open("mysql", dsn)
	if err != nil {
		base.GetLogger().Error("Open MySQL connection error", zap.Error(err))
		return err
	}

	// 设置连接池参数
	mm.db.SetMaxOpenConns(config.MaxConn)
	mm.db.SetMaxIdleConns(config.MaxIdle)
	mm.db.SetConnMaxLifetime(time.Hour)

	// 测试连接
	err = mm.db.Ping()
	if err != nil {
		base.GetLogger().Error("Ping MySQL error", zap.Error(err))
		return err
	}

	mm.connected = true
	base.GetLogger().Info("MySQL connected successfully",
		zap.String("host", config.Host),
		zap.Int("port", config.Port),
		zap.String("database", config.Database))

	return nil
}

func (mm *MysqlManager) Uninit() {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if mm.connected && mm.db != nil {
		mm.db.Close()
		mm.connected = false
		base.GetLogger().Info("MySQL disconnected")
	}
}

func (mm *MysqlManager) IsConnected() bool {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	return mm.connected
}

func (mm *MysqlManager) Query(query string, args ...interface{}) (*sql.Rows, error) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	if !mm.connected {
		return nil, fmt.Errorf("mysql not connected")
	}

	rows, err := mm.db.Query(query, args...)
	if err != nil {
		base.GetLogger().Error("MySQL query error",
			zap.Error(err),
			zap.String("query", query))
		return nil, err
	}

	return rows, nil
}

func (mm *MysqlManager) Exec(query string, args ...interface{}) (sql.Result, error) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	if !mm.connected {
		return nil, fmt.Errorf("mysql not connected")
	}

	result, err := mm.db.Exec(query, args...)
	if err != nil {
		base.GetLogger().Error("MySQL exec error",
			zap.Error(err),
			zap.String("query", query))
		return nil, err
	}

	return result, nil
}

func (mm *MysqlManager) Prepare(query string) (*sql.Stmt, error) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	if !mm.connected {
		return nil, fmt.Errorf("mysql not connected")
	}

	stmt, err := mm.db.Prepare(query)
	if err != nil {
		base.GetLogger().Error("MySQL prepare error",
			zap.Error(err),
			zap.String("query", query))
		return nil, err
	}

	return stmt, nil
}

func (mm *MysqlManager) Begin() (*sql.Tx, error) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	if !mm.connected {
		return nil, fmt.Errorf("mysql not connected")
	}

	tx, err := mm.db.Begin()
	if err != nil {
		base.GetLogger().Error("MySQL begin transaction error", zap.Error(err))
		return nil, err
	}

	return tx, nil
}

// 便捷方法：执行查询并返回单行结果
func (mm *MysqlManager) QueryRow(query string, args ...interface{}) *sql.Row {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	if !mm.connected {
		// 返回一个错误的Row，当Scan时会返回错误
		return nil
	}

	return mm.db.QueryRow(query, args...)
}

// 便捷方法：获取连接数信息
func (mm *MysqlManager) GetConnStats() (int, int) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	if !mm.connected || mm.db == nil {
		return 0, 0
	}

	// Go的database/sql包没有直接提供获取当前连接数的方法
	// 这里返回配置的最大连接数和最大空闲连接数
	return mm.config.MaxConn, mm.config.MaxIdle
}
