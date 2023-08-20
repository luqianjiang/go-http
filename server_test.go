package go_http

import (
	"testing"
)

func TestHTTP_ServeHTTP(t *testing.T) {
	h := NewHTTP(WithHTTPServerStop(nil))
	//err := h.Start(":8080")
	//if err != nil {
	//	t.Failed()
	//}
	go func() {
		err := h.Start(":8080")
		if err != nil {
			t.Fail()
		}
	}()

	err := h.stop()
	if err != nil {
		t.Fail()
	}
}
func Login(ctx *Context) {
	ctx.response.Write([]byte("login请求成功"))
}

func Register(ctx *Context) {
	ctx.response.Write([]byte("Register请求成功"))
}

func Index(ctx *Context) {
	ctx.response.Write([]byte("index请求成功"))
}

func TestHTTP_Start(t *testing.T) {
	h := NewHTTP()
	h.GET("/login", Login)
	h.POST("/register", Register)
	h.DELETE("/", Index)
	err := h.Start(":8080")
	if err != nil {
		panic(err)
	}
}

func TestHTTP_Stop(t *testing.T) {

}
