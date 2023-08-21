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
	part string
	// children 其实就是静态路由
	children map[string]*node
	// handleFunc 这里存的是当前节点上的视图函数
	// 就是咱们之前讲的data
	handleFunc HandleFunc

	// paramChild 参数路由
	// 问题一：为什么是一个纯的node节点？
	// 问题二：静态路由和动态路由的优先级问题
	// 注册的路由一 /study/golang
	// 注册的路由二 /study/:course
	// 请求的地址 /study/golang 匹配哪一个路由？ => 匹配第一个
	// 请求的地址 /study/golang1 匹配哪一个路由？ => 匹配第二个
	// 结论：静态路由优先级 > 动态路由
	paramChild *node
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
func (r *router) getRouter(method string, pattern string) (*node, map[string]string, bool) {
	// 问题：为什么是一个kv都是string类型的
	params := make(map[string]string)

	if pattern == "" {
		return nil, params, false
	}

	// 获取根节点
	root, ok := r.trees[method]
	if !ok {
		return nil, params, false
	}

	// TODO 跟路由怎么办
	if pattern == "/" {
		return root, params, true
	}
	// 切割pattern
	parts := strings.Split(strings.Trim(pattern, "/"), "/")
	for _, part := range parts {
		if part == "" {
			return nil, params, false
		}
		root = root.getNode(part)
		if root == nil {
			return nil, params, false
		}
		// 想一想：我们注册的路由是 /study/:course
		// 					    /study/golang
		// {"course": "golang"}
		// 节点找到了
		// 1. 是静态路由 pass（不处理）
		// 2. 是动态路由中的参数路由-特殊处理：把参数维护住
		if strings.HasPrefix(root.part, ":") {
			params[root.part[1:]] = part
		}
	}
	return root, params, root.handleFunc != nil
}

// addNode 这个方法是在服务启动前调用 n.paramChild == nil保证只能注册一个
func (n *node) addNode(part string) *node {
	if strings.HasPrefix(part, ":") && n.paramChild == nil {
		n.paramChild = &node{
			part: part,
		}
		return n.paramChild
	}
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

// 一个路由的同一个位置，不能同时有静态路由和动态路由
func (n *node) getNode(part string) *node {

	// n 的 children属性都不存在
	if n.children == nil {
		if n.paramChild != nil {
			return n.paramChild
		}
		return nil
	}
	// 正常思路
	child, ok := n.children[part]
	if !ok {
		// 到这里了，说明没有匹配到静态路由
		if n.paramChild != nil {
			return n.paramChild
		}
		return nil
	}
	return child
}

/**
路由分为动态的和静态的
- 静态路由
	/study/golang
	/user/login
	/register
	...

- 动态路由
	1. 参数路由
		/study/:course 这是咱们注册的路由
			匹配的时候能匹配到什么路由：
				/study/golang、/study/python:能匹配到
				/study/golang/action:匹配不到
	2. 通配符路由:贪婪匹配
		/static/*filepath 这是咱们注册的路由
			匹配的时候能匹配到什么路由：
				/static/css/stylADAHSDCJUVKJSSVDSEKJ FCDNVNe.css
				/static/js/index.js
			filepath = js/index.js

	3. 正则路由

**/
