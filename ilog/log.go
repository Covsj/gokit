// Package ilog 提供灵活的日志记录功能，支持自定义颜色、调用栈深度等配置
//
// 使用示例:
//
//  1. 基本使用:
//     ilog.Info("用户登录", "user", "张三", "ip", "192.168.1.1")
//     ilog.Error("数据库连接失败", "error", "connection timeout")
//
//  2. 自定义颜色:
//     ilog.SetColors(
//     struct { Debug, Info, Warn, Error int }{
//     Debug: ilog.ColorBrightCyan,  // 亮青色调试
//     Info:  ilog.ColorBrightGreen, // 亮绿色信息
//     Warn:  ilog.ColorBrightYellow, // 亮黄色警告
//     Error: ilog.ColorBrightRed,   // 亮红色错误
//     },
//     ilog.ColorBrightWhite, // 亮白色时间戳
//     ilog.ColorBrightCyan,  // 亮青色调用信息
//     ilog.ColorWhite,       // 白色消息
//     ilog.ColorBrightMagenta, // 亮紫色字段名
//     ilog.ColorBrightYellow, // 亮黄色字段值
//     )
//
//  3. 使用预设主题:
//     ilog.SetDarkTheme()   // 深色主题
//     ilog.SetLightTheme()  // 浅色主题
//     ilog.SetMonochrome()  // 单色主题
//
//  4. 自定义调用栈:
//     ilog.SetCallerConfig(2, true) // 跳过2层调用栈，启用调用栈信息
//
//  5. 自定义时间格式:
//     ilog.SetTimestampFormat("2006-01-02 15:04:05")
//
//  6. 自定义级别显示:
//     ilog.SetLevelDisplay("DBG", "INF", "WRN", "ERR")
package ilog

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

// 性能优化：使用对象池减少内存分配
var (
	fieldPairsPool = sync.Pool{
		New: func() interface{} {
			return make([]string, 0, 16) // 预分配容量
		},
	}

	orderedFieldsPool = sync.Pool{
		New: func() interface{} {
			return make([]orderedField, 0, 16) // 预分配容量
		},
	}
)

var Log *logrus.Logger
var ErrLog *logrus.Logger
var once sync.Once

const (
	FormatJSON = "json"
	FormatText = "text"
)

// ANSI 颜色代码常量 - 方便用户配置时参考
const (
	// 基础颜色
	ColorBlack   = 30 // 黑色
	ColorRed     = 31 // 红色
	ColorGreen   = 32 // 绿色
	ColorYellow  = 33 // 黄色
	ColorBlue    = 34 // 蓝色
	ColorMagenta = 35 // 紫色
	ColorCyan    = 36 // 青色
	ColorWhite   = 37 // 白色

	// 亮色 (高亮)
	ColorBrightBlack   = 90 // 亮黑色
	ColorBrightRed     = 91 // 亮红色
	ColorBrightGreen   = 92 // 亮绿色
	ColorBrightYellow  = 93 // 亮黄色
	ColorBrightBlue    = 94 // 亮蓝色
	ColorBrightMagenta = 95 // 亮紫色
	ColorBrightCyan    = 96 // 亮青色
	ColorBrightWhite   = 97 // 亮白色

	// 常用组合颜色
	ColorDebug = 94 // 亮蓝色 - 适合调试信息
	ColorInfo  = 92 // 亮绿色 - 适合一般信息
	ColorWarn  = 93 // 亮黄色 - 适合警告信息
	ColorError = 91 // 亮红色 - 适合错误信息

	// 特殊用途颜色
	ColorTimestamp  = 95 // 亮紫色 - 适合时间戳
	ColorCaller     = 96 // 亮青色 - 适合调用信息
	ColorMessage    = 94 // 蓝色 - 适合消息内容
	ColorFieldKey   = 95 // 亮紫色 - 适合字段名
	ColorFieldValue = 33 // 黄色 - 适合字段值
)

// 日志配置结构体
type LogConfig struct {
	// 颜色配置
	Colors struct {
		Level struct {
			Debug int // 调试级别颜色
			Info  int // 信息级别颜色
			Warn  int // 警告级别颜色
			Error int // 错误级别颜色
		}
		Timestamp  int // 时间戳颜色
		Caller     int // 调用信息颜色
		Message    int // 消息颜色
		FieldKey   int // 字段名颜色
		FieldValue int // 字段值颜色
	}

	// 调用栈配置
	Caller struct {
		SkipDepth int  // 调用栈跳过深度
		Enabled   bool // 是否启用调用栈信息
	}

	// 时间格式配置
	TimestampFormat string

	// 级别显示配置
	LevelDisplay struct {
		Debug string // 调试级别显示文本
		Info  string // 信息级别显示文本
		Warn  string // 警告级别显示文本
		Error string // 错误级别显示文本
	}
}

