package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
	ID      any    `json:"id,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
	ID      any             `json:"id,omitempty"`
}

type rpcError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func main() {
	endpoint := flag.String("server", "http://localhost:8080/rpc", "JSON-RPC endpoint URL")
	flag.Parse()

	log.SetFlags(0)

	examples := []struct {
		title  string
		req    rpcRequest
		isNote bool
	}{
		{
			title: "Add two numbers (object params)",
			req: rpcRequest{
				JSONRPC: "2.0",
				Method:  "math.add",
				Params: map[string]any{
					"a": 2,
					"b": 3,
				},
				ID: 1,
			},
		},
		{
			title: "Sum a slice of numbers (positional params)",
			req: rpcRequest{
				JSONRPC: "2.0",
				Method:  "math.sum",
				Params:  []float64{1, 2, 3, 4.5},
				ID:      2,
			},
		},
		{
			title: "Send a notification (no response expected)",
			req: rpcRequest{
				JSONRPC: "2.0",
				Method:  "text.concat",
				Params: map[string]any{
					"parts":     []string{"hello", "json-rpc"},
					"separator": ", ",
				},
			},
			isNote: true,
		},
		{
			title: "Trigger an error (unknown method)",
			req: rpcRequest{
				JSONRPC: "2.0",
				Method:  "math.divide",
				Params:  map[string]any{"a": 4, "b": 0},
				ID:      3,
			},
		},
	}

	for _, example := range examples {
		fmt.Println("\n==>", example.title)
		if err := sendRequest(*endpoint, example.req, example.isNote); err != nil {
			log.Printf("request failed: %v", err)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func sendRequest(endpoint string, req rpcRequest, isNotification bool) error {
	encoded, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return fmt.Errorf("encode request: %w", err)
	}

	fmt.Println("Request:")
	fmt.Println(string(encoded))

	httpReq, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(encoded))
	if err != nil {
		return fmt.Errorf("create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if isNotification {
		fmt.Printf("Response status: %d %s (notifications omit bodies)\n", resp.StatusCode, resp.Status)
		io.Copy(io.Discard, resp.Body)
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	fmt.Printf("Response status: %d %s\n", resp.StatusCode, resp.Status)
	if len(body) == 0 {
		fmt.Println("(no response body)")
		return nil
	}

	var rpcResp rpcResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		fmt.Println("Raw response:")
		fmt.Println(string(body))
		return fmt.Errorf("decode JSON-RPC response: %w", err)
	}

	pretty, err := json.MarshalIndent(rpcResp, "", "  ")
	if err != nil {
		fmt.Println("Raw response:")
		fmt.Println(string(body))
		return fmt.Errorf("pretty-print response: %w", err)
	}

	fmt.Println("Response body:")
	fmt.Println(string(pretty))

	if rpcResp.Error != nil {
		fmt.Printf("Server returned JSON-RPC error (code %d): %s\n", rpcResp.Error.Code, rpcResp.Error.Message)
		if len(rpcResp.Error.Data) > 0 {
			fmt.Printf("Error data: %s\n", string(rpcResp.Error.Data))
		}
	}

	return nil
}

func init() {
	if os.Getenv("HTTP_PROXY") != "" || os.Getenv("http_proxy") != "" {
		log.Println("Warning: HTTP proxy environment variables detected; direct localhost calls may bypass the proxy.")
	}
}
