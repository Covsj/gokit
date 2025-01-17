package idb

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Covsj/gokit/pkg/log"
	_ "github.com/sijms/go-ora/v2"
)

// OracleConfig 定义Oracle数据库连接配置
type OracleConfig struct {
	Host         string        // 数据库主机地址
	Port         int           // 数据库端口号
	Username     string        // 数据库用户名
	Password     string        // 数据库密码
	Service      string        // Oracle服务名
	MaxOpenConns int           // 最大连接数
	MaxIdleConns int           // 最大空闲连接数
	MaxLifetime  time.Duration // 连接最大生命周期
}

// OracleDB 封装Oracle数据库操作
type OracleDB struct {
	db *sql.DB
}

// NewOracleConnection 创建新的Oracle数据库连接
func NewOracleConnection(config OracleConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf("oracle://%s:%s@%s:%d/%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Service,
	)

	db, err := sql.Open("oracle", connStr)
	if err != nil {
		log.Error("连接数据库失败", "错误", err.Error())
		return nil, fmt.Errorf("连接Oracle数据库失败: %v", err)
	}

	// 设置连接池参数
	if config.MaxOpenConns > 0 {
		db.SetMaxOpenConns(config.MaxOpenConns)
	}
	if config.MaxIdleConns > 0 {
		db.SetMaxIdleConns(config.MaxIdleConns)
	}
	if config.MaxLifetime > 0 {
		db.SetConnMaxLifetime(config.MaxLifetime)
	}

	// 测试连接
	if err = db.Ping(); err != nil {
		log.Error("Ping数据库失败", "错误", err.Error())
		return nil, fmt.Errorf("ping Oracle数据库失败: %v", err)
	}

	log.Info("连接数据库成功",
		"用户", config.Username,
		"地址", fmt.Sprintf("%s:%d", config.Host, config.Port),
		"服务", config.Service,
		"最大连接数", config.MaxOpenConns,
		"最大空闲连接", config.MaxIdleConns,
		"连接生命周期", config.MaxLifetime,
	)
	return db, nil
}

// NewOracleDB 创建OracleDB实例
func NewOracleDB(config OracleConfig) (*OracleDB, error) {
	db, err := NewOracleConnection(config)
	if err != nil {
		return nil, err
	}
	return &OracleDB{db: db}, nil
}

// GetDB 获取原始数据库连接
func (o *OracleDB) GetDB() *sql.DB {
	return o.db
}

// Close 关闭数据库连接
func (o *OracleDB) Close() error {
	return o.db.Close()
}

// Stats 获取数据库连接池统计信息
func (o *OracleDB) Stats() sql.DBStats {
	return o.db.Stats()
}

// ExecContext 执行SQL语句
func (o *OracleDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	result, err := o.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("执行SQL失败: %v", err)
	}
	return result, nil
}

// QueryRowContext 查询单行数据
func (o *OracleDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return o.db.QueryRowContext(ctx, query, args...)
}

// QueryContext 查询多行数据
func (o *OracleDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := o.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %v", err)
	}
	return rows, nil
}

// Begin 开始事务
func (o *OracleDB) Begin() (*sql.Tx, error) {
	return o.db.Begin()
}

// BeginTx 开始事务（带选项）
func (o *OracleDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return o.db.BeginTx(ctx, opts)
}

