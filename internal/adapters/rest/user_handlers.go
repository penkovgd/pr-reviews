package rest

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/penkovgd/pr-reviews/internal/core"
)

type SetUserActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type SetUserActiveResponse struct {
	User UserDto `json:"user"`
}

type UserDto struct {
	ID       string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

func NewSetUserActiveHandler(log *slog.Logger, us core.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SetUserActiveRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Warn("invalid request body", "error", err)
			writeAPIError(w, http.StatusBadRequest, ErrorCodeNotFound, "invalid request body")
			return
		}

		if req.UserID == "" {
			writeAPIError(w, http.StatusBadRequest, ErrorCodeNotFound, "user_id is required")
			return
		}

		user, err := us.SetUserActive(r.Context(), req.UserID, req.IsActive)
		if err != nil {
			log.Error("set user active failed", "user", req.UserID, "error", err)

			status, code, message := toAPIError(err)
			writeAPIError(w, status, code, message)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		resp := SetUserActiveResponse{
			User: UserDto(*user),
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Error("encode response", "error", err)
		}
	}
}

type UserReviewResponse struct {
	UserID       string                `json:"user_id"`
	PullRequests []PullRequestShortDto `json:"pull_requests"`
}

type PullRequestShortDto struct {
	ID       string                 `json:"pull_request_id"`
	Name     string                 `json:"pull_request_name"`
	AuthorID string                 `json:"author_id"`
	Status   core.PullRequestStatus `json:"status"`
}

func ToPullRequestShortDto(pr *core.PullRequest) PullRequestShortDto {
	return PullRequestShortDto{
		ID:       pr.ID,
		Name:     pr.Name,
		AuthorID: pr.AuthorID,
		Status:   pr.Status,
	}
}

func NewGetUserReviewHandler(log *slog.Logger, us core.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			log.Warn("user_id parameter is required")
			writeAPIError(w, http.StatusBadRequest, ErrorCodeNotFound, "user_id parameter is required")
			return
		}

		prs, err := us.GetUserReviewRequests(r.Context(), userID)
		if err != nil {
			log.Error("get user review requests failed", "user", userID, "error", err)

			status, code, message := toAPIError(err)
			writeAPIError(w, status, code, message)
			return
		}

		prShorts := make([]PullRequestShortDto, len(prs))
		for i, pr := range prs {
			prShorts[i] = ToPullRequestShortDto(pr)
		}

		response := UserReviewResponse{
			UserID:       userID,
			PullRequests: prShorts,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Error("encode response", "error", err)
		}
	}
}
