package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pr-reviewer-assignment-service/internal/models"
)

// MockService - заглушка сервиса для тестирования обработчиков
type MockService struct {
	CreateTeamFunc            func(team *models.Team) (*models.Team, error)
	GetTeamFunc               func(teamName string) (*models.Team, error)
	SetUserActiveFunc         func(userID string, isActive bool) (*models.User, error)
	GetPRsForReviewerFunc     func(userID string) ([]*models.PullRequestShort, error)
	CreatePullRequestFunc     func(ctx context.Context, prID, prName, authorID string) (*models.PullRequest, error)
	MergePullRequestFunc      func(ctx context.Context, prID string) (*models.PullRequest, error)
	MassDeactivateUsersFunc   func(ctx context.Context, teamName string, userIDs []string, withReassignment bool) (*models.MassDeactivateResult, error)
	ReassignReviewerFunc      func(ctx context.Context, prID, oldUserID string) (*models.PullRequest, string, error)
}

func (m *MockService) CreateTeam(team *models.Team) (*models.Team, error) {
	if m.CreateTeamFunc != nil {
		return m.CreateTeamFunc(team)
	}
	return nil, nil
}

func (m *MockService) GetTeam(teamName string) (*models.Team, error) {
	if m.GetTeamFunc != nil {
		return m.GetTeamFunc(teamName)
	}
	return nil, nil
}

func (m *MockService) SetUserActive(userID string, isActive bool) (*models.User, error) {
	if m.SetUserActiveFunc != nil {
		return m.SetUserActiveFunc(userID, isActive)
	}
	return nil, nil
}

func (m *MockService) GetPRsForReviewer(userID string) ([]*models.PullRequestShort, error) {
	if m.GetPRsForReviewerFunc != nil {
		return m.GetPRsForReviewerFunc(userID)
	}
	return nil, nil
}

func (m *MockService) CreatePullRequest(ctx context.Context, prID, prName, authorID string) (*models.PullRequest, error) {
	if m.CreatePullRequestFunc != nil {
		return m.CreatePullRequestFunc(ctx, prID, prName, authorID)
	}
	return nil, nil
}

func (m *MockService) MergePullRequest(ctx context.Context, prID string) (*models.PullRequest, error) {
	if m.MergePullRequestFunc != nil {
		return m.MergePullRequestFunc(ctx, prID)
	}
	return nil, nil
}

func (m *MockService) MassDeactivateUsers(ctx context.Context, teamName string, userIDs []string, withReassignment bool) (*models.MassDeactivateResult, error) {
	if m.MassDeactivateUsersFunc != nil {
		return m.MassDeactivateUsersFunc(ctx, teamName, userIDs, withReassignment)
	}
	return nil, nil
}

func (m *MockService) ReassignReviewer(ctx context.Context, prID, oldUserID string) (*models.PullRequest, string, error) {
	if m.ReassignReviewerFunc != nil {
		return m.ReassignReviewerFunc(ctx, prID, oldUserID)
	}
	return nil, "", nil
}

// Тестирование создания нового обработчика
func TestNewHandler(t *testing.T) {
	mockService := &MockService{}
	logger := slog.New(slog.Default().Handler())

	handler := NewHandler(mockService, logger)

	if handler == nil {
		t.Error("Ожидался объект обработчика, получили nil")
	}

	if handler.service != mockService {
		t.Error("Сервис не был установлен корректно")
	}

	if handler.logger == nil {
		t.Error("Логгер не был установлен")
	}
}

