package models

import (
	"testing"
	"time"
)

// Тестирование метода ToPullRequestShort
func TestPullRequestToPullRequestShort(t *testing.T) {
	now := time.Now()
	pr := &PullRequest{
		PullRequestID:     "pr-123",
		PullRequestName:   "Test PR",
		AuthorID:          "user-1",
		Status:            PullRequestStatusOpen,
		AssignedReviewers: []string{"reviewer-1", "reviewer-2"},
		CreatedAt:         &now,
		MergedAt:          nil,
	}

	short := pr.ToPullRequestShort()

	// Проверяем, что все поля скопированы корректно
	if short.PullRequestID != pr.PullRequestID {
		t.Errorf("Ожидался PullRequestID %s, получили %s", pr.PullRequestID, short.PullRequestID)
	}

	if short.PullRequestName != pr.PullRequestName {
		t.Errorf("Ожидался PullRequestName %s, получили %s", pr.PullRequestName, short.PullRequestName)
	}

	if short.AuthorID != pr.AuthorID {
		t.Errorf("Ожидался AuthorID %s, получили %s", pr.AuthorID, short.AuthorID)
	}

	if short.Status != pr.Status {
		t.Errorf("Ожидался Status %s, получили %s", pr.Status, short.Status)
	}

	// Убеждаемся, что укороченная версия не содержит поле AssignedReviewers (у нее нет этого поля)
	// Проверяем, что остальные поля присутствуют
	if short.PullRequestID != pr.PullRequestID {
		t.Errorf("Ожидался PullRequestID %s, получили %s", pr.PullRequestID, short.PullRequestID)
	}
	if short.PullRequestName != pr.PullRequestName {
		t.Errorf("Ожидался PullRequestName %s, получили %s", pr.PullRequestName, short.PullRequestName)
	}
	if short.AuthorID != pr.AuthorID {
		t.Errorf("Ожидался AuthorID %s, получили %s", pr.AuthorID, short.AuthorID)
	}
	if short.Status != pr.Status {
		t.Errorf("Ожидался Status %s, получили %s", pr.Status, short.Status)
	}
}

// Тестирование структуры AppError
func TestAppError(t *testing.T) {
	err := NewAppError(ErrorCodeNotFound, "Test error message")

	if err.Code != ErrorCodeNotFound {
		t.Errorf("Ожидался ErrorCode %s, получили %s", string(ErrorCodeNotFound), string(err.Code))
	}

	if err.Message != "Test error message" {
		t.Errorf("Ожидалось сообщение 'Test error message', получили %s", err.Message)
	}

	expectedError := "ошибка: Test error message (код: NOT_FOUND)"
	if err.Error() != expectedError {
		t.Errorf("Ожидался метод Error() вернуть '%s', получили '%s'", expectedError, err.Error())
	}
}

// Тестирование констант статусов Pull Request
func TestPullRequestStatusStringValues(t *testing.T) {
	if string(PullRequestStatusOpen) != "OPEN" {
		t.Errorf("Ожидался PullRequestStatusOpen = 'OPEN', получили %s", PullRequestStatusOpen)
	}

	if string(PullRequestStatusMerged) != "MERGED" {
		t.Errorf("Ожидался PullRequestStatusMerged = 'MERGED', получили %s", PullRequestStatusMerged)
	}
}

// Тестирование структуры User
func TestUserStruct(t *testing.T) {
	user := &User{
		UserID:   "user-123",
		Username: "testuser",
		TeamName: "test-team",
		IsActive: true,
	}

	if user.UserID != "user-123" {
		t.Errorf("Ожидался UserID 'user-123', получили %s", user.UserID)
	}

	if user.Username != "testuser" {
		t.Errorf("Ожидался Username 'testuser', получили %s", user.Username)
	}

	if user.TeamName != "test-team" {
		t.Errorf("Ожидался TeamName 'test-team', получили %s", user.TeamName)
	}

	if !user.IsActive {
		t.Errorf("Ожидался IsActive = true, получили %v", user.IsActive)
	}
}

// Тестирование структуры TeamMember
func TestTeamMemberStruct(t *testing.T) {
	member := &TeamMember{
		UserID:   "user-456",
		Username: "memberuser",
		IsActive: false,
	}

	if member.UserID != "user-456" {
		t.Errorf("Ожидался UserID 'user-456', получили %s", member.UserID)
	}

	if member.Username != "memberuser" {
		t.Errorf("Ожидался Username 'memberuser', получили %s", member.Username)
	}

	if member.IsActive {
		t.Errorf("Ожидался IsActive = false, получили %v", member.IsActive)
	}
}

