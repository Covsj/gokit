package core

import (
	"context"
	"time"

	"github.com/Covsj/gokit/pkg/ilog"
	"gorm.io/gorm/logger"
)

// GormLogAdapter 实现GORM的logger.Interface接口
type GormLogAdapter struct {
	LogLevel      logger.LogLevel
	SlowThreshold time.Duration
}

// NewGormLogAdapter 创建GORM日志适配器
func NewGormLogAdapter(level logger.LogLevel, slowThreshold time.Duration) *GormLogAdapter {
	if slowThreshold == 0 {
		slowThreshold = 200 * time.Millisecond
	}

	return &GormLogAdapter{
		LogLevel:      level,
		SlowThreshold: slowThreshold,
	}
}

// LogMode 设置日志级别
func (l *GormLogAdapter) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info 输出信息级别日志
func (l *GormLogAdapter) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		ilog.InfoF(msg, data...)
	}
}

// Warn 输出警告级别日志
func (l *GormLogAdapter) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		ilog.WarnF(msg, data...)
	}
}

// Error 输出错误级别日志
func (l *GormLogAdapter) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		ilog.ErrorF(msg, data...)
	}
}

// Trace 跟踪SQL执行
func (l *GormLogAdapter) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// 根据执行结果和耗时决定日志级别
	if err != nil {
		// 出错记录为错误
		ilog.Error("SQL错误",
			"sql", sql,
			"rows", rows,
			"elapsed", elapsed,
			"error", err)
	} else if l.LogLevel >= logger.Info {
		// 慢查询记录为警告
		if elapsed > l.SlowThreshold {
			ilog.Warn("慢查询",
				"sql", sql,
				"rows", rows,
				"elapsed", elapsed)
		} else if l.LogLevel >= logger.Info {
			// 正常查询记录为信息
			ilog.Debug("SQL执行",
				"sql", sql,
				"rows", rows,
				"elapsed", elapsed)
		}
	}
}