// 默认配置
var currentConfig = LogConfig{
	Colors: struct {
		Level struct {
			Debug int
			Info  int
			Warn  int
			Error int
		}
		Timestamp  int
		Caller     int
		Message    int
		FieldKey   int
		FieldValue int
	}{
		Level: struct {
			Debug int
			Info  int
			Warn  int
			Error int
		}{
			Debug: ColorDebug, // 亮蓝色
			Info:  ColorInfo,  // 亮绿色
			Warn:  ColorWarn,  // 亮黄色
			Error: ColorError, // 亮红色
		},
		Timestamp:  ColorTimestamp,  // 亮紫色
		Caller:     ColorCaller,     // 亮青色
		Message:    ColorMessage,    // 蓝色
		FieldKey:   ColorFieldKey,   // 亮紫色
		FieldValue: ColorFieldValue, // 黄色
	},
	Caller: struct {
		SkipDepth int
		Enabled   bool
	}{
		SkipDepth: 3, // 默认跳过3层调用栈
		Enabled:   true,
	},
	TimestampFormat: "01-02 15:04:05",
	LevelDisplay: struct {
		Debug string
		Info  string
		Warn  string
		Error string
	}{
		Debug: "DBUG",
		Info:  "INFO",
		Warn:  "WARN",
		Error: "ERRO",
	},
}

var configMutex sync.RWMutex // 配置读写锁，保证线程安全

func init() {
	once.Do(func() {
		Log = logrus.New()
		SetLogFormat(Log, FormatText)
		Log.SetLevel(logrus.InfoLevel)
		Log.SetOutput(os.Stdout)

		ErrLog = logrus.New()
		SetLogFormat(ErrLog, FormatText)
		ErrLog.SetLevel(logrus.InfoLevel)
		ErrLog.SetOutput(os.Stderr)
	})
}

// 自定义文本格式化器
type customTextFormatter struct {
	logrus.TextFormatter
	config *LogConfig
}

func (f *customTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 使用配置中的时间戳格式
	f.TimestampFormat = f.config.TimestampFormat

	// 获取调用信息
	caller := ""
	if c, ok := entry.Data["caller"]; ok {
		caller = fmt.Sprintf("[%s]", c)
		delete(entry.Data, "caller")
	}

	// 格式化级别和时间戳
	timestamp := entry.Time.Format(f.TimestampFormat)
	level := strings.ToUpper(entry.Level.String())

	// 使用配置中的级别显示文本
	switch level {
	case "DEBUG":
		level = f.config.LevelDisplay.Debug
	case "INFO":
		level = f.config.LevelDisplay.Info
	case "WARNING":
		level = f.config.LevelDisplay.Warn
	case "ERROR":
		level = f.config.LevelDisplay.Error
	}

	// 根据日志级别选择颜色 - 使用配置中的颜色
	var levelColor int
	switch entry.Level {
	case logrus.DebugLevel:
		levelColor = f.config.Colors.Level.Debug
	case logrus.InfoLevel:
		levelColor = f.config.Colors.Level.Info
	case logrus.WarnLevel:
		levelColor = f.config.Colors.Level.Warn
	case logrus.ErrorLevel:
		levelColor = f.config.Colors.Level.Error
	default:
		levelColor = 37 // 白色
	}

	// 构建额外字段的字符串
	var fields string
	if len(entry.Data) > 0 {
		// 使用对象池减少内存分配
		pairs := fieldPairsPool.Get().([]string)
		pairs = pairs[:0] // 重置长度但保留容量
		defer fieldPairsPool.Put(pairs)

		// 检查是否有有序字段信息
		if orderedFields, ok := entry.Data["_ordered_fields"]; ok {
			// 使用有序字段
			if fieldsList, ok := orderedFields.([]orderedField); ok {
				for _, field := range fieldsList {
					if field.key != "caller" && field.key != "_ordered_fields" {
						formattedValue := formatValue(field.value)
						pairs = append(pairs, fmt.Sprintf("\x1b[%dm%s\x1b[0m=\x1b[%dm%s\x1b[0m",
							f.config.Colors.FieldKey, field.key, // 字段名使用配置的颜色
							f.config.Colors.FieldValue, formattedValue)) // 字段值使用配置的颜色
					}
				}
			}
		} else {
			// 回退到原来的逻辑（按字母顺序排序）
			keys := make([]string, 0, len(entry.Data))
			for k := range entry.Data {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				if k != "caller" { // 跳过caller字段，因为已经单独处理
					formattedValue := formatValue(entry.Data[k])
					pairs = append(pairs, fmt.Sprintf("\x1b[%dm%s\x1b[0m=\x1b[%dm%s\x1b[0m",
						f.config.Colors.FieldKey, k, // 字段名使用配置的颜色
						f.config.Colors.FieldValue, formattedValue)) // 字段值使用配置的颜色
				}
			}
		}

		if len(pairs) > 0 {
			fields = " │ " + strings.Join(pairs, " ")
		}
	}

	// 修改消息格式，添加字段信息 - 更美观的布局
	msg := fmt.Sprintf("\x1b[%dm%s\x1b[0m \x1b[%dm%s\x1b[0m \x1b[%dm%s\x1b[0m \x1b[%dm%s\x1b[0m%s\n",
		levelColor, level, // 日志级别颜色
		f.config.Colors.Timestamp, timestamp, // 使用配置的时间戳颜色
		f.config.Colors.Caller, caller, // 使用配置的调用信息颜色
		f.config.Colors.Message, entry.Message, // 使用配置的消息颜色
		fields, // 添加额外字段
	)

	return []byte(msg), nil
}

