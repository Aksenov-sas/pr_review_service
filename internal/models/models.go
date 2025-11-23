package models

import "time"

// User - пользователь (участник команды)
type User struct {
	UserID   string `json:"user_id"`   // Уникальный идентификатор пользователя
	Username string `json:"username"`  // Имя пользователя
	TeamName string `json:"team_name"` // Название команды пользователя
	IsActive bool   `json:"is_active"` // Флаг активности пользователя
}

// TeamMember - участник команды
type TeamMember struct {
	UserID   string `json:"user_id"`   // Уникальный идентификатор пользователя
	Username string `json:"username"`  // Имя пользователя
	IsActive bool   `json:"is_active"` // Флаг активности пользователя
}

// Team - команда с участниками
type Team struct {
	Name    string       `json:"team_name"` // Уникальное имя команды
	Members []TeamMember `json:"members"`   // Участники команды
}

// PullRequestStatus - статус Pull Request'а
type PullRequestStatus string

const (
	// PullRequestStatusOpen - статус "открыт"
	PullRequestStatusOpen PullRequestStatus = "OPEN"

	// PullRequestStatusMerged - статус "слит"
	PullRequestStatusMerged PullRequestStatus = "MERGED"
)

// PullRequest - сущность Pull Request
type PullRequest struct {
	PullRequestID     string            `json:"pull_request_id"`     // Уникальный идентификатор PR
	PullRequestName   string            `json:"pull_request_name"`   // Название PR
	AuthorID          string            `json:"author_id"`           // Идентификатор автора PR
	Status            PullRequestStatus `json:"status"`              // Статус PR (OPEN/MERGED)
	AssignedReviewers []string          `json:"assigned_reviewers"`  // Назначенные ревьюверы (до 2)
	CreatedAt         *time.Time        `json:"createdAt,omitempty"` // Дата создания PR
	MergedAt          *time.Time        `json:"mergedAt,omitempty"`  // Дата слияния PR (только для MERGED)
}

// PullRequestShort - укороченная информация о Pull Request
type PullRequestShort struct {
	PullRequestID   string            `json:"pull_request_id"`   // Уникальный идентификатор PR
	PullRequestName string            `json:"pull_request_name"` // Название PR
	AuthorID        string            `json:"author_id"`         // Идентификатор автора PR
	Status          PullRequestStatus `json:"status"`            // Статус PR (OPEN/MERGED)
}

// ToPullRequestShort - преобразует PullRequest в PullRequestShort
func (pr *PullRequest) ToPullRequestShort() *PullRequestShort {
	return &PullRequestShort{
		PullRequestID:   pr.PullRequestID,
		PullRequestName: pr.PullRequestName,
		AuthorID:        pr.AuthorID,
		Status:          pr.Status,
	}
}

// MassDeactivateRequest - структура запроса для массовой деактивации пользователей
type MassDeactivateRequest struct {
	TeamName         string   `json:"team_name"`         // Имя команды
	UserIDs          []string `json:"user_ids"`          // Список ID пользователей для деактивации
	WithReassignment bool     `json:"with_reassignment"` // Флаг необходимости переназначения PR
}

// MassDeactivateResult - структура результата массовой деактивации пользователей
type MassDeactivateResult struct {
	DeactivatedCount    int               `json:"deactivated_count"`    // Количество деактивированных пользователей
	ReassignedCount     int               `json:"reassigned_count"`     // Количество переназначенных PR
	FailedDeactivations map[string]string `json:"failed_deactivations"` // Словарь с ошибками деактивации (user_id -> error)
	FailedReassignments map[string]string `json:"failed_reassignments"` // Словарь с ошибками переназначения (pr_id -> error)
}
