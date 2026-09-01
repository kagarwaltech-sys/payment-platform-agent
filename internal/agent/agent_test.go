package agent

import (
	"context"
	"testing"
)

func TestRefundPaymentRequiresConfirmation(t *testing.T) {
	var runner Agent
	if _, err := runner.RefundPayment(context.Background(), "pay-123", 100, "", false); err == nil {
		t.Fatal("expected confirmation error")
	}
}
