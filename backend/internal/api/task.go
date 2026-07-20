// Package api provides HTTP handlers and route registration for the WMS API.
package api

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
	"github.com/ai-wms/ai-wms/backend/internal/repository"
	pkgerrors "github.com/ai-wms/ai-wms/backend/pkg/errors"

	"github.com/ai-wms/ai-wms/backend/internal/service"
)

// TaskHandler handles HTTP requests for task resources (used by both Admin and PDA).
type TaskHandler struct {
	svc *service.TaskService
	log *slog.Logger
}

// NewTaskHandler creates a new TaskHandler.
func NewTaskHandler(svc *service.TaskService, log *slog.Logger) *TaskHandler {
	return &TaskHandler{svc: svc, log: log}
}

// ── Response Types ─────────────────────────────────────────────────────────────────────

// taskResponse is the JSON shape returned for task endpoints.
type taskResponse struct {
	ID           string  `json:"id"`
	TaskNo       string  `json:"task_no"`
	TaskType     string  `json:"task_type"`
	WarehouseID  string  `json:"warehouse_id"`
	OrderID      string  `json:"order_id,omitempty"`
	OrderLineID  string  `json:"order_line_id,omitempty"`
	Priority     string  `json:"priority"`
	Status       string  `json:"status"`
	AssignedTo   string  `json:"assigned_to,omitempty"`
	FromLocation string  `json:"from_location_id,omitempty"`
	ToLocation   string  `json:"to_location_id,omitempty"`
	SKUID        string  `json:"sku_id"`
	ExpectedQty  float64 `json:"expected_qty"`
	ActualQty    float64 `json:"actual_qty"`
	UOM          string  `json:"uom"`
	BatchNo      string  `json:"batch_no,omitempty"`
	Instructions string  `json:"instructions,omitempty"`
	CreatedAt    string  `json:"created_at"`
	StartedAt    string  `json:"started_at,omitempty"`
	CompletedAt  string  `json:"completed_at,omitempty"`
	CancelledAt  string  `json:"cancelled_at,omitempty"`
}

func toTaskResponse(t *domain.Task) taskResponse {
	r := taskResponse{
		ID:           t.ID.String(),
		TaskNo:       t.TaskNo,
		TaskType:     string(t.TaskType),
		WarehouseID:  t.WarehouseID.String(),
		Priority:     string(t.Priority),
		Status:       string(t.Status),
		AssignedTo:   t.AssignedTo,
		SKUID:        t.SKUID.String(),
		ExpectedQty:  t.ExpectedQty,
		ActualQty:    t.ActualQty,
		UOM:          t.UOM,
		BatchNo:      t.BatchNo,
		Instructions: t.Instructions,
		CreatedAt:    t.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if t.OrderID != nil {
		r.OrderID = t.OrderID.String()
	}
	if t.OrderLineID != nil {
		r.OrderLineID = t.OrderLineID.String()
	}
	if t.FromLocation != nil {
		r.FromLocation = t.FromLocation.String()
	}
	if t.ToLocation != nil {
		r.ToLocation = t.ToLocation.String()
	}
	if t.StartedAt != nil {
		r.StartedAt = t.StartedAt.Format("2006-01-02T15:04:05Z")
	}
	if t.CompletedAt != nil {
		r.CompletedAt = t.CompletedAt.Format("2006-01-02T15:04:05Z")
	}
	if t.CancelledAt != nil {
		r.CancelledAt = t.CancelledAt.Format("2006-01-02T15:04:05Z")
	}
	return r
}

// ── Task Handlers ─────────────────────────────────────────────────────────────────────

// CreateTask handles POST /api/v1/tasks
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var input service.CreateTaskInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	task, err := h.svc.CreateTask(r.Context(), input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteCreated(w, toTaskResponse(task))
}

// GetTask handles GET /api/v1/tasks/{id}
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	task, err := h.svc.GetTask(r.Context(), id)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toTaskResponse(task))
}

// ListTasks handles GET /api/v1/tasks
func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	filter := repository.TaskFilter{
		Limit:  QueryParamInt(r, "limit", 50),
		Offset: QueryParamInt(r, "offset", 0),
	}

	if raw := QueryParam(r, "warehouse_id", ""); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			WriteError(w, r, pkgerrors.NewInvalidInput("invalid warehouse_id UUID"))
			return
		}
		filter.WarehouseID = id
	}
	if raw := QueryParam(r, "task_type", ""); raw != "" {
		filter.TaskType = domain.TaskType(raw)
	}
	if raw := QueryParam(r, "status", ""); raw != "" {
		filter.Status = domain.TaskStatus(raw)
	}
	if raw := QueryParam(r, "assigned_to", ""); raw != "" {
		filter.AssignedTo = raw
	}

	tasks, err := h.svc.ListTasks(r.Context(), filter)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	resp := make([]taskResponse, 0, len(tasks))
	for _, t := range tasks {
		resp = append(resp, toTaskResponse(t))
	}

	WriteJSON(w, http.StatusOK, resp)
}

// AssignTask handles POST /api/v1/tasks/{id}/assign (PDA: assign task to worker)
func (h *TaskHandler) AssignTask(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	var input service.AssignTaskInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	task, err := h.svc.AssignTask(r.Context(), id, input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toTaskResponse(task))
}

// UpdateTaskStatus handles PUT /api/v1/tasks/{id}/status (PDA: start, pause, resume, cancel)
func (h *TaskHandler) UpdateTaskStatus(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	var input service.UpdateTaskStatusInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	task, err := h.svc.UpdateTaskStatus(r.Context(), id, input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toTaskResponse(task))
}

// CompleteTask handles POST /api/v1/tasks/{id}/complete (PDA: complete with actual qty)
func (h *TaskHandler) CompleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := PathUUID(r, "id")
	if err != nil {
		WriteError(w, r, err)
		return
	}

	var input service.CompleteTaskInput
	if err := ReadJSON(r, &input); err != nil {
		WriteError(w, r, err)
		return
	}

	task, err := h.svc.CompleteTask(r.Context(), id, input)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	WriteJSON(w, http.StatusOK, toTaskResponse(task))
}
