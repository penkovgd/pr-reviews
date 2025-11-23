package rest

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/penkovgd/pr-reviews/internal/core"
)

type UserAssignmentStatsResponse struct {
	UserAssignments map[string]int `json:"user_assignments"`
}

func NewUserAssignmentStatsHandler(log *slog.Logger, stats core.Statistics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := stats.GetUserAssignmentStats(r.Context())
		if err != nil {
			log.Error("get user assignment stats failed", "error", err)
			writeAPIError(w, http.StatusInternalServerError, ErrorCodeNotFound, "internal server error")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		response := UserAssignmentStatsResponse{
			UserAssignments: stats,
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Error("encode response", "error", err)
		}
	}
}
