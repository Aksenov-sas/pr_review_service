package interfaces

import (
	"context"
	"time"

	"pr-reviewer-assignment-service/internal/models"
)

// Repository - интерфейс репозитория для работы с данными
type Repository interface {
	// Team Methods
	CreateTeam(team *models.Team) error
	GetTeam(teamName string) (*models.Team, error)

	// User Methods
	CreateUser(user *models.User) error
	UpdateUser(user *models.User) error
	GetUser(userID string) (*models.User, error)
	GetUsersByTeam(teamName string) ([]*models.User, error)
	GetPRsForReviewer(userID string) ([]*models.PullRequest, error)

	// Pull Request Methods
	CreatePullRequest(pr *models.PullRequest) error
	UpdatePullRequest(pr *models.PullRequest) error
	GetPullRequest(prID string) (*models.PullRequest, error)
	GetOpenPRsAssignedToUsers(userIDs []string) ([]*models.PullRequest, error)

	// Close метод для закрытия соединения
	Close() error
}

// ReviewerService - интерфейс бизнес-логики сервиса назначения ревьюверов
type ReviewerService interface {
	// Team Methods
	CreateTeam(team *models.Team) (*models.Team, error)
	GetTeam(teamName string) (*models.Team, error)

	// User Methods
	SetUserActive(userID string, isActive bool) (*models.User, error)
	GetPRsForReviewer(userID string) ([]*models.PullRequestShort, error)
	MassDeactivateUsers(ctx context.Context, teamName string, userIDs []string, withReassignment bool) (*models.MassDeactivateResult, error)

	// Pull Request Methods
	CreatePullRequest(ctx context.Context, prID, prName, authorID string) (*models.PullRequest, error)
	MergePullRequest(ctx context.Context, prID string) (*models.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldUserID string) (*models.PullRequest, string, error)
}

// Cache - интерфейс кэширования
type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Close() error
}
