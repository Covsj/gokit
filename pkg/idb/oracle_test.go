package idb

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/Covsj/gokit/pkg/log"
)

// 测试配置，根据实际环境修改
var testConfig = OracleConfig{
	Host:     "localhost",
	Port:     1521,
	Username: "system",
	Password: "nanfang66",
	Service:  "XE",
}

// TestNewOracleDB 测试创建数据库连接
func TestNewOracleDB(t *testing.T) {
	log.Info("开始测试数据库连接")
	db, err := NewOracleDB(testConfig)
	if err != nil {
		log.ErrorF("创建数据库连接失败: %v", err)
		return
	}
	defer db.Close()

	// 测试连接是否正常
	ctx := context.Background()
	var dummy string
	err = db.QueryRowContext(ctx, "SELECT 'TEST' FROM DUAL").Scan(&dummy)
	if err != nil {
		log.ErrorF("执行测试查询失败: %v", err)
		return
	}
	if dummy != "TEST" {
		log.ErrorF("期望得到 'TEST'，实际得到 %s", dummy)
		return
	}
	log.Info("数据库连接测试成功", "结果", dummy)
}

// TestCRUD 测试基本的CRUD操作
func TestCRUD(t *testing.T) {
	log.Info("开始测试CRUD操作")
	db, err := NewOracleDB(testConfig)
	if err != nil {
		log.ErrorF("创建数据库连接失败: %v", err)
		return
	}
	defer db.Close()

	ctx := context.Background()

	// 创建测试表
	log.Info("创建测试表")
	_, err = db.ExecContext(ctx, `
		CREATE TABLE test_users (
			id NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			name VARCHAR2(100),
			age NUMBER(3),
			birth_date DATE,
			created_at TIMESTAMP,
			updated_at TIMESTAMP,
			salary NUMBER(10,2),
			bonus NUMBER,
			performance NUMBER(5,4),
			score NUMBER(3),
			amount NUMBER(20,6),
			is_active NUMBER(1),
			description CLOB,
			email VARCHAR2(200),
			phone VARCHAR2(20)
		)
	`)
	if err != nil {
		log.ErrorF("创建测试表失败: %v", err)
		return
	}
	defer func() {
		log.Info("清理测试表")
		if _, err := db.ExecContext(ctx, "DROP TABLE test_users"); err != nil {
			log.ErrorF("删除测试表失败: %v", err)
		}
	}()

	// 测试插入
	t.Run("Insert", func(t *testing.T) {
		log.Info("测试插入操作")
		// 准备测试数据
		birthDateInit := time.Date(1990, 1, 1, 0, 0, 0, 0, time.Local)
		now := time.Now()

		data := map[string]interface{}{
			"name":        "张三",
			"age":         25,
			"birth_date":  birthDateInit,
			"created_at":  now,
			"updated_at":  now,
			"salary":      8888.88,
			"bonus":       123456.789123,
			"performance": 0.9999,
			"score":       100,
			"amount":      123456789.123456,
			"is_active":   1,
			"description": "这是一段测试描述",
			"email":       "zhangsan@example.com",
			"phone":       "13800138000",
		}

		rows, err := db.InsertRow(ctx, "test_users", data)
		if err != nil {
			log.ErrorF("插入数据失败: %v", err)
			return
		}
		if rows != 1 {
			log.ErrorF("期望插入1行，实际插入%d行", rows)
			return
		}
		log.Info("插入数据成功", "影响行数", rows)

		// 验证插入结果
		var (
			name        string
			age         int
			birthDate   time.Time
			createdAt   time.Time
			updatedAt   time.Time
			salary      float64
			bonus       float64
			performance float64
			score       int
			amount      float64
			isActive    int
			description string
			email       string
			phone       string
		)

		err = db.QueryRowContext(ctx, `
			SELECT name, age, birth_date, created_at, updated_at, 
				   salary, bonus, performance, score, amount,
				   is_active, description, email, phone 
			FROM test_users WHERE name = :v1`, "张三").
			Scan(&name, &age, &birthDate, &createdAt, &updatedAt,
				&salary, &bonus, &performance, &score, &amount,
				&isActive, &description, &email, &phone)

		if err != nil {
			log.ErrorF("验证插入结果失败: %v", err)
			return
		}

		log.Info("验证插入结果",
			"姓名", name,
			"年龄", age,
			"出生日期", birthDate.Format("2006-01-02"),
			"创建时间", createdAt.Format("2006-01-02 15:04:05"),
			"更新时间", updatedAt.Format("2006-01-02 15:04:05"),
			"薪资", salary,
			"奖金", bonus,
			"绩效", performance,
			"分数", score,
			"金额", amount,
			"是否激活", isActive,
			"描述", description,
			"邮箱", email,
			"电话", phone,
		)

		// 验证各字段值
		if name != "张三" || age != 25 || salary != 8888.88 || bonus != 123456.789123 ||
			performance != 0.9999 || score != 100 || amount != 123456789.123456 ||
			isActive != 1 || description != "这是一段测试描述" || email != "zhangsan@example.com" ||
			phone != "13800138000" {
			log.ErrorF("插入数据验证失败，数据不匹配")
			return
		}
	})

	// 测试更新
	t.Run("Update", func(t *testing.T) {
		log.Info("测试更新操作")
		now := time.Now()
		updateData := map[string]interface{}{
			"age":         26,
			"updated_at":  now,
			"salary":      9999.99,
			"description": "更新后的描述",
		}
		rows, err := db.UpdateRows(ctx, "test_users", updateData, "name = :v1", "张三")
		if err != nil {
			log.ErrorF("更新数据失败: %v", err)
			return
		}
		if rows != 1 {
			log.ErrorF("期望更新1行，实际更新%d行", rows)
			return
		}
		log.Info("更新数据成功", "影响行数", rows)

		// 验证更新结果
		var (
			age         int
			updatedAt   time.Time
			salary      float64
			description string
		)

		err = db.QueryRowContext(ctx, `
			SELECT age, updated_at, salary, description 
			FROM test_users WHERE name = :v1`, "张三").
			Scan(&age, &updatedAt, &salary, &description)

		if err != nil {
			log.ErrorF("验证更新结果失败: %v", err)
			return
		}

		log.Info("验证更新结果",
			"年龄", age,
			"更新时间", updatedAt.Format("2006-01-02 15:04:05"),
			"薪资", salary,
			"描述", description,
		)

		if age != 26 || salary != 9999.99 || description != "更新后的描述" {
			log.ErrorF("更新数据验证失败，数据不匹配")
			return
		}
	})

	// 测试查询
	t.Run("Query", func(t *testing.T) {
		log.Info("测试查询操作")
		// 查询单行
		var name string
		var age int
		err := db.QueryRowContext(ctx, "SELECT name, age FROM test_users WHERE name = :v1", "张三").Scan(&name, &age)
		if err != nil {
			log.ErrorF("查询单行数据失败: %v", err)
			return
		}
		log.Info("查询单行结果", "姓名", name, "年龄", age)

		// 插入更多测试数据
		log.Info("插入更多测试数据")
		testData := []map[string]interface{}{
			{"name": "李四", "age": 30},
			{"name": "王五", "age": 35},
		}
		for _, data := range testData {
			_, err := db.InsertRow(ctx, "test_users", data)
			if err != nil {
				log.ErrorF("插入测试数据失败: %v", err)
				return
			}
		}

		// 查询多行
		log.Info("测试查询多行数据")
		rows, err := db.QueryContext(ctx, "SELECT name, age FROM test_users ORDER BY age")
		if err != nil {
			log.ErrorF("查询多行数据失败: %v", err)
			return
		}
		defer rows.Close()

		var users []struct {
			Name string
			Age  int
		}
		for rows.Next() {
			var u struct {
				Name string
				Age  int
			}
			if err := rows.Scan(&u.Name, &u.Age); err != nil {
				log.ErrorF("扫描行数据失败: %v", err)
				return
			}
			users = append(users, u)
		}
		if err = rows.Err(); err != nil {
			log.ErrorF("遍历结果集失败: %v", err)
			return
		}

		log.Info("查询到的所有用户", "数量", len(users))
		for _, u := range users {
			log.Info("用户信息", "姓名", u.Name, "年龄", u.Age)
		}

		// 验证结果数量
		if len(users) != 3 {
			log.ErrorF("期望查询到3条记录，实际查询到%d条", len(users))
			return
		}
	})

	// 测试删除
	t.Run("Delete", func(t *testing.T) {
		log.Info("测试删除操作")
		// 先统计总数
		var count int
		err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM test_users").Scan(&count)
		if err != nil {
			log.ErrorF("统计删除前数据失败: %v", err)
			return
		}
		log.Info("删除前总数", "count", count)

		// 删除数据
		rows, err := db.DeleteRows(ctx, "test_users", "name = :v1", "张三")
		if err != nil {
			log.ErrorF("删除数据失败: %v", err)
			return
		}
		if rows != 1 {
			log.ErrorF("期望删除1行，实际删除%d行", rows)
			return
		}
		log.Info("删除数据成功", "影响行数", rows)

		// 验证删除结果
		err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM test_users").Scan(&count)
		if err != nil {
			log.ErrorF("统计删除后数据失败: %v", err)
			return
		}
		log.Info("删除后总数", "count", count)

		var name string
		// 确认数据已被删除
		err = db.QueryRowContext(ctx, "SELECT name FROM test_users WHERE name = :v1", "张三").Scan(&name)
		if err == nil {
			log.ErrorF("数据未被正确删除，仍能查询到记录")
			return
		}
		if err != sql.ErrNoRows {
			log.ErrorF("预期查询无结果，但发生其他错误: %v", err)
			return
		}
		log.Info("验证删除成功，查询不到已删除的记录")
	})
}

