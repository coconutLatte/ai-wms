package domain

import (
	"testing"

	"github.com/google/uuid"
)

// ── Task State Machine Tests ─────────────────────────────────────────────────

func TestTask_CanTransitionTo_ValidTransitions(t *testing.T) {
	tests := []struct {
		name    string
		current TaskStatus
		target  TaskStatus
		want    bool
	}{
		// pending → assigned, cancelled
		{name: "pending → assigned", current: TaskStatusPending, target: TaskStatusAssigned, want: true},
		{name: "pending → cancelled", current: TaskStatusPending, target: TaskStatusCancelled, want: true},
		// assigned → in_progress, cancelled
		{name: "assigned → in_progress", current: TaskStatusAssigned, target: TaskStatusInProgress, want: true},
		{name: "assigned → cancelled", current: TaskStatusAssigned, target: TaskStatusCancelled, want: true},
		// in_progress → completed, paused, cancelled
		{name: "in_progress → completed", current: TaskStatusInProgress, target: TaskStatusCompleted, want: true},
		{name: "in_progress → paused", current: TaskStatusInProgress, target: TaskStatusPaused, want: true},
		{name: "in_progress → cancelled", current: TaskStatusInProgress, target: TaskStatusCancelled, want: true},
		// paused → in_progress, cancelled
		{name: "paused → in_progress (resume)", current: TaskStatusPaused, target: TaskStatusInProgress, want: true},
		{name: "paused → cancelled", current: TaskStatusPaused, target: TaskStatusCancelled, want: true},
		// exception → in_progress, cancelled
		{name: "exception → in_progress (resolve)", current: TaskStatusException, target: TaskStatusInProgress, want: true},
		{name: "exception → cancelled", current: TaskStatusException, target: TaskStatusCancelled, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Status: tt.current}
			got := task.CanTransitionTo(tt.target)
			if got != tt.want {
				t.Errorf("CanTransitionTo(%q) = %v, want %v", tt.target, got, tt.want)
			}
		})
	}
}

func TestTask_CanTransitionTo_InvalidTransitions(t *testing.T) {
	tests := []struct {
		name    string
		current TaskStatus
		target  TaskStatus
	}{
		// Can't skip statuses
		{name: "pending → in_progress (skip assigned)", current: TaskStatusPending, target: TaskStatusInProgress},
		{name: "pending → completed (skip all)", current: TaskStatusPending, target: TaskStatusCompleted},
		{name: "pending → paused", current: TaskStatusPending, target: TaskStatusPaused},
		{name: "pending → exception", current: TaskStatusPending, target: TaskStatusException},
		{name: "assigned → completed (skip in_progress)", current: TaskStatusAssigned, target: TaskStatusCompleted},
		{name: "assigned → paused", current: TaskStatusAssigned, target: TaskStatusPaused},
		{name: "assigned → exception", current: TaskStatusAssigned, target: TaskStatusException},
		// Can't go backwards
		{name: "assigned → pending", current: TaskStatusAssigned, target: TaskStatusPending},
		{name: "in_progress → assigned", current: TaskStatusInProgress, target: TaskStatusAssigned},
		{name: "in_progress → pending", current: TaskStatusInProgress, target: TaskStatusPending},
		// Paused can only go to in_progress or cancelled
		{name: "paused → assigned", current: TaskStatusPaused, target: TaskStatusAssigned},
		{name: "paused → completed", current: TaskStatusPaused, target: TaskStatusCompleted},
		{name: "paused → pending", current: TaskStatusPaused, target: TaskStatusPending},
		// Exception can only go to in_progress or cancelled
		{name: "exception → completed", current: TaskStatusException, target: TaskStatusCompleted},
		{name: "exception → assigned", current: TaskStatusException, target: TaskStatusAssigned},
		{name: "exception → pending", current: TaskStatusException, target: TaskStatusPending},
		{name: "exception → paused", current: TaskStatusException, target: TaskStatusPaused},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Status: tt.current}
			if task.CanTransitionTo(tt.target) {
				t.Errorf("CanTransitionTo(%q) = true, want false", tt.target)
			}
		})
	}
}

func TestTask_CanTransitionTo_NoopRejected(t *testing.T) {
	task := &Task{Status: TaskStatusPending}
	if task.CanTransitionTo(TaskStatusPending) {
		t.Error("same-status transition should return false")
	}
}

