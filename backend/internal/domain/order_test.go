package domain

import (
	"testing"
)

// ── Order State Machine Tests ────────────────────────────────────────────────

func TestOrder_CanTransitionTo_ValidTransitions(t *testing.T) {
	tests := []struct {
		name    string
		current OrderStatus
		target  OrderStatus
		want    bool
	}{
		// draft → confirmed, cancelled
		{name: "draft → confirmed", current: OrderStatusDraft, target: OrderStatusConfirmed, want: true},
		{name: "draft → cancelled", current: OrderStatusDraft, target: OrderStatusCancelled, want: true},
		// confirmed → processing, cancelled
		{name: "confirmed → processing", current: OrderStatusConfirmed, target: OrderStatusProcessing, want: true},
		{name: "confirmed → cancelled", current: OrderStatusConfirmed, target: OrderStatusCancelled, want: true},
		// processing → completed, partial, cancelled
		{name: "processing → completed", current: OrderStatusProcessing, target: OrderStatusCompleted, want: true},
		{name: "processing → partial", current: OrderStatusProcessing, target: OrderStatusPartial, want: true},
		{name: "processing → cancelled", current: OrderStatusProcessing, target: OrderStatusCancelled, want: true},
		// partial → completed, cancelled
		{name: "partial → completed", current: OrderStatusPartial, target: OrderStatusCompleted, want: true},
		{name: "partial → cancelled", current: OrderStatusPartial, target: OrderStatusCancelled, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{Status: tt.current}
			got := o.CanTransitionTo(tt.target)
			if got != tt.want {
				t.Errorf("CanTransitionTo(%q) = %v, want %v", tt.target, got, tt.want)
			}
		})
	}
}

func TestOrder_CanTransitionTo_InvalidTransitions(t *testing.T) {
	tests := []struct {
		name    string
		current OrderStatus
		target  OrderStatus
	}{
		// Can't skip statuses
		{name: "draft → processing", current: OrderStatusDraft, target: OrderStatusProcessing},
		{name: "draft → completed", current: OrderStatusDraft, target: OrderStatusCompleted},
		{name: "draft → partial", current: OrderStatusDraft, target: OrderStatusPartial},
		{name: "confirmed → completed", current: OrderStatusConfirmed, target: OrderStatusCompleted},
		{name: "confirmed → partial", current: OrderStatusConfirmed, target: OrderStatusPartial},
		{name: "confirmed → draft", current: OrderStatusConfirmed, target: OrderStatusDraft},
		{name: "processing → confirmed", current: OrderStatusProcessing, target: OrderStatusConfirmed},
		{name: "processing → draft", current: OrderStatusProcessing, target: OrderStatusDraft},
		{name: "partial → processing", current: OrderStatusPartial, target: OrderStatusProcessing},
		{name: "partial → draft", current: OrderStatusPartial, target: OrderStatusDraft},
		{name: "partial → confirmed", current: OrderStatusPartial, target: OrderStatusConfirmed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{Status: tt.current}
			if o.CanTransitionTo(tt.target) {
				t.Errorf("CanTransitionTo(%q) = true, want false", tt.target)
			}
		})
	}
}

func TestOrder_CanTransitionTo_NoopRejected(t *testing.T) {
	o := &Order{Status: OrderStatusDraft}
	if o.CanTransitionTo(OrderStatusDraft) {
		t.Error("same-status transition should return false")
	}
}

func TestOrder_CanTransitionTo_TerminalStates(t *testing.T) {
	// Cancelled is terminal — no transitions out.
	cancelled := &Order{Status: OrderStatusCancelled}
	for _, target := range []OrderStatus{
		OrderStatusDraft, OrderStatusConfirmed, OrderStatusProcessing,
		OrderStatusPartial, OrderStatusCompleted,
	} {
		if cancelled.CanTransitionTo(target) {
			t.Errorf("cancelled → %s: should be false", target)
		}
	}
	// Cancelled → cancelled is no-op.
	if cancelled.CanTransitionTo(OrderStatusCancelled) {
		t.Error("cancelled → cancelled (no-op) should be false")
	}

	// Completed is terminal — no transitions out.
	completed := &Order{Status: OrderStatusCompleted}
	for _, target := range []OrderStatus{
		OrderStatusDraft, OrderStatusConfirmed, OrderStatusProcessing,
		OrderStatusPartial, OrderStatusCancelled,
	} {
		if completed.CanTransitionTo(target) {
			t.Errorf("completed → %s: should be false", target)
		}
	}
	if completed.CanTransitionTo(OrderStatusCompleted) {
		t.Error("completed → completed (no-op) should be false")
	}
}

