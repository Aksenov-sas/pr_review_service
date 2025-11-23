package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/lib/pq"

	"pr-reviewer-assignment-service/internal/models"
)

// PostgresStorage реализация хранилища на PostgreSQL
type PostgresStorage struct {
	db     *sql.DB
	logger *slog.Logger
	psql   squirrel.StatementBuilderType
}

// NewPostgresStorage создает новое PostgreSQL хранилище
func NewPostgresStorage(dsn string, logger *slog.Logger) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к PostgreSQL: %w", err)
	}

	// Установим максимальное количество соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Проверим соединение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ошибка пинга PostgreSQL: %w", err)
	}

	storage := &PostgresStorage{
		db:     db,
		logger: logger,
		psql:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	return storage, nil
}

// Close закрывает соединение с базой данных
func (s *PostgresStorage) Close() error {
	return s.db.Close()
}

// CreateTeam создает новую команду
func (s *PostgresStorage) CreateTeam(team *models.Team) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Проверим, существует ли уже команда с таким именем
	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM teams WHERE name = $1)", team.Name).Scan(&exists)
	if err != nil {
		s.logger.Error("Ошибка проверки существования команды", slog.String("team_name", team.Name), slog.String("error", err.Error()))
		return models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера")
	}

	if exists {
		return models.NewAppError(models.ErrorCodeTeamExists, "Команда с таким именем уже существует")
	}

	// Начнем транзакцию
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error("Ошибка начала транзакции", slog.String("team_name", team.Name), slog.String("error", err.Error()))
		return models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера")
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Создаем команду
	_, err = tx.ExecContext(ctx, "INSERT INTO teams (name) VALUES ($1)", team.Name)
	if err != nil {
		s.logger.Error("Ошибка создания команды", slog.String("team_name", team.Name), slog.String("error", err.Error()))
		return models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера")
	}

	// Добавляем пользователей команды (или обновляем, если уже существуют)
	for _, member := range team.Members {
		_, err = tx.ExecContext(ctx, "INSERT INTO users (user_id, username, team_name, is_active) VALUES ($1, $2, $3, $4) ON CONFLICT (user_id) DO UPDATE SET team_name = EXCLUDED.team_name, username = EXCLUDED.username, is_active = EXCLUDED.is_active",
			member.UserID, member.Username, team.Name, member.IsActive)
		if err != nil {
			s.logger.Error("Ошибка создания/обновления пользователя", slog.String("user_id", member.UserID), slog.String("error", err.Error()))
			return models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера")
		}
	}

	// Зафиксируем транзакцию
	if err = tx.Commit(); err != nil {
		s.logger.Error("Ошибка фиксации транзакции", slog.String("team_name", team.Name), slog.String("error", err.Error()))
		return models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера")
	}

	s.logger.Info("Команда успешно создана", slog.String("team_name", team.Name))
	return nil
}

// GetTeam возвращает команду по имени
func (s *PostgresStorage) GetTeam(teamName string) (*models.Team, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var team models.Team
	team.Name = teamName

	rows, err := s.db.QueryContext(ctx, "SELECT user_id, username, is_active FROM users WHERE team_name = $1", teamName)
	if err != nil {
		s.logger.Error("Ошибка получения пользователей команды", slog.String("team_name", teamName), slog.String("error", err.Error()))
		return nil, models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера")
	}
	defer func() {
		_ = rows.Close()
	}()

	var members []models.TeamMember
	for rows.Next() {
		var member models.TeamMember
		err := rows.Scan(&member.UserID, &member.Username, &member.IsActive)
		if err != nil {
			s.logger.Error("Ошибка сканирования пользователя", slog.String("team_name", teamName), slog.String("error", err.Error()))
			return nil, models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера")
		}
		members = append(members, member)
	}

	if err = rows.Err(); err != nil {
		s.logger.Error("Ошибка итерации по пользователям", slog.String("team_name", teamName), slog.String("error", err.Error()))
		return nil, models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера")
	}

	if len(members) == 0 {
		return nil, models.NewAppError(models.ErrorCodeNotFound, "Команда не найдена")
	}

	team.Members = members
	return &team, nil
}