func TestTask_CanTransitionTo_TerminalStates(t *testing.T) {
	// Completed is terminal.
	completed := &Task{Status: TaskStatusCompleted}
	for _, target := range []TaskStatus{
		TaskStatusPending, TaskStatusAssigned, TaskStatusInProgress,
		TaskStatusPaused, TaskStatusCancelled, TaskStatusException,
	} {
		if completed.CanTransitionTo(target) {
			t.Errorf("completed → %s: should be false", target)
		}
	}

	// Cancelled is terminal.
	cancelled := &Task{Status: TaskStatusCancelled}
	for _, target := range []TaskStatus{
		TaskStatusPending, TaskStatusAssigned, TaskStatusInProgress,
		TaskStatusPaused, TaskStatusCompleted, TaskStatusException,
	} {
		if cancelled.CanTransitionTo(target) {
			t.Errorf("cancelled → %s: should be false", target)
		}
	}
}

func TestTask_IsTerminal(t *testing.T) {
	terminal := []TaskStatus{TaskStatusCompleted, TaskStatusCancelled}
	nonTerminal := []TaskStatus{
		TaskStatusPending, TaskStatusAssigned, TaskStatusInProgress,
		TaskStatusPaused, TaskStatusException,
	}

	for _, s := range terminal {
		task := &Task{Status: s}
		if !task.IsTerminal() {
			t.Errorf("status %q should be terminal", s)
		}
	}
	for _, s := range nonTerminal {
		task := &Task{Status: s}
		if task.IsTerminal() {
			t.Errorf("status %q should NOT be terminal", s)
		}
	}
}

func TestTask_CancellationFromAllNonTerminal(t *testing.T) {
	nonTerminal := []TaskStatus{
		TaskStatusPending, TaskStatusAssigned, TaskStatusInProgress,
		TaskStatusPaused, TaskStatusException,
	}

	for _, s := range nonTerminal {
		task := &Task{Status: s}
		if !task.CanTransitionTo(TaskStatusCancelled) {
			t.Errorf("%s → cancelled should be valid", s)
		}
	}
}

func TestTask_FullLifecycle(t *testing.T) {
	task := &Task{Status: TaskStatusPending}

	transitions := []TaskStatus{
		TaskStatusAssigned,
		TaskStatusInProgress,
		TaskStatusCompleted,
	}

	for _, target := range transitions {
		if !task.CanTransitionTo(target) {
			t.Fatalf("transition to %s should be valid from %s", target, task.Status)
		}
		task.Status = target
	}

	if task.Status != TaskStatusCompleted {
		t.Errorf("final status = %s, want %s", task.Status, TaskStatusCompleted)
	}
}

func TestTask_PauseResumeFlow(t *testing.T) {
	task := &Task{Status: TaskStatusPending}

	// pending → assigned → in_progress.
	task.Status = TaskStatusAssigned
	task.Status = TaskStatusInProgress

	// Pause.
	if !task.CanTransitionTo(TaskStatusPaused) {
		t.Fatal("in_progress → paused should be valid")
	}
	task.Status = TaskStatusPaused

	// Resume.
	if !task.CanTransitionTo(TaskStatusInProgress) {
		t.Fatal("paused → in_progress should be valid")
	}
	task.Status = TaskStatusInProgress

	// Complete.
	if !task.CanTransitionTo(TaskStatusCompleted) {
		t.Fatal("in_progress → completed should be valid")
	}
	task.Status = TaskStatusCompleted

	if !task.IsTerminal() {
		t.Error("completed task should be terminal")
	}
}

func TestTask_ExceptionResolveFlow(t *testing.T) {
	task := &Task{Status: TaskStatusInProgress}

	// Exception can't be reached via CanTransitionTo — it's set by system events.
	// But we can test resolving from exception.
	task.Status = TaskStatusException

	if !task.CanTransitionTo(TaskStatusInProgress) {
		t.Error("exception → in_progress should be valid")
	}
	if !task.CanTransitionTo(TaskStatusCancelled) {
		t.Error("exception → cancelled should be valid")
	}

	// Resolve exception.
	task.Status = TaskStatusInProgress

	if !task.CanTransitionTo(TaskStatusCompleted) {
		t.Error("in_progress → completed should be valid after resolving exception")
	}
}

// ── Task Convenience Methods ─────────────────────────────────────────────────

