package go_http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type server interface {
	http.Handler //组合方式，将http包内置的引擎嵌入进来
	// Start 启动服务
	Start(add string) error
	// Stop 停止服务
	Stop(add string) error
	// addRouter 路由注册，核心api，不能对外暴露
	// 造一些衍生api给外部使用
	addRouter(method string, pattern string, handleFunc HandleFunc)
}

// Option
type HTTPOption func(h *HTTPServer)

// HandleFunc 视图函数签名
type HandleFunc func(w http.ResponseWriter, r *http.Request)

type HTTPServer struct {
	srv  *http.Server
	stop func() error
	// routers 临时存放路由的位置，数据格式例如："GET-login":HandleFunc1
	routers map[string]HandleFunc
}

func WithHTTPServerStop(fn func() error) HTTPOption {
	return func(h *HTTPServer) {
		if fn == nil {
			fn = func() error {
				fmt.Println("1231231312")
				quit := make(chan os.Signal)
				signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // 比如按住ctrl+c停止，会发送信号到quit,不然会一直阻塞
				<-quit
				log.Println("Shutdown Server ...")

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				// 关闭之前：需要做某些操作
				if err := h.srv.Shutdown(ctx); err != nil {
					log.Fatal("Server Shutdown:", err)
				}
				// 关闭之后，需要做某些操作
				select {
				case <-ctx.Done():
					log.Println("timeout of 5 seconds.")
				}
				return nil
			}
		}
		h.stop = fn
	}
}
func NewHTTP(opts ...HTTPOption) *HTTPServer {
	h := &HTTPServer{
		routers: map[string]HandleFunc{},
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// ServeHTTP 接受前端请求，转发前端请求到自己的框架中
func (h *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. 匹配路由
	key := fmt.Sprintf("%s-%s", r.Method, r.URL.Path)
	handler, ok := h.routers[key]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 not found"))
		return
	}
	// 转发请求
	handler(w, r)

}

// Start 启动服务
func (h *HTTPServer) Start(addr string) error {
	h.srv = &http.Server{
		Addr:    addr,
		Handler: h,
	}
	return h.srv.ListenAndServe()
}

// Stop 停止服务
func (h *HTTPServer) Stop(add string) error {
	return h.stop()
}

// addRouter 注册路由
// 注册路由的时机： 就是项目启动的时候注册，启动后不能就再注册
// 问题一： 注册路由放哪里
func (h *HTTPServer) addRouter(method string, pattern string, handleFunc HandleFunc) {
	// 构建唯一的key
	key := fmt.Sprintf("%s-%s", method, pattern)
	fmt.Printf("add router %s - %s\n", method, pattern)
	h.routers[key] = handleFunc
}

// GET 请求
func (h *HTTPServer) GET(pattern string, handleFunc HandleFunc) {
	h.addRouter(http.MethodGet, pattern, handleFunc)
}

// POST 请求
func (h *HTTPServer) POST(pattern string, handleFunc HandleFunc) {
	h.addRouter(http.MethodPost, pattern, handleFunc)
}

// DELETE 请求
func (h *HTTPServer) DELETE(pattern string, handleFunc HandleFunc) {
	h.addRouter(http.MethodDelete, pattern, handleFunc)
}

// PUT 请求
func (h *HTTPServer) PUT(pattern string, handleFunc HandleFunc) {
	h.addRouter(http.MethodPut, pattern, handleFunc)
}

// 一个server需要什么功能
// 1. 启动
// 2. 关闭
// 3. 注册路由

func main() {
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		// 业务逻辑
	})
	http.ListenAndServe(":8080", nil)
	h := NewHTTP(WithHTTPServerStop(nil))
	go func() {
		err := h.Start(":8080")
		if err != nil && err != http.ErrServerClosed {
			panic("启动失败")
		}
	}()

	err := h.stop()
	if err != nil {
		panic("关闭失败")
	}

}
