# Payment Platform Agent

Initial agent runtime for the payment platform. It launches the independent MCP server, discovers its tools, and executes only approved operations through MCP.

## Current capabilities

- Connect to `payment-platform-mcp` over stdio
- Discover MCP tools
- Read a payment with `get <payment-id>`
- Request a payment with `pay <amount> <currency> <reference>`
- Refund a captured payment with `refund <payment-id> <amount>`
- Require explicit `yes` confirmation before payment execution
- Never access PostgreSQL, Stripe, or Adyen directly

This first milestone uses a deterministic CLI command interpreter. A later milestone can replace the interpreter with an LLM planner while keeping the MCP tool boundary and approval gate unchanged.

## Codebase

```text
payment-platform-agent/
├── cmd/agent/main.go             # CLI loop, commands, and confirmations
├── internal/agent/agent.go       # MCP connection, calls, and formatting
└── internal/agent/agent_test.go  # approval guard coverage
```

The agent is deliberately independent of payment implementation details. Payment validation, idempotency, processor selection, and final state remain in the payment API behind MCP.

## Run locally

Build the MCP server as an independent binary first:

```powershell
Push-Location ..\payment-platform-mcp
go build -o payment-platform-mcp.exe ./cmd/server
Pop-Location
```

Configure the agent to launch that binary:

```powershell
$env:MCP_SERVER_COMMAND='C:\Users\16122\learn\payment-platform-mcp\payment-platform-mcp.exe'
Remove-Item Env:MCP_SERVER_ARGS -ErrorAction SilentlyContinue
go run ./cmd/agent
```

The MCP binary uses `PAYMENT_API_URL` to reach the payment API. The agent requires explicit `yes` confirmation before invoking `pay_payment`.

Available commands:

```text
get <payment-id>
pay <amount> <currency> <reference>
refund <payment-id> <amount>
quit
```

Amounts are integer minor units, so `10000` means USD 100.00. The `pay` and `refund` commands execute only after a separate confirmation response of `yes`.

## Docker

Build the image:

```powershell
docker build -t payment-platform-agent:local .
```

The image expects an MCP server command to be available in its runtime environment. For local independent containers, run the MCP server separately and add a network transport in a later milestone; stdio is used for the current development path.

The container contains only the agent binary and runs as a non-root user. Because stdio is the current transport, the MCP server must be launched in the same runtime or be available as the configured command.

## Validation

```powershell
go test ./...
go vet ./...
go build ./...
```

The unit tests cover the approval guard. For an end-to-end smoke test, start the payment API, point the MCP server at it, launch the agent, then verify `get` and an explicitly confirmed `pay` or `refund` operation.