// InsertRow 插入单行数据
func (o *OracleDB) InsertRow(ctx context.Context, tableName string, data map[string]interface{}) (int64, error) {
	columns := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))
	placeholders := make([]string, 0, len(data))

	i := 1
	for col, val := range data {
		columns = append(columns, col)
		values = append(values, val)
		placeholders = append(placeholders, fmt.Sprintf(":v%d", i))
		i++
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	result, err := o.ExecContext(ctx, query, values...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// BatchInsert 批量插入数据
func (o *OracleDB) BatchInsert(ctx context.Context, tableName string, dataList []map[string]interface{}) (int64, error) {
	if len(dataList) == 0 {
		return 0, nil
	}

	// 获取所有列名
	columns := make([]string, 0)
	for col := range dataList[0] {
		columns = append(columns, col)
	}

	// 构建 Oracle 的批量插入语句
	var insertValues []string
	var values []interface{}
	paramCount := 1

	// Oracle 的批量插入使用 UNION ALL
	for _, data := range dataList {
		placeholders := make([]string, len(columns))
		for i, col := range columns {
			placeholders[i] = fmt.Sprintf(":v%d", paramCount)
			values = append(values, data[col])
			paramCount++
		}
		insertValues = append(insertValues, fmt.Sprintf("SELECT %s FROM DUAL", strings.Join(placeholders, ", ")))
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) %s",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(insertValues, " UNION ALL "),
	)

	log.Info("批量插入SQL",
		"表名", tableName,
		"SQL", query,
		"参数数量", len(values),
	)

	result, err := o.ExecContext(ctx, query, values...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// UpdateRows 更新数据
func (o *OracleDB) UpdateRows(ctx context.Context, tableName string, data map[string]interface{}, where string, args ...interface{}) (int64, error) {
	setClauses := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data)+len(args))

	i := 1
	for col, val := range data {
		setClauses = append(setClauses, fmt.Sprintf("%s = :v%d", col, i))
		values = append(values, val)
		i++
	}
	values = append(values, args...)

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s",
		tableName,
		strings.Join(setClauses, ", "),
		where,
	)

	result, err := o.ExecContext(ctx, query, values...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// DeleteRows 删除数据
func (o *OracleDB) DeleteRows(ctx context.Context, tableName string, where string, args ...interface{}) (int64, error) {
	query := fmt.Sprintf("DELETE FROM %s WHERE %s", tableName, where)
	result, err := o.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// CountRows 统计行数
func (o *OracleDB) CountRows(ctx context.Context, tableName string, where string, args ...interface{}) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	if where != "" {
		query += " WHERE " + where
	}

	var count int64
	err := o.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("统计行数失败: %v", err)
	}
	return count, nil
}

// ExistsRow 检查记录是否存在
func (o *OracleDB) ExistsRow(ctx context.Context, tableName string, where string, args ...interface{}) (bool, error) {
	count, err := o.CountRows(ctx, tableName, where, args...)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// TruncateTable 清空表
func (o *OracleDB) TruncateTable(ctx context.Context, tableName string) error {
	_, err := o.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE %s", tableName))
	return err
}

// GetTableColumns 获取表的列信息
func (o *OracleDB) GetTableColumns(ctx context.Context, tableName string) ([]string, error) {
	query := `
		SELECT column_name 
		FROM user_tab_columns 
		WHERE table_name = UPPER(:v1)
		ORDER BY column_id`

	rows, err := o.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("获取表列信息失败: %v", err)
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var column string
		if err := rows.Scan(&column); err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}

	return columns, nil
}

// TransactionFunc 定义事务执行的函数类型
type TransactionFunc func(*sql.Tx) error

// WithTransaction 封装事务操作
func (o *OracleDB) WithTransaction(ctx context.Context, fn TransactionFunc) error {
	// 开启事务
	tx, err := o.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开启事务失败: %v", err)
	}

	// 延迟处理事务提交或回滚
	defer func() {
		if p := recover(); p != nil {
			// 发生panic时回滚事务
			tx.Rollback()
			panic(p) // 重新抛出panic
		}
	}()

	// 执行事务操作
	if err := fn(tx); err != nil {
		// 执行失败时回滚事务
		tx.Rollback()
		return err
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// SavePoint 创建保存点
func (o *OracleDB) SavePoint(ctx context.Context, tx *sql.Tx, name string) error {
	_, err := tx.ExecContext(ctx, fmt.Sprintf("SAVEPOINT %s", name))
	if err != nil {
		return fmt.Errorf("创建保存点失败: %v", err)
	}
	return nil
}

// RollbackToSavePoint 回滚到指定保存点
func (o *OracleDB) RollbackToSavePoint(ctx context.Context, tx *sql.Tx, name string) error {
	_, err := tx.ExecContext(ctx, fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", name))
	if err != nil {
		return fmt.Errorf("回滚到保存点失败: %v", err)
	}
	return nil
}

// ReleaseSavePoint 释放保存点
func (o *OracleDB) ReleaseSavePoint(ctx context.Context, tx *sql.Tx, name string) error {
	_, err := tx.ExecContext(ctx, fmt.Sprintf("RELEASE SAVEPOINT %s", name))
	if err != nil {
		return fmt.Errorf("释放保存点失败: %v", err)
	}
	return nil
}

// SQLTemplate 定义常用SQL模板
// 提供了一组预定义的SQL模板，用于常见的数据库操作
// 包括：查询、更新、删除、分页等操作
type SQLTemplate struct {
	// 查询类模板
	SelectByID     string // 根据ID查询单条记录
	SelectAll      string // 查询表中所有记录
	SelectWithPage string // 分页查询（使用ROWNUM实现）
	SelectCount    string // 统计表中记录总数

	// 更新类模板
	UpdateByID     string // 根据ID更新记录
	DeleteByID     string // 根据ID删除记录
	SoftDeleteByID string // 根据ID软删除记录（设置删除标记）

	// 插入类模板
	BatchInsert string // 批量插入记录
	UpsertByID  string // 插入或更新记录（MERGE INTO）
}

// NewSQLTemplate 创建SQL模板实例
// 参数：
//   - tableName: 表名
//
// 返回：
//   - *SQLTemplate: SQL模板实例
func NewSQLTemplate(tableName string) *SQLTemplate {
	return &SQLTemplate{
		// 基本查询模板
		SelectByID: fmt.Sprintf("SELECT * FROM %s WHERE id = :v1", tableName),
		SelectAll:  fmt.Sprintf("SELECT * FROM %s", tableName),

		// 修改分页查询模板，使用正确的Oracle分页语法
		SelectWithPage: fmt.Sprintf(`
			SELECT a.* FROM (
				SELECT t.*, ROWNUM AS rn 
				FROM (
					SELECT t.* 
					FROM %s t 
					ORDER BY t.id
				) t WHERE ROWNUM <= :page_end
			) a WHERE rn > :page_start`, tableName),

		// 统计查询
		SelectCount: fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName),

		// 更新和删除模板
		UpdateByID:     fmt.Sprintf("UPDATE %s SET %%s WHERE id = :v1", tableName),
		DeleteByID:     fmt.Sprintf("DELETE FROM %s WHERE id = :v1", tableName),
		SoftDeleteByID: fmt.Sprintf("UPDATE %s SET is_deleted = 1, deleted_at = SYSTIMESTAMP WHERE id = :v1", tableName),

		// 插入相关模板
		BatchInsert: fmt.Sprintf("INSERT INTO %s (%%s) VALUES %%s", tableName),

		// Oracle MERGE语法实现的UPSERT
		UpsertByID: fmt.Sprintf(`
			MERGE INTO %s t
			USING (SELECT :v1 as id, %%s FROM dual) s
			ON (t.id = s.id)
			WHEN MATCHED THEN
				UPDATE SET %%s
			WHEN NOT MATCHED THEN
				INSERT (%%s) VALUES (%%s)`, tableName),
	}
}

// PageQuery 分页查询参数
type PageQuery struct {
	PageNum  int    // 页码（从1开始）
	PageSize int    // 每页记录数
	OrderBy  string // 排序字段
	Desc     bool   // 是否降序
}

// PageResult 分页查询结果
type PageResult struct {
	Total    int64       `json:"total"`     // 总记录数
	PageNum  int         `json:"page_num"`  // 当前页码
	PageSize int         `json:"page_size"` // 每页大小
	Data     interface{} `json:"data"`      // 数据列表
}

// QueryPage 执行分页查询
func (o *OracleDB) QueryPage(ctx context.Context, dest interface{}, baseSQL string, page PageQuery, args ...interface{}) (*PageResult, error) {
	// 1. 构建查询语句
	countSQL := fmt.Sprintf("SELECT COUNT(1) FROM (%s)", baseSQL)

	// 添加排序
	orderClause := ""
	if page.OrderBy != "" {
		orderClause = fmt.Sprintf(" ORDER BY %s", page.OrderBy)
		if page.Desc {
			orderClause += " DESC"
		}
	}

	// 构建分页SQL，修改为不返回RN列
	pageSQL := fmt.Sprintf(`
		SELECT a.* FROM (
			SELECT t.*, ROWNUM AS rn 
			FROM (
				%s%s
			) t WHERE ROWNUM <= :page_end
		) a WHERE rn > :page_start`, baseSQL, orderClause)

	log.Info("分页SQL", "SQL", pageSQL)

	// 2. 查询总记录数
	var total int64
	err := o.QueryRowContext(ctx, countSQL, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("查询总数失败: %v", err)
	}
	log.Info("查询总数", "total", total)

	// 3. 计算分页参数
	startRow := (page.PageNum - 1) * page.PageSize
	endRow := page.PageNum * page.PageSize

	// 4. 添加分页参数
	pageArgs := make([]interface{}, len(args)+2)
	copy(pageArgs, args)
	pageArgs[len(args)] = sql.Named("page_start", startRow)
	pageArgs[len(args)+1] = sql.Named("page_end", endRow)

	log.Info("分页参数",
		"startRow", startRow,
		"endRow", endRow,
		"pageNum", page.PageNum,
		"pageSize", page.PageSize,
	)

	// 5. 执行分页查询
	rows, err := o.QueryContext(ctx, pageSQL, pageArgs...)
	if err != nil {
		return nil, fmt.Errorf("执行分页查询失败: %v", err)
	}
	defer rows.Close()

	// 6. 扫描结果到目标对象
	err = scanStructs(dest, rows)
	if err != nil {
		return nil, fmt.Errorf("扫描数据失败: %v", err)
	}

	// 获取实际数据长度
	destVal := reflect.ValueOf(dest).Elem()
	dataLen := destVal.Len()
	log.Info("查询结果", "数据条数", dataLen)

	return &PageResult{
		Total:    total,
		PageNum:  page.PageNum,
		PageSize: page.PageSize,
		Data:     dest,
	}, nil
}

// QueryStruct 查询单行数据并映射到结构体
func (o *OracleDB) QueryStruct(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	rows, err := o.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}

	values := make([]interface{}, len(columns))
	for i := range values {
		values[i] = new(interface{})
	}

	err = rows.Scan(values...)
	if err != nil {
		return err
	}

	return scanStruct(dest, columns, values)
}

// QueryStructs 查询多行数据并映射到结构体切片
func (o *OracleDB) QueryStructs(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	rows, err := o.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return scanStructs(dest, rows)
}

// QueryPageResult 执行分页查询并返回结果
// 参数：
//   - ctx: 上下文
//   - dest: 结果集接收对象（指针类型）
//   - countQuery: 统计总数的SQL
//   - dataQuery: 查询数据的SQL
//   - pageNum: 页码
//   - pageSize: 每页大小
//   - args: 查询参数
//
// 返回：
//   - *PageResult: 分页结果
//   - error: 错误信息
func (o *OracleDB) QueryPageResult(ctx context.Context, dest interface{}, countQuery, dataQuery string, pageNum, pageSize int, args ...interface{}) (*PageResult, error) {
	// 1. 查询总数
	var total int64
	err := o.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("查询总数失败: %v", err)
	}
	log.Info("查询总数", "total", total)

	// 2. 执行分页查询
	rows, err := o.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("查询数据失败: %v", err)
	}
	defer rows.Close()

	// 3. 扫描数据到目标对象
	if err := scanStructs(dest, rows); err != nil {
		return nil, fmt.Errorf("扫描数据失败: %v", err)
	}

	// 4. 返回分页结果
	result := &PageResult{
		Total:    total,
		PageNum:  pageNum,
		PageSize: pageSize,
		Data:     dest,
	}

	return result, nil
}

