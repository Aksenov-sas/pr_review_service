package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"pr-reviewer-assignment-service/internal/interfaces"
	"pr-reviewer-assignment-service/internal/models"
)

// Handler - структура обработчика HTTP-запросов
type Handler struct {
	service interfaces.ReviewerService
	logger  *slog.Logger
}

// NewHandler - создает новый экземпляр обработчика
func NewHandler(service interfaces.ReviewerService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes - регистрирует все маршруты API
func (h *Handler) RegisterRoutes(r chi.Router) {
	// Маршруты для команд
	r.Post("/team/add", h.createTeam)
	r.Get("/team/get", h.getTeam)

	// Маршруты для пользователей
	r.Post("/users/setIsActive", h.setUserActive)
	r.Post("/users/massDeactivate", h.massDeactivateUsers)
	r.Get("/users/getReview", h.getPRsForReviewer)

	// Маршруты для Pull Request
	r.Post("/pullRequest/create", h.createPullRequest)
	r.Post("/pullRequest/merge", h.mergePullRequest)
	r.Post("/pullRequest/reassign", h.reassignReviewer)
}

// errorResponse - вспомогательная функция для формирования ответа об ошибке
func (h *Handler) errorResponse(w http.ResponseWriter, statusCode int, appErr *models.AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]*models.AppError{
		"error": appErr,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Ошибка при кодировании ответа об ошибке", slog.String("error", err.Error()))
	}
}

// Response структура для успешных ответов
type Response struct {
	Data interface{} `json:"data,omitempty"`
}

// jsonResponse - вспомогательная функция для формирования успешного JSON-ответа
func (h *Handler) jsonResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Ошибка при кодировании JSON-ответа", slog.String("error", err.Error()))
	}
}

// createTeam - обработчик создания команды
func (h *Handler) createTeam(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Обработка запроса на создание команды")

	var req models.Team

	// Декодируем тело запроса
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Ошибка при декодировании запроса на создание команды", slog.String("error", err.Error()))
		h.errorResponse(w, http.StatusBadRequest, models.NewAppError(models.ErrorCodeNotFound, "Некорректный формат JSON"))
		return
	}

	// Вызываем сервис для создания команды
	team, err := h.service.CreateTeam(&req)
	if err != nil {
		appErr, ok := err.(*models.AppError)
		if !ok {
			h.logger.Error("Неизвестная ошибка при создании команды", slog.String("error", err.Error()))
			h.errorResponse(w, http.StatusInternalServerError, models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера"))
			return
		}

		// Обработка специфичных ошибок
		statusCode := http.StatusInternalServerError
		if appErr.Code == models.ErrorCodeTeamExists {
			statusCode = http.StatusBadRequest
		}

		h.errorResponse(w, statusCode, appErr)
		return
	}

	// Возвращаем успешный ответ
	h.jsonResponse(w, http.StatusCreated, map[string]*models.Team{
		"team": team,
	})

	h.logger.Info("Команда успешно создана", slog.String("team_name", team.Name))
}

// getTeam - обработчик получения команды
func (h *Handler) getTeam(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Обработка запроса на получение команды")

	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		h.logger.Error("Отсутствует параметр team_name в запросе на получение команды")
		h.errorResponse(w, http.StatusBadRequest, models.NewAppError(models.ErrorCodeNotFound, "Отсутствует параметр team_name"))
		return
	}

	// Вызываем сервис для получения команды
	team, err := h.service.GetTeam(teamName)
	if err != nil {
		appErr, ok := err.(*models.AppError)
		if !ok {
			h.logger.Error("Неизвестная ошибка при получении команды", slog.String("team_name", teamName), slog.String("error", err.Error()))
			h.errorResponse(w, http.StatusInternalServerError, models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера"))
			return
		}

		h.errorResponse(w, http.StatusNotFound, appErr)
		return
	}

	// Возвращаем успешный ответ
	h.jsonResponse(w, http.StatusOK, team)

	h.logger.Info("Команда успешно получена", slog.String("team_name", teamName))
}

