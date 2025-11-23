package postgres

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Masterminds/squirrel"
	"github.com/lib/pq"

	"pr-reviewer-assignment-service/internal/models"
)

// MockLogger - заглушка для логгера в тестах
type MockLogger struct{}

func (m *MockLogger) Enabled(ctx context.Context, level slog.Level) bool {
	return false
}

func (m *MockLogger) Handle(ctx context.Context, record slog.Record) error {
	return nil
}

func (m *MockLogger) WithGroup(name string) slog.Handler {
	return m
}

func (m *MockLogger) WithAttrs(attrs []slog.Attr) slog.Handler {
	return m
}

// Тестирование создания PostgreSQL хранилища
func TestPostgresStorageCreation(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Не удалось создать мок: %v", err)
	}
	defer db.Close()

	logger := slog.New(&MockLogger{})

	// Создаем хранилище с моком
	storage := &PostgresStorage{
		db:     db,
		logger: logger,
		psql:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	// Проверяем, что хранилище создано
	if storage == nil {
		t.Error("Ожидался объект хранилища, получили nil")
	}
}

// Тестирование метода Close
func TestPostgresStorageClose(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Не удалось создать мок: %v", err)
	}

	// Ожидаем вызов Close
	mock.ExpectClose()

	logger := slog.New(&MockLogger{})
	storage := &PostgresStorage{
		db:     db,
		logger: logger,
		psql:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	err = storage.Close()
	if err != nil {
		t.Errorf("Ожидалось успешное закрытие, получили ошибку: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Не все ожидания были выполнены: %v", err)
	}
}

// Тестирование создания команды
func TestPostgresStorageCreateTeam(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Не удалось создать мок: %v", err)
	}
	defer db.Close()

	logger := slog.New(&MockLogger{})
	storage := &PostgresStorage{
		db:     db,
		logger: logger,
		psql:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	// Подготовим команду для теста
	team := &models.Team{
		Name: "test-team",
		Members: []models.TeamMember{
			{UserID: "user1", Username: "username1", IsActive: true},
			{UserID: "user2", Username: "username2", IsActive: false},
		},
	}

	// Ожидаем проверку существования команды
	rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)
	mock.ExpectQuery("SELECT EXISTS").WithArgs("test-team").WillReturnRows(rows)

	// Ожидаем начало транзакции
	mock.ExpectBegin()

	// Ожидаем создание команды
	mock.ExpectExec("INSERT INTO teams").WithArgs("test-team").WillReturnResult(sqlmock.NewResult(1, 1))

	// Ожидаем добавление первого участника
	mock.ExpectExec("INSERT INTO users").WithArgs("user1", "username1", "test-team", true).WillReturnResult(sqlmock.NewResult(1, 1))

	// Ожидаем добавление второго участника
	mock.ExpectExec("INSERT INTO users").WithArgs("user2", "username2", "test-team", false).WillReturnResult(sqlmock.NewResult(1, 1))

	// Ожидаем фиксацию транзакции
	mock.ExpectCommit()

	err = storage.CreateTeam(team)
	if err != nil {
		t.Errorf("Ожидалось успешное создание команды, получили ошибку: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Не все ожидания были выполнены: %v", err)
	}
}

// Тестирование получения команды
func TestPostgresStorageGetTeam(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Не удалось создать мок: %v", err)
	}
	defer db.Close()

	logger := slog.New(&MockLogger{})
	storage := &PostgresStorage{
		db:     db,
		logger: logger,
		psql:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	// Подготовим результаты запроса
	rows := sqlmock.NewRows([]string{"user_id", "username", "is_active"}).
		AddRow("user1", "username1", true).
		AddRow("user2", "username2", false)

	mock.ExpectQuery("SELECT user_id, username, is_active FROM users WHERE team_name =").WithArgs("test-team").WillReturnRows(rows)

	team, err := storage.GetTeam("test-team")
	if err != nil {
		t.Errorf("Ожидалось успешное получение команды, получили ошибку: %v", err)
	}

	if team.Name != "test-team" {
		t.Errorf("Ожидалось имя команды 'test-team', получили %s", team.Name)
	}

	if len(team.Members) != 2 {
		t.Errorf("Ожидалось 2 участника, получили %d", len(team.Members))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Не все ожидания были выполнены: %v", err)
	}
}

