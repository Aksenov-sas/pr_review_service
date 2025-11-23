package services

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"pr-reviewer-assignment-service/internal/models"
)

// MockRepository - заглушка для репозитория в тестах
type MockRepository struct {
	CreateTeamFunc            func(team *models.Team) error
	GetTeamFunc               func(teamName string) (*models.Team, error)
	CreateUserFunc            func(user *models.User) error
	UpdateUserFunc            func(user *models.User) error
	GetUserFunc               func(userID string) (*models.User, error)
	GetUsersByTeamFunc        func(teamName string) ([]*models.User, error)
	GetPRsForReviewerFunc     func(userID string) ([]*models.PullRequest, error)
	CreatePullRequestFunc     func(pr *models.PullRequest) error
	UpdatePullRequestFunc     func(pr *models.PullRequest) error
	GetPullRequestFunc        func(prID string) (*models.PullRequest, error)
	GetOpenPRsAssignedToUsersFunc func(userIDs []string) ([]*models.PullRequest, error)
	CloseFunc                 func() error
}

func (m *MockRepository) CreateTeam(team *models.Team) error {
	if m.CreateTeamFunc != nil {
		return m.CreateTeamFunc(team)
	}
	return nil
}

func (m *MockRepository) GetTeam(teamName string) (*models.Team, error) {
	if m.GetTeamFunc != nil {
		return m.GetTeamFunc(teamName)
	}
	return nil, nil
}

func (m *MockRepository) CreateUser(user *models.User) error {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(user)
	}
	return nil
}

func (m *MockRepository) UpdateUser(user *models.User) error {
	if m.UpdateUserFunc != nil {
		return m.UpdateUserFunc(user)
	}
	return nil
}

func (m *MockRepository) GetUser(userID string) (*models.User, error) {
	if m.GetUserFunc != nil {
		return m.GetUserFunc(userID)
	}
	return nil, nil
}

func (m *MockRepository) GetUsersByTeam(teamName string) ([]*models.User, error) {
	if m.GetUsersByTeamFunc != nil {
		return m.GetUsersByTeamFunc(teamName)
	}
	return nil, nil
}

func (m *MockRepository) GetPRsForReviewer(userID string) ([]*models.PullRequest, error) {
	if m.GetPRsForReviewerFunc != nil {
		return m.GetPRsForReviewerFunc(userID)
	}
	return nil, nil
}

func (m *MockRepository) CreatePullRequest(pr *models.PullRequest) error {
	if m.CreatePullRequestFunc != nil {
		return m.CreatePullRequestFunc(pr)
	}
	return nil
}

func (m *MockRepository) UpdatePullRequest(pr *models.PullRequest) error {
	if m.UpdatePullRequestFunc != nil {
		return m.UpdatePullRequestFunc(pr)
	}
	return nil
}

func (m *MockRepository) GetPullRequest(prID string) (*models.PullRequest, error) {
	if m.GetPullRequestFunc != nil {
		return m.GetPullRequestFunc(prID)
	}
	return nil, nil
}

func (m *MockRepository) GetOpenPRsAssignedToUsers(userIDs []string) ([]*models.PullRequest, error) {
	if m.GetOpenPRsAssignedToUsersFunc != nil {
		return m.GetOpenPRsAssignedToUsersFunc(userIDs)
	}
	return nil, nil
}

func (m *MockRepository) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// MockCache - заглушка для кэша в тестах
type MockCache struct {
	GetFunc    func(ctx context.Context, key string, dest interface{}) error
	SetFunc    func(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	DeleteFunc func(ctx context.Context, key string) error
	ExistsFunc func(ctx context.Context, key string) (bool, error)
	CloseFunc  func() error
}

func (m *MockCache) Get(ctx context.Context, key string, dest interface{}) error {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, key, dest)
	}
	return nil
}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if m.SetFunc != nil {
		return m.SetFunc(ctx, key, value, expiration)
	}
	return nil
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, key)
	}
	return nil
}

func (m *MockCache) Exists(ctx context.Context, key string) (bool, error) {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(ctx, key)
	}
	return false, nil
}

func (m *MockCache) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// Тестирование создания нового сервиса ревьюверов
func TestNewReviewerService(t *testing.T) {
	mockRepo := &MockRepository{}
	mockCache := &MockCache{}
	logger := slog.New(slog.Default().Handler())

	service := NewReviewerService(mockRepo, mockCache, nil, logger)

	if service == nil {
		t.Error("Ожидался объект сервиса ревьюверов, получили nil")
	}
}