// Тестирование обработчика создания команды - успешный случай
func TestHandlerCreateTeamSuccess(t *testing.T) {
	expectedTeam := &models.Team{
		Name: "test-team",
		Members: []models.TeamMember{
			{UserID: "user1", Username: "User1", IsActive: true},
		},
	}

	mockService := &MockService{
		CreateTeamFunc: func(team *models.Team) (*models.Team, error) {
			return expectedTeam, nil
		},
	}
	logger := slog.New(slog.Default().Handler())
	handler := NewHandler(mockService, logger)

	// Подготовим тело запроса
	teamData := models.Team{
		Name: "test-team",
		Members: []models.TeamMember{
			{UserID: "user1", Username: "User1", IsActive: true},
		},
	}
	jsonData, _ := json.Marshal(teamData)

	req := httptest.NewRequest("POST", "/team/add", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.createTeam(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Ожидался статус %d, получили %d", http.StatusCreated, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var response map[string]*models.Team
	json.Unmarshal(body, &response)

	if response["team"] == nil {
		t.Error("Ожидался объект команды в ответе")
	} else if response["team"].Name != "test-team" {
		t.Errorf("Ожидалось имя команды 'test-team', получили %s", response["team"].Name)
	}
}

// Тестирование обработчика создания команды - ошибка валидации
func TestHandlerCreateTeamValidationError(t *testing.T) {
	mockService := &MockService{}
	logger := slog.New(slog.Default().Handler())
	handler := NewHandler(mockService, logger)

	// Некорректный JSON
	req := httptest.NewRequest("POST", "/team/add", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.createTeam(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Ожидался статус %d, получили %d", http.StatusBadRequest, resp.StatusCode)
	}
}

// Тестирование обработчика получения команды - успешный случай
func TestHandlerGetTeamSuccess(t *testing.T) {
	expectedTeam := &models.Team{
		Name: "test-team",
		Members: []models.TeamMember{
			{UserID: "user1", Username: "User1", IsActive: true},
		},
	}

	mockService := &MockService{
		GetTeamFunc: func(teamName string) (*models.Team, error) {
			if teamName == "test-team" {
				return expectedTeam, nil
			}
			return nil, models.NewAppError(models.ErrorCodeNotFound, "Команда не найдена")
		},
	}
	logger := slog.New(slog.Default().Handler())
	handler := NewHandler(mockService, logger)

	req := httptest.NewRequest("GET", "/team/get?team_name=test-team", nil)
	w := httptest.NewRecorder()

	handler.getTeam(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Ожидался статус %d, получили %d", http.StatusOK, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var response models.Team
	json.Unmarshal(body, &response)

	if response.Name != "test-team" {
		t.Errorf("Ожидалось имя команды 'test-team', получили %s", response.Name)
	}
}

// Тестирование обработчика получения команды - отсутствует параметр
func TestHandlerGetTeamMissingParam(t *testing.T) {
	mockService := &MockService{}
	logger := slog.New(slog.Default().Handler())
	handler := NewHandler(mockService, logger)

	req := httptest.NewRequest("GET", "/team/get", nil)
	w := httptest.NewRecorder()

	handler.getTeam(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Ожидался статус %d, получили %d", http.StatusBadRequest, resp.StatusCode)
	}
}

// Тестирование обработчика изменения статуса активности - успешный случай
func TestHandlerSetUserActiveSuccess(t *testing.T) {
	expectedUser := &models.User{
		UserID:   "user1",
		Username: "User1",
		TeamName: "test-team",
		IsActive: false,
	}

	mockService := &MockService{
		SetUserActiveFunc: func(userID string, isActive bool) (*models.User, error) {
			return expectedUser, nil
		},
	}
	logger := slog.New(slog.Default().Handler())
	handler := NewHandler(mockService, logger)

	requestData := struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}{
		UserID:   "user1",
		IsActive: false,
	}
	jsonData, _ := json.Marshal(requestData)

	req := httptest.NewRequest("POST", "/users/setIsActive", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.setUserActive(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Ожидался статус %d, получили %d", http.StatusOK, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var response map[string]*models.User
	json.Unmarshal(body, &response)

	if response["user"] == nil {
		t.Error("Ожидался объект пользователя в ответе")
	} else if response["user"].UserID != "user1" {
		t.Errorf("Ожидался UserID 'user1', получили %s", response["user"].UserID)
	}
}

// Тестирование обработчика изменения статуса активности - ошибка валидации
func TestHandlerSetUserActiveValidationError(t *testing.T) {
	mockService := &MockService{}
	logger := slog.New(slog.Default().Handler())
	handler := NewHandler(mockService, logger)

	// Отсутствует UserID
	requestData := struct {
		IsActive bool `json:"is_active"`
	}{
		IsActive: true,
	}
	jsonData, _ := json.Marshal(requestData)

	req := httptest.NewRequest("POST", "/users/setIsActive", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.setUserActive(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Ожидался статус %d, получили %d", http.StatusBadRequest, resp.StatusCode)
	}
}

// Тестирование обработчика получения PR для ревьювера - успешный случай
func TestHandlerGetPRsForReviewerSuccess(t *testing.T) {
	expectedPRs := []*models.PullRequestShort{
		{
			PullRequestID:   "pr-1",
			PullRequestName: "Test PR 1",
			AuthorID:        "author1",
			Status:          models.PullRequestStatusOpen,
		},
	}

	mockService := &MockService{
		GetPRsForReviewerFunc: func(userID string) ([]*models.PullRequestShort, error) {
			return expectedPRs, nil
		},
	}
	logger := slog.New(slog.Default().Handler())
	handler := NewHandler(mockService, logger)

	req := httptest.NewRequest("GET", "/users/getReview?user_id=user1", nil)
	w := httptest.NewRecorder()

	handler.getPRsForReviewer(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Ожидался статус %d, получили %d", http.StatusOK, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var response map[string]interface{}
	json.Unmarshal(body, &response)

	if response["user_id"] != "user1" {
		t.Errorf("Ожидался user_id 'user1', получили %v", response["user_id"])
	}

	prs, ok := response["pull_requests"].([]interface{})
	if !ok {
		t.Error("Ожидался массив pull_requests в ответе")
	} else if len(prs) != 1 {
		t.Errorf("Ожидался 1 PR, получили %d", len(prs))
	}
}

// Тестирование обработчика получения PR для ревьювера - отсутствует параметр
func TestHandlerGetPRsForReviewerMissingParam(t *testing.T) {
	mockService := &MockService{}
	logger := slog.New(slog.Default().Handler())
	handler := NewHandler(mockService, logger)

	req := httptest.NewRequest("GET", "/users/getReview", nil)
	w := httptest.NewRecorder()

	handler.getPRsForReviewer(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Ожидался статус %d, получили %d", http.StatusBadRequest, resp.StatusCode)
	}
}

// Тестирование обработчика создания PR - успешный случай
func TestHandlerCreatePullRequestSuccess(t *testing.T) {
	expectedPR := &models.PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Test PR",
		AuthorID:          "author1",
		Status:            models.PullRequestStatusOpen,
		AssignedReviewers: []string{"reviewer1", "reviewer2"},
	}

	mockService := &MockService{
		CreatePullRequestFunc: func(ctx context.Context, prID, prName, authorID string) (*models.PullRequest, error) {
			return expectedPR, nil
		},
	}
	logger := slog.New(slog.Default().Handler())
	handler := NewHandler(mockService, logger)

	requestData := struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
	}{
		PullRequestID:   "pr-1",
		PullRequestName: "Test PR",
		AuthorID:        "author1",
	}
	jsonData, _ := json.Marshal(requestData)

	req := httptest.NewRequest("POST", "/pullRequest/create", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.createPullRequest(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Ожидался статус %d, получили %d", http.StatusCreated, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var response map[string]*models.PullRequest
	json.Unmarshal(body, &response)

	if response["pr"] == nil {
		t.Error("Ожидался объект PR в ответе")
	} else if response["pr"].PullRequestID != "pr-1" {
		t.Errorf("Ожидался PullRequestID 'pr-1', получили %s", response["pr"].PullRequestID)
	}
}

// Тестирование обработчика создания PR - отсутствует параметр
func TestHandlerCreatePullRequestMissingParam(t *testing.T) {
	mockService := &MockService{}
	logger := slog.New(slog.Default().Handler())
	handler := NewHandler(mockService, logger)

	requestData := struct {
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
	}{
		PullRequestName: "Test PR",
		AuthorID:        "author1",
	}
	jsonData, _ := json.Marshal(requestData)

	req := httptest.NewRequest("POST", "/pullRequest/create", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.createPullRequest(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Ожидался статус %d, получили %d", http.StatusBadRequest, resp.StatusCode)
	}
}

// Тестирование обработчика слияния PR - успешный случай
func TestHandlerMergePullRequestSuccess(t *testing.T) {
	expectedPR := &models.PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Test PR",
		AuthorID:          "author1",
		Status:            models.PullRequestStatusMerged,
		AssignedReviewers: []string{"reviewer1"},
	}

	mockService := &MockService{
		MergePullRequestFunc: func(ctx context.Context, prID string) (*models.PullRequest, error) {
			return expectedPR, nil
		},
	}
	logger := slog.New(slog.Default().Handler())
	handler := NewHandler(mockService, logger)

	requestData := struct {
		PullRequestID string `json:"pull_request_id"`
	}{
		PullRequestID: "pr-1",
	}
	jsonData, _ := json.Marshal(requestData)

	req := httptest.NewRequest("POST", "/pullRequest/merge", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.mergePullRequest(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Ожидался статус %d, получили %d", http.StatusOK, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var response map[string]*models.PullRequest
	json.Unmarshal(body, &response)

	if response["pr"] == nil {
		t.Error("Ожидался объект PR в ответе")
	} else if response["pr"].Status != models.PullRequestStatusMerged {
		t.Errorf("Ожидался статус 'MERGED', получили %s", response["pr"].Status)
	}
}

// Тестирование обработчика массовой деактивации - успешный случай
func TestHandlerMassDeactivateUsersSuccess(t *testing.T) {
	expectedResult := &models.MassDeactivateResult{
		DeactivatedCount:    2,
		ReassignedCount:     1,
		FailedDeactivations: make(map[string]string),
		FailedReassignments: make(map[string]string),
	}

	mockService := &MockService{
		MassDeactivateUsersFunc: func(ctx context.Context, teamName string, userIDs []string, withReassignment bool) (*models.MassDeactivateResult, error) {
			return expectedResult, nil
		},
	}
	logger := slog.New(slog.Default().Handler())
	handler := NewHandler(mockService, logger)

	requestData := models.MassDeactivateRequest{
		TeamName:         "test-team",
		UserIDs:          []string{"user1", "user2"},
		WithReassignment: true,
	}
	jsonData, _ := json.Marshal(requestData)

	req := httptest.NewRequest("POST", "/users/massDeactivate", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.massDeactivateUsers(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Ожидался статус %d, получили %d", http.StatusOK, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var response map[string]interface{}
	json.Unmarshal(body, &response)

	if response["deactivated_count"] != 2.0 {
		t.Errorf("Ожидалось deactivated_count 2, получили %v", response["deactivated_count"])
	}

	if response["reassigned_count"] != 1.0 {
		t.Errorf("Ожидалось reassigned_count 1, получили %v", response["reassigned_count"])
	}
}

// Тестирование обработчика переназначения ревьювера - успешный случай
func TestHandlerReassignReviewerSuccess(t *testing.T) {
	expectedPR := &models.PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Test PR",
		AuthorID:          "author1",
		Status:            models.PullRequestStatusOpen,
		AssignedReviewers: []string{"new-reviewer", "reviewer2"},
	}

	mockService := &MockService{
		ReassignReviewerFunc: func(ctx context.Context, prID, oldUserID string) (*models.PullRequest, string, error) {
			return expectedPR, "new-reviewer", nil
		},
	}
	logger := slog.New(slog.Default().Handler())
	handler := NewHandler(mockService, logger)

	requestData := struct {
		PullRequestID string `json:"pull_request_id"`
		OldUserID     string `json:"old_user_id"`
	}{
		PullRequestID: "pr-1",
		OldUserID:     "old-reviewer",
	}
	jsonData, _ := json.Marshal(requestData)

	req := httptest.NewRequest("POST", "/pullRequest/reassign", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.reassignReviewer(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Ожидался статус %d, получили %d", http.StatusOK, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var response map[string]interface{}
	json.Unmarshal(body, &response)

	if response["replaced_by"] != "new-reviewer" {
		t.Errorf("Ожидался replaced_by 'new-reviewer', получили %s", response["replaced_by"])
	}
}

// Тестирование вспомогательной функции errorResponse
func TestHandlerErrorResponse(t *testing.T) {
	mockService := &MockService{}
	logger := slog.New(slog.Default().Handler())
	handler := NewHandler(mockService, logger)

	appErr := models.NewAppError(models.ErrorCodeNotFound, "Тестовая ошибка")

	w := httptest.NewRecorder()
	handler.errorResponse(w, http.StatusNotFound, appErr)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Ожидался статус %d, получили %d", http.StatusNotFound, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var response map[string]*models.AppError
	json.Unmarshal(body, &response)

	if response["error"] == nil {
		t.Error("Ожидался объект ошибки в ответе")
	} else if response["error"].Message != "Тестовая ошибка" {
		t.Errorf("Ожидалось сообщение ошибки 'Тестовая ошибка', получили %s", response["error"].Message)
	}
}

// Тестирование обработки ошибки в сервисе
func TestHandlerServiceError(t *testing.T) {
	appError := models.NewAppError(models.ErrorCodeNotFound, "Пользователь не найден")

	mockService := &MockService{
		GetTeamFunc: func(teamName string) (*models.Team, error) {
			return nil, appError
		},
	}
	logger := slog.New(slog.Default().Handler())
	handler := NewHandler(mockService, logger)

	req := httptest.NewRequest("GET", "/team/get?team_name=nonexistent", nil)
	w := httptest.NewRecorder()

	handler.getTeam(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Ожидался статус %d, получили %d", http.StatusNotFound, resp.StatusCode)
	}
}

// Тестирование неизвестной ошибки в сервисе
func TestHandlerUnknownServiceError(t *testing.T) {
	unknownError := errors.New("неизвестная ошибка")

	mockService := &MockService{
		GetTeamFunc: func(teamName string) (*models.Team, error) {
			return nil, unknownError
		},
	}
	logger := slog.New(slog.Default().Handler())
	handler := NewHandler(mockService, logger)

	req := httptest.NewRequest("GET", "/team/get?team_name=any", nil)
	w := httptest.NewRecorder()

	handler.getTeam(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Ожидался статус %d, получили %d", http.StatusInternalServerError, resp.StatusCode)
	}
}