// Тестирование создания пользователя
func TestPostgresStorageCreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Не удалось создать мок: %v", err)
	}
	defer db.Close()

	logger := slog.New(&MockLogger{})
	storage := &PostgresStorage{
		db:     db,
		logger: logger,
		psql:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	user := &models.User{
		UserID:   "user1",
		Username: "username1",
		TeamName: "test-team",
		IsActive: true,
	}

	// Ожидаем проверку существования пользователя
	rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)
	mock.ExpectQuery("SELECT EXISTS").WithArgs("user1").WillReturnRows(rows)

	// Ожидаем создание пользователя
	mock.ExpectExec("INSERT INTO users").WithArgs("user1", "username1", "test-team", true).WillReturnResult(sqlmock.NewResult(1, 1))

	err = storage.CreateUser(user)
	if err != nil {
		t.Errorf("Ожидалось успешное создание пользователя, получили ошибку: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Не все ожидания были выполнены: %v", err)
	}
}

// Тестирование обновления пользователя
func TestPostgresStorageUpdateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Не удалось создать мок: %v", err)
	}
	defer db.Close()

	logger := slog.New(&MockLogger{})
	storage := &PostgresStorage{
		db:     db,
		logger: logger,
		psql:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	user := &models.User{
		UserID:   "user1",
		Username: "new-username",
		TeamName: "new-team",
		IsActive: false,
	}

	// Ожидаем обновление пользователя
	mock.ExpectExec("UPDATE users SET username =").WithArgs("new-username", "new-team", false, "user1").WillReturnResult(sqlmock.NewResult(1, 1))

	err = storage.UpdateUser(user)
	if err != nil {
		t.Errorf("Ожидалось успешное обновление пользователя, получили ошибку: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Не все ожидания были выполнены: %v", err)
	}
}

// Тестирование получения пользователя
func TestPostgresStorageGetUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Не удалось создать мок: %v", err)
	}
	defer db.Close()

	logger := slog.New(&MockLogger{})
	storage := &PostgresStorage{
		db:     db,
		logger: logger,
		psql:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	rows := sqlmock.NewRows([]string{"user_id", "username", "team_name", "is_active"}).
		AddRow("user1", "username1", "test-team", true)

	mock.ExpectQuery("SELECT user_id, username, team_name, is_active FROM users WHERE user_id").WithArgs("user1").WillReturnRows(rows)

	user, err := storage.GetUser("user1")
	if err != nil {
		t.Errorf("Ожидалось успешное получение пользователя, получили ошибку: %v", err)
	}

	if user.UserID != "user1" {
		t.Errorf("Ожидался UserID 'user1', получили %s", user.UserID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Не все ожидания были выполнены: %v", err)
	}
}

// Тестирование получения пользователей команды
func TestPostgresStorageGetUsersByTeam(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Не удалось создать мок: %v", err)
	}
	defer db.Close()

	logger := slog.New(&MockLogger{})
	storage := &PostgresStorage{
		db:     db,
		logger: logger,
		psql:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	rows := sqlmock.NewRows([]string{"user_id", "username", "team_name", "is_active"}).
		AddRow("user1", "username1", "test-team", true).
		AddRow("user2", "username2", "test-team", false)

	mock.ExpectQuery("SELECT user_id, username, team_name, is_active FROM users WHERE team_name").WithArgs("test-team").WillReturnRows(rows)

	users, err := storage.GetUsersByTeam("test-team")
	if err != nil {
		t.Errorf("Ожидалось успешное получение пользователей команды, получили ошибку: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Ожидалось 2 пользователя, получили %d", len(users))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Не все ожидания были выполнены: %v", err)
	}
}

// Тестирование получения PR для ревьювера
func TestPostgresStorageGetPRsForReviewer(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Не удалось создать мок: %v", err)
	}
	defer db.Close()

	logger := slog.New(&MockLogger{})
	storage := &PostgresStorage{
		db:     db,
		logger: logger,
		psql:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	now := time.Now()
	assignedReviewers := pq.Array([]string{"user1", "user2"})
	
	rows := sqlmock.NewRows([]string{"pull_request_id", "pull_request_name", "author_id", "status", "assigned_reviewers", "created_at", "merged_at"}).
		AddRow("pr-1", "PR 1", "author1", "OPEN", assignedReviewers, now, nil)

	mock.ExpectQuery("SELECT pull_request_id, pull_request_name, author_id, status, assigned_reviewers, created_at, merged_at FROM pull_requests WHERE").WithArgs("user1").WillReturnRows(rows)

	prs, err := storage.GetPRsForReviewer("user1")
	if err != nil {
		t.Errorf("Ожидалось успешное получение PR, получили ошибку: %v", err)
	}

	if len(prs) != 1 {
		t.Errorf("Ожидался 1 PR, получили %d", len(prs))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Не все ожидания были выполнены: %v", err)
	}
}

// Тестирование создания Pull Request
func TestPostgresStorageCreatePullRequest(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Не удалось создать мок: %v", err)
	}
	defer db.Close()

	logger := slog.New(&MockLogger{})
	storage := &PostgresStorage{
		db:     db,
		logger: logger,
		psql:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	now := time.Now()
	pr := &models.PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Test PR",
		AuthorID:          "author1",
		Status:            models.PullRequestStatusOpen,
		AssignedReviewers: []string{"user1", "user2"},
		CreatedAt:         &now,
		MergedAt:          nil,
	}

	// Ожидаем проверку существования PR
	rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)
	mock.ExpectQuery("SELECT EXISTS").WithArgs("pr-1").WillReturnRows(rows)

	// Ожидаем создание PR
	mock.ExpectExec("INSERT INTO pull_requests").WithArgs("pr-1", "Test PR", "author1", models.PullRequestStatusOpen, pq.Array([]string{"user1", "user2"}), &now, nil).WillReturnResult(sqlmock.NewResult(1, 1))

	err = storage.CreatePullRequest(pr)
	if err != nil {
		t.Errorf("Ожидалось успешное создание PR, получили ошибку: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Не все ожидания были выполнены: %v", err)
	}
}

