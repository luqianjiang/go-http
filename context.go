package go_http

import "net/http"

// Context 上下文
type Context struct {
	// 响应
	response http.ResponseWriter
	// 请求
	request *http.Request
	// Method 请求方式
	Method string
	// URL 请求路径
	Pattern string
}

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		response: w,
		request:  r,
		Method:   r.Method,
		Pattern:  r.URL.Path,
	}
}