// setUserActive - обработчик изменения статуса активности пользователя
func (h *Handler) setUserActive(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Обработка запроса на изменение статуса активности пользователя")

	var req struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}

	// Декодируем тело запроса
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Ошибка при декодировании запроса на изменение статуса активности", slog.String("error", err.Error()))
		h.errorResponse(w, http.StatusBadRequest, models.NewAppError(models.ErrorCodeNotFound, "Некорректный формат JSON"))
		return
	}

	// Проверяем, что предоставлены необходимые данные
	if req.UserID == "" {
		h.logger.Error("Отсутствует user_id в запросе на изменение статуса активности")
		h.errorResponse(w, http.StatusBadRequest, models.NewAppError(models.ErrorCodeNotFound, "Отсутствует user_id"))
		return
	}

	// Вызываем сервис для изменения статуса пользователя
	user, err := h.service.SetUserActive(req.UserID, req.IsActive)
	if err != nil {
		appErr, ok := err.(*models.AppError)
		if !ok {
			h.logger.Error("Неизвестная ошибка при изменении статуса пользователя", slog.String("user_id", req.UserID), slog.String("error", err.Error()))
			h.errorResponse(w, http.StatusInternalServerError, models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера"))
			return
		}

		h.errorResponse(w, http.StatusNotFound, appErr)
		return
	}

	// Возвращаем успешный ответ
	h.jsonResponse(w, http.StatusOK, map[string]*models.User{
		"user": user,
	})

	h.logger.Info("Статус активности пользователя успешно обновлен", slog.String("user_id", req.UserID), slog.Bool("is_active", req.IsActive))
}

// getPRsForReviewer - обработчик получения PR для ревьювера
func (h *Handler) getPRsForReviewer(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Обработка запроса на получение PR для ревьювера")

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		h.logger.Error("Отсутствует параметр user_id в запросе на получение PR для ревьювера")
		h.errorResponse(w, http.StatusBadRequest, models.NewAppError(models.ErrorCodeNotFound, "Отсутствует параметр user_id"))
		return
	}

	// Вызываем сервис для получения PR для ревьювера
	prs, err := h.service.GetPRsForReviewer(userID)
	if err != nil {
		appErr, ok := err.(*models.AppError)
		if !ok {
			h.logger.Error("Неизвестная ошибка при получении PR для ревьювера", slog.String("user_id", userID), slog.String("error", err.Error()))
			h.errorResponse(w, http.StatusInternalServerError, models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера"))
			return
		}

		h.errorResponse(w, http.StatusNotFound, appErr)
		return
	}

	// Возвращаем успешный ответ
	response := map[string]interface{}{
		"user_id":       userID,
		"pull_requests": prs,
	}
	h.jsonResponse(w, http.StatusOK, response)

	h.logger.Info("PR для ревьювера успешно получены", slog.String("user_id", userID), slog.Int("count", len(prs)))
}

