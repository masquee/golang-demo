package main

import (
	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm"
	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm/types"
	"github.com/higress-group/wasm-go/pkg/wrapper"
	"github.com/tidwall/gjson"
)

func main() {}

func init() {
	wrapper.SetCtx(
		// 插件名称
		"my-plugin",
		// 为解析插件配置，设置自定义函数
		wrapper.ParseConfig(parseConfig),
		// 为处理请求头，设置自定义函数
		wrapper.ProcessRequestHeaders(onHttpRequestHeaders),
	)
}

// 自定义插件配置
type MyConfig struct {
	mockEnable bool
}

// 在控制台插件配置中填写的 yaml 配置会自动转换为 json，此处直接从 json 这个参数里解析配置即可
func parseConfig(json gjson.Result, config *MyConfig) error {
	// 解析出配置，更新到 config 中
	config.mockEnable = json.Get("mockEnable").Bool()
	return nil
}

func onHttpRequestHeaders(ctx wrapper.HttpContext, config MyConfig) types.Action {
	proxywasm.AddHttpRequestHeader("hello", "world")
	if config.mockEnable {
		proxywasm.SendHttpResponse(200, nil, []byte("hello world"), -1)
	}
	return types.HeaderContinue
}
