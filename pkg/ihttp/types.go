package ihttp

import (
	"mime/multipart"
	"time"

	"github.com/valyala/fasthttp"
)

var (
	DefaultUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.360"
)

// Options 定义 HTTP 请求的配置选项
type Options struct {
	// 基本请求信息
	Method  string            // HTTP 方法: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS 等
	URL     string            // 请求 URL，完整的目标地址
	Headers map[string]string // 请求头，自定义 HTTP 头部信息
	Query   map[string]string // URL 查询参数

	// 请求体数据 - 各种格式支持
	Body           interface{}            // 通用请求体，可以是任意类型
	JSONBody       interface{}            // JSON 格式的请求体，会自动设置 Content-Type: application/json
	XMLBody        interface{}            // XML 格式的请求体，会自动设置 Content-Type: application/xml
	FormData       map[string]interface{} // multipart/form-data 格式的表单数据
	FormURLEncoded map[string]string      // application/x-www-form-urlencoded 格式的表单数据
	RawBody        string                 // 原始字符串请求体
	BinaryBody     []byte                 // 二进制请求体
	GraphQLQuery   string                 // GraphQL 查询语句
	GraphQLVars    map[string]interface{} // GraphQL 变量
	MsgpackBody    interface{}            // MessagePack 格式的请求体

	// 文件上传
	Files       map[string]string                // 要上传的文件，键为字段名，值为文件路径
	FileReaders map[string]*multipart.FileHeader // 文件读取器，用于流式上传

	BearerToken string // Bearer Token 认证

	// 响应处理
	ResponseOut    interface{} // 用于解析响应的结构体指针
	ExpectedStatus []int       // 预期的 HTTP 状态码，如果响应码不在此列表中，将返回错误

	// 高级请求配置
	Timeout       time.Duration // 请求超时时间
	RetryCount    int           // 重试次数
	RetryInterval time.Duration // 重试延迟时间
	// TLS 和代理配置
	SkipTLSVerify bool // 是否跳过 TLS 证书验证

	// 高级功能
	DisableCompression bool   // 禁用响应压缩
	UserAgent          string // 自定义 User-Agent

	DebugMode bool // 是否启用调试模式，输出请求和响应的详细信息

}

// Client 是 HTTP 客户端的包装器，提供链式调用 API
type Client struct {
	client  *fasthttp.Client // 底层的 fasthttp 客户端
	options *Options         // 客户端选项
}

// Request 表示一个 HTTP 请求
type Request struct {
	client  *Client  // 客户端引用
	options *Options // 请求选项
}

// Response 封装 HTTP 响应
type Response struct {
	StatusCode    int                // HTTP 状态码
	Body          []byte             // 响应体
	Headers       map[string]string  // 响应头
	ContentLength int                // 内容长度
	Raw           *fasthttp.Response // 原始响应对象
}
