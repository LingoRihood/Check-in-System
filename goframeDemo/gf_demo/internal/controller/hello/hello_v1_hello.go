package hello

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/text/gstr"

	v1 "gf_demo/api/hello/v1"
)

// JsonOutputsForLogger is for JSON marshaling in sequence.
// 自定义的“日志 JSON 模板”
// {"time":"...","level":"...","content":"..."}
type JsonOutputsForLogger struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Content string `json:"content"`
}

// LoggingJsonHandler is a example handler for logging JSON format content.
// 日志处理器（Handler）
var LoggingJsonHandler glog.Handler = func(ctx context.Context, in *glog.HandlerInput) {
	jsonForLogger := JsonOutputsForLogger{
		Time: in.TimeFormat,
		// 用 gstr.Trim(in.LevelFormat, "[]") 把 [] 去掉，变成 INFO
		Level: gstr.Trim(in.LevelFormat, "[]"),
		// 把内容前后空格去掉。
		Content: gstr.Trim(in.ValuesContent()), // 2.7以上版本用in.ValuesContent()
	}

	// 把结构体转成 JSON 字节
	jsonBytes, err := json.Marshal(jsonForLogger)

	// 如果转 JSON 失败，就把错误写到 stderr
	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error())
		return
	}

	// 把 JSON 写入日志缓冲区
	in.Buffer.Write(jsonBytes)
	in.Buffer.WriteString("\n")

	// 继续走后续处理链
	in.Next(ctx)
}

func (c *ControllerV1) Hello(ctx context.Context, req *v1.HelloReq) (res *v1.HelloRes, err error) {
	fmt.Println(g.Cfg().Get(ctx, "name"))
	fmt.Println(g.Cfg().Get(ctx, "info.version"))
	g.Log().SetHandlers(LoggingJsonHandler)
	g.Log().Infof(ctx, "Hello %s", "Simon")
	g.Log().Debug(ctx, "Debug Log")
	g.Log().Debug(ctx, g.Map{
		"uid":  100,
		"name": "john",
	})
	g.RequestFromCtx(ctx).Response.Writeln("Hello World!")
	return
}
