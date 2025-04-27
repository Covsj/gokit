package ioracle

import (
	"fmt"
	"time"

	"github.com/Covsj/gokit/pkg/idb/core"
	"github.com/dzwvip/oracle"
	_ "github.com/godror/godror"
	"gorm.io/gorm"
)

// OracleConfig 定义Oracle数据库配置
type OracleConfig struct {
	core.BaseConfig               // 嵌入基础配置
	Username        string        // 用户名
	Password        string        // 密码
	Host            string        // 主机地址
	Port            int           // 端口号
	ServiceName     string        // 服务名
	SID             string        // SID
	ConnectString   string        // 自定义连接字符串（优先级高于其他连接参数）
	ConnTimeout     time.Duration // 连接超时时间
}

// DefaultConfig 返回默认配置
func DefaultConfig() OracleConfig {
	return OracleConfig{
		BaseConfig:  core.NewBaseConfig(),
		Host:        "localhost",
		Port:        1521,
		ConnTimeout: 10 * time.Second,
	}
}

// BuildDSN 构建数据源连接字符串
func (c *OracleConfig) BuildDSN() string {
	if c.ConnectString != "" {
		return c.ConnectString
	}

	// 优先使用ServiceName
	if c.ServiceName != "" {
		return fmt.Sprintf("%s/%s@%s:%d/%s",
			c.Username, c.Password, c.Host, c.Port, c.ServiceName)
	}

	// 使用SID
	return fmt.Sprintf("%s/%s@%s:%d/%s",
		c.Username, c.Password, c.Host, c.Port, c.SID)
}

// Client Oracle客户端结构体
type Client struct {
	*core.BaseClient // 嵌入基础客户端
}

// New 创建新的Oracle客户端
func New(config OracleConfig) (*Client, error) {
	// 构建DSN
	dsn := config.BuildDSN()

	// 使用核心包的日志适配器
	gormLogger := core.NewGormLogAdapter(config.LogLevel, config.SlowThreshold)

	// 构建Oracle方言配置
	oracleDialector := oracle.Open(dsn)

	// 连接数据库
	db, err := gorm.Open(oracleDialector, &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("连接Oracle数据库失败: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取底层数据库连接失败: %w", err)
	}

	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 创建基础客户端
	baseClient := core.NewBaseClient(db)

	return &Client{
		BaseClient: baseClient,
	}, nil
}

// TableExists 检查表是否存在 (Oracle特定实现)
func (c *Client) TableExists(tableName string) (bool, error) {
	var count int64
	err := c.DB().Raw("SELECT COUNT(*) FROM ALL_TABLES WHERE TABLE_NAME = ?", tableName).Count(&count).Error
	return count > 0, err
}

// Oracle特有功能

// CreateSequence 创建序列
func (c *Client) CreateSequence(sequenceName string, startWith int, incrementBy int) error {
	sql := fmt.Sprintf("CREATE SEQUENCE %s START WITH %d INCREMENT BY %d", sequenceName, startWith, incrementBy)
	return c.Exec(sql)
}

// NextSequenceValue 获取序列的下一个值
func (c *Client) NextSequenceValue(sequenceName string) (int64, error) {
	var value int64
	err := c.RawSQL(&value, fmt.Sprintf("SELECT %s.NEXTVAL FROM DUAL", sequenceName))
	return value, err
}

// CurrentSequenceValue 获取序列的当前值
func (c *Client) CurrentSequenceValue(sequenceName string) (int64, error) {
	var value int64
	err := c.RawSQL(&value, fmt.Sprintf("SELECT %s.CURRVAL FROM DUAL", sequenceName))
	return value, err
}

// DropSequence 删除序列
func (c *Client) DropSequence(sequenceName string) error {
	return c.Exec(fmt.Sprintf("DROP SEQUENCE %s", sequenceName))
}

// ExecuteProcedure 执行存储过程
func (c *Client) ExecuteProcedure(procedureName string, args ...interface{}) error {
	placeholders := ""
	for i := 0; i < len(args); i++ {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += "?"
	}

	sql := fmt.Sprintf("BEGIN %s(%s); END;", procedureName, placeholders)
	return c.Exec(sql, args...)
}

// GetDBVersion 获取数据库版本
func (c *Client) GetDBVersion() (string, error) {
	var version string
	err := c.RawSQL(&version, "SELECT VERSION FROM PRODUCT_COMPONENT_VERSION WHERE PRODUCT LIKE 'Oracle%' AND ROWNUM = 1")
	return version, err
}

// GetTableColumns 获取表的列信息
func (c *Client) GetTableColumns(tableName string) ([]map[string]interface{}, error) {
	var columns []map[string]interface{}
	err := c.DB().Raw("SELECT COLUMN_NAME, DATA_TYPE, DATA_LENGTH, NULLABLE FROM ALL_TAB_COLUMNS WHERE TABLE_NAME = ? ORDER BY COLUMN_ID", tableName).Scan(&columns).Error
	return columns, err
}

// GetTableConstraints 获取表的约束信息
func (c *Client) GetTableConstraints(tableName string) ([]map[string]interface{}, error) {
	var constraints []map[string]interface{}
	sql := `
		SELECT
			c.CONSTRAINT_NAME,
			c.CONSTRAINT_TYPE,
			cc.COLUMN_NAME,
			c.R_CONSTRAINT_NAME
		FROM
			ALL_CONSTRAINTS c
		JOIN
			ALL_CONS_COLUMNS cc ON c.CONSTRAINT_NAME = cc.CONSTRAINT_NAME
		WHERE
			c.TABLE_NAME = ?
		ORDER BY
			c.CONSTRAINT_NAME
	`
	err := c.DB().Raw(sql, tableName).Scan(&constraints).Error
	return constraints, err
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
