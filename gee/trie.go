package gee

import "strings"

type node struct {
	pattern string //待匹配路由
	part string //路由的一部分
	children []*node //子节点
	isWild bool //是否精准匹配
}

// 第一个匹配成功的节点, 用于插入
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {  // 到了有'*'和':'的地方,就直接返回节点,由此节点向下
			return child
		}
	}
	return nil
}

// 所有匹配成功的节点, 用于查找
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0) // 可能多个中间节点都能匹配  通配符也能匹配上
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

func (n *node) insert(pattern string, parts []string, height int) {
	if len(parts) == height { // 到了最后一个part,将这个路由pattern绑定到这个节点,返回
		n.pattern = pattern
		return
	}
	part := parts[height]
	child := n.matchChild(part) // trie中是否已经存在这个part
	if child == nil {  // 不存在,则生成该part的节点
		child = &node{part: part, isWild: part[0] == ':' || part[0] == '*'}  // isWild生成后,这一层不会再有其他几点 见15行
		n.children = append(n.children, child)
	}
	//从该节点继续向下匹配
	child.insert(pattern, parts, height+1)
}

func (n *node) search(parts []string, height int) *node {
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			return nil
		}
		return n
	}
	part := parts[height]
	children := n.matchChildren(part) // 虽然可以匹配多个,但只能返回1个
	for _,child := range children {
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}
	return nil
}