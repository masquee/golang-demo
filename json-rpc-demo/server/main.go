package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

const (
	jsonRPCVersion    = "2.0"
	defaultServerAddr = ":8080"
)

type rpcRequest struct {
	JSONRPC string           `json:"jsonrpc"`
	Method  string           `json:"method"`
	Params  json.RawMessage  `json:"params,omitempty"`
	ID      *json.RawMessage `json:"id,omitempty"`
}

type rpcResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	Result  any       `json:"result,omitempty"`
	Error   *rpcError `json:"error,omitempty"`
	ID      any       `json:"id"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type methodFunc func(ctx context.Context, params json.RawMessage) (any, *rpcError)

// methodRegistry maps method names to their Go handlers.
var methodRegistry = map[string]methodFunc{}

func init() {
	methodRegistry["math.add"] = addNumbers
	methodRegistry["math.sum"] = sumSlice
	methodRegistry["text.concat"] = concatText

	methodNames := make([]string, 0, len(methodRegistry))
	for name := range methodRegistry {
		methodNames = append(methodNames, name)
	}
	sort.Strings(methodNames)
	log.Printf("Registered JSON-RPC methods: %s", strings.Join(methodNames, ", "))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/rpc", rpcHandler)

	server := &http.Server{
		Addr:              defaultServerAddr,
		Handler:           loggingMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("JSON-RPC server listening on http://localhost%s/rpc", defaultServerAddr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}

func rpcHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "JSON-RPC endpoint only accepts POST requests", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, rpcError{Code: -32700, Message: "failed to read request"}, nil)
		return
	}

	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		writeError(w, http.StatusBadRequest, rpcError{Code: -32700, Message: "empty request body"}, nil)
		return
	}

	if strings.HasPrefix(trimmed, "[") {
		handleBatch(w, r.Context(), body)
		return
	}

	var req rpcRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, rpcError{Code: -32700, Message: "invalid JSON"}, nil)
		return
	}

	resp := dispatchRequest(r.Context(), req)
	if resp == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func handleBatch(w http.ResponseWriter, ctx context.Context, body []byte) {
	var requests []rpcRequest
	if err := json.Unmarshal(body, &requests); err != nil {
		writeError(w, http.StatusBadRequest, rpcError{Code: -32700, Message: "invalid JSON batch"}, nil)
		return
	}

	responses := make([]rpcResponse, 0, len(requests))
	for _, req := range requests {
		if resp := dispatchRequest(ctx, req); resp != nil {
			responses = append(responses, *resp)
		}
	}

	if len(responses) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	writeJSON(w, http.StatusOK, responses)
}

func dispatchRequest(ctx context.Context, req rpcRequest) *rpcResponse {
	idValue := decodeID(req.ID)

	if req.JSONRPC != jsonRPCVersion {
		return &rpcResponse{JSONRPC: jsonRPCVersion, Error: &rpcError{Code: -32600, Message: "jsonrpc field must be \"2.0\""}, ID: idValue}
	}

	if req.Method == "" {
		return &rpcResponse{JSONRPC: jsonRPCVersion, Error: &rpcError{Code: -32600, Message: "method is required"}, ID: idValue}
	}

	handler, ok := methodRegistry[req.Method]
	if !ok {
		return &rpcResponse{JSONRPC: jsonRPCVersion, Error: &rpcError{Code: -32601, Message: "method not found"}, ID: idValue}
	}

	result, rpcErr := handler(ctx, req.Params)
	if rpcErr != nil {
		return &rpcResponse{JSONRPC: jsonRPCVersion, Error: rpcErr, ID: idValue}
	}

	if req.ID == nil {
		return nil
	}
	return &rpcResponse{JSONRPC: jsonRPCVersion, Result: result, ID: idValue}
}

func addNumbers(_ context.Context, params json.RawMessage) (any, *rpcError) {
	var numbers struct {
		A float64 `json:"a"`
		B float64 `json:"b"`
	}
	if err := json.Unmarshal(params, &numbers); err != nil {
		return nil, &rpcError{Code: -32602, Message: "expected params object with 'a' and 'b'"}
	}
	return map[string]float64{"sum": numbers.A + numbers.B}, nil
}

func sumSlice(_ context.Context, params json.RawMessage) (any, *rpcError) {
	var values []float64
	if err := json.Unmarshal(params, &values); err != nil {
		return nil, &rpcError{Code: -32602, Message: "expected params array of numbers"}
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return map[string]float64{"sum": sum}, nil
}

func concatText(_ context.Context, params json.RawMessage) (any, *rpcError) {
	var args struct {
		Parts     []string `json:"parts"`
		Separator string   `json:"separator"`
	}
	if err := json.Unmarshal(params, &args); err != nil {
		return nil, &rpcError{Code: -32602, Message: "expected params object with 'parts' and optional 'separator'"}
	}

	if len(args.Parts) == 0 {
		return nil, &rpcError{Code: -32602, Message: "parts must contain at least one string"}
	}

	if args.Separator == "" {
		args.Separator = " "
	}

	return map[string]string{"text": strings.Join(args.Parts, args.Separator)}, nil
}

func decodeID(raw *json.RawMessage) any {
	if raw == nil {
		return nil
	}
	var v any
	if err := json.Unmarshal(*raw, &v); err != nil {
		return nil
	}
	switch val := v.(type) {
	case float64:
		if val == float64(int64(val)) {
			return int64(val)
		}
	}
	return v
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, err rpcError, id any) {
	writeJSON(w, status, rpcResponse{JSONRPC: jsonRPCVersion, Error: &err, ID: id})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)
		duration := time.Since(started)
		log.Printf("%s %s -> %d (%s)", r.Method, r.URL.Path, recorder.status, duration.Truncate(time.Millisecond))
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func init() {
	methodNames := make([]string, 0, len(methodRegistry))
	for name := range methodRegistry {
		methodNames = append(methodNames, name)
	}
	sort.Strings(methodNames)
	log.Printf("Registered JSON-RPC methods: %s", strings.Join(methodNames, ", "))
}
