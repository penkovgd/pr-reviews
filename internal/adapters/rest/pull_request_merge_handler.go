package rest

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/penkovgd/pr-reviews/internal/core"
)

type MergePRRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

type MergePRResponse struct {
	PR PullRequestMergedDto `json:"pr"`
}

type PullRequestMergedDto struct {
	PullRequestID     string                 `json:"pull_request_id"`
	PullRequestName   string                 `json:"pull_request_name"`
	AuthorID          string                 `json:"author_id"`
	Status            core.PullRequestStatus `json:"status"`
	AssignedReviewers []string               `json:"assigned_reviewers"`
	MergedAt          string                 `json:"mergedAt"`
}

func ToPullRequestMergedDto(pr *core.PullRequest) PullRequestMergedDto {
	var mergedAtStr string
	if pr.MergedAt != nil {
		mergedAtStr = pr.MergedAt.Format(time.RFC3339)
	}

	return PullRequestMergedDto{
		PullRequestID:     pr.ID,
		PullRequestName:   pr.Name,
		AuthorID:          pr.AuthorID,
		Status:            pr.Status,
		AssignedReviewers: pr.AssignedReviewers,
		MergedAt:          mergedAtStr,
	}
}

func NewMergePRHandler(log *slog.Logger, prs core.PullRequestService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req MergePRRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Warn("invalid request body", "error", err)
			writeAPIError(w, http.StatusBadRequest, ErrorCodeNotFound, "invalid request body")
			return
		}

		if req.PullRequestID == "" {
			writeAPIError(w, http.StatusBadRequest, ErrorCodeNotFound, "pull_request_id is required")
			return
		}

		pr, err := prs.MergePR(r.Context(), req.PullRequestID)
		if err != nil {
			log.Error("merge PR failed", "pr", req.PullRequestID, "error", err)

			status, code, message := toAPIError(err)
			writeAPIError(w, status, code, message)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		resp := MergePRResponse{
			PR: ToPullRequestMergedDto(pr),
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Error("encode response", "error", err)
		}
	}
}