// TestTransaction 测试事务操作
func TestTransaction(t *testing.T) {
	log.Info("开始测试事务操作")
	db, err := NewOracleDB(testConfig)
	if err != nil {
		log.ErrorF("创建数据库连接失败: %v", err)
		return
	}
	defer db.Close()

	ctx := context.Background()

	// 创建测试表
	log.Info("创建测试表")
	_, err = db.ExecContext(ctx, `
		CREATE TABLE test_accounts (
			id NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			name VARCHAR2(100),
			balance NUMBER
		)
	`)
	if err != nil {
		log.ErrorF("创建测试表失败: %v", err)
		return
	}
	defer func() {
		log.Info("清理测试表")
		if _, err := db.ExecContext(ctx, "DROP TABLE test_accounts"); err != nil {
			log.ErrorF("删除测试表失败: %v", err)
		}
	}()

	// 插入测试数据
	log.Info("插入初始测试数据")
	_, err = db.ExecContext(ctx, `
		INSERT INTO test_accounts (name, balance) VALUES ('账户A', 1000)
	`)
	if err != nil {
		log.ErrorF("插入测试数据失败: %v", err)
		return
	}

	// 测试事务
	t.Run("Transaction", func(t *testing.T) {
		log.Info("开始事务测试")
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			log.ErrorF("开始事务失败: %v", err)
			return
		}

		// 更新余额
		log.Info("执行事务更新")
		_, err = tx.ExecContext(ctx, "UPDATE test_accounts SET balance = balance - 100 WHERE name = :v1", "账户A")
		if err != nil {
			tx.Rollback()
			log.ErrorF("执行事务失败: %v", err)
			return
		}

		// 提交事务
		log.Info("提交事务")
		err = tx.Commit()
		if err != nil {
			log.ErrorF("提交事务失败: %v", err)
			return
		}

		// 验证结果
		var balance float64
		err = db.QueryRowContext(ctx, "SELECT balance FROM test_accounts WHERE name = :v1", "账户A").Scan(&balance)
		if err != nil {
			log.ErrorF("查询结果失败: %v", err)
			return
		}
		if balance != 900 {
			log.ErrorF("期望余额为900，实际为%.2f", balance)
			return
		}
		log.Info("事务测试成功", "最终余额", balance)
	})
}

