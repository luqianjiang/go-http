package trie

import (
	"errors"
	"strings"
)

type Router struct {
	// 根节点
	root map[string]*node
}

// AddRouter 最开始的数据肯定是类似于这样的字符串
// /user/login  含义就是保存的data 例如 hello，user这个节点是没有数据的，login才有数据hello，下面一样的道理
// /user/register world
// 就是将/user/login 等字符串进行分割，分块保存到前缀树上
func (r *Router) AddRouter(pattern, data string) {
	// 初始化逻辑
	if r.root == nil {
		r.root = make(map[string]*node)
	}
	root, ok := r.root["/"]
	// 创建跟路由
	if !ok {
		root = &node{
			part: "/",
		}
		r.root["/"] = root
	}
	// /user/login => ["user","login"]
	parts := strings.Split(strings.Trim(pattern, "/"), "/")
	for _, part := range parts {
		if part == "" { // 传的url中有//
			panic("pattern不符合格式")
		}
		root = root.addNode(part)
	}
	// 循环结束后，此时的root是什么？ => 叶子节点，将data放在叶子节点
	// 即统一设置data
	root.data = data
}

func (r *Router) GetRouter(pattern string) (*node, error) {
	root, ok := r.root["/"]
	//	创建根路由
	if !ok {
		return nil, errors.New("根节点不存在")
	}

	// 切割pattern
	// ["user", "login"]
	parts := strings.Split(strings.Trim(pattern, "/"), "/")
	for _, part := range parts {
		if part == "" {
			return nil, errors.New("pattern格式不对")
		}
		root = root.getNode(part) // 例如先找到张三节点，然后再for循环找
		if root == nil {
			return nil, errors.New("pattern不存在")
		}
	}
	return root, nil
}

type node struct {
	// part 当前节点的唯一标识
	part string
	// children 维护子节点数据
	// 怎么保存，或者说用什么结构保存
	// 1. map 2. slice
	children map[string]*node
	// data 当前节点需要保存的数据
	data string
}

// 这个节点有什么功能
// 1. 注册节点：新建一个Node节点
// 2. 查找节点

// addNode 注册节点
// 问题：创建节点的时候，是将data直接赋值好还是最后赋值
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

// getNode 查找节点
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
