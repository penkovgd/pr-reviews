package rest

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/penkovgd/pr-reviews/internal/core"
)

type CreatePRRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

type CreatePRResponse struct {
	PR PullRequestDto `json:"pr"`
}

type PullRequestDto struct {
	PullRequestID     string                 `json:"pull_request_id"`
	PullRequestName   string                 `json:"pull_request_name"`
	AuthorID          string                 `json:"author_id"`
	Status            core.PullRequestStatus `json:"status"`
	AssignedReviewers []string               `json:"assigned_reviewers"`
}

func ToPullRequestDto(pr *core.PullRequest) PullRequestDto {
	return PullRequestDto{
		PullRequestID:     pr.ID,
		PullRequestName:   pr.Name,
		AuthorID:          pr.AuthorID,
		Status:            pr.Status,
		AssignedReviewers: pr.AssignedReviewers,
	}
}

func NewCreatePRHandler(log *slog.Logger, prs core.PullRequestService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreatePRRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Warn("invalid request body", "error", err)
			writeAPIError(w, http.StatusBadRequest, ErrorCodeNotFound, "invalid request body")
			return
		}

		if req.PullRequestID == "" {
			writeAPIError(w, http.StatusBadRequest, ErrorCodeNotFound, "pull_request_id is required")
			return
		}
		if req.PullRequestName == "" {
			writeAPIError(w, http.StatusBadRequest, ErrorCodeNotFound, "pull_request_name is required")
			return
		}
		if req.AuthorID == "" {
			writeAPIError(w, http.StatusBadRequest, ErrorCodeNotFound, "author_id is required")
			return
		}

		pr, err := prs.CreatePR(r.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
		if err != nil {
			log.Error("create PR failed", "pr", req.PullRequestID, "error", err)

			status, code, message := toAPIError(err)
			writeAPIError(w, status, code, message)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		resp := CreatePRResponse{
			PR: ToPullRequestDto(pr),
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Error("encode response", "error", err)
		}
	}
}
