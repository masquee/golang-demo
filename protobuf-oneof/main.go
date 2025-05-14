package main

import (
	"fmt"
	"log"

	"google.golang.org/protobuf/proto"
	"protobuf-oneof/pb"
)

func main() {
	msg1 := &pb.MyMessage{
		Data: &pb.MyMessage_Name{Name: "Alice"},
	}

	data, err := proto.Marshal(msg1)
	if err != nil {
		log.Fatalf("序列化消息失败: %v", err)
	}
	fmt.Println("--------")

	var msg2 pb.MyMessage
	err = proto.Unmarshal(data, &msg2)
	if err != nil {
		log.Fatalf("反序列化消息失败: %v", err)
	}

	switch v := msg2.GetData().(type) {
	case *pb.MyMessage_Name:
		log.Printf("Name: %s", v.Name)
	case *pb.MyMessage_Age:
		log.Printf("Age: %d", v.Age)
	case *pb.MyMessage_IsStudent:
		log.Printf("IsStudent: %v", v.IsStudent)
	default:
		log.Printf("未知类型")
	}
}
