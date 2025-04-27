package isqllite

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Covsj/gokit/pkg/idb/core"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// SQLiteConfig 定义SQLite数据库配置
type SQLiteConfig struct {
	core.BaseConfig        // 嵌入基础配置
	Path            string // 数据库文件路径
}

// DefaultConfig 返回默认配置
func DefaultConfig() SQLiteConfig {
	return SQLiteConfig{
		BaseConfig: core.NewBaseConfig(),
		Path:       "data.db",
	}
}

// Client SQLite客户端结构体
type Client struct {
	*core.BaseClient // 嵌入基础客户端
}

// New 创建新的SQLite客户端
func New(config SQLiteConfig) (*Client, error) {
	// 确保数据库目录存在
	dbDir := filepath.Dir(config.Path)
	if dbDir != "." && dbDir != "" {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return nil, fmt.Errorf("创建数据库目录失败: %w", err)
		}
	}

	// 使用核心包的日志适配器
	gormLogger := core.NewGormLogAdapter(config.LogLevel, config.SlowThreshold)

	// 连接数据库
	dialector := sqlite.Open(config.Path)
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取底层数据库连接失败: %w", err)
	}

	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)

	// 创建基础客户端
	baseClient := core.NewBaseClient(db)

	return &Client{
		BaseClient: baseClient,
	}, nil
}

// TableExists 检查表是否存在 (SQLite特定实现)
func (c *Client) TableExists(tableName string) (bool, error) {
	var count int64
	err := c.DB().Raw("SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?", tableName).Count(&count).Error
	return count > 0, err
}

// Backup 备份数据库到指定路径 (SQLite特定功能)
func (c *Client) Backup(backupPath string) error {
	// 获取原始数据库路径
	_, err := c.DB().DB()
	if err != nil {
		return err
	}

	// 使用原始SQL执行备份
	return c.DB().Exec("VACUUM INTO ?", backupPath).Error
}
