# JSON-RPC Go Demo

This mini project shows how JSON-RPC 2.0 works by pairing a small HTTP server with a console client, both written in Go. The goal is to help you understand the protocol concepts as well as how to implement and call JSON-RPC services.

```
json-rpc-demo/
├── README.md
├── server/
│   └── main.go
└── client/
    └── main.go
```

## What is JSON-RPC?

JSON-RPC is a lightweight *remote procedure call* protocol. A client sends a JSON object describing which method it wants to invoke and with which parameters. The server responds with a JSON object that either contains the method result or an error. A few key ideas:

- It is **transport agnostic**. You can ship JSON-RPC over HTTP, WebSocket, TCP, etc. In this demo we use HTTP POST.
- Requests contain four important fields:
  - `jsonrpc`: must be the string `"2.0"` for JSON-RPC 2.0.
  - `method`: the name of the remote procedure to call.
  - `params` (optional): data passed to the method. It can be either an object or an array.
  - `id` (optional): used by the client to match responses to requests. If omitted, the request becomes a **notification** and the server must not reply.
- Responses always echo `jsonrpc` and `id`, and they include either a `result` or an `error` object.

Example request/response pair:

```json
// request from client to server
{
  "jsonrpc": "2.0",
  "method": "math.add",
  "params": {"a": 2, "b": 3},
  "id": 1
}
```

```json
// response from server back to the client
{
  "jsonrpc": "2.0",
  "result": {"sum": 5},
  "id": 1
}
```

If something goes wrong the server sends an error instead of a result:

```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32601,
    "message": "method not found"
  },
  "id": 3
}
```

Common error codes:

- `-32600`: invalid request (missing fields, wrong structure).
- `-32601`: method not found.
- `-32602`: invalid params.
- `-32603`: internal error.
- `-32700`: parse error (malformed JSON).

## Running the demo

From the repository root:

```bash
# terminal window 1 – start the server
go run ./server
```

The server listens on `http://localhost:8080/rpc` and registers three example methods:

- `math.add`: expects an object with `a` and `b`, returns their sum.
- `math.sum`: expects an array of numbers, returns their sum.
- `text.concat`: expects an object with `parts` (array of strings) and optional `separator`.

In another terminal:

```bash
# terminal window 2 – run the sample client
go run ./client
```

The client sends four requests so you can see different behaviors:

1. Calling `math.add` using named parameters.
2. Calling `math.sum` using positional parameters (an array).
3. Sending a **notification** to `text.concat` (no `id`, so the server returns `204 No Content`).
4. Requesting an unknown method to trigger a JSON-RPC error response.

Sample output from the client (abbreviated):

```
==> Add two numbers (object params)
Request:
{
  "jsonrpc": "2.0",
  "method": "math.add",
  "params": {
    "a": 2,
    "b": 3
  },
  "id": 1
}
Response status: 200 200 OK
Response body:
{
  "jsonrpc": "2.0",
  "result": {
    "sum": 5
  },
  "id": 1
}
```

Try editing `client/main.go` to send your own methods or parameters. You can also use `curl` or `httpie` to experiment manually:

```bash
curl -X POST http://localhost:8080/rpc \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","method":"text.concat","params":{"parts":["Go","JSON-RPC"],"separator":" + "},"id":"demo"}'
```

## Understanding the server code

`server/main.go` keeps a registry of method handlers. Each handler receives the raw JSON params and returns either a result (any JSON-serializable value) or a JSON-RPC error. The server code demonstrates how to:

- validate `jsonrpc` and `method` fields
- distinguish between calls and notifications (`id` present vs missing)
- decode object and array-style params
- send errors that follow the JSON-RPC 2.0 specification
- support basic batch requests (an array of request objects)

Request logging is implemented with a simple middleware so that newcomers can see when requests arrive and how the server responds.

## Understanding the client code

`client/main.go` constructs JSON-RPC request objects, prints them, sends them via `http.Client`, and then pretty-prints the response. It also shows how to detect notifications (no `id`), and how to surface errors returned by the server.

## Next steps

- Add your own method to the server and call it from the client.
- Extend the client to accept method names and params from command-line flags or standard input.
- Switch the transport from HTTP to raw TCP or WebSocket to see how transport-agnostic JSON-RPC really is.

Learning by modifying the code and re-running the client is the fastest way to get comfortable with the protocol.
