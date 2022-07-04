package gee

import (
    "fmt"
    "log"
    "net/http"
    "runtime"
    "strings"
)

func Recovery() HandlerFunc {
    return func(ctx *Context) {
        defer func() {
            if err := recover(); err != nil {
                msg := fmt.Sprintf("%s", err)
                log.Printf("%s\n\n", trace(msg))
                ctx.Fail(http.StatusInternalServerError, "Internal Server Error")
            }
        }()
        ctx.Next()
    }
}

// print stack trace
func trace(msg string) string {
    var pcs [32]uintptr
    /* 为日志简洁跳过前三个调用者
       Callers 用来返回调用栈的程序计数器,
       第 0 个 Caller 是 Callers 本身，
       第 1 个是上一层 trace，
       第 2 个是再上一层的 defer func
    */
    n := runtime.Callers(3, pcs[:])

    var str strings.Builder
    str.WriteString(msg + "\nTraceback:")
    for _, pc := range pcs[:n] {
        fn := runtime.FuncForPC(pc)
        file, line := fn.FileLine(pc)
        str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
    }
    return str.String()
}