func TestOrder_IsTerminal(t *testing.T) {
	terminal := []OrderStatus{OrderStatusCancelled, OrderStatusCompleted}
	nonTerminal := []OrderStatus{
		OrderStatusDraft, OrderStatusConfirmed,
		OrderStatusProcessing, OrderStatusPartial,
	}

	for _, s := range terminal {
		o := &Order{Status: s}
		if !o.IsTerminal() {
			t.Errorf("status %q should be terminal", s)
		}
	}

	for _, s := range nonTerminal {
		o := &Order{Status: s}
		if o.IsTerminal() {
			t.Errorf("status %q should NOT be terminal", s)
		}
	}
}

func TestOrder_CancellationFromAllNonTerminal(t *testing.T) {
	// Any non-terminal status can be cancelled.
	nonTerminal := []OrderStatus{
		OrderStatusDraft, OrderStatusConfirmed,
		OrderStatusProcessing, OrderStatusPartial,
	}

	for _, s := range nonTerminal {
		o := &Order{Status: s}
		if !o.CanTransitionTo(OrderStatusCancelled) {
			t.Errorf("%s → cancelled should be valid", s)
		}
	}
}

func TestOrder_FullLifecycle(t *testing.T) {
	// Simulate full happy path: draft → confirmed → processing → completed.
	o := &Order{Status: OrderStatusDraft}

	transitions := []OrderStatus{
		OrderStatusConfirmed,
		OrderStatusProcessing,
		OrderStatusCompleted,
	}

	for _, target := range transitions {
		if !o.CanTransitionTo(target) {
			t.Fatalf("transition to %s should be valid from %s", target, o.Status)
		}
		o.Status = target
	}

	if o.Status != OrderStatusCompleted {
		t.Errorf("final status = %s, want %s", o.Status, OrderStatusCompleted)
	}
}

func TestOrder_FullLifecycleWithPartial(t *testing.T) {
	// Simulate: draft → confirmed → processing → partial → completed.
	o := &Order{Status: OrderStatusDraft}

	transitions := []OrderStatus{
		OrderStatusConfirmed,
		OrderStatusProcessing,
		OrderStatusPartial,
		OrderStatusCompleted,
	}

	for _, target := range transitions {
		if !o.CanTransitionTo(target) {
			t.Fatalf("transition to %s should be valid from %s", target, o.Status)
		}
		o.Status = target
	}

	if o.Status != OrderStatusCompleted {
		t.Errorf("final status = %s, want %s", o.Status, OrderStatusCompleted)
	}
}

// ── OrderLine State Machine Tests ────────────────────────────────────────────

func TestOrderLine_CanTransitionTo_ValidTransitions(t *testing.T) {
	tests := []struct {
		name    string
		current OrderLineStatus
		target  OrderLineStatus
		want    bool
	}{
		// pending → allocated, cancelled
		{name: "pending → allocated", current: OrderLineStatusPending, target: OrderLineStatusAllocated, want: true},
		{name: "pending → cancelled", current: OrderLineStatusPending, target: OrderLineStatusCancelled, want: true},
		// allocated → partial, fulfilled, cancelled
		{name: "allocated → partial", current: OrderLineStatusAllocated, target: OrderLineStatusPartial, want: true},
		{name: "allocated → fulfilled", current: OrderLineStatusAllocated, target: OrderLineStatusFulfilled, want: true},
		{name: "allocated → cancelled", current: OrderLineStatusAllocated, target: OrderLineStatusCancelled, want: true},
		// partial → fulfilled, cancelled
		{name: "partial → fulfilled", current: OrderLineStatusPartial, target: OrderLineStatusFulfilled, want: true},
		{name: "partial → cancelled", current: OrderLineStatusPartial, target: OrderLineStatusCancelled, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ol := &OrderLine{Status: tt.current}
			got := ol.CanTransitionTo(tt.target)
			if got != tt.want {
				t.Errorf("CanTransitionTo(%q) = %v, want %v", tt.target, got, tt.want)
			}
		})
	}
}

