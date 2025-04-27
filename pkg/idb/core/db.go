package core

import (
	"gorm.io/gorm"
)

// DBOperator 定义数据库基本操作接口
type DBOperator interface {
	// 基本操作
	DB() *gorm.DB
	AutoMigrate(models ...interface{}) error
	Close() error

	// 事务操作
	Transaction(fn func(tx *gorm.DB) error) error

	// 增删改查操作
	Create(value interface{}) error
	Save(value interface{}) error
	Update(model interface{}, updates interface{}) error
	UpdateAll(model interface{}, updates interface{}) error
	Delete(value interface{}, conds ...interface{}) error
	First(dest interface{}, conds ...interface{}) error
	Find(dest interface{}, conds ...interface{}) error
	Where(query interface{}, args ...interface{}) *gorm.DB
	Count(model interface{}, count *int64, conds ...interface{}) error

	// 分页查询
	Paginate(page, pageSize int, dest interface{}, conds ...interface{}) (int64, error)

	// 原始SQL操作
	RawSQL(dest interface{}, sql string, values ...interface{}) error
	Exec(sql string, values ...interface{}) error

	// 表操作
	TableExists(tableName string) (bool, error)
}

// BaseClient 基础客户端实现
type BaseClient struct {
	db *gorm.DB
}

// NewBaseClient 创建基础客户端
func NewBaseClient(db *gorm.DB) *BaseClient {
	return &BaseClient{
		db: db,
	}
}

// DB 返回原始的gorm.DB实例
func (c *BaseClient) DB() *gorm.DB {
	return c.db
}

// AutoMigrate 自动迁移表结构
func (c *BaseClient) AutoMigrate(models ...interface{}) error {
	return c.db.AutoMigrate(models...)
}

// Transaction 执行事务
func (c *BaseClient) Transaction(fn func(tx *gorm.DB) error) error {
	return c.db.Transaction(fn)
}

// Create 创建记录
func (c *BaseClient) Create(value interface{}) error {
	result := c.db.Create(value)
	return result.Error
}

// Save 保存记录（创建或更新）
func (c *BaseClient) Save(value interface{}) error {
	result := c.db.Save(value)
	return result.Error
}

// Update 更新指定字段
func (c *BaseClient) Update(model interface{}, updates interface{}) error {
	result := c.db.Model(model).Updates(updates)
	return result.Error
}

// UpdateAll 更新所有字段，包括零值字段
func (c *BaseClient) UpdateAll(model interface{}, updates interface{}) error {
	result := c.db.Model(model).UpdateColumns(updates)
	return result.Error
}

// Delete 删除记录
func (c *BaseClient) Delete(value interface{}, conds ...interface{}) error {
	result := c.db.Delete(value, conds...)
	return result.Error
}

// First 查找第一条记录
func (c *BaseClient) First(dest interface{}, conds ...interface{}) error {
	result := c.db.First(dest, conds...)
	return result.Error
}

// Find 查找多条记录
func (c *BaseClient) Find(dest interface{}, conds ...interface{}) error {
	result := c.db.Find(dest, conds...)
	return result.Error
}

// Where 条件查询
func (c *BaseClient) Where(query interface{}, args ...interface{}) *gorm.DB {
	return c.db.Where(query, args...)
}

// Count 计数
func (c *BaseClient) Count(model interface{}, count *int64, conds ...interface{}) error {
	query := c.db.Model(model)
	if len(conds) > 0 {
		query = query.Where(conds[0], conds[1:]...)
	}
	result := query.Count(count)
	return result.Error
}

// Paginate 分页查询
func (c *BaseClient) Paginate(page, pageSize int, dest interface{}, conds ...interface{}) (int64, error) {
	var total int64
	query := c.db.Model(dest)

	if len(conds) > 0 {
		query = query.Where(conds[0], conds[1:]...)
	}

	err := query.Count(&total).Error
	if err != nil {
		return 0, err
	}

	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	err = query.Offset(offset).Limit(pageSize).Find(dest).Error
	return total, err
}

// RawSQL 执行原始SQL查询
func (c *BaseClient) RawSQL(dest interface{}, sql string, values ...interface{}) error {
	return c.db.Raw(sql, values...).Scan(dest).Error
}

// Exec 执行原始SQL语句
func (c *BaseClient) Exec(sql string, values ...interface{}) error {
	return c.db.Exec(sql, values...).Error
}

// Close 关闭数据库连接
func (c *BaseClient) Close() error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// TableExists 是一个需要各数据库具体实现的方法
// 基类提供一个通用实现，但建议子类重写为特定数据库的高效实现
func (c *BaseClient) TableExists(tableName string) (bool, error) {
	var count int64
	// 这个SQL在不同数据库中可能不兼容，子类应该重写这个方法
	err := c.db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = ?", tableName).Count(&count).Error
	return count > 0, err
}
