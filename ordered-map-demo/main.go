package main

import (
	"encoding/json"
	"fmt"
	"github.com/iancoleman/orderedmap"
)

func main() {
	jsonStr := `{"name": "张三", "age": 25, "city": "北京", "email": "zhangsan@example.com"}`
	data := orderedmap.New()
	err := json.Unmarshal([]byte(jsonStr), data)
	if err != nil {
		fmt.Println("JSON 解析错误:", err)
		return
	}

	for _, key := range data.Keys() {
		value, _ := data.Get(key)
		fmt.Printf("%s: %v\n", key, value)
	}
}
