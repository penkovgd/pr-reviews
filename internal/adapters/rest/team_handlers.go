package rest

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/penkovgd/pr-reviews/internal/core"
)

type TeamDto struct {
	TeamName string      `json:"team_name"`
	Members  []MemberDto `json:"members"`
}

type MemberDto struct {
	ID       string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

func ToTeam(dto TeamDto) *core.Team {
	t := core.Team{Name: dto.TeamName}

	for _, member := range dto.Members {
		t.Members = append(t.Members, core.User{
			ID:       member.ID,
			Username: member.Username,
			TeamName: dto.TeamName,
			IsActive: member.IsActive,
		})
	}
	return &t
}
func ToTeamDto(t *core.Team) TeamDto {
	dto := TeamDto{TeamName: t.Name}

	for _, member := range t.Members {
		dto.Members = append(dto.Members, MemberDto{
			ID:       member.ID,
			Username: member.Username,
			IsActive: member.IsActive,
		})
	}

	return dto
}

type AddTeamResponse struct {
	Team TeamDto `json:"team"`
}

func NewAddTeamHandler(log *slog.Logger, ts core.TeamService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var team TeamDto

		if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
			log.Warn("invalid request body", "error", err)
			writeAPIError(w, http.StatusBadRequest, ErrorCodeNotFound, "invalid request body")
			return
		}

		if team.TeamName == "" {
			writeAPIError(w, http.StatusBadRequest, ErrorCodeNotFound, "team_name is required")
			return
		}

		if err := ts.CreateTeam(r.Context(), ToTeam(team)); err != nil {
			log.Error("create team", "team", team.TeamName, "error", err)

			status, code, message := toAPIError(err)
			writeAPIError(w, status, code, message)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(AddTeamResponse{Team: team}); err != nil {
			log.Error("encode response", "error", err)
		}
	}
}

func NewGetTeamHandler(log *slog.Logger, ts core.TeamService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		teamName := r.URL.Query().Get("team_name")
		if teamName == "" {
			log.Warn("team_name param is required")
			writeAPIError(w, http.StatusBadRequest, ErrorCodeNotFound, "team_name param is required")
			return
		}

		team, err := ts.GetTeam(r.Context(), teamName)
		if err != nil {
			log.Error("get team", "team", teamName, "error", err)

			status, code, message := toAPIError(err)
			writeAPIError(w, status, code, message)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		resp := ToTeamDto(team)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Error("encode response", "error", err)
		}
	}
}
