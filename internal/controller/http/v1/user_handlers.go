// internal/controller/http/v1/user_handlers.go
package v1

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/PaulLocust/Avito-review/internal/dto"
	"github.com/PaulLocust/Avito-review/internal/entity"
	"github.com/PaulLocust/Avito-review/internal/usecase"
	"github.com/PaulLocust/Avito-review/pkg/logger"
)

type userHandlers struct {
	userUC usecase.UserUseCase
	logger logger.Interface
}

func newUserHandlers(userUC usecase.UserUseCase, l logger.Interface) *userHandlers {
	return &userHandlers{
		userUC: userUC,
		logger: l,
	}
}

// SetIsActive устанавливает флаг активности пользователя
// @Summary Установить флаг активности пользователя
// @Description Изменяет статус активности пользователя
// @Tags Users
// @Accept json
// @Produce json
// @Param request body dto.PostUsersSetIsActiveJSONBody true "Данные пользователя"
// @Success 200 {object} map[string]interface{} "Обновлённый пользователь"
// @Failure 404 {object} dto.ErrorResponse "Пользователь не найден"
// @Router /users/setIsActive [post]
func (h *userHandlers) setIsActive(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("POST /api/v1/users/setIsActive")

	var req dto.PostUsersSetIsActiveJSONBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body: %v", err)
		writeErrorResponse(w, http.StatusBadRequest, entity.ErrorInvalidInput, "invalid request body")
		return
	}

	user, err := h.userUC.SetUserActive(r.Context(), req.UserId, req.IsActive)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Конвертируем в DTO
	response := dto.User{
		UserId:   user.ID,
		Username: user.Username,
		TeamName: user.TeamName,
		IsActive: user.IsActive,
	}

	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"user": response,
	})
}

// GetReviews возвращает PR'ы, где пользователь назначен ревьювером
// @Summary Получить PR'ы, где пользователь назначен ревьювером
// @Description Возвращает список pull requests, назначенных пользователю на ревью
// @Tags Users
// @Accept json
// @Produce json
// @Param user_id query string true "Идентификатор пользователя"
// @Success 200 {object} map[string]interface{} "Список PR'ов пользователя"
// @Router /users/getReview [get]
func (h *userHandlers) getReviews(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("GET /api/v1/users/getReview")

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeErrorResponse(w, http.StatusBadRequest, entity.ErrorInvalidInput, "user_id is required")
		return
	}

	prs, err := h.userUC.GetUserReviews(r.Context(), userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Конвертируем в DTO
	prShorts := make([]dto.PullRequestShort, len(prs))
	for i, pr := range prs {
		prShorts[i] = dto.PullRequestShort{
			PullRequestId:   pr.ID,
			PullRequestName: pr.Name,
			AuthorId:        pr.AuthorID,
			Status:          dto.PullRequestShortStatus(pr.Status),
		}
	}

	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"user_id":        userID,
		"pull_requests": prShorts,
	})
}

func (h *userHandlers) handleError(w http.ResponseWriter, err error) {
	var appErr entity.AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {
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