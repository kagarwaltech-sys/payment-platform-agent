package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Agent struct {
	session *mcp.ClientSession
}

func Connect(ctx context.Context, command string, args ...string) (*Agent, error) {
	client := mcp.NewClient(&mcp.Implementation{Name: "payment-platform-agent", Version: "0.1.0"}, nil)
	transport := &mcp.CommandTransport{Command: exec.CommandContext(ctx, command, args...)}
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, err
	}
	return &Agent{session: session}, nil
}

func (a *Agent) Close() error { return a.session.Close() }

func (a *Agent) Tools(ctx context.Context) ([]*mcp.Tool, error) {
	result, err := a.session.ListTools(ctx, nil)
	if err != nil {
		return nil, err
	}
	return result.Tools, nil
}

func (a *Agent) GetPayment(ctx context.Context, paymentID string) (json.RawMessage, error) {
	return a.call(ctx, "get_payment", map[string]any{"payment_id": paymentID})
}

func (a *Agent) PayPayment(ctx context.Context, amount int64, currency, reference string, confirmed bool) (json.RawMessage, error) {
	if !confirmed {
		return nil, fmt.Errorf("payment requires explicit confirmation")
	}
	return a.call(ctx, "pay_payment", map[string]any{
		"amount": amount, "currency": currency, "capture_method": "manual", "reference": reference,
	})
}

func (a *Agent) RefundPayment(ctx context.Context, paymentID string, amount int64, idempotencyKey string, confirmed bool) (json.RawMessage, error) {
	if !confirmed {
		return nil, fmt.Errorf("refund requires explicit confirmation")
	}
	return a.call(ctx, "refund_payment", map[string]any{
		"payment_id": paymentID, "amount": amount, "idempotency_key": idempotencyKey,
	})
}

func (a *Agent) call(ctx context.Context, name string, arguments map[string]any) (json.RawMessage, error) {
	result, err := a.session.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: arguments})
	if err != nil {
		return nil, err
	}
	if result.IsError {
		return nil, fmt.Errorf("MCP tool %s returned an error", name)
	}
	if result.StructuredContent != nil {
		return json.Marshal(result.StructuredContent)
	}
	for _, content := range result.Content {
		if text, ok := content.(*mcp.TextContent); ok {
			return []byte(text.Text), nil
		}
	}
	return nil, fmt.Errorf("MCP tool %s returned no content", name)
}

func FormatPayment(data json.RawMessage) string {
	var value map[string]any
	if json.Unmarshal(data, &value) != nil {
		return strings.TrimSpace(string(data))
	}
	encoded, _ := json.MarshalIndent(value, "", "  ")
	return string(encoded)
}
