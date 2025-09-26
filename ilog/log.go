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

var Log *logrus.Logger
var ErrLog *logrus.Logger
var once sync.Once

const (
	FormatJSON = "json"
	FormatText = "text"

	// 颜色代码 - 更美观的配色方案
	colorWhite = 37 // 白色 - 默认

	// 字段显示的颜色 - 更醒目的配色
	colorFieldKey   = 95 // 亮紫色用于字段名，更醒目
	colorFieldValue = 33 // 黄色用于字段值，更突出
)

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
}

func (f *customTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 时间戳格式
	f.TimestampFormat = "01-02 15:04:05"

	// 获取调用信息
	caller := ""
	if c, ok := entry.Data["caller"]; ok {
		caller = fmt.Sprintf("[%s]", c)
		delete(entry.Data, "caller")
	}

	// 格式化级别和时间戳
	timestamp := entry.Time.Format(f.TimestampFormat)
	level := strings.ToUpper(entry.Level.String())
	// 保持Level为四个字符，自定义Level
	switch level {
	case "DEBUG":
		level = "DBUG"
	case "WARNING":
		level = "WARN"
	case "ERROR":
		level = "ERRO"
	}
	// 根据日志级别选择颜色 - 更鲜明的配色
	var levelColor int
	switch entry.Level {
	case logrus.DebugLevel:
		levelColor = 94 // 亮蓝色 - 调试信息
	case logrus.InfoLevel:
		levelColor = 92 // 亮绿色 - 一般信息
	case logrus.WarnLevel:
		levelColor = 93 // 亮黄色 - 警告信息
	case logrus.ErrorLevel:
		levelColor = 91 // 亮红色 - 错误信息
	default:
		levelColor = colorWhite
	}

	// 构建额外字段的字符串
	var fields string
	if len(entry.Data) > 0 {
		pairs := make([]string, 0, len(entry.Data))

		// 检查是否有有序字段信息
		if orderedFields, ok := entry.Data["_ordered_fields"]; ok {
			// 使用有序字段
			if fieldsList, ok := orderedFields.([]orderedField); ok {
				for _, field := range fieldsList {
					if field.key != "caller" && field.key != "_ordered_fields" {
						formattedValue := formatValue(field.value)
						pairs = append(pairs, fmt.Sprintf("\x1b[%dm%s\x1b[0m=\x1b[%dm%s\x1b[0m",
							colorFieldKey, field.key, // 字段名使用亮青色
							colorFieldValue, formattedValue)) // 字段值使用亮黄色
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
						colorFieldKey, k, // 字段名使用亮青色
						colorFieldValue, formattedValue)) // 字段值使用亮黄色
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
		95, timestamp, // 亮紫色用于时间戳
		96, caller, // 亮青色用于调用信息
		94, entry.Message, // 蓝色消息，不加粗
		fields, // 添加额外字段
	)

	return []byte(msg), nil
}

// 添加新的辅助函数来格式化值
func formatValue(v interface{}) string {
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
		})
	default:
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "01-02 15:04:05",
		})
	}
}

func addCallerFields(fields map[string]interface{}) logrus.Fields {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	pc, fileName, line, ok := runtime.Caller(3)
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

// 有序字段结构
type orderedField struct {
	key   string
	value interface{}
}

func argsToFields(args ...interface{}) (map[string]interface{}, []orderedField) {
	fields := make(map[string]interface{})
	orderedFields := make([]orderedField, 0)

	for i := 0; i < len(args); i += 2 {
		var key string
		switch k := args[i].(type) {
		case string:
			key = k
		default:
			key = fmt.Sprintf("%v", k)
		}

		var value interface{}
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
func logWithFields(logger *logrus.Logger, level logrus.Level, key interface{}, args ...interface{}) {
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

func logWithFieldsFormat(logger *logrus.Logger, level logrus.Level, format string, args ...interface{}) {
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
func Info(key interface{}, args ...interface{}) {
	logWithFields(Log, logrus.InfoLevel, key, args...)
}
func InfoF(format string, args ...interface{}) {
	logWithFieldsFormat(Log, logrus.InfoLevel, format, args...)
}
func Error(key string, args ...interface{}) {
	logWithFields(ErrLog, logrus.ErrorLevel, key, args...)
}
func ErrorF(format string, args ...interface{}) {
	logWithFieldsFormat(ErrLog, logrus.ErrorLevel, format, args...)
}
func Warn(key string, args ...interface{}) {
	logWithFields(Log, logrus.WarnLevel, key, args...)
}
func WarnF(format string, args ...interface{}) {
	logWithFieldsFormat(Log, logrus.WarnLevel, format, args...)
}
func Debug(key string, args ...interface{}) {
	logWithFields(Log, logrus.DebugLevel, key, args...)
}
func DebugF(format string, args ...interface{}) {
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