// CreateUser создает нового пользователя
func (s *PostgresStorage) CreateUser(user *models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)", user.UserID).Scan(&exists)
	if err != nil {
		s.logger.Error("Ошибка проверки существования пользователя", slog.String("user_id", user.UserID), slog.String("error", err.Error()))
		return models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера")
	}

	if exists {
		return models.NewAppError(models.ErrorCodeNotFound, "Пользователь с таким ID уже существует")
	}

	_, err = s.db.ExecContext(ctx, "INSERT INTO users (user_id, username, team_name, is_active) VALUES ($1, $2, $3, $4)",
		user.UserID, user.Username, user.TeamName, user.IsActive)
	if err != nil {
		s.logger.Error("Ошибка создания пользователя", slog.String("user_id", user.UserID), slog.String("error", err.Error()))
		return models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера")
	}

	return nil
}

// UpdateUser обновляет пользователя
func (s *PostgresStorage) UpdateUser(user *models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := s.db.ExecContext(ctx, "UPDATE users SET username = $1, team_name = $2, is_active = $3 WHERE user_id = $4",
		user.Username, user.TeamName, user.IsActive, user.UserID)
	if err != nil {
		s.logger.Error("Ошибка обновления пользователя", slog.String("user_id", user.UserID), slog.String("error", err.Error()))
		return models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		s.logger.Error("Ошибка получения количества измененных строк", slog.String("user_id", user.UserID), slog.String("error", err.Error()))
		return models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера")
	}

	if rowsAffected == 0 {
		return models.NewAppError(models.ErrorCodeNotFound, "Пользователь не найден")
	}

	return nil
}

// GetUser возвращает пользователя по ID
func (s *PostgresStorage) GetUser(userID string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := s.db.QueryRowContext(ctx, "SELECT user_id, username, team_name, is_active FROM users WHERE user_id = $1", userID).Scan(
		&user.UserID, &user.Username, &user.TeamName, &user.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.NewAppError(models.ErrorCodeNotFound, "Пользователь не найден")
		}
		s.logger.Error("Ошибка получения пользователя", slog.String("user_id", userID), slog.String("error", err.Error()))
		return nil, models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера")
	}

	return &user, nil
}

// GetUsersByTeam возвращает всех пользователей из команды
func (s *PostgresStorage) GetUsersByTeam(teamName string) ([]*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, "SELECT user_id, username, team_name, is_active FROM users WHERE team_name = $1", teamName)
	if err != nil {
		s.logger.Error("Ошибка получения пользователей команды", slog.String("team_name", teamName), slog.String("error", err.Error()))
		return nil, models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера")
	}
	defer func() {
		_ = rows.Close()
	}()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)
		if err != nil {
			s.logger.Error("Ошибка сканирования пользователя", slog.String("team_name", teamName), slog.String("error", err.Error()))
			return nil, models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера")
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		s.logger.Error("Ошибка итерации по пользователям", slog.String("team_name", teamName), slog.String("error", err.Error()))
		return nil, models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера")
	}

	if len(users) == 0 {
		return nil, models.NewAppError(models.ErrorCodeNotFound, "Пользователи команды не найдены")
	}

	return users, nil
}

// GetPRsForReviewer возвращает список PR для ревьювера
func (s *PostgresStorage) GetPRsForReviewer(userID string) ([]*models.PullRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT pull_request_id, pull_request_name, author_id, status, assigned_reviewers, created_at, merged_at
		FROM pull_requests
		WHERE $1 = ANY(assigned_reviewers)
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		s.logger.Error("Ошибка получения PR для ревьювера", slog.String("user_id", userID), slog.String("error", err.Error()))
		return nil, models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера")
	}
	defer func() {
		_ = rows.Close()
	}()

	var prs []*models.PullRequest
	for rows.Next() {
		pr, err := s.scanPullRequest(rows, "user_id", userID)
		if err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}

	if err = rows.Err(); err != nil {
		s.logger.Error("Ошибка итерации по PR", slog.String("user_id", userID), slog.String("error", err.Error()))
		return nil, models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера")
	}

	return prs, nil
}

// CreatePullRequest создает новый Pull Request
func (s *PostgresStorage) CreatePullRequest(pr *models.PullRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Проверим, существует ли уже PR с таким ID
	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)", pr.PullRequestID).Scan(&exists)
	if err != nil {
		s.logger.Error("Ошибка проверки существования PR", slog.String("pr_id", pr.PullRequestID), slog.String("error", err.Error()))
		return models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера")
	}

	if exists {
		return models.NewAppError(models.ErrorCodePRExists, "PR с таким ID уже существует")
	}

	_, err = s.db.ExecContext(ctx,
		"INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, assigned_reviewers, created_at, merged_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		pr.PullRequestID, pr.PullRequestName, pr.AuthorID, pr.Status, pq.Array(pr.AssignedReviewers), pr.CreatedAt, pr.MergedAt)
	if err != nil {
		s.logger.Error("Ошибка создания PR", slog.String("pr_id", pr.PullRequestID), slog.String("error", err.Error()))
		return models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера")
	}

	return nil
}

