package domain

import (
	"time"

	"github.com/google/uuid"
)

// Task represents a warehouse operation task assigned to a worker or automated system.
type Task struct {
	ID            uuid.UUID  `json:"id"`
	TaskNo        string     `json:"task_no"`        // e.g. "TASK-20260720-001"
	TaskType      TaskType   `json:"task_type"`
	WarehouseID   uuid.UUID  `json:"warehouse_id"`
	OrderID       *uuid.UUID `json:"order_id,omitempty"`   // Source order (if any)
	OrderLineID   *uuid.UUID `json:"order_line_id,omitempty"`
	Priority      TaskPriority `json:"priority"`
	Status        TaskStatus `json:"status"`
	AssignedTo    string     `json:"assigned_to,omitempty"` // Worker ID or robot ID
	FromLocation  *uuid.UUID `json:"from_location_id,omitempty"` // Source location
	ToLocation    *uuid.UUID `json:"to_location_id,omitempty"`   // Target location
	SKUID         uuid.UUID  `json:"sku_id"`
	ExpectedQty   float64    `json:"expected_qty"`
	ActualQty     float64    `json:"actual_qty"`
	UOM           string     `json:"uom"`
	BatchNo       string     `json:"batch_no,omitempty"`
	Instructions  string     `json:"instructions,omitempty"` // Human-readable instructions
	CreatedAt     time.Time  `json:"created_at"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	CancelledAt   *time.Time `json:"cancelled_at,omitempty"`
}

// TaskType classifies the type of warehouse operation.
type TaskType string

const (
	TaskTypePutaway    TaskType = "putaway"    // Move goods from receiving to storage
	TaskTypePick       TaskType = "pick"       // Pick goods from storage for order
	TaskTypeReplenish  TaskType = "replenish"  // Refill pick locations from reserve
	TaskTypeTransfer   TaskType = "transfer"   // Move goods between locations
	TaskTypeCycleCount TaskType = "cycle_count" // Inventory count task
	TaskTypeLoad       TaskType = "load"       // Load goods onto truck
	TaskTypeUnload     TaskType = "unload"     // Unload goods from truck
)

// TaskPriority indicates the urgency of a task.
type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityNormal TaskPriority = "normal"
	TaskPriorityHigh   TaskPriority = "high"
	TaskPriorityUrgent TaskPriority = "urgent"
)

// TaskStatus represents the lifecycle state of a task.
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"    // Waiting to be assigned
	TaskStatusAssigned   TaskStatus = "assigned"   // Assigned to a worker/robot
	TaskStatusInProgress TaskStatus = "in_progress" // Worker has started
	TaskStatusPaused     TaskStatus = "paused"     // Temporarily paused
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusCancelled  TaskStatus = "cancelled"
	TaskStatusException  TaskStatus = "exception"  // Error/issue needs human intervention
)

// TaskAllocationStrategy defines how tasks are assigned to resources.
type TaskAllocationStrategy string

const (
	TaskAllocationFIFO        TaskAllocationStrategy = "fifo"        // First-in-first-out
	TaskAllocationPriority    TaskAllocationStrategy = "priority"    // By priority then age
	TaskAllocationShortestPath TaskAllocationStrategy = "shortest_path" // Optimize travel distance
	TaskAllocationZonePick    TaskAllocationStrategy = "zone_pick"   // Zone-based picking
)

// Wave groups multiple orders into a single picking wave for efficiency.
type Wave struct {
	ID            uuid.UUID   `json:"id"`
	WaveNo        string      `json:"wave_no"`
	WarehouseID   uuid.UUID   `json:"warehouse_id"`
	WaveType      WaveType    `json:"wave_type"`
	Status        WaveStatus  `json:"status"`
	OrderIDs      []uuid.UUID `json:"order_ids"`
	TaskIDs       []uuid.UUID `json:"task_ids"`
	TotalOrders   int         `json:"total_orders"`
	TotalLines    int         `json:"total_lines"`
	TotalQty      float64     `json:"total_qty"`
	CreatedAt     time.Time   `json:"created_at"`
	ReleasedAt    *time.Time  `json:"released_at,omitempty"`
	CompletedAt   *time.Time  `json:"completed_at,omitempty"`
}

// WaveType classifies the purpose of a picking wave.
type WaveType string

const (
	WaveTypeSingleOrder  WaveType = "single_order"  // One order per wave
	WaveTypeBatch        WaveType = "batch"          // Multiple orders batched
	WaveTypeZone         WaveType = "zone"           // By zone
	WaveTypeCarrier      WaveType = "carrier"        // By shipping carrier
)

// WaveStatus represents the lifecycle of a wave.
type WaveStatus string

const (
	WaveStatusCreated   WaveStatus = "created"
	WaveStatusReleased  WaveStatus = "released"  // Tasks generated
	WaveStatusInProgress WaveStatus = "in_progress"
	WaveStatusCompleted WaveStatus = "completed"
)
