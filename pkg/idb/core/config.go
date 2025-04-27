package core

import (
	"time"

	"gorm.io/gorm/logger"
)

// DBConfig 数据库配置的通用接口
type DBConfig interface {
	// 获取日志级别
	GetLogLevel() logger.LogLevel

	// 获取慢查询阈值
	GetSlowThreshold() time.Duration

	// 获取连接池配置
	GetMaxOpenConns() int
	GetMaxIdleConns() int
}

// BaseConfig 基础配置结构体，可被嵌入到具体数据库配置中
type BaseConfig struct {
	LogLevel      logger.LogLevel // 日志级别
	SlowThreshold time.Duration   // 慢查询阈值
	MaxOpenConns  int             // 最大打开连接数
	MaxIdleConns  int             // 最大空闲连接数
}

// GetLogLevel 获取日志级别
func (c *BaseConfig) GetLogLevel() logger.LogLevel {
	return c.LogLevel
}

// GetSlowThreshold 获取慢查询阈值
func (c *BaseConfig) GetSlowThreshold() time.Duration {
	return c.SlowThreshold
}

// GetMaxOpenConns 获取最大连接数
func (c *BaseConfig) GetMaxOpenConns() int {
	return c.MaxOpenConns
}

// GetMaxIdleConns 获取最大空闲连接数
func (c *BaseConfig) GetMaxIdleConns() int {
	return c.MaxIdleConns
}

// NewBaseConfig 创建默认的基础配置
func NewBaseConfig() BaseConfig {
	return BaseConfig{
		LogLevel:      logger.Silent,
		SlowThreshold: 1 * time.Minute,
		MaxOpenConns:  100,
		MaxIdleConns:  10,
	}
}