// createPullRequest - обработчик создания Pull Request
func (h *Handler) createPullRequest(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Обработка запроса на создание PR")

	var req struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
	}

	// Декодируем тело запроса
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Ошибка при декодировании запроса на создание PR", slog.String("error", err.Error()))
		h.errorResponse(w, http.StatusBadRequest, models.NewAppError(models.ErrorCodeNotFound, "Некорректный формат JSON"))
		return
	}

	// Проверяем, что предоставлены необходимые данные
	if req.PullRequestID == "" {
		h.logger.Error("Отсутствует pull_request_id в запросе на создание PR")
		h.errorResponse(w, http.StatusBadRequest, models.NewAppError(models.ErrorCodeNotFound, "Отсутствует pull_request_id"))
		return
	}

	if req.PullRequestName == "" {
		h.logger.Error("Отсутствует pull_request_name в запросе на создание PR")
		h.errorResponse(w, http.StatusBadRequest, models.NewAppError(models.ErrorCodeNotFound, "Отсутствует pull_request_name"))
		return
	}

	if req.AuthorID == "" {
		h.logger.Error("Отсутствует author_id в запросе на создание PR")
		h.errorResponse(w, http.StatusBadRequest, models.NewAppError(models.ErrorCodeNotFound, "Отсутствует author_id"))
		return
	}

	// Вызываем сервис для создания PR
	pr, err := h.service.CreatePullRequest(r.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		appErr, ok := err.(*models.AppError)
		if !ok {
			h.logger.Error("Неизвестная ошибка при создании PR", slog.String("pr_id", req.PullRequestID), slog.String("error", err.Error()))
			h.errorResponse(w, http.StatusInternalServerError, models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера"))
			return
		}

		// Обработка специфичных ошибок
		statusCode := http.StatusInternalServerError
		if appErr.Code == models.ErrorCodePRExists {
			statusCode = http.StatusConflict
		} else if appErr.Code == models.ErrorCodeNotFound {
			statusCode = http.StatusNotFound
		}

		h.errorResponse(w, statusCode, appErr)
		return
	}

	// Возвращаем успешный ответ
	h.jsonResponse(w, http.StatusCreated, map[string]*models.PullRequest{
		"pr": pr,
	})

	h.logger.Info("PR успешно создан", slog.String("pr_id", pr.PullRequestID), slog.Any("reviewers", pr.AssignedReviewers))
}

// mergePullRequest - обработчик слияния Pull Request
func (h *Handler) mergePullRequest(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Обработка запроса на слияние PR")

	var req struct {
		PullRequestID string `json:"pull_request_id"`
	}

	// Декодируем тело запроса
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Ошибка при декодировании запроса на слияние PR", slog.String("error", err.Error()))
		h.errorResponse(w, http.StatusBadRequest, models.NewAppError(models.ErrorCodeNotFound, "Некорректный формат JSON"))
		return
	}

	// Проверяем, что предоставлены необходимые данные
	if req.PullRequestID == "" {
		h.logger.Error("Отсутствует pull_request_id в запросе на слияние PR")
		h.errorResponse(w, http.StatusBadRequest, models.NewAppError(models.ErrorCodeNotFound, "Отсутствует pull_request_id"))
		return
	}

	// Вызываем сервис для слияния PR
	pr, err := h.service.MergePullRequest(r.Context(), req.PullRequestID)
	if err != nil {
		appErr, ok := err.(*models.AppError)
		if !ok {
			h.logger.Error("Неизвестная ошибка при слиянии PR", slog.String("pr_id", req.PullRequestID), slog.String("error", err.Error()))
			h.errorResponse(w, http.StatusInternalServerError, models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера"))
			return
		}

		h.errorResponse(w, http.StatusNotFound, appErr)
		return
	}

	// Возвращаем успешный ответ
	h.jsonResponse(w, http.StatusOK, map[string]*models.PullRequest{
		"pr": pr,
	})

	h.logger.Info("PR успешно слит", slog.String("pr_id", pr.PullRequestID))
}