// Тестирование создания команды
func TestReviewerServiceCreateTeam(t *testing.T) {
	mockRepo := &MockRepository{
		CreateTeamFunc: func(team *models.Team) error {
			return nil
		},
	}
	mockCache := &MockCache{
		DeleteFunc: func(ctx context.Context, key string) error {
			return nil
		},
	}
	logger := slog.New(slog.Default().Handler())

	service := NewReviewerService(mockRepo, mockCache, nil, logger)

	team := &models.Team{
		Name: "test-team",
		Members: []models.TeamMember{
			{UserID: "user1", Username: "user1", IsActive: true},
		},
	}

	resultTeam, err := service.CreateTeam(team)
	if err != nil {
		t.Errorf("Ожидалось успешное создание команды, получили ошибку: %v", err)
	}

	if resultTeam.Name != "test-team" {
		t.Errorf("Ожидалось имя команды 'test-team', получили %s", resultTeam.Name)
	}
}

// Тестирование получения команды
func TestReviewerServiceGetTeam(t *testing.T) {
	expectedTeam := &models.Team{
		Name: "test-team",
		Members: []models.TeamMember{
			{UserID: "user1", Username: "user1", IsActive: true},
		},
	}

	mockRepo := &MockRepository{
		GetTeamFunc: func(teamName string) (*models.Team, error) {
			if teamName == "test-team" {
				return expectedTeam, nil
			}
			return nil, models.NewAppError(models.ErrorCodeNotFound, "Команда не найдена")
		},
	}
	mockCache := &MockCache{
		GetFunc: func(ctx context.Context, key string, dest interface{}) error {
			// Имитируем отсутствие в кэше
			return errors.New("key not found")
		},
		SetFunc: func(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
			return nil
		},
	}
	logger := slog.New(slog.Default().Handler())

	service := NewReviewerService(mockRepo, mockCache, nil, logger)

	resultTeam, err := service.GetTeam("test-team")
	if err != nil {
		t.Errorf("Ожидалось успешное получение команды, получили ошибку: %v", err)
	}

	if resultTeam.Name != "test-team" {
		t.Errorf("Ожидалось имя команды 'test-team', получили %s", resultTeam.Name)
	}

	if len(resultTeam.Members) != 1 {
		t.Errorf("Ожидался 1 участник, получили %d", len(resultTeam.Members))
	}
}

// Тестирование установки статуса активности пользователя
func TestReviewerServiceSetUserActive(t *testing.T) {
	user := &models.User{
		UserID:   "user1",
		Username: "username1",
		TeamName: "test-team",
		IsActive: true,
	}

	mockRepo := &MockRepository{
		GetUserFunc: func(userID string) (*models.User, error) {
			if userID == "user1" {
				return user, nil
			}
			return nil, models.NewAppError(models.ErrorCodeNotFound, "Пользователь не найден")
		},
		UpdateUserFunc: func(user *models.User) error {
			return nil
		},
	}
	mockCache := &MockCache{
		DeleteFunc: func(ctx context.Context, key string) error {
			return nil
		},
	}
	logger := slog.New(slog.Default().Handler())

	service := NewReviewerService(mockRepo, mockCache, nil, logger)

	resultUser, err := service.SetUserActive("user1", false)
	if err != nil {
		t.Errorf("Ожидалось успешное изменение статуса активности, получили ошибку: %v", err)
	}

	if resultUser.IsActive {
		t.Errorf("Ожидался статус активности false, получили %v", resultUser.IsActive)
	}
}

// Тестирование получения PR для ревьювера
func TestReviewerServiceGetPRsForReviewer(t *testing.T) {
	mockPRs := []*models.PullRequest{
		{
			PullRequestID:     "pr-1",
			PullRequestName:   "Test PR",
			AuthorID:          "author1",
			Status:            models.PullRequestStatusOpen,
			AssignedReviewers: []string{"reviewer1", "reviewer2"},
		},
	}

	mockRepo := &MockRepository{
		GetPRsForReviewerFunc: func(userID string) ([]*models.PullRequest, error) {
			return mockPRs, nil
		},
	}
	mockCache := &MockCache{}
	logger := slog.New(slog.Default().Handler())

	service := NewReviewerService(mockRepo, mockCache, nil, logger)

	resultPRs, err := service.GetPRsForReviewer("reviewer1")
	if err != nil {
		t.Errorf("Ожидалось успешное получение PR, получили ошибку: %v", err)
	}

	if len(resultPRs) != 1 {
		t.Errorf("Ожидался 1 PR, получили %d", len(resultPRs))
	}

	if resultPRs[0].PullRequestID != "pr-1" {
		t.Errorf("Ожидался PullRequestID 'pr-1', получили %s", resultPRs[0].PullRequestID)
	}
}