func TestTask_CanBeAssigned(t *testing.T) {
	// Only pending tasks can be assigned.
	pending := &Task{Status: TaskStatusPending}
	if !pending.CanBeAssigned() {
		t.Error("pending task should be assignable")
	}

	nonPending := []TaskStatus{
		TaskStatusAssigned, TaskStatusInProgress,
		TaskStatusPaused, TaskStatusCompleted,
		TaskStatusCancelled, TaskStatusException,
	}
	for _, s := range nonPending {
		task := &Task{Status: s}
		if task.CanBeAssigned() {
			t.Errorf("task in %s should NOT be assignable", s)
		}
	}
}

func TestTask_CanBeStarted(t *testing.T) {
	// Assigned and exception tasks can be started (moved to in_progress).
	canStart := []TaskStatus{TaskStatusAssigned, TaskStatusException}
	for _, s := range canStart {
		task := &Task{Status: s}
		if !task.CanBeStarted() {
			t.Errorf("task in %s should be startable", s)
		}
	}

	cannotStart := []TaskStatus{
		TaskStatusPending, TaskStatusInProgress,
		TaskStatusPaused, TaskStatusCompleted,
		TaskStatusCancelled,
	}
	for _, s := range cannotStart {
		task := &Task{Status: s}
		if task.CanBeStarted() {
			t.Errorf("task in %s should NOT be startable", s)
		}
	}
}

func TestTask_CanBeCompleted(t *testing.T) {
	// Only in_progress tasks can be completed.
	inProgress := &Task{Status: TaskStatusInProgress}
	if !inProgress.CanBeCompleted() {
		t.Error("in_progress task should be completable")
	}

	notInProgress := []TaskStatus{
		TaskStatusPending, TaskStatusAssigned,
		TaskStatusPaused, TaskStatusCompleted,
		TaskStatusCancelled, TaskStatusException,
	}
	for _, s := range notInProgress {
		task := &Task{Status: s}
		if task.CanBeCompleted() {
			t.Errorf("task in %s should NOT be completable", s)
		}
	}
}

func TestTask_CanBePaused(t *testing.T) {
	// Only in_progress can be paused.
	inProgress := &Task{Status: TaskStatusInProgress}
	if !inProgress.CanBePaused() {
		t.Error("in_progress task should be pausable")
	}

	notInProgress := []TaskStatus{
		TaskStatusPending, TaskStatusAssigned,
		TaskStatusPaused, TaskStatusCompleted,
		TaskStatusCancelled, TaskStatusException,
	}
	for _, s := range notInProgress {
		task := &Task{Status: s}
		if task.CanBePaused() {
			t.Errorf("task in %s should NOT be pausable", s)
		}
	}
}

func TestTask_CanBeResumed(t *testing.T) {
	// Paused and exception tasks can be resumed.
	canResume := []TaskStatus{TaskStatusPaused, TaskStatusException}
	for _, s := range canResume {
		task := &Task{Status: s}
		if !task.CanBeResumed() {
			t.Errorf("task in %s should be resumable", s)
		}
	}

	cannotResume := []TaskStatus{
		TaskStatusPending, TaskStatusAssigned,
		TaskStatusInProgress, TaskStatusCompleted,
		TaskStatusCancelled,
	}
	for _, s := range cannotResume {
		task := &Task{Status: s}
		if task.CanBeResumed() {
			t.Errorf("task in %s should NOT be resumable", s)
		}
	}
}

// ── Wave State Machine Tests ─────────────────────────────────────────────────

func TestWave_CanTransitionTo_ValidTransitions(t *testing.T) {
	tests := []struct {
		name    string
		current WaveStatus
		target  WaveStatus
		want    bool
	}{
		{name: "created → released", current: WaveStatusCreated, target: WaveStatusReleased, want: true},
		{name: "released → in_progress", current: WaveStatusReleased, target: WaveStatusInProgress, want: true},
		{name: "in_progress → completed", current: WaveStatusInProgress, target: WaveStatusCompleted, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Wave{Status: tt.current}
			got := w.CanTransitionTo(tt.target)
			if got != tt.want {
				t.Errorf("CanTransitionTo(%q) = %v, want %v", tt.target, got, tt.want)
			}
		})
	}
}

