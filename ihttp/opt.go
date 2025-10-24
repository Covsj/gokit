package ihttp

import "io"

// Opt 请求配置选项
type Opt struct {
	URL     string
	Proxy   string
	Method  string
	TimeOut int

	// 请求体类型 - 互斥使用
	Data  map[string]any  // Form数据
	Json  any             // JSON格式数据
	Files map[string]File // 文件上传

	Headers map[string]string
	Cookies *map[string]string // 使用指针类型，支持自动更新

	RespOut any // 响应体反序列化目标

	// 安全设置
	AllowRedirects bool // 是否允许重定向，默认true
	SkipVerify     bool // 是否跳过验证SSL证书，默认false
	FollowRedirect bool // 是否跟随重定向，默认true

	NotLog bool // 是否不记录日志
}

// File 文件上传结构
type File struct {
	Path        string    // 文件路径
	ContentType string    // MIME类型，空则自动检测
	Reader      io.Reader // 文件读取器，优先级高于Path
	FileName    string    // 文件名，空则使用Path的文件名
}

// NewOpt 创建新的请求配置
func NewOpt() *Opt {
	return &Opt{
		SkipVerify: false,
		TimeOut:    30, // 默认30秒超时
		Headers:    map[string]string{},
		Cookies:    &map[string]string{},
	}
}