func TestOrderLine_CanTransitionTo_InvalidTransitions(t *testing.T) {
	tests := []struct {
		name    string
		current OrderLineStatus
		target  OrderLineStatus
	}{
		// pending → partial, fulfilled (skip allocated)
		{name: "pending → partial", current: OrderLineStatusPending, target: OrderLineStatusPartial},
		{name: "pending → fulfilled", current: OrderLineStatusPending, target: OrderLineStatusFulfilled},
		// allocated → pending (can't go backwards)
		{name: "allocated → pending", current: OrderLineStatusAllocated, target: OrderLineStatusPending},
		// partial → pending, allocated (can't go backwards)
		{name: "partial → pending", current: OrderLineStatusPartial, target: OrderLineStatusPending},
		{name: "partial → allocated", current: OrderLineStatusPartial, target: OrderLineStatusAllocated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ol := &OrderLine{Status: tt.current}
			if ol.CanTransitionTo(tt.target) {
				t.Errorf("CanTransitionTo(%q) = true, want false", tt.target)
			}
		})
	}
}

func TestOrderLine_CanTransitionTo_TerminalStates(t *testing.T) {
	fulfilled := &OrderLine{Status: OrderLineStatusFulfilled}
	for _, target := range []OrderLineStatus{
		OrderLineStatusPending, OrderLineStatusAllocated,
		OrderLineStatusPartial, OrderLineStatusCancelled,
	} {
		if fulfilled.CanTransitionTo(target) {
			t.Errorf("fulfilled → %s: should be false", target)
		}
	}

	cancelled := &OrderLine{Status: OrderLineStatusCancelled}
	for _, target := range []OrderLineStatus{
		OrderLineStatusPending, OrderLineStatusAllocated,
		OrderLineStatusPartial, OrderLineStatusFulfilled,
	} {
		if cancelled.CanTransitionTo(target) {
			t.Errorf("cancelled → %s: should be false", target)
		}
	}
}

func TestOrderLine_IsTerminal(t *testing.T) {
	terminal := []OrderLineStatus{OrderLineStatusFulfilled, OrderLineStatusCancelled}
	nonTerminal := []OrderLineStatus{
		OrderLineStatusPending, OrderLineStatusAllocated, OrderLineStatusPartial,
	}

	for _, s := range terminal {
		ol := &OrderLine{Status: s}
		if !ol.IsTerminal() {
			t.Errorf("status %q should be terminal", s)
		}
	}
	for _, s := range nonTerminal {
		ol := &OrderLine{Status: s}
		if ol.IsTerminal() {
			t.Errorf("status %q should NOT be terminal", s)
		}
	}
}

func TestOrderLine_FullLifecycle(t *testing.T) {
	ol := &OrderLine{Status: OrderLineStatusPending}

	transitions := []OrderLineStatus{
		OrderLineStatusAllocated,
		OrderLineStatusPartial,
		OrderLineStatusFulfilled,
	}

	for _, target := range transitions {
		if !ol.CanTransitionTo(target) {
			t.Fatalf("transition to %s should be valid from %s", target, ol.Status)
		}
		ol.Status = target
	}

	if ol.Status != OrderLineStatusFulfilled {
		t.Errorf("final status = %s, want %s", ol.Status, OrderLineStatusFulfilled)
	}
}

// ── ASN State Machine Tests ──────────────────────────────────────────────────

func TestASN_CanTransitionTo_ValidTransitions(t *testing.T) {
	tests := []struct {
		name    string
		current ASNStatus
		target  ASNStatus
		want    bool
	}{
		{name: "pending → arrived", current: ASNStatusPending, target: ASNStatusArrived, want: true},
		{name: "arrived → receiving", current: ASNStatusArrived, target: ASNStatusReceiving, want: true},
		{name: "receiving → partial", current: ASNStatusReceiving, target: ASNStatusPartial, want: true},
		{name: "receiving → received", current: ASNStatusReceiving, target: ASNStatusReceived, want: true},
		{name: "partial → received", current: ASNStatusPartial, target: ASNStatusReceived, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &ASN{Status: tt.current}
			got := a.CanTransitionTo(tt.target)
			if got != tt.want {
				t.Errorf("CanTransitionTo(%q) = %v, want %v", tt.target, got, tt.want)
			}
		})
	}
}