// UpdatePullRequest обновляет Pull Request
func (s *PostgresStorage) UpdatePullRequest(pr *models.PullRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := s.db.ExecContext(ctx,
		"UPDATE pull_requests SET pull_request_name = $1, author_id = $2, status = $3, assigned_reviewers = $4, created_at = $5, merged_at = $6 WHERE pull_request_id = $7",
		pr.PullRequestName, pr.AuthorID, pr.Status, pq.Array(pr.AssignedReviewers), pr.CreatedAt, pr.MergedAt, pr.PullRequestID)
	if err != nil {
		s.logger.Error("Ошибка обновления PR", slog.String("pr_id", pr.PullRequestID), slog.String("error", err.Error()))
		return models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		s.logger.Error("Ошибка получения количества измененных строк", slog.String("pr_id", pr.PullRequestID), slog.String("error", err.Error()))
		return models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера")
	}

	if rowsAffected == 0 {
		return models.NewAppError(models.ErrorCodeNotFound, "PR не найден")
	}

	return nil
}

// GetPullRequest возвращает Pull Request по ID
func (s *PostgresStorage) GetPullRequest(prID string) (*models.PullRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var pr models.PullRequest
	var assignedReviewers []string
	var createdAt, mergedAt *time.Time

	err := s.db.QueryRowContext(ctx,
		"SELECT pull_request_id, pull_request_name, author_id, status, assigned_reviewers, created_at, merged_at FROM pull_requests WHERE pull_request_id = $1",
		prID).Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, pq.Array(&assignedReviewers), &createdAt, &mergedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.NewAppError(models.ErrorCodeNotFound, "PR не найден")
		}
		s.logger.Error("Ошибка получения PR", slog.String("pr_id", prID), slog.String("error", err.Error()))
		return nil, models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера")
	}

	// Присваиваем массив напрямую
	pr.AssignedReviewers = assignedReviewers

	pr.CreatedAt = createdAt
	pr.MergedAt = mergedAt

	return &pr, nil
}

// Вспомогательный метод для сканирования строк результатов в структуру PullRequest
func (s *PostgresStorage) scanPullRequest(rows *sql.Rows, logField string, logFieldVal interface{}) (*models.PullRequest, error) {
	var pr models.PullRequest
	var assignedReviewers []string
	var createdAt, mergedAt *time.Time

	err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, pq.Array(&assignedReviewers), &createdAt, &mergedAt)
	if err != nil {
		s.logger.Error("Ошибка сканирования PR", slog.Any(logField, logFieldVal), slog.String("error", err.Error()))
		return nil, models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера")
	}

	pr.AssignedReviewers = assignedReviewers
	pr.CreatedAt = createdAt
	pr.MergedAt = mergedAt

	return &pr, nil
}

// GetOpenPRsAssignedToUsers возвращает открытые PR, назначенные на указанных пользователей
func (s *PostgresStorage) GetOpenPRsAssignedToUsers(userIDs []string) ([]*models.PullRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if len(userIDs) == 0 {
		return []*models.PullRequest{}, nil
	}

	// Используем готовый запрос с параметрами PostgreSQL
	query := `SELECT pull_request_id, pull_request_name, author_id, status, assigned_reviewers, created_at, merged_at
		FROM pull_requests
		WHERE status = $1 AND (
			assigned_reviewers && $2::text[]
		)`

	rows, err := s.db.QueryContext(ctx, query, models.PullRequestStatusOpen, pq.Array(userIDs))
	if err != nil {
		s.logger.Error("Ошибка получения открытых PR, назначенных на пользователей",
			slog.Any("user_ids", userIDs),
			slog.String("error", err.Error()))
		return nil, models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера")
	}
	defer func() {
		_ = rows.Close()
	}()

	var prs []*models.PullRequest
	for rows.Next() {
		pr, err := s.scanPullRequest(rows, "user_ids", userIDs)
		if err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}

	if err = rows.Err(); err != nil {
		s.logger.Error("Ошибка итерации по PR", slog.Any("user_ids", userIDs), slog.String("error", err.Error()))
		return nil, models.NewAppError(models.ErrorCodeNotFound, "Внутренняя ошибка сервера")
	}

	return prs, nil
}
