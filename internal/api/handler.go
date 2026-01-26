package api

import (
	"encoding/json"
	"net/http"
	"sql-sharding-v2/internal/executor"
	"sql-sharding-v2/pkg/logger"
)

type Handler struct {
	app interface {
		ExecuteSQL(projectID string, sql string) ([]executor.ExecutionResult, error)
	}
}

func NewHandler(app interface {
	ExecuteSQL(projectID string, sql string) ([]executor.ExecutionResult, error)
}) *Handler {
	return &Handler{app: app}
}

func (h *Handler) ExecuteQuery(w http.ResponseWriter, r *http.Request) {

	var req ExecuteQueryRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.ProjectID == "" || req.SQL == "" {
		http.Error(w, "project_id and sql are required", http.StatusBadRequest)
		return
	}

	results, err := h.app.ExecuteSQL(req.ProjectID, req.SQL)
	if err != nil {
		logger.Logger.Error("query execution failed", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := ExecuteQueryResponse{
		Results: make([]ShardResultResponse, 0, len(results)),
	}

	for _, r := range results {
		out := ShardResultResponse{
			ShardID:      r.ShardID,
			Columns:      r.Columns,
			Rows:         r.Rows,
			RowsAffected: r.RowsAffected,
		}

		if r.Err != nil {
			out.Error = r.Err.Error()
		}

		resp.Results = append(resp.Results, out)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