func TestASN_CanTransitionTo_InvalidTransitions(t *testing.T) {
	tests := []struct {
		name    string
		current ASNStatus
		target  ASNStatus
	}{
		{name: "pending → receiving (skip arrived)", current: ASNStatusPending, target: ASNStatusReceiving},
		{name: "pending → received (skip ahead)", current: ASNStatusPending, target: ASNStatusReceived},
		{name: "arrived → received (skip receiving)", current: ASNStatusArrived, target: ASNStatusReceived},
		{name: "arrived → pending (backwards)", current: ASNStatusArrived, target: ASNStatusPending},
		{name: "receiving → arrived (backwards)", current: ASNStatusReceiving, target: ASNStatusArrived},
		{name: "partial → receiving (backwards)", current: ASNStatusPartial, target: ASNStatusReceiving},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &ASN{Status: tt.current}
			if a.CanTransitionTo(tt.target) {
				t.Errorf("CanTransitionTo(%q) = true, want false", tt.target)
			}
		})
	}
}

func TestASN_CanTransitionTo_Terminal(t *testing.T) {
	received := &ASN{Status: ASNStatusReceived}
	for _, target := range []ASNStatus{
		ASNStatusPending, ASNStatusArrived, ASNStatusReceiving, ASNStatusPartial,
	} {
		if received.CanTransitionTo(target) {
			t.Errorf("received → %s: should be false", target)
		}
	}
}

func TestASN_IsTerminal(t *testing.T) {
	received := &ASN{Status: ASNStatusReceived}
	if !received.IsTerminal() {
		t.Error("received should be terminal")
	}

	nonTerminal := []ASNStatus{
		ASNStatusPending, ASNStatusArrived,
		ASNStatusReceiving, ASNStatusPartial,
	}
	for _, s := range nonTerminal {
		a := &ASN{Status: s}
		if a.IsTerminal() {
			t.Errorf("status %q should NOT be terminal", s)
		}
	}
}

func TestASN_FullLifecycle(t *testing.T) {
	a := &ASN{Status: ASNStatusPending}

	transitions := []ASNStatus{
		ASNStatusArrived,
		ASNStatusReceiving,
		ASNStatusReceived,
	}

	for _, target := range transitions {
		if !a.CanTransitionTo(target) {
			t.Fatalf("transition to %s should be valid from %s", target, a.Status)
		}
		a.Status = target
	}

	if a.Status != ASNStatusReceived {
		t.Errorf("final status = %s, want %s", a.Status, ASNStatusReceived)
	}
}

func TestASN_PartialFlow(t *testing.T) {
	// pending → arrived → receiving → partial → received
	a := &ASN{Status: ASNStatusPending}

	transitions := []ASNStatus{
		ASNStatusArrived,
		ASNStatusReceiving,
		ASNStatusPartial,
		ASNStatusReceived,
	}

	for _, target := range transitions {
		if !a.CanTransitionTo(target) {
			t.Fatalf("transition to %s should be valid from %s", target, a.Status)
		}
		a.Status = target
	}

	if a.Status != ASNStatusReceived {
		t.Errorf("final status = %s, want %s", a.Status, ASNStatusReceived)
	}
}

// ── Order Constants Tests ────────────────────────────────────────────────────

func TestOrderTypeValues(t *testing.T) {
	all := []OrderType{
		OrderTypeInbound, OrderTypeOutbound,
		OrderTypeTransfer, OrderTypeReturn,
	}

	for _, ot := range all {
		if ot == "" {
			t.Error("order type should not be empty")
		}
	}
}

func TestOrderStatusValues(t *testing.T) {
	all := []OrderStatus{
		OrderStatusDraft, OrderStatusConfirmed, OrderStatusProcessing,
		OrderStatusPartial, OrderStatusCompleted, OrderStatusCancelled,
	}

	for _, s := range all {
		if s == "" {
			t.Error("order status should not be empty")
		}
	}
}

func TestOrderPriorityValues(t *testing.T) {
	all := []OrderPriority{
		OrderPriorityLow, OrderPriorityNormal,
		OrderPriorityHigh, OrderPriorityUrgent,
	}

	for _, p := range all {
		if p == "" {
			t.Error("order priority should not be empty")
		}
	}
}
