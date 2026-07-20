package domain

import (
	"testing"
)

// ── Permission Tests ─────────────────────────────────────────────────────────

func TestPermission_Can_ExactMatch(t *testing.T) {
	p := Permission{
		Resource: "warehouse",
		Actions:  []string{"read", "create"},
	}

	if !p.Can("warehouse", "read") {
		t.Error("should allow read on warehouse")
	}
	if !p.Can("warehouse", "create") {
		t.Error("should allow create on warehouse")
	}
	if p.Can("warehouse", "delete") {
		t.Error("should deny delete on warehouse (not in action list)")
	}
	if p.Can("inventory", "read") {
		t.Error("should deny read on inventory (wrong resource)")
	}
}

func TestPermission_Can_WildcardResource(t *testing.T) {
	p := Permission{
		Resource: "*",
		Actions:  []string{"read"},
	}

	if !p.Can("warehouse", "read") {
		t.Error("wildcard resource should allow read on any resource")
	}
	if !p.Can("inventory", "read") {
		t.Error("wildcard resource should allow read on any resource")
	}
	if p.Can("warehouse", "write") {
		t.Error("wildcard resource should only allow listed actions")
	}
}

func TestPermission_Can_WildcardAction(t *testing.T) {
	p := Permission{
		Resource: "order",
		Actions:  []string{"*"},
	}

	if !p.Can("order", "read") {
		t.Error("wildcard action should allow any action on order")
	}
	if !p.Can("order", "create") {
		t.Error("wildcard action should allow any action on order")
	}
	if !p.Can("order", "delete") {
		t.Error("wildcard action should allow any action on order")
	}
	if p.Can("warehouse", "read") {
		t.Error("wildcard action on order should not extend to other resources")
	}
}

func TestPermission_Can_FullWildcard(t *testing.T) {
	p := Permission{
		Resource: "*",
		Actions:  []string{"*"},
	}

	if !p.Can("anything", "anything-else") {
		t.Error("full wildcard should allow any action on any resource")
	}
	if !p.Can("warehouse", "delete") {
		t.Error("full wildcard should allow any action on any resource")
	}
}

func TestPermission_Can_EmptyActions(t *testing.T) {
	p := Permission{
		Resource: "warehouse",
		Actions:  []string{},
	}

	if p.Can("warehouse", "read") {
		t.Error("empty actions should deny everything")
	}
	if p.Can("warehouse", "any") {
		t.Error("empty actions should deny everything")
	}
}

func TestPermission_Can_EmptyResource(t *testing.T) {
	p := Permission{
		Resource: "",
		Actions:  []string{"read"},
	}

	// Empty resource matches empty resource (exact match), action matches.
	if !p.Can("", "read") {
		t.Error("empty resource should match empty resource with matching action")
	}
	// Empty resource != wildcard — should NOT match a non-empty resource.
	if p.Can("warehouse", "read") {
		t.Error("empty resource should not match as wildcard")
	}
}

func TestPermission_Can_MultipleActions(t *testing.T) {
	p := Permission{
		Resource: "task",
		Actions:  []string{"read", "create", "update", "delete"},
	}

	for _, action := range []string{"read", "create", "update", "delete"} {
		if !p.Can("task", action) {
			t.Errorf("should allow %s on task", action)
		}
	}
}

// ── User Tests ───────────────────────────────────────────────────────────────

func TestUserStatusValues(t *testing.T) {
	all := []UserStatus{
		UserStatusActive, UserStatusInactive, UserStatusLocked,
	}

	for _, s := range all {
		if s == "" {
			t.Error("user status should not be empty")
		}
	}
}

// ── Role Tests ───────────────────────────────────────────────────────────────

func TestRole_HasPermissions(t *testing.T) {
	r := Role{
		Name: "admin",
		Permissions: []Permission{
			{Resource: "*", Actions: []string{"*"}},
		},
	}

	if len(r.Permissions) != 1 {
		t.Errorf("expected 1 permission, got %d", len(r.Permissions))
	}
	if !r.Permissions[0].Can("anything", "anything") {
		t.Error("admin role should have full wildcard permission")
	}
}

func TestRole_OperatorPermissions(t *testing.T) {
	r := Role{
		Name: "operator",
		Permissions: []Permission{
			{Resource: "task", Actions: []string{"read", "create", "update"}},
			{Resource: "inventory", Actions: []string{"read"}},
		},
	}

	if !r.Permissions[0].Can("task", "read") {
		t.Error("operator should read tasks")
	}
	if !r.Permissions[0].Can("task", "create") {
		t.Error("operator should create tasks")
	}
	if r.Permissions[0].Can("task", "delete") {
		t.Error("operator should NOT delete tasks")
	}
	if r.Permissions[0].Can("inventory", "read") {
		t.Error("task permission should not apply to inventory")
	}
	if !r.Permissions[1].Can("inventory", "read") {
		t.Error("inventory permission should allow read")
	}
}

// ── AuditLog Tests ───────────────────────────────────────────────────────────

func TestAuditLog_Struct(t *testing.T) {
	log := AuditLog{
		Action:     "order.create",
		Resource:   "order",
		ResourceID: "order-123",
		Username:   "admin",
	}

	if log.Action != "order.create" {
		t.Errorf("Action = %s, want order.create", log.Action)
	}
	if log.Resource != "order" {
		t.Errorf("Resource = %s, want order", log.Resource)
	}
}
