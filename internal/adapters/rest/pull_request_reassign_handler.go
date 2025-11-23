package rest

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/penkovgd/pr-reviews/internal/core"
)

type ReassignReviewerRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_reviewer_id"`
}

type ReassignReviewerResponse struct {
	PR         PullRequestDto `json:"pr"`
	ReplacedBy string         `json:"replaced_by"`
}

func NewReassignReviewerHandler(log *slog.Logger, prs core.PullRequestService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReassignReviewerRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Warn("invalid request body", "error", err)
			writeAPIError(w, http.StatusBadRequest, ErrorCodeNotFound, "invalid request body")
			return
		}

		if req.PullRequestID == "" {
			writeAPIError(w, http.StatusBadRequest, ErrorCodeNotFound, "pull_request_id is required")
			return
		}
		if req.OldUserID == "" {
			writeAPIError(w, http.StatusBadRequest, ErrorCodeNotFound, "old_reviewer_id is required")
			return
		}

		reassignment, err := prs.ReassignReviewer(r.Context(), req.PullRequestID, req.OldUserID)
		if err != nil {
			log.Error("reassign reviewer failed", "pr", req.PullRequestID, "old_reviewer_id", req.OldUserID, "error", err)

			status, code, message := toAPIError(err)
			writeAPIError(w, status, code, message)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		resp := ReassignReviewerResponse{
			PR:         ToPullRequestDto(reassignment.PR),
			ReplacedBy: reassignment.NewReviewerID,
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Error("encode response", "error", err)
		}
	}
}