// 添加新的辅助函数来格式化值
func formatValue(v any) string {
	if v == nil {
		return "nil"
	}

	// 特殊处理 error 类型
	if err, ok := v.(error); ok {
		return err.Error()
	}

	// 处理 fmt.Stringer 接口
	if s, ok := v.(fmt.Stringer); ok {
		return s.String()
	}

	// 对基本类型直接使用 fmt.Sprint
	switch v.(type) {
	case string, bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return fmt.Sprint(v)
	}

	// 对复杂类型使用 json 序列化
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}

	// 如果是简单的字符串，去掉多余的引号
	str := string(jsonBytes)
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}
	return str
}

func SetLogFormat(log *logrus.Logger, format string) {
	switch format {
	case FormatText:
		log.SetFormatter(&customTextFormatter{
			TextFormatter: logrus.TextFormatter{
				ForceColors:            true,
				DisableColors:          false,
				FullTimestamp:          true,
				DisableLevelTruncation: false,
			},
			config: &currentConfig,
		})
	default:
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: currentConfig.TimestampFormat,
		})
	}
}

func addCallerFields(fields map[string]any) logrus.Fields {
	if fields == nil {
		fields = make(map[string]any)
	}

	// 如果调用栈信息被禁用，直接返回
	if !currentConfig.Caller.Enabled {
		return fields
	}

	pc, fileName, line, ok := runtime.Caller(currentConfig.Caller.SkipDepth)
	if ok {
		function := runtime.FuncForPC(pc)
		if function != nil {
			parts := strings.Split(function.Name(), ".")
			fileName = strings.TrimSuffix(strings.Split(fileName, "/")[len(strings.Split(fileName, "/"))-1], ".go")
			funcName := parts[len(parts)-1]
			fields["caller"] = fmt.Sprintf("%s:%s:%d", fileName, funcName, line)
		}
	}
	return fields
}

// 配置设置函数

// SetConfig 设置全局日志配置（线程安全）
func SetConfig(config LogConfig) {
	configMutex.Lock()
	defer configMutex.Unlock()
	currentConfig = config
	// 重新设置格式化器以应用新配置
	SetLogFormat(Log, FormatText)
	SetLogFormat(ErrLog, FormatText)
}

// GetConfig 获取当前配置（线程安全）
func GetConfig() LogConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return currentConfig
}

// SetColors 设置颜色配置
func SetColors(levelColors struct {
	Debug int
	Info  int
	Warn  int
	Error int
}, timestamp, caller, message, fieldKey, fieldValue int) {
	currentConfig.Colors.Level.Debug = levelColors.Debug
	currentConfig.Colors.Level.Info = levelColors.Info
	currentConfig.Colors.Level.Warn = levelColors.Warn
	currentConfig.Colors.Level.Error = levelColors.Error
	currentConfig.Colors.Timestamp = timestamp
	currentConfig.Colors.Caller = caller
	currentConfig.Colors.Message = message
	currentConfig.Colors.FieldKey = fieldKey
	currentConfig.Colors.FieldValue = fieldValue
	// 重新设置格式化器以应用新配置
	SetLogFormat(Log, FormatText)
	SetLogFormat(ErrLog, FormatText)
}

// SetCallerConfig 设置调用栈配置
func SetCallerConfig(skipDepth int, enabled bool) {
	currentConfig.Caller.SkipDepth = skipDepth
	currentConfig.Caller.Enabled = enabled
}

// SetTimestampFormat 设置时间戳格式
func SetTimestampFormat(format string) {
	currentConfig.TimestampFormat = format
	// 重新设置格式化器以应用新配置
	SetLogFormat(Log, FormatText)
	SetLogFormat(ErrLog, FormatText)
}

// SetLevelDisplay 设置级别显示文本
func SetLevelDisplay(debug, info, warn, error string) {
	currentConfig.LevelDisplay.Debug = debug
	currentConfig.LevelDisplay.Info = info
	currentConfig.LevelDisplay.Warn = warn
	currentConfig.LevelDisplay.Error = error
	// 重新设置格式化器以应用新配置
	SetLogFormat(Log, FormatText)
	SetLogFormat(ErrLog, FormatText)
}