// Тестирование структуры Team
func TestTeamStruct(t *testing.T) {
	members := []TeamMember{
		{UserID: "user-1", Username: "user1", IsActive: true},
		{UserID: "user-2", Username: "user2", IsActive: false},
	}
	team := &Team{
		Name:    "test-team",
		Members: members,
	}

	if team.Name != "test-team" {
		t.Errorf("Ожидалось имя команды 'test-team', получили %s", team.Name)
	}

	if len(team.Members) != 2 {
		t.Errorf("Ожидалось 2 участника, получили %d", len(team.Members))
	}

	if team.Members[0].UserID != "user-1" {
		t.Errorf("Ожидался первый участник с UserID 'user-1', получили %s", team.Members[0].UserID)
	}
}

// Тестирование структуры PullRequest
func TestPullRequestStruct(t *testing.T) {
	now := time.Now()
	pr := &PullRequest{
		PullRequestID:     "pr-789",
		PullRequestName:   "Another PR",
		AuthorID:          "author-1",
		Status:            PullRequestStatusMerged,
		AssignedReviewers: []string{"reviewer-1"},
		CreatedAt:         &now,
		MergedAt:          &now,
	}

	if pr.PullRequestID != "pr-789" {
		t.Errorf("Ожидался PullRequestID 'pr-789', получили %s", pr.PullRequestID)
	}

	if pr.Status != PullRequestStatusMerged {
		t.Errorf("Ожидался Status 'MERGED', получили %s", pr.Status)
	}

	if len(pr.AssignedReviewers) != 1 {
		t.Errorf("Ожидался 1 ревьювер, получили %d", len(pr.AssignedReviewers))
	}

	if pr.AssignedReviewers[0] != "reviewer-1" {
		t.Errorf("Ожидался первый ревьювер 'reviewer-1', получили %s", pr.AssignedReviewers[0])
	}
}

// Тестирование структуры PullRequestShort
func TestPullRequestShortStruct(t *testing.T) {
	prs := &PullRequestShort{
		PullRequestID:   "prs-101",
		PullRequestName: "Short PR",
		AuthorID:        "author-2",
		Status:          PullRequestStatusOpen,
	}

	if prs.PullRequestID != "prs-101" {
		t.Errorf("Ожидался PullRequestID 'prs-101', получили %s", prs.PullRequestID)
	}

	if prs.Status != PullRequestStatusOpen {
		t.Errorf("Ожидался Status 'OPEN', получили %s", prs.Status)
	}
}

// Тестирование структуры MassDeactivateRequest
func TestMassDeactivateRequestStruct(t *testing.T) {
	req := &MassDeactivateRequest{
		TeamName:         "test-team",
		UserIDs:          []string{"user-1", "user-2"},
		WithReassignment: true,
	}

	if req.TeamName != "test-team" {
		t.Errorf("Ожидалось имя команды 'test-team', получили %s", req.TeamName)
	}

	if len(req.UserIDs) != 2 {
		t.Errorf("Ожидалось 2 ID пользователей, получили %d", len(req.UserIDs))
	}

	if !req.WithReassignment {
		t.Errorf("Ожидался WithReassignment = true, получили %v", req.WithReassignment)
	}
}

// Тестирование структуры MassDeactivateResult
func TestMassDeactivateResultStruct(t *testing.T) {
	result := &MassDeactivateResult{
		DeactivatedCount:    5,
		ReassignedCount:     3,
		FailedDeactivations: map[string]string{"user-1": "error1"},
		FailedReassignments: map[string]string{"pr-1": "error2"},
	}

	if result.DeactivatedCount != 5 {
		t.Errorf("Ожидалось количество деактивированных 5, получили %d", result.DeactivatedCount)
	}

	if result.ReassignedCount != 3 {
		t.Errorf("Ожидалось количество переназначенных 3, получили %d", result.ReassignedCount)
	}

	if len(result.FailedDeactivations) != 1 {
		t.Errorf("Ожидалась 1 ошибка деактивации, получили %d", len(result.FailedDeactivations))
	}

	if len(result.FailedReassignments) != 1 {
		t.Errorf("Ожидалась 1 ошибка переназначения, получили %d", len(result.FailedReassignments))
	}
}