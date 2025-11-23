package rest

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/penkovgd/pr-reviews/internal/core"
)

type ErrorCode string

const (
	ErrorCodeTeamExists  ErrorCode = "TEAM_EXISTS"
	ErrorCodePRExists    ErrorCode = "PR_EXISTS"
	ErrorCodePRMerged    ErrorCode = "PR_MERGED"
	ErrorCodeNotAssigned ErrorCode = "NOT_ASSIGNED"
	ErrorCodeNoCandidate ErrorCode = "NO_CANDIDATE"
	ErrorCodeNotFound    ErrorCode = "NOT_FOUND"
)

type ErrorResponse struct {
	Error struct {
		Code    ErrorCode `json:"code"`
		Message string    `json:"message"`
	} `json:"error"`
}

func writeAPIError(w http.ResponseWriter, status int, code ErrorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := ErrorResponse{}
	response.Error.Code = code
	response.Error.Message = message

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("encode json response", "error", err)
	}
}

func toAPIError(err error) (int, ErrorCode, string) {
	switch {
	case errors.Is(err, core.ErrTeamExists):
		return http.StatusBadRequest, ErrorCodeTeamExists, "team_name already exists"
	case errors.Is(err, core.ErrTeamNotFound):
		return http.StatusNotFound, ErrorCodeNotFound, "resource not found"
	case errors.Is(err, core.ErrUserNotFound):
		return http.StatusNotFound, ErrorCodeNotFound, "resource not found"
	case errors.Is(err, core.ErrUserNotActive):
		return http.StatusNotFound, ErrorCodeNotFound, "resource not found"
	case errors.Is(err, core.ErrPRExists):
		return http.StatusConflict, ErrorCodePRExists, "PR id already exists"
	case errors.Is(err, core.ErrPRNotFound):
		return http.StatusNotFound, ErrorCodeNotFound, "resource not found"
	case errors.Is(err, core.ErrPRMerged):
		return http.StatusConflict, ErrorCodePRMerged, "cannot reassign on merged PR"
	case errors.Is(err, core.ErrReviewerNotAssigned):
		return http.StatusConflict, ErrorCodeNotAssigned, "reviewer is not assigned to this PR"
	case errors.Is(err, core.ErrNoCandidate):
		return http.StatusConflict, ErrorCodeNoCandidate, "no active replacement candidate in team"
	default:
		return http.StatusInternalServerError, ErrorCodeNotFound, "internal server error"
	}
}