// Тестирование получения Pull Request
func TestPostgresStorageGetPullRequest(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Не удалось создать мок: %v", err)
	}
	defer db.Close()

	logger := slog.New(&MockLogger{})
	storage := &PostgresStorage{
		db:     db,
		logger: logger,
		psql:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	now := time.Now()
	assignedReviewers := pq.Array([]string{"user1", "user2"})
	
	rows := sqlmock.NewRows([]string{"pull_request_id", "pull_request_name", "author_id", "status", "assigned_reviewers", "created_at", "merged_at"}).
		AddRow("pr-1", "Test PR", "author1", "OPEN", assignedReviewers, now, nil)

	mock.ExpectQuery("SELECT pull_request_id, pull_request_name, author_id, status, assigned_reviewers, created_at, merged_at FROM pull_requests WHERE pull_request_id").WithArgs("pr-1").WillReturnRows(rows)

	pr, err := storage.GetPullRequest("pr-1")
	if err != nil {
		t.Errorf("Ожидалось успешное получение PR, получили ошибку: %v", err)
	}

	if pr.PullRequestID != "pr-1" {
		t.Errorf("Ожидался PullRequestID 'pr-1', получили %s", pr.PullRequestID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Не все ожидания были выполнены: %v", err)
	}
}

// Тестирование получения открытых PR, назначенных на пользователей
func TestPostgresStorageGetOpenPRsAssignedToUsers(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Не удалось создать мок: %v", err)
	}
	defer db.Close()

	logger := slog.New(&MockLogger{})
	storage := &PostgresStorage{
		db:     db,
		logger: logger,
		psql:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	now := time.Now()
	assignedReviewers := pq.Array([]string{"user1", "user2"})
	
	rows := sqlmock.NewRows([]string{"pull_request_id", "pull_request_name", "author_id", "status", "assigned_reviewers", "created_at", "merged_at"}).
		AddRow("pr-1", "Test PR", "author1", "OPEN", assignedReviewers, now, nil)

	mock.ExpectQuery("SELECT pull_request_id, pull_request_name, author_id, status, assigned_reviewers, created_at, merged_at FROM pull_requests WHERE status").WithArgs(models.PullRequestStatusOpen, pq.Array([]string{"user1", "user2"})).WillReturnRows(rows)

	prs, err := storage.GetOpenPRsAssignedToUsers([]string{"user1", "user2"})
	if err != nil {
		t.Errorf("Ожидалось успешное получение открытых PR, получили ошибку: %v", err)
	}

	if len(prs) != 1 {
		t.Errorf("Ожидался 1 PR, получили %d", len(prs))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Не все ожидания были выполнены: %v", err)
	}
}