// Тестирование создания Pull Request
func TestReviewerServiceCreatePullRequest(t *testing.T) {
	author := &models.User{
		UserID:   "author1",
		Username: "author1",
		TeamName: "test-team",
		IsActive: true,
	}

	members := []*models.User{
		author,
		{UserID: "user1", Username: "user1", TeamName: "test-team", IsActive: true},
		{UserID: "user2", Username: "user2", TeamName: "test-team", IsActive: true},
	}

	mockRepo := &MockRepository{
		GetUserFunc: func(userID string) (*models.User, error) {
			if userID == "author1" {
				return author, nil
			}
			return nil, models.NewAppError(models.ErrorCodeNotFound, "Пользователь не найден")
		},
		GetUsersByTeamFunc: func(teamName string) ([]*models.User, error) {
			if teamName == "test-team" {
				return members, nil
			}
			return nil, models.NewAppError(models.ErrorCodeNotFound, "Пользователи не найдены")
		},
		CreatePullRequestFunc: func(pr *models.PullRequest) error {
			return nil
		},
	}
	mockCache := &MockCache{}
	logger := slog.New(slog.Default().Handler())

	service := NewReviewerService(mockRepo, mockCache, nil, logger)

	ctx := context.Background()
	resultPR, err := service.CreatePullRequest(ctx, "pr-1", "Test PR", "author1")
	if err != nil {
		t.Errorf("Ожидалось успешное создание PR, получили ошибку: %v", err)
	}

	if resultPR.PullRequestID != "pr-1" {
		t.Errorf("Ожидался PullRequestID 'pr-1', получили %s", resultPR.PullRequestID)
	}

	// Проверяем, что ревьюверы были назначены (максимум 2, исключая автора)
	if len(resultPR.AssignedReviewers) > 2 {
		t.Errorf("Ожидалось не более 2 ревьюверов, получили %d", len(resultPR.AssignedReviewers))
	}

	// Проверяем, что автор не включен в список ревьюверов
	for _, reviewerID := range resultPR.AssignedReviewers {
		if reviewerID == "author1" {
			t.Errorf("Автор 'author1' не должен быть в списке ревьюверов")
		}
	}
}

// Тестирование слияния Pull Request
func TestReviewerServiceMergePullRequest(t *testing.T) {
	pr := &models.PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Test PR",
		AuthorID:          "author1",
		Status:            models.PullRequestStatusOpen,
		AssignedReviewers: []string{"reviewer1"},
	}

	mockRepo := &MockRepository{
		GetPullRequestFunc: func(prID string) (*models.PullRequest, error) {
			if prID == "pr-1" {
				return pr, nil
			}
			return nil, models.NewAppError(models.ErrorCodeNotFound, "PR не найден")
		},
		UpdatePullRequestFunc: func(pr *models.PullRequest) error {
			return nil
		},
	}
	mockCache := &MockCache{}
	logger := slog.New(slog.Default().Handler())

	service := NewReviewerService(mockRepo, mockCache, nil, logger)

	ctx := context.Background()
	resultPR, err := service.MergePullRequest(ctx, "pr-1")
	if err != nil {
		t.Errorf("Ожидалось успешное слияние PR, получили ошибку: %v", err)
	}

	if resultPR.Status != models.PullRequestStatusMerged {
		t.Errorf("Ожидался статус MERGED, получили %s", resultPR.Status)
	}
}

