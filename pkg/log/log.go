package log

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger
var errLog *logrus.Logger
var once sync.Once

const (
	FormatJSON = "json"
	FormatText = "text"

	// 颜色代码
	colorRed     = 31
	colorGreen   = 32
	colorYellow  = 33
	colorBlue    = 36
	colorGray    = 37
	colorMagenta = 35 // 添加紫色用于时间戳
	colorCyan    = 36 // 添加青色用于调用信息

	// 修改字段显示的颜色
	colorFieldKey   = 36 // 青色用于字段名
	colorFieldValue = 33 // 黄色用于字段值
)

func init() {
	once.Do(func() {
		log = logrus.New()
		SetLogFormat(log, FormatText)
		log.SetLevel(logrus.InfoLevel)
		log.SetOutput(os.Stdout)

		errLog = logrus.New()
		SetLogFormat(errLog, FormatText)
		errLog.SetLevel(logrus.InfoLevel)
		errLog.SetOutput(os.Stderr)
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
	if len(level) < 7 {
		level = level + strings.Repeat(" ", 7-len(level))
	}

	// 根据日志级别选择颜色
	var levelColor int
	switch entry.Level {
	case logrus.DebugLevel:
		levelColor = colorBlue
	case logrus.InfoLevel:
		levelColor = colorGreen
	case logrus.WarnLevel:
		levelColor = colorYellow
	case logrus.ErrorLevel:
		levelColor = colorRed
	default:
		levelColor = colorGray
	}

	// 构建额外字段的字符串
	var fields string
	if len(entry.Data) > 0 {
		pairs := make([]string, 0, len(entry.Data))
		for k, v := range entry.Data {
			if k != "caller" { // 跳过caller字段，因为已经单独处理
				formattedValue := formatValue(v)
				pairs = append(pairs, fmt.Sprintf("\x1b[%dm%v\x1b[0m=\x1b[%dm%v\x1b[0m",
					colorFieldKey, k, // 字段名使用青色
					colorFieldValue, formattedValue)) // 字段值使用黄色
			}
		}
		if len(pairs) > 0 {
			fields = " " + strings.Join(pairs, " ")
		}
	}

	// 修改消息格式，添加字段信息
	msg := fmt.Sprintf("\x1b[%dm%s\x1b[0m [\x1b[%dm%s\x1b[0m] \x1b[%dm%s\x1b[0m %s%s\n",
		levelColor, level, // 日志级别颜色
		colorMagenta, timestamp, // 时间戳使用紫色
		colorCyan, caller, // 调用信息使用青色
		entry.Message,
		fields, // 添加额外字段
	)

	return []byte(msg), nil
}

// 添加新的辅助函数来格式化值
func formatValue(v interface{}) string {
	if v == nil {
		return "nil"
	}

	// 使用反射获取值的类型信息
	val := reflect.ValueOf(v)
	typ := val.Type()

	// 处理指针类型
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return "nil"
		}
		val = val.Elem()
		typ = val.Type()
	}

	switch val.Kind() {
	case reflect.Struct:
		// 构建结构体字段
		fields := make([]string, 0, val.NumField())
		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			// 获取json标签名，如果没有则使用字段名
			fieldName := field.Tag.Get("json")
			if fieldName == "" {
				fieldName = field.Name
			}
			// 去掉json标签中的omitempty等选项
			fieldName = strings.Split(fieldName, ",")[0]
			fields = append(fields, fmt.Sprintf("%s:%v", fieldName, val.Field(i).Interface()))
		}
		return "{" + strings.Join(fields, " ") + "}"
	case reflect.Map:
		// 处理map类型
		pairs := make([]string, 0, val.Len())
		for _, k := range val.MapKeys() {
			pairs = append(pairs, fmt.Sprintf("%v:%v", k.Interface(), val.MapIndex(k).Interface()))
		}
		return "{" + strings.Join(pairs, " ") + "}"
	case reflect.Slice, reflect.Array:
		// 处理切片和数组
		elements := make([]string, val.Len())
		for i := 0; i < val.Len(); i++ {
			elements[i] = fmt.Sprintf("%v", val.Index(i).Interface())
		}
		return "[" + strings.Join(elements, " ") + "]"
	default:
		return fmt.Sprintf("%v", v)
	}
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
	pc, _, line, ok := runtime.Caller(3)
	if ok {
		function := runtime.FuncForPC(pc)
		if function != nil {
			parts := strings.Split(function.Name(), ".")
			funcName := parts[len(parts)-1]
			fields["caller"] = fmt.Sprintf("%s:%d", funcName, line)
		}
	}
	return fields
}

func argsToFields(args ...interface{}) map[string]interface{} {
	fields := make(map[string]interface{})
	for i := 0; i < len(args); i += 2 {
		var key string
		switch k := args[i].(type) {
		case string:
			key = k
		default:
			key = fmt.Sprintf("%v", k)
		}
		if i+1 >= len(args) {
			fields[key] = ""
		} else {
			fields[key] = args[i+1]
		}
	}
	return fields
}
func logWithFields(logger *logrus.Logger, level logrus.Level, key interface{}, args ...interface{}) {
	fields := addCallerFields(argsToFields(args...))
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
	logWithFields(log, logrus.InfoLevel, key, args...)
}
func InfoF(format string, args ...interface{}) {
	logWithFieldsFormat(log, logrus.InfoLevel, format, args...)
}
func Error(key string, args ...interface{}) {
	logWithFields(errLog, logrus.ErrorLevel, key, args...)
}
func ErrorF(format string, args ...interface{}) {
	logWithFieldsFormat(errLog, logrus.ErrorLevel, format, args...)
}
func Warn(key string, args ...interface{}) {
	logWithFields(log, logrus.WarnLevel, key, args...)
}
func WarnF(format string, args ...interface{}) {
	logWithFieldsFormat(log, logrus.WarnLevel, format, args...)
}
func Debug(key string, args ...interface{}) {
	logWithFields(log, logrus.DebugLevel, key, args...)
}
func DebugF(format string, args ...interface{}) {
	logWithFieldsFormat(log, logrus.DebugLevel, format, args...)
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