func TestWave_CanTransitionTo_InvalidTransitions(t *testing.T) {
	tests := []struct {
		name    string
		current WaveStatus
		target  WaveStatus
	}{
		// Skip statuses
		{name: "created → in_progress (skip released)", current: WaveStatusCreated, target: WaveStatusInProgress},
		{name: "created → completed (skip all)", current: WaveStatusCreated, target: WaveStatusCompleted},
		{name: "released → completed (skip in_progress)", current: WaveStatusReleased, target: WaveStatusCompleted},
		// Backwards
		{name: "released → created", current: WaveStatusReleased, target: WaveStatusCreated},
		{name: "in_progress → released", current: WaveStatusInProgress, target: WaveStatusReleased},
		{name: "in_progress → created", current: WaveStatusInProgress, target: WaveStatusCreated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Wave{Status: tt.current}
			if w.CanTransitionTo(tt.target) {
				t.Errorf("CanTransitionTo(%q) = true, want false", tt.target)
			}
		})
	}
}

func TestWave_CanTransitionTo_Terminal(t *testing.T) {
	completed := &Wave{Status: WaveStatusCompleted}
	for _, target := range []WaveStatus{
		WaveStatusCreated, WaveStatusReleased, WaveStatusInProgress,
	} {
		if completed.CanTransitionTo(target) {
			t.Errorf("completed → %s: should be false", target)
		}
	}
}

func TestWave_IsTerminal(t *testing.T) {
	completed := &Wave{Status: WaveStatusCompleted}
	if !completed.IsTerminal() {
		t.Error("completed wave should be terminal")
	}

	nonTerminal := []WaveStatus{
		WaveStatusCreated, WaveStatusReleased, WaveStatusInProgress,
	}
	for _, s := range nonTerminal {
		w := &Wave{Status: s}
		if w.IsTerminal() {
			t.Errorf("status %q should NOT be terminal", s)
		}
	}
}

func TestWave_FullLifecycle(t *testing.T) {
	w := &Wave{Status: WaveStatusCreated}

	transitions := []WaveStatus{
		WaveStatusReleased,
		WaveStatusInProgress,
		WaveStatusCompleted,
	}

	for _, target := range transitions {
		if !w.CanTransitionTo(target) {
			t.Fatalf("transition to %s should be valid from %s", target, w.Status)
		}
		w.Status = target
	}

	if !w.IsTerminal() {
		t.Error("completed wave should be terminal")
	}
}

// ── Task Constants Tests ─────────────────────────────────────────────────────

func TestTaskTypeValues(t *testing.T) {
	all := []TaskType{
		TaskTypePutaway, TaskTypePick, TaskTypeReplenish,
		TaskTypeTransfer, TaskTypeCycleCount, TaskTypeLoad, TaskTypeUnload,
	}

	for _, tt := range all {
		if tt == "" {
			t.Error("task type should not be empty")
		}
	}
}

func TestTaskStatusValues(t *testing.T) {
	all := []TaskStatus{
		TaskStatusPending, TaskStatusAssigned, TaskStatusInProgress,
		TaskStatusPaused, TaskStatusCompleted, TaskStatusCancelled,
		TaskStatusException,
	}

	for _, s := range all {
		if s == "" {
			t.Error("task status should not be empty")
		}
	}
}

func TestTaskPriorityValues(t *testing.T) {
	all := []TaskPriority{
		TaskPriorityLow, TaskPriorityNormal,
		TaskPriorityHigh, TaskPriorityUrgent,
	}

	for _, p := range all {
		if p == "" {
			t.Error("task priority should not be empty")
		}
	}
}

func TestWaveTypeValues(t *testing.T) {
	all := []WaveType{
		WaveTypeSingleOrder, WaveTypeBatch, WaveTypeZone, WaveTypeCarrier,
	}

	for _, wt := range all {
		if wt == "" {
			t.Error("wave type should not be empty")
		}
	}
}

// ── Task Struct Tests ────────────────────────────────────────────────────────

func TestTask_DefaultValues(t *testing.T) {
	task := &Task{
		TaskType:    TaskTypePick,
		WarehouseID: uuid.New(),
		SKUID:       uuid.New(),
		ExpectedQty: 10,
	}
	// Default status should be empty until explicitly set.
	if task.Status != "" {
		t.Errorf("zero-value status should be empty, got %q", task.Status)
	}
	if task.Priority != "" {
		t.Errorf("zero-value priority should be empty, got %q", task.Priority)
	}
}

func TestWave_DefaultValues(t *testing.T) {
	w := &Wave{}
	if w.Status != "" {
		t.Errorf("zero-value wave status should be empty, got %q", w.Status)
	}
}
