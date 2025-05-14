package main

import (
	"encoding/json"
	"fmt"
	"log"

	clusterpb "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes"
)

func main() {
	// 原始消息
	cluster := &clusterpb.Cluster{Name: "v1-cluster"}

	// 封装为 Any
	content, err := ptypes.MarshalAny(cluster)
	if err != nil {
		log.Fatal(err)
	}

	// 序列化成 JSON
	marshaler := &jsonpb.Marshaler{Indent: "  "}
	jsonStr, err := marshaler.MarshalToString(content)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== v1 JSON ===")
	fmt.Println(jsonStr)

	// 验证 JSON 可被标准库解析（可选）
	var generic map[string]any
	_ = json.Unmarshal([]byte(jsonStr), &generic)
}