// Тестирование повторного слияния (идемпотентность)
func TestReviewerServiceMergePullRequestIdempotent(t *testing.T) {
	pr := &models.PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Test PR",
		AuthorID:          "author1",
		Status:            models.PullRequestStatusMerged,
		AssignedReviewers: []string{"reviewer1"},
	}

	mockRepo := &MockRepository{
		GetPullRequestFunc: func(prID string) (*models.PullRequest, error) {
			if prID == "pr-1" {
				return pr, nil
			}
			return nil, models.NewAppError(models.ErrorCodeNotFound, "PR не найден")
		},
	}
	mockCache := &MockCache{}
	logger := slog.New(slog.Default().Handler())

	service := NewReviewerService(mockRepo, mockCache, nil, logger)

	ctx := context.Background()
	resultPR, err := service.MergePullRequest(ctx, "pr-1")
	if err != nil {
		t.Errorf("Ожидалось успешное выполнение (идемпотентность), получили ошибку: %v", err)
	}

	if resultPR.Status != models.PullRequestStatusMerged {
		t.Errorf("Ожидался статус MERGED, получили %s", resultPR.Status)
	}
}

// Тестирование переназначения ревьювера
func TestReviewerServiceReassignReviewer(t *testing.T) {
	pr := &models.PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Test PR",
		AuthorID:          "author1",
		Status:            models.PullRequestStatusOpen,
		AssignedReviewers: []string{"old-reviewer", "reviewer2"},
	}

	oldReviewer := &models.User{
		UserID:   "old-reviewer",
		Username: "Old Reviewer",
		TeamName: "test-team",
		IsActive: true,
	}

	newReviewer := &models.User{
		UserID:   "new-reviewer",
		Username: "New Reviewer",
		TeamName: "test-team",
		IsActive: true,
	}

	members := []*models.User{
		oldReviewer,
		newReviewer,
		{UserID: "author1", Username: "Author", TeamName: "test-team", IsActive: true},
	}

	mockRepo := &MockRepository{
		GetPullRequestFunc: func(prID string) (*models.PullRequest, error) {
			if prID == "pr-1" {
				return pr, nil
			}
			return nil, models.NewAppError(models.ErrorCodeNotFound, "PR не найден")
		},
		GetUserFunc: func(userID string) (*models.User, error) {
			switch userID {
			case "old-reviewer":
				return oldReviewer, nil
			case "new-reviewer":
				return newReviewer, nil
			default:
				return nil, models.NewAppError(models.ErrorCodeNotFound, "Пользователь не найден")
			}
		},
		GetUsersByTeamFunc: func(teamName string) ([]*models.User, error) {
			if teamName == "test-team" {
				return members, nil
			}
			return nil, models.NewAppError(models.ErrorCodeNotFound, "Пользователи не найдены")
		},
		UpdatePullRequestFunc: func(pr *models.PullRequest) error {
			return nil
		},
	}
	mockCache := &MockCache{}
	logger := slog.New(slog.Default().Handler())

	service := NewReviewerService(mockRepo, mockCache, nil, logger)

	ctx := context.Background()
	resultPR, newUserID, err := service.ReassignReviewer(ctx, "pr-1", "old-reviewer")
	if err != nil {
		t.Errorf("Ожидалось успешное переназначение ревьювера, получили ошибку: %v", err)
	}

	// Проверяем, что старый ревьювер больше не в списке
	oldReviewerFound := false
	for _, reviewer := range resultPR.AssignedReviewers {
		if reviewer == "old-reviewer" {
			oldReviewerFound = true
			break
		}
	}
	if oldReviewerFound {
		t.Errorf("Старый ревьювер 'old-reviewer' все еще в списке ревьюверов")
	}

	// Проверяем, что новый ревьювер в списке
	newReviewerFound := false
	for _, reviewer := range resultPR.AssignedReviewers {
		if reviewer == "new-reviewer" {
			newReviewerFound = true
			break
		}
	}
	if !newReviewerFound {
		t.Errorf("Новый ревьювер 'new-reviewer' не найден в списке ревьюверов")
	}

	if newUserID != "new-reviewer" {
		t.Errorf("Ожидался ID нового ревьювера 'new-reviewer', получили %s", newUserID)
	}
}

