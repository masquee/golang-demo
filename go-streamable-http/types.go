package main

// StreamData 表示流式传输的数据结构
type StreamData struct {
	Timestamp int64  `json:"timestamp"`
	Message   string `json:"message"`
	Count     int    `json:"count"`
}