// massDeactivateUsers - обработчик массовой деактивации пользователей
func (h *Handler) massDeactivateUsers(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Обработка запроса на массовую деактивацию пользователей")

	var req models.MassDeactivateRequest

	// Декодируем тело запроса
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Ошибка при декодировании запроса на массовую деактивацию пользователей", slog.String("error", err.Error()))
		h.errorResponse(w, http.StatusBadRequest, models.NewAppError(models.ErrorCodeNotFound, "Некорректный формат JSON"))
		return
	}

	// Проверяем, что предоставлены необходимые данные
	if req.TeamName == "" {
		h.logger.Error("Отсутствует team_name в запросе на массовую деактивацию пользователей")
		h.errorResponse(w, http.StatusBadRequest, models.NewAppError(models.ErrorCodeNotFound, "Отсутствует team_name"))
		return
	}

	if len(req.UserIDs) == 0 {
		h.logger.Error("Отсутствуют user_ids в запросе на массовую деактивацию пользователей")
		h.errorResponse(w, http.StatusBadRequest, models.NewAppError(models.ErrorCodeNotFound, "Отсутствуют user_ids"))
		return
	}

	// Вызываем сервис для массовой деактивации пользователей
	result, err := h.service.MassDeactivateUsers(r.Context(), req.TeamName, req.UserIDs, req.WithReassignment)
	if err != nil {
		appErr, ok := err.(*models.AppError)
		if !ok {
			h.logger.Error("Неизвестная ошибка при массовой деактивации пользователей", slog.String("team_name", req.TeamName), slog.Any("user_ids", req.UserIDs), slog.String("error", err.Error()))
			h.errorResponse(w, http.StatusInternalServerError, models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера"))
			return
		}

		h.errorResponse(w, http.StatusNotFound, appErr)
		return
	}

	// Возвращаем успешный ответ
	h.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"deactivated_count":    result.DeactivatedCount,
		"reassigned_count":     result.ReassignedCount,
		"failed_deactivations": result.FailedDeactivations,
		"failed_reassignments": result.FailedReassignments,
	})

	h.logger.Info("Массовая деактивация пользователей завершена",
		slog.String("team_name", req.TeamName),
		slog.Int("deactivated_count", result.DeactivatedCount),
		slog.Bool("with_reassignment", req.WithReassignment),
	)
}

// reassignReviewer - обработчик переназначения ревьювера
func (h *Handler) reassignReviewer(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Обработка запроса на переназначение ревьювера")

	var req struct {
		PullRequestID string `json:"pull_request_id"`
		OldUserID     string `json:"old_user_id"`
	}

	// Декодируем тело запроса
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Ошибка при декодировании запроса на переназначение ревьювера", slog.String("error", err.Error()))
		h.errorResponse(w, http.StatusBadRequest, models.NewAppError(models.ErrorCodeNotFound, "Некорректный формат JSON"))
		return
	}

	// Проверяем, что предоставлены необходимые данные
	if req.PullRequestID == "" {
		h.logger.Error("Отсутствует pull_request_id в запросе на переназначение ревьювера")
		h.errorResponse(w, http.StatusBadRequest, models.NewAppError(models.ErrorCodeNotFound, "Отсутствует pull_request_id"))
		return
	}

	if req.OldUserID == "" {
		h.logger.Error("Отсутствует old_user_id в запросе на переназначение ревьювера")
		h.errorResponse(w, http.StatusBadRequest, models.NewAppError(models.ErrorCodeNotFound, "Отсутствует old_user_id"))
		return
	}

	// Вызываем сервис для переназначения ревьювера
	pr, newUserID, err := h.service.ReassignReviewer(r.Context(), req.PullRequestID, req.OldUserID)
	if err != nil {
		appErr, ok := err.(*models.AppError)
		if !ok {
			h.logger.Error("Неизвестная ошибка при переназначении ревьювера", slog.String("pr_id", req.PullRequestID), slog.String("old_user_id", req.OldUserID), slog.String("error", err.Error()))
			h.errorResponse(w, http.StatusInternalServerError, models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера"))
			return
		}

		// Обработка специфичных ошибок
		statusCode := http.StatusInternalServerError
		if appErr.Code == models.ErrorCodePRMerged || appErr.Code == models.ErrorCodeNotAssigned || appErr.Code == models.ErrorCodeNoCandidate {
			statusCode = http.StatusConflict
		} else if appErr.Code == models.ErrorCodeNotFound {
			statusCode = http.StatusNotFound
		}

		h.errorResponse(w, statusCode, appErr)
		return
	}

	// Возвращаем успешный ответ
	response := map[string]interface{}{
		"pr":          pr,
		"replaced_by": newUserID,
	}
	h.jsonResponse(w, http.StatusOK, response)

	h.logger.Info("Ревьювер успешно переназначен", slog.String("pr_id", req.PullRequestID), slog.String("old_user_id", req.OldUserID), slog.String("new_user_id", newUserID))
}
