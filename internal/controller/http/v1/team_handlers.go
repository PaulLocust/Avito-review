// internal/controller/http/v1/team_handlers.go
package v1

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/PaulLocust/Avito-review/internal/entity"
	"github.com/PaulLocust/Avito-review/internal/usecase"
	"github.com/PaulLocust/Avito-review/internal/dto"
	"github.com/PaulLocust/Avito-review/pkg/logger"
)

type teamHandlers struct {
	teamUC usecase.TeamUseCase
	logger logger.Interface
}

func newTeamHandlers(teamUC usecase.TeamUseCase, l logger.Interface) *teamHandlers {
	return &teamHandlers{
		teamUC: teamUC,
		logger: l,
	}
}

// AddTeam создает команду с участниками
// @Summary Создать команду с участниками (создаёт/обновляет пользователей)
// @Description Создает новую команду и обновляет/создает пользователей
// @Tags Teams
// @Accept json
// @Produce json
// @Param team body dto.Team true "Данные команды"
// @Success 201 {object} map[string]interface{} "Команда создана"
// @Failure 400 {object} dto.ErrorResponse "Команда уже существует"
// @Router /team/add [post]
func (h *teamHandlers) addTeam(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("POST /api/v1/team/add")

	var req dto.Team
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body: %v", err)
		writeErrorResponse(w, http.StatusBadRequest, entity.ErrorInvalidInput, "invalid request body")
		return
	}

	// Конвертируем DTO в entity
	team := entity.Team{
		Name:    req.TeamName,
		Members: make([]entity.TeamMember, len(req.Members)),
	}
	for i, member := range req.Members {
		team.Members[i] = entity.TeamMember{
			UserID:   member.UserId,
			Username: member.Username,
			IsActive: member.IsActive,
		}
	}

	// Создаем команду
	if err := h.teamUC.CreateTeam(r.Context(), team); err != nil {
		h.handleError(w, err)
		return
	}

	// Конвертируем обратно в DTO для ответа
	response := dto.Team{
		TeamName: team.Name,
		Members:  make([]dto.TeamMember, len(team.Members)),
	}
	for i, member := range team.Members {
		response.Members[i] = dto.TeamMember{
			UserId:   member.UserID,
			Username: member.Username,
			IsActive: member.IsActive,
		}
	}

	writeJSONResponse(w, http.StatusCreated, map[string]interface{}{
		"team": response,
	})
}

// GetTeam возвращает команду с участниками
// @Summary Получить команду с участниками
// @Description Возвращает информацию о команде и её участниках
// @Tags Teams
// @Accept json
// @Produce json
// @Param team_name query string true "Уникальное имя команды"
// @Success 200 {object} dto.Team "Объект команды"
// @Router /team/get [get]
func (h *teamHandlers) getTeam(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("GET /api/v1/team/get")

	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		writeErrorResponse(w, http.StatusBadRequest, entity.ErrorInvalidInput, "team_name is required")
		return
	}

	team, err := h.teamUC.GetTeam(r.Context(), teamName)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Конвертируем в DTO
	response := dto.Team{
		TeamName: team.Name,
		Members:  make([]dto.TeamMember, len(team.Members)),
	}
	for i, member := range team.Members {
		response.Members[i] = dto.TeamMember{
			UserId:   member.UserID,
			Username: member.Username,
			IsActive: member.IsActive,
		}
	}

	writeJSONResponse(w, http.StatusOK, response)
}

func (h *teamHandlers) handleError(w http.ResponseWriter, err error) {
	var appErr entity.AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case entity.ErrorTeamExists:
			writeErrorResponse(w, http.StatusBadRequest, appErr.Code, appErr.Message)
		case entity.ErrorNotFound:
			writeErrorResponse(w, http.StatusNotFound, appErr.Code, appErr.Message)
		default:
			writeErrorResponse(w, http.StatusInternalServerError, entity.ErrorInvalidInput, appErr.Message)
		}
	} else {
		h.logger.Error("Internal server error: %v", err)
		writeErrorResponse(w, http.StatusInternalServerError, entity.ErrorInvalidInput, "internal server error")
	}
}
