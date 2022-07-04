package gee

import (
    "html/template"
    "net/http"
    "path"
    "strings"
)

type HandlerFunc func(ctx *Context)

type Engine struct {
    *RouterGroup  // 将 Engine 抽象成 RouterGroup
    router        *router
    groups        []*RouterGroup
    htmlTemplates *template.Template // html template, not text template 所有的模板加载进内存
    funcMap       template.FuncMap   // 所有的自定义模板渲染函数
}

func New() *Engine {
    engine := &Engine{
        router: newRouter(),
    }
    // 将 engine 指向自己，使得 addRouter() 可以像原来一样操作
    engine.RouterGroup = &RouterGroup{
        engine: engine,
    }
    engine.groups = []*RouterGroup{engine.RouterGroup}
    return engine
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // 查看该请求适用于哪些 Group，取这些 Group 的中间件
    var middlewares []HandlerFunc
    for _, group := range e.groups {
        if strings.HasPrefix(r.URL.Path, group.prefix) {
            middlewares = append(middlewares, group.middlewares...)
        }
    }

    ctx := newContext(w, r)
    ctx.handlers = middlewares
    ctx.engine = e
    e.router.handler(ctx)
}

func (e *Engine) Run(addr string) error {
    return http.ListenAndServe(addr, e)
}

func (e *Engine) SetFuncMap(funcMap template.FuncMap) {
    e.funcMap = funcMap
}

func (e *Engine) LoadHTMLGlob(pattern string) {
    // 看不懂
    e.htmlTemplates = template.Must(template.New("").Funcs(e.funcMap).ParseGlob(pattern))
}

func (g *RouterGroup) GET(pattern string, handler HandlerFunc) {
    g.addRouter("GET", pattern, handler)
}

func (g *RouterGroup) POST(pattern string, handler HandlerFunc) {
    g.addRouter("POST", pattern, handler)
}

func (g *RouterGroup) addRouter(method, pattern string, handler HandlerFunc) {
    pattern = g.prefix + pattern
    g.engine.router.addRouter(method, pattern, handler)
}

type RouterGroup struct {
    prefix      string
    middlewares []HandlerFunc // 用于支持中间件
    parent      *RouterGroup  // 可以删除该属性
    engine      *Engine       // 指向最初的 Engine
}

func (g *RouterGroup) Group(prefix string) *RouterGroup {
    engine := g.engine
    newGroup := &RouterGroup{
        prefix: g.prefix + prefix,
        parent: g,
        engine: engine,
    }
    engine.groups = append(engine.groups, newGroup)
    return newGroup
}

func (g *RouterGroup) Use(middlewares ...HandlerFunc) {
    g.middlewares = append(g.middlewares, middlewares...)
}

func (g *RouterGroup) Static(relativePath string, root string) {
    handler := g.createStaticHandler(relativePath, http.Dir(root))
    urlPattern := path.Join(relativePath, "/*filepath")
    g.GET(urlPattern, handler)
}

func (g *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
    absPath := path.Join(g.prefix, relativePath)
    // 需要了解一下
    fileServer := http.StripPrefix(absPath, http.FileServer(fs))
    return func(ctx *Context) {
        file := ctx.Param("filepath")
        if _, err := fs.Open(file); err != nil {
            ctx.Status(http.StatusNotFound)
            return
        }

        fileServer.ServeHTTP(ctx.Writer, ctx.Req)
    }
}