// 内部辅助函数
func scanStruct(dest interface{}, columns []string, values []interface{}) error {
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer")
	}
	v = v.Elem()
	t := v.Type()

	for i, column := range columns {
		val := reflect.ValueOf(values[i]).Elem().Interface()

		// 先尝试通过字段名匹配
		field := v.FieldByNameFunc(func(name string) bool {
			return strings.EqualFold(name, column)
		})

		// 如果没找到，尝试通过 db 标签匹配
		if !field.IsValid() {
			for j := 0; j < t.NumField(); j++ {
				if strings.EqualFold(t.Field(j).Tag.Get("db"), column) {
					field = v.Field(j)
					break
				}
			}
		}

		if !field.IsValid() {
			log.Info("字段未找到", "column", column)
			continue
		}

		if err := setValue(field, val); err != nil {
			return fmt.Errorf("设置字段值失败 [%s]: %v", column, err)
		}
	}
	return nil
}

func scanStructs(dest interface{}, rows *sql.Rows) error {
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer")
	}
	if v.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("dest must be a pointer to slice")
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	sliceType := v.Elem().Type()
	elemType := sliceType.Elem()
	slice := reflect.MakeSlice(sliceType, 0, 0)

	for rows.Next() {
		elem := reflect.New(elemType).Elem()
		values := make([]interface{}, len(columns))
		for i := range values {
			values[i] = new(interface{})
		}

		if err := rows.Scan(values...); err != nil {
			return err
		}

		if err := scanStruct(elem.Addr().Interface(), columns, values); err != nil {
			return err
		}

		slice = reflect.Append(slice, elem)
	}

	v.Elem().Set(slice)
	return rows.Err()
}

func setValue(field reflect.Value, value interface{}) error {
	if value == nil {
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(fmt.Sprint(value))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, err := strconv.ParseInt(fmt.Sprint(value), 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(fmt.Sprint(value), 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(val)
	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(fmt.Sprint(value), 64)
		if err != nil {
			return err
		}
		field.SetFloat(val)
	case reflect.Bool:
		val, err := strconv.ParseBool(fmt.Sprint(value))
		if err != nil {
			return err
		}
		field.SetBool(val)
	case reflect.Struct:
		if field.Type() == reflect.TypeOf(time.Time{}) {
			if t, ok := value.(time.Time); ok {
				field.Set(reflect.ValueOf(t))
			}
		}
	}
	return nil
}
