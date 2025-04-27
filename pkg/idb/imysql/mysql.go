package imysql

import (
	"fmt"
	"time"

	"github.com/Covsj/gokit/pkg/idb/core"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MySQLConfig 定义MySQL数据库配置
type MySQLConfig struct {
	core.BaseConfig               // 嵌入基础配置
	Username        string        // 用户名
	Password        string        // 密码
	Host            string        // 主机地址
	Port            int           // 端口号
	Database        string        // 数据库名
	Charset         string        // 字符集
	ParseTime       bool          // 是否解析时间
	Loc             string        // 时区
	Timeout         time.Duration // 连接超时时间
	ReadTimeout     time.Duration // 读取超时时间
	WriteTimeout    time.Duration // 写入超时时间
	ConnMaxIdle     time.Duration // 连接最大空闲时间
	ConnMaxLife     time.Duration // 连接最大生命周期
}

// DefaultConfig 返回默认配置
func DefaultConfig() MySQLConfig {
	return MySQLConfig{
		BaseConfig:   core.NewBaseConfig(),
		Host:         "localhost",
		Port:         3306,
		Charset:      "utf8mb4",
		ParseTime:    true,
		Loc:          "Local",
		Timeout:      10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		ConnMaxIdle:  time.Hour,
		ConnMaxLife:  24 * time.Hour,
	}
}

// BuildDSN 构建数据源连接字符串
func (c *MySQLConfig) BuildDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s&timeout=%s&readTimeout=%s&writeTimeout=%s",
		c.Username, c.Password, c.Host, c.Port, c.Database, c.Charset,
		c.ParseTime, c.Loc, c.Timeout, c.ReadTimeout, c.WriteTimeout)
}

// Client MySQL客户端结构体
type Client struct {
	*core.BaseClient // 嵌入基础客户端
}

// New 创建新的MySQL客户端
func New(config MySQLConfig) (*Client, error) {
	// 构建DSN
	dsn := config.BuildDSN()

	// 使用核心包的日志适配器
	gormLogger := core.NewGormLogAdapter(config.LogLevel, config.SlowThreshold)

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("连接MySQL数据库失败: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取底层数据库连接失败: %w", err)
	}

	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdle)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLife)

	// 创建基础客户端
	baseClient := core.NewBaseClient(db)

	return &Client{
		BaseClient: baseClient,
	}, nil
}

// TableExists 检查表是否存在 (MySQL特定实现)
func (c *Client) TableExists(tableName string) (bool, error) {
	var count int64
	err := c.DB().Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", tableName).Count(&count).Error
	return count > 0, err
}

// MySQL特有功能

// GetTableStatus 获取表状态信息
func (c *Client) GetTableStatus(tableName string) (map[string]interface{}, error) {
	var status map[string]interface{}
	err := c.DB().Raw("SHOW TABLE STATUS WHERE Name = ?", tableName).Scan(&status).Error
	return status, err
}

// CreateDatabase 创建数据库
func (c *Client) CreateDatabase(dbName string, charset string) error {
	if charset == "" {
		charset = "utf8mb4"
	}
	sql := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET %s", dbName, charset)
	return c.Exec(sql)
}

// DropDatabase 删除数据库
func (c *Client) DropDatabase(dbName string) error {
	sql := fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", dbName)
	return c.Exec(sql)
}

// Truncate 清空表
func (c *Client) Truncate(tableName string) error {
	return c.Exec(fmt.Sprintf("TRUNCATE TABLE `%s`", tableName))
}

// GetTableSchema 获取表结构信息
func (c *Client) GetTableSchema(tableName string) ([]map[string]interface{}, error) {
	var columns []map[string]interface{}
	err := c.DB().Raw("SHOW FULL COLUMNS FROM `" + tableName + "`").Scan(&columns).Error
	return columns, err
}

// GetDatabases 获取所有数据库
func (c *Client) GetDatabases() ([]string, error) {
	var databases []string
	err := c.DB().Raw("SHOW DATABASES").Scan(&databases).Error
	return databases, err
}

// GetTables 获取所有表
func (c *Client) GetTables() ([]string, error) {
	var tables []string
	err := c.DB().Raw("SHOW TABLES").Scan(&tables).Error
	return tables, err
}

// GetIndexes 获取表的索引信息
func (c *Client) GetIndexes(tableName string) ([]map[string]interface{}, error) {
	var indexes []map[string]interface{}
	err := c.DB().Raw("SHOW INDEX FROM `" + tableName + "`").Scan(&indexes).Error
	return indexes, err
}

// OptimizeTable 优化表
func (c *Client) OptimizeTable(tableName string) error {
	return c.Exec("OPTIMIZE TABLE `" + tableName + "`")
}

// LockTable 锁表
func (c *Client) LockTable(tableName string, lockType string) error {
	if lockType == "" {
		lockType = "WRITE"
	}
	return c.Exec(fmt.Sprintf("LOCK TABLES `%s` %s", tableName, lockType))
}

// UnlockTables 解锁表
func (c *Client) UnlockTables() error {
	return c.Exec("UNLOCK TABLES")
}

// GetServerVersion 获取MySQL服务器版本
func (c *Client) GetServerVersion() (string, error) {
	var version string
	err := c.RawSQL(&version, "SELECT VERSION()")
	return version, err
}

// CreateEvent 创建事件
func (c *Client) CreateEvent(eventName, schedule, statement string) error {
	sql := fmt.Sprintf("CREATE EVENT `%s` ON SCHEDULE %s DO %s", eventName, schedule, statement)
	return c.Exec(sql)
}

// DropEvent 删除事件
func (c *Client) DropEvent(eventName string) error {
	return c.Exec(fmt.Sprintf("DROP EVENT IF EXISTS `%s`", eventName))
}

// ExecuteBatch 批量执行多条SQL语句
func (c *Client) ExecuteBatch(sqls []string) error {
	return c.Transaction(func(tx *gorm.DB) error {
		for _, sql := range sqls {
			if err := tx.Exec(sql).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// SetCharset 设置连接字符集
func (c *Client) SetCharset(charset string) error {
	return c.Exec(fmt.Sprintf("SET NAMES %s", charset))
}
