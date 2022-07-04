package gee

import (
    "encoding/json"
    "fmt"
    "net/http"
)

type H map[string]interface{}

type Context struct {
    Writer http.ResponseWriter
    Req    *http.Request
    // request
    Path   string
    Method string
    Params map[string]string
    // response
    StatusCode int
    // middleware
    handlers []HandlerFunc // 顺序放置 M1, M2, M3..., Our Handler
    index    int
}

func newContext(w http.ResponseWriter, r *http.Request) *Context {
    return &Context{
        Writer: w,
        Req:    r,
        Path:   r.URL.Path,
        Method: r.Method,
        index:  -1,
    }
}

func (c *Context) Param(key string) string {
    value, _ := c.Params[key]
    return value
}

func (c *Context) PostForm(key string) string {
    return c.Req.FormValue(key)
}

func (c *Context) Query(key string) string {
    return c.Req.URL.Query().Get(key)
}

func (c *Context) Status(code int) {
    c.StatusCode = code
    c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
    c.Writer.Header().Set(key, value)
}

func (c *Context) String(code int, format string, value ...interface{}) {
    c.SetHeader("Content-Type", "text/plain")
    c.Status(code)
    c.Writer.Write([]byte(fmt.Sprintf(format, value...)))
}

func (c *Context) Json(code int, obj interface{}) {
    c.SetHeader("Content-Type", "application/json")
    c.Status(code)
    encoder := json.NewEncoder(c.Writer)
    if err := encoder.Encode(obj); err != nil {
        http.Error(c.Writer, err.Error(), 500)
    }
}

func (c *Context) Data(code int, data []byte) {
    c.Status(code)
    c.Writer.Write(data)
}

func (c *Context) HTML(code int, html string) {
    c.SetHeader("Content-Type", "text/html")
    c.Status(code)
    c.Writer.Write([]byte(html))
}

func (c *Context) Next() {
    s := len(c.handlers)
    for c.index++; c.index < s; c.index++ {
        c.handlers[c.index](c)
    }
}

func (c *Context) Fail(code int, err string) {
    c.index = len(c.handlers)
    c.Json(code, H{
        "message": err,
    })
}