// 日志级别过滤和批量设置功能

// SetLogLevel 设置日志级别
func SetLogLevel(level logrus.Level) {
	configMutex.Lock()
	defer configMutex.Unlock()
	Log.SetLevel(level)
	ErrLog.SetLevel(level)
}

// SetLogLevelString 通过字符串设置日志级别
func SetLogLevelString(level string) {
	var logLevel logrus.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = logrus.DebugLevel
	case "info":
		logLevel = logrus.InfoLevel
	case "warn", "warning":
		logLevel = logrus.WarnLevel
	case "error":
		logLevel = logrus.ErrorLevel
	default:
		logLevel = logrus.InfoLevel
	}
	SetLogLevel(logLevel)
}

// SetOutputFile 设置日志输出到文件
func SetOutputFile(filename string, log *logrus.Logger) error {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	log.SetOutput(file)
	return nil
}

// 有序字段结构
type orderedField struct {
	key   string
	value any
}

func argsToFields(args ...any) (map[string]any, []orderedField) {
	fields := make(map[string]any)
	// 使用对象池减少内存分配
	orderedFields := orderedFieldsPool.Get().([]orderedField)
	orderedFields = orderedFields[:0] // 重置长度但保留容量
	defer orderedFieldsPool.Put(orderedFields)

	for i := 0; i < len(args); i += 2 {
		var key string
		switch k := args[i].(type) {
		case string:
			key = k
		default:
			key = fmt.Sprintf("%v", k)
		}

		var value any
		if i+1 >= len(args) {
			value = ""
		} else {
			value = args[i+1]
		}

		fields[key] = value
		orderedFields = append(orderedFields, orderedField{key: key, value: value})
	}
	return fields, orderedFields
}
func logWithFields(logger *logrus.Logger, level logrus.Level, key any, args ...any) {
	fields, orderedFields := argsToFields(args...)
	fields = addCallerFields(fields)

	// 将有序字段信息存储到 Data 中，使用特殊键名
	fields["_ordered_fields"] = orderedFields

	switch level {
	case logrus.InfoLevel:
		logger.WithFields(fields).Info(key)
	case logrus.ErrorLevel:
		logger.WithFields(fields).Error(key)
	case logrus.WarnLevel:
		logger.WithFields(fields).Warn(key)
	case logrus.DebugLevel:
		logger.WithFields(fields).Debug(key)
	default:
		logger.WithFields(fields).Info(key)
	}
}

func logWithFieldsFormat(logger *logrus.Logger, level logrus.Level, format string, args ...any) {
	fields := addCallerFields(nil)

	switch level {
	case logrus.InfoLevel:
		logger.WithFields(fields).Infof(format, args...)
	case logrus.ErrorLevel:
		logger.WithFields(fields).Errorf(format, args...)
	case logrus.WarnLevel:
		logger.WithFields(fields).Warnf(format, args...)
	case logrus.DebugLevel:
		logger.WithFields(fields).Debugf(format, args...)
	default:
		logger.WithFields(fields).Infof(format, args...)
	}
}
func Info(key any, args ...any) {
	logWithFields(Log, logrus.InfoLevel, key, args...)
}
func InfoF(format string, args ...any) {
	logWithFieldsFormat(Log, logrus.InfoLevel, format, args...)
}
func Error(key string, args ...any) {
	logWithFields(ErrLog, logrus.ErrorLevel, key, args...)
}
func ErrorF(format string, args ...any) {
	logWithFieldsFormat(ErrLog, logrus.ErrorLevel, format, args...)
}
func Warn(key string, args ...any) {
	logWithFields(Log, logrus.WarnLevel, key, args...)
}
func WarnF(format string, args ...any) {
	logWithFieldsFormat(Log, logrus.WarnLevel, format, args...)
}
func Debug(key string, args ...any) {
	logWithFields(Log, logrus.DebugLevel, key, args...)
}
func DebugF(format string, args ...any) {
	logWithFieldsFormat(Log, logrus.DebugLevel, format, args...)
}

type Logger struct {
	*logrus.Logger
	file *os.File
}

func NewLogger(level logrus.Level, format logrus.Formatter, outputFilePath string) (*Logger, error) {
	l := &Logger{
		Logger: logrus.New(),
	}
	l.SetLevel(level)

	if format == nil {
		l.SetFormatter(&logrus.JSONFormatter{})
	} else {
		l.SetFormatter(format)
	}

	if outputFilePath != "" {
		file, err := os.OpenFile(outputFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		l.SetOutput(file)
		l.file = file
	} else {
		l.SetOutput(os.Stdout)
	}

	return l, nil
}

func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
