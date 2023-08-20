package go_http

import (
	"fmt"
	"strings"
)

// router 路由数
/**
{
	"GET": node{},
	"POST": node{},
	"DELETE": node{}
	....
}
根节点，咱们直接用 / 代替
**/
type router struct {
	trees map[string]*node
}

func newRouter() *router {
	// trees要初始化，否则会报错
	return &router{trees: map[string]*node{}}
}

type node struct {
	part     string
	children map[string]*node
	// handleFunc 这里存的是当前节点上的视图函数
	// 就是咱们之前讲的data
	handleFunc HandleFunc
}

// addRouter 注册路由
// 考虑点
// 1. 什么样的pattern是合法的
func (r *router) addRouter(method, pattern string, handleFunc HandleFunc) {
	if pattern == "" {
		panic("web:路由不能为空")
	}

	// 获取节点
	root, ok := r.trees[method]
	if !ok {
		//根节点不存在
		// 1. 创建根节点
		// 2. 把根节点放到trees中
		root = &node{
			part: "/",
		}
		r.trees[method] = root
	}

	// TODO 如果是根路由怎么办？
	// method = "get",pattern = "/"，为节点绑定一个视图函数
	if pattern == "/" {
		root.handleFunc = handleFunc
		return
	}
	if !strings.HasPrefix(pattern, "/") {
		panic("web: 路由必须 / 开头")
	}
	if strings.HasSuffix(pattern, "/") {
		panic("web: 路由不准 / 结尾")
	}
	fmt.Printf("add router %s - %s \n", method, pattern)

	// 切割pattern
	// /user/login => ["", "user", "login"]
	parts := strings.Split(pattern[1:], "/")
	for _, part := range parts {
		if part == "" {
			panic("web: 路由不能连续出现 /")
		}
		root = root.addNode(part)
	}
	root.handleFunc = handleFunc

}

// getRouter 匹配路由
// pattern校验： /user/login/获取是可以的，但是add是不行
func (r *router) getRouter(method, pattern string) (*node, bool) {
	if pattern == "" {
		return nil, false
	}

	// 获取根节点
	root, ok := r.trees[method]
	if !ok {
		return nil, false
	}

	// TODO 跟路由怎么办
	if pattern == "/" {
		return root, true
	}
	// 切割pattern
	parts := strings.Split(strings.Trim(pattern, "/"), "/")
	for _, part := range parts {
		if part == "" {
			return nil, false
		}
		root = root.getNode(part)
		if root == nil {
			return nil, false
		}
	}
	return root, true
}

func (n *node) addNode(part string) *node {
	// 判断当前节点有没有children属性，就是说，是不是nil
	if n.children == nil {
		n.children = make(map[string]*node)
	}
	// 找到就直接返回，找不到创建
	child, ok := n.children[part]
	if !ok {
		child = &node{
			part: part,
		}
		n.children[part] = child
	}
	return child
}

func (n *node) getNode(part string) *node {
	// n 的 children属性都不存在
	if n.children == nil {
		return nil
	}
	// 正常思路
	child, ok := n.children[part]
	if !ok {
		return nil
	}
	return child
}
