package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm"
	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm/types"
	logs "github.com/higress-group/wasm-go/pkg/log"
	"github.com/higress-group/wasm-go/pkg/wrapper"
	"github.com/tidwall/gjson"
)

func main() {}

func init() {
	wrapper.SetCtx(
		"http-call",
		wrapper.ParseConfigBy(parseConfig),
		wrapper.ProcessRequestHeadersBy(onHttpRequestHeaders),
	)
}

type MyConfig struct {
	// 用于发起 HTTP 调用 client
	client wrapper.HttpClient
	// 请求 url
	requestPath string
	// 根据这个 key 取出调用服务的应答头对应字段，再设置到原始请求的请求头，key 为此配置项
	tokenHeader string
}

func parseConfig(json gjson.Result, config *MyConfig, log logs.Log) error {
	config.tokenHeader = json.Get("tokenHeader").String()
	if config.tokenHeader == "" {
		return errors.New("missing tokenHeader in config")
	}
	config.requestPath = json.Get("requestPath").String()
	if config.requestPath == "" {
		return errors.New("missing requestPath in config")
	}
	// 带服务类型的完整 FQDN 名称，例如 my-svc.dns, my-svc.static, service-provider.DEFAULT-GROUP.public.nacos, httpbin.my-ns.svc.cluster.local
	serviceName := json.Get("serviceName").String()
	servicePort := json.Get("servicePort").Int()
	if servicePort == 0 {
		if strings.HasSuffix(serviceName, ".static") {
			// 静态 IP 类型服务的逻辑端口是 80
			servicePort = 80
		}
	}
	config.client = wrapper.NewClusterClient(wrapper.FQDNCluster{
		FQDN: serviceName,
		Port: servicePort,
	})
	return nil
}

func onHttpRequestHeaders(ctx wrapper.HttpContext, config MyConfig, log logs.Log) types.Action {
	// 使用 client 的 Get 方法发起 HTTP Get 调用，此处省略了 timeout 参数，默认超时时间 500 毫秒
	err := config.client.Get(config.requestPath, nil,
		// 回调函数，将在响应异步返回时被执行
		func(statusCode int, responseHeaders http.Header, responseBody []byte) {
			// 请求没有返回 200 状态码，进行处理
			if statusCode != http.StatusOK {
				log.Errorf("http call failed, status: %d", statusCode)
				proxywasm.SendHttpResponse(http.StatusInternalServerError, nil,
					[]byte("http call failed"), -1)
				return
			}
			// 打印响应的 HTTP 状态码和应答 body
			log.Infof("get status: %d, response body: %s", statusCode, responseBody)
			// 从应答头中解析 token 字段设置到原始请求头中
			token := responseHeaders.Get(config.tokenHeader)
			if token != "" {
				proxywasm.AddHttpRequestHeader(config.tokenHeader, token)
			}
			// 恢复原始请求流程，继续往下处理，才能正常转发给后端服务
			proxywasm.ResumeHttpRequest()
		})

	if err != nil {
		// 由于调用外部服务失败，放行请求，记录日志
		log.Errorf("Error occured while calling http, it seems cannot find the service cluster.")
		return types.ActionContinue
	} else {
		// 需要等待异步回调完成，返回 HeaderStopAllIterationAndWatermark 状态，可以被 ResumeHttpRequest 恢复
		return types.HeaderStopAllIterationAndWatermark
	}
}
