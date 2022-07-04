package gee

import (
    "net/http"
    "strings"
)

type router struct {
    roots    map[string]*node // 每种 HTTP 操作有一棵前缀树
    handlers map[string]HandlerFunc
}

func newRouter() *router {
    return &router{
        roots:    make(map[string]*node),
        handlers: make(map[string]HandlerFunc),
    }
}

func (r *router) addRouter(method, pattern string, handler HandlerFunc) {
    parts := parsePattern(pattern)

    if _, ok := r.roots[method]; !ok {
        r.roots[method] = &node{}
    }
    r.roots[method].insert(pattern, parts, 0)

    key := method + "-" + pattern
    r.handlers[key] = handler
}

// 返回匹配成功的前缀树节点和动态路由对应的相关参数
func (r *router) getRoute(method, path string) (*node, map[string]string) {
    searchParts := parsePattern(path)
    root, ok := r.roots[method]
    if !ok {
        return nil, nil
    }

    params := make(map[string]string)
    n := root.search(searchParts, 0)
    // 比对请求的路径与匹配成功的路径的对应关系
    if n != nil {
        parts := parsePattern(n.pattern)
        for idx, part := range parts {
            if part[0] == ':' {
                params[part[1:]] = searchParts[idx]
            }
            if part[0] == '*' && len(part) > 1 {
                params[part[1:]] = strings.Join(searchParts[idx:], "/")
                break
            }
        }
        return n, params
    }
    return nil, nil
}

func (r *router) handler(c *Context) {
    n, params := r.getRoute(c.Method, c.Path)
    if n != nil {
        c.Params = params
        //key := c.Method + "-" + c.Path
        key := c.Method + "-" + n.pattern
        c.handlers = append(c.handlers, r.handlers[key])
    } else {
        c.handlers = append(c.handlers, func(ctx *Context) {
            ctx.String(http.StatusNotFound, "404 NOTFOUND: %s\n", c.Path)
        })
    }
    c.Next()
}

// Only one * is allowed
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