// Тестирование массовой деактивации пользователей
func TestReviewerServiceMassDeactivateUsers(t *testing.T) {
	teamMembers := []*models.User{
		{UserID: "user1", Username: "User1", TeamName: "test-team", IsActive: true},
		{UserID: "user2", Username: "User2", TeamName: "test-team", IsActive: true},
		{UserID: "user3", Username: "User3", TeamName: "test-team", IsActive: true},
		{UserID: "author1", Username: "Author", TeamName: "test-team", IsActive: true},
	}

	mockRepo := &MockRepository{
		GetUsersByTeamFunc: func(teamName string) ([]*models.User, error) {
			if teamName == "test-team" {
				return teamMembers, nil
			}
			return nil, models.NewAppError(models.ErrorCodeNotFound, "Пользователи не найдены")
		},
		UpdateUserFunc: func(user *models.User) error {
			return nil
		},
		GetOpenPRsAssignedToUsersFunc: func(userIDs []string) ([]*models.PullRequest, error) {
			// Возвращаем пустой список, чтобы не тестировать переназначение в этом тесте
			return []*models.PullRequest{}, nil
		},
	}
	mockCache := &MockCache{
		DeleteFunc: func(ctx context.Context, key string) error {
			return nil
		},
	}
	logger := slog.New(slog.Default().Handler())

	service := NewReviewerService(mockRepo, mockCache, nil, logger)

	ctx := context.Background()
	result, err := service.MassDeactivateUsers(ctx, "test-team", []string{"user1", "user2"}, false)
	if err != nil {
		t.Errorf("Ожидалось успешное выполнение массовой деактивации, получили ошибку: %v", err)
	}

	if result.DeactivatedCount != 2 {
		t.Errorf("Ожидалось 2 деактивированных пользователя, получили %d", result.DeactivatedCount)
	}
}

// Тестирование метода assignReviewers
func TestReviewerServiceAssignReviewers(t *testing.T) {
	logger := slog.New(slog.Default().Handler())
	service := &reviewerService{logger: logger}

	users := []*models.User{
		{UserID: "author", Username: "Author", IsActive: true},
		{UserID: "user1", Username: "User1", IsActive: true},
		{UserID: "user2", Username: "User2", IsActive: true},
		{UserID: "user3", Username: "User3", IsActive: true},
		{UserID: "inactive-user", Username: "Inactive User", IsActive: false},
	}

	reviewers := service.assignReviewers(users, "author", 2)

	// Проверяем, что ревьюверов не более 2
	if len(reviewers) > 2 {
		t.Errorf("Ожидалось не более 2 ревьюверов, получили %d", len(reviewers))
	}

	// Проверяем, что автор не включен в список ревьюверов
	for _, reviewer := range reviewers {
		if reviewer == "author" {
			t.Errorf("Автор 'author' не должен быть назначен ревьювером")
		}
	}

	// Проверяем, что неактивные пользователи не включены
	for _, reviewer := range reviewers {
		if reviewer == "inactive-user" {
			t.Errorf("Неактивный пользователь 'inactive-user' не должен быть назначен ревьювером")
		}
	}
}

// Тестирование метода isInList
func TestReviewerServiceIsInList(t *testing.T) {
	logger := slog.New(slog.Default().Handler())
	service := &reviewerService{logger: logger}

	userIDs := []string{"user1", "user2", "user3"}

	if !service.isInList("user1", userIDs) {
		t.Errorf("Ожидалось, что 'user1' содержится в списке")
	}

	if service.isInList("user4", userIDs) {
		t.Errorf("Ожидалось, что 'user4' не содержится в списке")
	}
}

// Тестирование метода findActiveReviewer
func TestReviewerServiceFindActiveReviewer(t *testing.T) {
	logger := slog.New(slog.Default().Handler())
	service := &reviewerService{logger: logger}

	pr := &models.PullRequest{
		AuthorID:          "author1",
		AssignedReviewers: []string{"reviewer1"},
	}

	availableUsers := []*models.User{
		{UserID: "author1", Username: "Author", IsActive: true},
		{UserID: "reviewer1", Username: "Reviewer1", IsActive: true},
		{UserID: "reviewer2", Username: "Reviewer2", IsActive: true},
		{UserID: "inactive-user", Username: "Inactive", IsActive: false},
	}

	candidate, err := service.findActiveReviewer(pr, availableUsers)

	if err != nil {
		t.Errorf("Ожидался успешный поиск ревьювера, получили ошибку: %v", err)
	}

	if candidate == nil {
		t.Errorf("Ожидался объект кандидата, получили nil")
	} else if candidate.UserID != "reviewer2" {
		t.Errorf("Ожидался ревьювер 'reviewer2', получили %s", candidate.UserID)
	}
}