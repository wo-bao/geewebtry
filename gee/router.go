package gee

import (
	"net/http"
	"strings"
)

type router struct {
	roots map[string]*node
	handlers map[string]HandlerFunc
}

func newRouter() *router {
	return &router{
		roots: make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")
	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}

func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	parts := parsePattern(pattern)
	key := method + "-" + pattern
	_, ok := r.roots[method]
	if !ok {
		r.roots[method] = &node{}
	}
	r.roots[method].insert(pattern,parts,0)  // 注册路由
	r.handlers[key] = handler  // 路由与处理函数匹配
}

func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	searchParts := parsePattern(path) // 将传入路由解析为parts[part]
	params := make(map[string]string)
	root, ok := r.roots[method]
	if !ok {
		return nil, nil
	}
	n := root.search(searchParts,0) // 到trie中进行匹配,找到绑定的节点
	if n != nil {
		parts := parsePattern(n.pattern)  // 可能是模糊匹配的,所以此处parts可能与searchParts不一样
		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[index]
			}
			if part[0] == '*' {
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n, params
	}
	return nil, nil
}

func (r *router) handle(c *Context) {
	n, params := r.getRoute(c.Method, c.Path) // 返回的trie中找到的对应的节点
	if n != nil {
		c.Params = params
		key := c.Method + "-" + n.pattern // 节点可能是模糊匹配上的,n.pattern可能与c.path不一样,实现了动态路由
		c.handlers = append(c.handlers, r.handlers[key]) // 将执行请求的func添加到handlers
	} else {
		c.handlers = append(c.handlers, func(c *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
		})
	}
	c.Next() // 这里开始执行handlers
}
