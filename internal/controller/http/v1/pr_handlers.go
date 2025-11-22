// internal/controller/http/v1/pr_handlers.go
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

type prHandlers struct {
	prUC   usecase.PRUseCase
	logger logger.Interface
}

func newPRHandlers(prUC usecase.PRUseCase, l logger.Interface) *prHandlers {
	return &prHandlers{
		prUC:   prUC,
		logger: l,
	}
}

// CreatePR создает PR и автоматически назначает ревьюверов
// @Summary Создать PR и автоматически назначить до 2 ревьюверов из команды автора
// @Description Создает новый pull request и автоматически назначает до двух активных ревьюверов из команды автора
// @Tags PullRequests
// @Accept json
// @Produce json
// @Param request body dto.PostPullRequestCreateJSONBody true "Данные PR"
// @Success 201 {object} map[string]interface{} "PR создан"
// @Failure 404 {object} dto.ErrorResponse "Автор/команда не найдены"
// @Failure 409 {object} dto.ErrorResponse "PR уже существует"
// @Router /pullRequest/create [post]
func (h *prHandlers) createPR(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("POST /api/v1/pullRequest/create")

	var req dto.PostPullRequestCreateJSONBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body: %v", err)
		writeErrorResponse(w, http.StatusBadRequest, entity.ErrorInvalidInput, "invalid request body")
		return
	}

	pr, err := h.prUC.CreatePR(r.Context(), req.PullRequestId, req.PullRequestName, req.AuthorId)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Конвертируем в DTO
	response := dto.PullRequest{
		PullRequestId:   pr.ID,
		PullRequestName: pr.Name,
		AuthorId:        pr.AuthorID,
		Status:          dto.PullRequestStatus(pr.Status),
		AssignedReviewers: pr.AssignedReviewers,
		CreatedAt:       &pr.CreatedAt,
		MergedAt:        pr.MergedAt,
	}

	writeJSONResponse(w, http.StatusCreated, map[string]interface{}{
		"pr": response,
	})
}

// MergePR помечает PR как MERGED
// @Summary Пометить PR как MERGED (идемпотентная операция)
// @Description Изменяет статус PR на MERGED. Операция идемпотентна - повторный вызов не приводит к ошибке
// @Tags PullRequests
// @Accept json
// @Produce json
// @Param request body dto.PostPullRequestMergeJSONBody true "Данные PR"
// @Success 200 {object} map[string]interface{} "PR в состоянии MERGED"
// @Failure 404 {object} dto.ErrorResponse "PR не найден"
// @Router /pullRequest/merge [post]
func (h *prHandlers) mergePR(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("POST /api/v1/pullRequest/merge")

	var req dto.PostPullRequestMergeJSONBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body: %v", err)
		writeErrorResponse(w, http.StatusBadRequest, entity.ErrorInvalidInput, "invalid request body")
		return
	}

	pr, err := h.prUC.MergePR(r.Context(), req.PullRequestId)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Конвертируем в DTO
	response := dto.PullRequest{
		PullRequestId:   pr.ID,
		PullRequestName: pr.Name,
		AuthorId:        pr.AuthorID,
		Status:          dto.PullRequestStatus(pr.Status),
		AssignedReviewers: pr.AssignedReviewers,
		CreatedAt:       &pr.CreatedAt,
		MergedAt:        pr.MergedAt,
	}

	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"pr": response,
	})
}

// ReassignReviewer переназначает ревьювера
// @Summary Переназначить конкретного ревьювера на другого из его команды
// @Description Заменяет одного ревьювера на случайного активного участника из команды заменяемого ревьювера
// @Tags PullRequests
// @Accept json
// @Produce json
// @Param request body dto.PostPullRequestReassignJSONBody true "Данные для переназначения"
// @Success 200 {object} map[string]interface{} "Переназначение выполнено"
// @Failure 404 {object} dto.ErrorResponse "PR или пользователь не найден"
// @Failure 409 {object} dto.ErrorResponse "Нарушение доменных правил переназначения"
// @Router /pullRequest/reassign [post]
func (h *prHandlers) reassignReviewer(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("POST /api/v1/pullRequest/reassign")

	var req dto.PostPullRequestReassignJSONBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body: %v", err)
		writeErrorResponse(w, http.StatusBadRequest, entity.ErrorInvalidInput, "invalid request body")
		return
	}

	pr, replacedBy, err := h.prUC.ReassignReviewer(r.Context(), req.PullRequestId, req.OldUserId)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Конвертируем в DTO
	response := dto.PullRequest{
		PullRequestId:   pr.ID,
		PullRequestName: pr.Name,
		AuthorId:        pr.AuthorID,
		Status:          dto.PullRequestStatus(pr.Status),
		AssignedReviewers: pr.AssignedReviewers,
		CreatedAt:       &pr.CreatedAt,
		MergedAt:        pr.MergedAt,
	}

	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"pr":          response,
		"replaced_by": replacedBy,
	})
}

func (h *prHandlers) handleError(w http.ResponseWriter, err error) {
	var appErr entity.AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case entity.ErrorPRExists:
			writeErrorResponse(w, http.StatusConflict, appErr.Code, appErr.Message)
		case entity.ErrorPRMerged:
			writeErrorResponse(w, http.StatusConflict, appErr.Code, appErr.Message)
		case entity.ErrorNotAssigned:
			writeErrorResponse(w, http.StatusConflict, appErr.Code, appErr.Message)
		case entity.ErrorNoCandidate:
			writeErrorResponse(w, http.StatusConflict, appErr.Code, appErr.Message)
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