package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestShipment_StateMachine(t *testing.T) {
	s := &Shipment{
		ID:     uuid.New(),
		Status: ShipmentStatusPending,
	}

	t.Run("pending → in_transit", func(t *testing.T) {
		if !s.CanTransitionTo(ShipmentStatusInTransit) {
			t.Error("pending should transition to in_transit")
		}
	})

	t.Run("pending → delivered (invalid)", func(t *testing.T) {
		if s.CanTransitionTo(ShipmentStatusDelivered) {
			t.Error("pending should NOT transition directly to delivered")
		}
	})

	t.Run("pending → cancelled", func(t *testing.T) {
		if !s.CanTransitionTo(ShipmentStatusCancelled) {
			t.Error("pending should allow cancellation")
		}
	})

	t.Run("can be shipped", func(t *testing.T) {
		if !s.CanBeShipped() {
			t.Error("pending should be shippable")
		}
	})

	t.Run("cannot be delivered in pending", func(t *testing.T) {
		if s.CanBeDelivered() {
			t.Error("pending should NOT be deliverable")
		}
	})

	// Transition to in_transit
	s.Status = ShipmentStatusInTransit

	t.Run("in_transit → delivered", func(t *testing.T) {
		if !s.CanTransitionTo(ShipmentStatusDelivered) {
			t.Error("in_transit should transition to delivered")
		}
	})

	t.Run("in_transit → pending (invalid - backward)", func(t *testing.T) {
		if s.CanTransitionTo(ShipmentStatusPending) {
			t.Error("in_transit should NOT transition backward to pending")
		}
	})

	t.Run("in_transit → cancelled", func(t *testing.T) {
		if !s.CanTransitionTo(ShipmentStatusCancelled) {
			t.Error("in_transit should allow cancellation")
		}
	})

	t.Run("can be delivered", func(t *testing.T) {
		if !s.CanBeDelivered() {
			t.Error("in_transit should be deliverable")
		}
	})

	t.Run("cannot be shipped in in_transit", func(t *testing.T) {
		if s.CanBeShipped() {
			t.Error("in_transit should NOT be shippable")
		}
	})

	// Transition to delivered (terminal)
	s.Status = ShipmentStatusDelivered

	t.Run("delivered → in_transit (invalid - terminal)", func(t *testing.T) {
		if s.CanTransitionTo(ShipmentStatusInTransit) {
			t.Error("delivered should be terminal, no further transitions")
		}
	})

	t.Run("delivered → cancelled (invalid - terminal)", func(t *testing.T) {
		if s.CanTransitionTo(ShipmentStatusCancelled) {
			t.Error("delivered should be terminal, cannot cancel")
		}
	})

	t.Run("delivered is terminal", func(t *testing.T) {
		if !s.IsTerminal() {
			t.Error("delivered should be terminal")
		}
	})

	// Test cancelled is terminal
	s2 := &Shipment{ID: uuid.New(), Status: ShipmentStatusCancelled}
	t.Run("cancelled is terminal", func(t *testing.T) {
		if !s2.IsTerminal() {
			t.Error("cancelled should be terminal")
		}
	})

	t.Run("cancelled → in_transit (invalid)", func(t *testing.T) {
		if s2.CanTransitionTo(ShipmentStatusInTransit) {
			t.Error("cancelled should be terminal, no further transitions")
		}
	})
}

func TestShipment_SameStatusTransition(t *testing.T) {
	s := &Shipment{ID: uuid.New(), Status: ShipmentStatusPending}
	if s.CanTransitionTo(ShipmentStatusPending) {
		t.Error("should not allow transition to same status")
	}
}