// TestTransactionWithSavePoint 测试带保存点的事务操作
func TestTransactionWithSavePoint(t *testing.T) {
	log.Info("开始测试带保存点的事务操作")
	db, err := NewOracleDB(testConfig)
	if err != nil {
		log.ErrorF("创建数据库连接失败: %v", err)
		return
	}
	defer db.Close()

	ctx := context.Background()

	// 创建测试表
	log.Info("创建测试表")
	_, err = db.ExecContext(ctx, `
		CREATE TABLE test_accounts (
			id NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			name VARCHAR2(100),
			balance NUMBER
		)
	`)
	if err != nil {
		log.ErrorF("创建测试表失败: %v", err)
		return
	}
	defer func() {
		log.Info("清理测试表")
		if _, err := db.ExecContext(ctx, "DROP TABLE test_accounts"); err != nil {
			log.ErrorF("删除测试表失败: %v", err)
		}
	}()

	// 测试带保存点的事务
	t.Run("TransactionWithSavePoint", func(t *testing.T) {
		err := db.WithTransaction(ctx, func(tx *sql.Tx) error {
			// 插入初始数据
			_, err := tx.ExecContext(ctx, `
				INSERT INTO test_accounts (name, balance) VALUES ('账户A', 1000)
			`)
			if err != nil {
				return err
			}

			// 创建保存点
			if err := db.SavePoint(ctx, tx, "sp1"); err != nil {
				return err
			}

			// 执行第一次更新
			_, err = tx.ExecContext(ctx, `
				UPDATE test_accounts SET balance = balance - 200 WHERE name = '账户A'
			`)
			if err != nil {
				return err
			}

			// 创建第二个保存点
			if err := db.SavePoint(ctx, tx, "sp2"); err != nil {
				return err
			}

			// 执行第二次更新
			_, err = tx.ExecContext(ctx, `
				UPDATE test_accounts SET balance = balance - 300 WHERE name = '账户A'
			`)
			if err != nil {
				return err
			}

			// 回滚到第一个保存点
			if err := db.RollbackToSavePoint(ctx, tx, "sp1"); err != nil {
				return err
			}

			// 执行最终更新
			_, err = tx.ExecContext(ctx, `
				UPDATE test_accounts SET balance = balance - 100 WHERE name = '账户A'
			`)
			if err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			log.ErrorF("事务执行失败: %v", err)
			return
		}

		// 验证最终结果
		var balance float64
		err = db.QueryRowContext(ctx, `
			SELECT balance FROM test_accounts WHERE name = '账户A'
		`).Scan(&balance)
		if err != nil {
			log.ErrorF("查询结果失败: %v", err)
			return
		}

		// 期望余额为900（初始1000 - 最后更新的100）
		if balance != 900 {
			log.ErrorF("期望余额为900，实际为%.2f", balance)
			return
		}
		log.Info("事务测试成功", "最终余额", balance)
	})
}

// User 结构体优化
type User struct {
	ID        int64     `db:"ID"`
	Name      string    `db:"NAME"`
	Age       int       `db:"AGE"`
	CreatedAt time.Time `db:"CREATED_AT"`
	UpdatedAt time.Time `db:"UPDATED_AT"`
	RN        int64     `db:"RN"` // 用于分页
}

// TestPageQuery 测试分页查询
func TestPageQuery(t *testing.T) {
	log.Info("开始测试分页查询")
	db, err := NewOracleDB(testConfig)
	if err != nil {
		t.Fatalf("创建数据库连接失败: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// 创建测试表
	log.Info("创建测试表")
	_, err = db.ExecContext(ctx, `
		CREATE TABLE test_users (
			id NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			name VARCHAR2(100) NOT NULL,
			age NUMBER(3) NOT NULL,
			created_at TIMESTAMP DEFAULT SYSTIMESTAMP NOT NULL,
			updated_at TIMESTAMP DEFAULT SYSTIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}
	defer func() {
		log.Info("清理测试表")
		if _, err := db.ExecContext(ctx, "DROP TABLE test_users"); err != nil {
			t.Errorf("清理测试表失败: %v", err)
		}
	}()

	// 插入测试数据
	log.Info("插入测试数据")
	testData := []map[string]interface{}{
		{"name": "张三", "age": 25},
		{"name": "李四", "age": 30},
		{"name": "王五", "age": 35},
		{"name": "赵六", "age": 40},
		{"name": "钱七", "age": 45},
	}

	// 批量插入测试数据
	_, err = db.BatchInsert(ctx, "test_users", testData)
	if err != nil {
		t.Fatalf("批量插入测试数据失败: %v", err)
	}

	// 验证数据总数
	count, err := db.CountRows(ctx, "test_users", "")
	if err != nil {
		t.Fatalf("统计数据总数失败: %v", err)
	}
	if count != int64(len(testData)) {
		t.Errorf("期望总数为%d，实际为%d", len(testData), count)
	}
	log.Info("验证插入数据", "总数", count)

	// 测试分页查询 - 第一页
	t.Run("FirstPage", func(t *testing.T) {
		var users []User
		baseSQL := "SELECT * FROM test_users"
		page := PageQuery{
			PageNum:  1,
			PageSize: 2,
			OrderBy:  "id",
		}

		result, err := db.QueryPage(ctx, &users, baseSQL, page)
		if err != nil {
			t.Fatalf("分页查询失败: %v", err)
		}

		// 验证分页结果
		assertPageResult(t, result, len(testData), page.PageNum, page.PageSize, len(users))

		// 验证数据内容
		for i, user := range users {
			log.Info("用户数据",
				"序号", i+1,
				"ID", user.ID,
				"姓名", user.Name,
				"年龄", user.Age,
				"创建时间", user.CreatedAt.Format("2006-01-02 15:04:05"),
				"更新时间", user.UpdatedAt.Format("2006-01-02 15:04:05"),
			)

			// 验证数据完整性
			if user.Name == "" || user.Age == 0 || user.CreatedAt.IsZero() || user.UpdatedAt.IsZero() {
				t.Errorf("数据不完整: %+v", user)
			}
		}
	})

	// 测试分页查询 - 最后一页
	t.Run("LastPage", func(t *testing.T) {
		var users []User
		baseSQL := "SELECT * FROM test_users"
		lastPage := (len(testData) + 1) / 2 // 计算最后一页
		page := PageQuery{
			PageNum:  lastPage,
			PageSize: 2,
			OrderBy:  "id",
		}

		result, err := db.QueryPage(ctx, &users, baseSQL, page)
		if err != nil {
			t.Fatalf("查询最后一页失败: %v", err)
		}

		// 验证分页结果
		expectedLastPageSize := len(testData) - (lastPage-1)*2 // 计算最后一页应有的记录数
		assertPageResult(t, result, len(testData), lastPage, page.PageSize, expectedLastPageSize)

		// 验证数据排序
		for i := 1; i < len(users); i++ {
			if users[i].ID <= users[i-1].ID {
				t.Errorf("数据未正确排序: id[%d]=%d <= id[%d]=%d",
					i, users[i].ID, i-1, users[i-1].ID)
			}
		}
	})
}

// assertPageResult 验证分页结果
func assertPageResult(t *testing.T, result *PageResult, expectedTotal, pageNum, pageSize, dataSize int) {
	t.Helper()

	if result.Total != int64(expectedTotal) {
		t.Errorf("总数不匹配: 期望=%d, 实际=%d", expectedTotal, result.Total)
	}
	if result.PageNum != pageNum {
		t.Errorf("页码不匹配: 期望=%d, 实际=%d", pageNum, result.PageNum)
	}
	if result.PageSize != pageSize {
		t.Errorf("每页大小不匹配: 期望=%d, 实际=%d", pageSize, result.PageSize)
	}
	if dataSize > pageSize {
		t.Errorf("返回数据超出每页大小: 数据大小=%d, 每页大小=%d", dataSize, pageSize)
	}

	log.Info("分页结果验证",
		"总数", result.Total,
		"当前页", result.PageNum,
		"每页大小", result.PageSize,
		"当前数据量", dataSize,
	)
}
