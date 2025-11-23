package models

import "fmt"

// ErrorCode - тип для кодов ошибок
type ErrorCode string

const (
	// ErrorCodeTeamExists - команда уже существует
	ErrorCodeTeamExists ErrorCode = "TEAM_EXISTS"

	// ErrorCodePRExists - PR уже существует
	ErrorCodePRExists ErrorCode = "PR_EXISTS"

	// ErrorCodePRMerged - PR уже слит, нельзя изменять
	ErrorCodePRMerged ErrorCode = "PR_MERGED"

	// ErrorCodeNotAssigned - пользователь не назначен ревьювером
	ErrorCodeNotAssigned ErrorCode = "NOT_ASSIGNED"

	// ErrorCodeNoCandidate - нет доступных кандидатов для переназначения
	ErrorCodeNoCandidate ErrorCode = "NO_CANDIDATE"

	// ErrorCodeNotFound - ресурс не найден
	ErrorCodeNotFound ErrorCode = "NOT_FOUND"
)

// AppError - ошибка приложения с кодом
type AppError struct {
	Code    ErrorCode `json:"code"`    // Код ошибки
	Message string    `json:"message"` // Сообщение об ошибке
}

// Error возвращает строковое представление ошибки
func (e *AppError) Error() string {
	return fmt.Sprintf("ошибка: %s (код: %s)", e.Message, e.Code)
}

// NewAppError создает новую ошибку приложения
func NewAppError(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}
