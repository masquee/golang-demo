package main

import (
	"fmt"
	"log"

	clusterpb "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
)

func main() {
	cluster := &clusterpb.Cluster{Name: "v2-cluster"}

	// 包装成 Any（会带 type.googleapis.com/... 的 TypeUrl）
	clusterAny, err := anypb.New(cluster)
	if err != nil {
		log.Fatal(err)
	}

	// 试图序列化
	jsonBytes, err := protojson.MarshalOptions{
		Multiline: true,
	}.Marshal(clusterAny)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== v2 JSON ===")
	fmt.Println(string(jsonBytes))
}
