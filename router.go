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

	// 通配符参数
	starChild *node
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
		root, ok = root.addNode(part)
		if !ok {
			panic(fmt.Sprintf("web: 路由冲突 - %s", pattern))
		}
	}

	if root.handleFunc != nil {
		panic(fmt.Sprintf("web: 路由冲突 - %s", pattern))
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

		// /study/:course/action
		// /study/*filepath

		// 参数路由和通配符路由是特殊的静态路由
		// 既然是特殊的路由，那咱们就得特殊处理
		// 参数路由和通配符路由还有点区别，就是通配符路由是贪婪匹配的
		if strings.HasPrefix(root.part, "*") {
			// /assets/*filepath
			// /assets/css/index.css
			index := strings.Index(pattern, part)
			params[root.part[1:]] = pattern[index:] // => css/index.css
			// 直接return就表示后面的不在匹配节点了
			return root, params, root.handleFunc != nil
		}
	}
	return root, params, root.handleFunc != nil
}

// addNode 这个方法是在服务启动前调用 n.paramChild == nil保证只能注册一个
func (n *node) addNode(part string) (*node, bool) {

	if strings.HasPrefix(part, "*") {
		// 这里是通配符路由
		if n.paramChild != nil {
			// 当前节点的参数路由是有值的，直接判定是：路由冲突
			return nil, false
		}
		n.starChild = &node{
			part: part,
		}
		return n.starChild, n.paramChild == nil
	}

	if strings.HasPrefix(part, ":") {
		// 这里是参数路由
		if n.starChild != nil {
			// 当前节点的通配符路由上是有值的，直接判定是冲突路由
			// study/*filepath
			// study/:course
			return nil, false
		}

		if n.paramChild == nil {
			// 创建参数路由
			n.paramChild = &node{part: part}
		}
		if n.paramChild.part != part {
			// /study/:course
			// /study/:action
			// 冲突路由，直接返回false
			//解释：第一次进来，n.paramChild.part 肯定等于part，第二次进来，n.paramChild.part 为course，part为action，冲突
			return nil, false
		}
		// /study/:course
		// /study/:course/action
		// 上面条件不成立，走到下一步去创建静态路由

		return n.paramChild, n.starChild == nil

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
	return child, true
}

// 一个路由的同一个位置，不能同时有静态路由和动态路由
func (n *node) getNode(part string) *node {

	// n 的 children属性都不存在
	if n.children == nil {
		if n.paramChild != nil {
			return n.paramChild
		}
		if n.starChild != nil {
			return n.starChild
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
		if n.starChild != nil {
			return n.starChild
		}
		return nil
	}
	return child
}

// 我们目前的添加节点的逻辑存在些问题
// 就是说，我们的添加节点的逻辑处理路由冲突的情况

// 路由冲突有哪些情况
/*
/study/login
/study/login
这是一个冲突的路由

/study/:course
/study/:action
/study/golang进来，到底是匹配那个呢？
这也是一个冲突的路由

同一个位置，参数路由和通配符路由不能同时存在
/study/*filepath
/study/:course
/study/golang进来，到底是匹配那个呢？
这也是一个冲突的路由
*/

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

/**
1. 静态路由
2. 动态
	- 参数路由
	- 通配符路由

问题一：这三个路由之间的优先级？
结论：静态路由>参数路由>通配符路由
1. /study/golang
2. /study/:course
3. /study/*action

/study/:course 其实完全可以当成是一个静态路由
/study/*action 也可以认为是一个静态路由
只不过上述两个路由是我们人为地设置为动态路由

问题二：参数路由和通配符路由的优先级？
完全取决于设计者
咱们的设计是：参数路由的优先级高于通配符的优先级

**/

// 添加节点操作
/*
1. 如果一条路由能够成功添加成一个通配符路由，是不是就意味着它也能添加到参数路由。保证没错的话就是
如果能够添加成是参数路由，那一定能够添加成是一个静态路由。所以咱们先从小范围判断
*/

/*
匹配路由
1. /study/login
2. /study/:course

现在进来/study/login路由，是匹配1号还是2号。肯定是1号
现在进来/study/register路由，是匹配1号还是2号。肯定是2号

同理
1. /study/login
2. /study/*filepath

抛出结论：就是说，优先判断是否是静态路由，在判断是否是参数路由。最后判断是否是通配符路由
*/
