package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pr-reviewer-assignment-service/internal/models"
)

const (
	testBaseURL = "http://localhost:8080"
)

// TestE2EWorkflow тестирует полный цикл работы сервиса
func TestE2EWorkflow(t *testing.T) {
	// Пропускаем тест, если не включен E2E режим
	if os.Getenv("E2E_TESTS") != "true" {
		t.Skip("E2E tests are disabled. Set E2E_TESTS=true to run them.")
	}

	t.Run("full service workflow", func(t *testing.T) {
		// Генерируем уникальные имена для теста
		teamName := fmt.Sprintf("e2e_test_team_%d", time.Now().Unix())
		prID := fmt.Sprintf("e2e_pr_%d", time.Now().Unix())
		prID2 := fmt.Sprintf("e2e_pr_%d_2", time.Now().Unix())

		// Шаг 1: Создание команды с 5 участниками
		teamReq := models.Team{
			Name: teamName,
			Members: []models.TeamMember{
				{UserID: "e2e_user_1", Username: "E2E User 1", IsActive: true},
				{UserID: "e2e_user_2", Username: "E2E User 2", IsActive: true},
				{UserID: "e2e_user_3", Username: "E2E User 3", IsActive: true},
				{UserID: "e2e_user_4", Username: "E2E User 4", IsActive: true},
				{UserID: "e2e_user_5", Username: "E2E User 5", IsActive: true},
			},
		}

		resp, err := http.Post(
			fmt.Sprintf("%s/team/add", testBaseURL),
			"application/json",
			toJSON(teamReq),
		)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var createTeamResp struct {
			Team *models.Team `json:"team"`
		}
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		err = json.Unmarshal(body, &createTeamResp)
		require.NoError(t, err)

		assert.Equal(t, teamName, createTeamResp.Team.Name)
		assert.Len(t, createTeamResp.Team.Members, 5)

		// Шаг 2: Получение команды
		getResp, err := http.Get(fmt.Sprintf("%s/team/get?team_name=%s", testBaseURL, teamName))
		require.NoError(t, err)
		defer getResp.Body.Close()

		assert.Equal(t, http.StatusOK, getResp.StatusCode)

		var getTeamResp struct {
			TeamName string              `json:"team_name"`
			Members  []models.TeamMember `json:"members"`
		}
		body, err = io.ReadAll(getResp.Body)
		require.NoError(t, err)
		err = json.Unmarshal(body, &getTeamResp)
		require.NoError(t, err)

		assert.Equal(t, teamName, getTeamResp.TeamName)
		assert.Len(t, getTeamResp.Members, 5)

		// Шаг 3: Создание PR с автором e2e_user_1
		prReq := struct {
			PullRequestID   string `json:"pull_request_id"`
			PullRequestName string `json:"pull_request_name"`
			AuthorID        string `json:"author_id"`
		}{
			PullRequestID:   prID,
			PullRequestName: "E2E Test Pull Request",
			AuthorID:        "e2e_user_1",
		}

		resp, err = http.Post(
			fmt.Sprintf("%s/pullRequest/create", testBaseURL),
			"application/json",
			toJSON(prReq),
		)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var createPRResp struct {
			PR *models.PullRequest `json:"pr"`
		}
		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err)
		err = json.Unmarshal(body, &createPRResp)
		require.NoError(t, err)

		assert.Equal(t, prID, createPRResp.PR.PullRequestID)
		assert.Equal(t, "e2e_user_1", createPRResp.PR.AuthorID)
		assert.Equal(t, models.PullRequestStatusOpen, createPRResp.PR.Status)
		// Должно быть назначено 2 ревьювера из 4 доступных (5 участников минус 1 автор)
		assert.Len(t, createPRResp.PR.AssignedReviewers, 2)

		// Проверяем, что автор не назначен ревьювером
		for _, reviewerID := range createPRResp.PR.AssignedReviewers {
			assert.NotEqual(t, "e2e_user_1", reviewerID, "Автор не должен быть назначен ревьювером")
		}

		// Шаг 4: Переназначение одного из ревьюверов
		oldReviewer := createPRResp.PR.AssignedReviewers[0]
		reassignReq := struct {
			PullRequestID string `json:"pull_request_id"`
			OldUserID     string `json:"old_user_id"`
		}{
			PullRequestID: prID,
			OldUserID:     oldReviewer,
		}

		resp, err = http.Post(
			fmt.Sprintf("%s/pullRequest/reassign", testBaseURL),
			"application/json",
			toJSON(reassignReq),
		)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var reassignResp struct {
			PR         *models.PullRequest `json:"pr"`
			ReplacedBy string              `json:"replaced_by"`
		}
		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err)
		err = json.Unmarshal(body, &reassignResp)
		require.NoError(t, err)

		// Проверяем, что старый ревьювер заменен на нового
		assert.Equal(t, prID, reassignResp.PR.PullRequestID)
		assert.Contains(t, reassignResp.PR.AssignedReviewers, reassignResp.ReplacedBy, "Новый ревьювер должен быть в списке назначенных")
		assert.Len(t, reassignResp.PR.AssignedReviewers, 2)

		// Шаг 5: Слияние PR
		mergeReq := struct {
			PullRequestID string `json:"pull_request_id"`
		}{
			PullRequestID: prID,
		}

		resp, err = http.Post(
			fmt.Sprintf("%s/pullRequest/merge", testBaseURL),
			"application/json",
			toJSON(mergeReq),
		)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var mergeResp struct {
			PR *models.PullRequest `json:"pr"`
		}
		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err)
		err = json.Unmarshal(body, &mergeResp)
		require.NoError(t, err)

		assert.Equal(t, prID, mergeResp.PR.PullRequestID)
		assert.Equal(t, models.PullRequestStatusMerged, mergeResp.PR.Status)
		assert.NotNil(t, mergeResp.PR.MergedAt)

		// Шаг 6: Попытка переназначить ревьювера у слитого PR (должно вернуть ошибку)
		resp, err = http.Post(
			fmt.Sprintf("%s/pullRequest/reassign", testBaseURL),
			"application/json",
			toJSON(struct {
				PullRequestID string `json:"pull_request_id"`
				OldUserID     string `json:"old_user_id"`
			}{
				PullRequestID: prID,
				OldUserID:     oldReviewer, // используем старого ревьювера
			}),
		)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)

		// Шаг 7: Изменение статуса активности пользователя
		setActiveReq := struct {
			UserID   string `json:"user_id"`
			IsActive bool   `json:"is_active"`
		}{
			UserID:   "e2e_user_5",
			IsActive: false,
		}

		resp, err = http.Post(
			fmt.Sprintf("%s/users/setIsActive", testBaseURL),
			"application/json",
			toJSON(setActiveReq),
		)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var setActiveResp struct {
			User *models.User `json:"user"`
		}
		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err)
		err = json.Unmarshal(body, &setActiveResp)
		require.NoError(t, err)

		assert.Equal(t, "e2e_user_5", setActiveResp.User.UserID)
		assert.False(t, setActiveResp.User.IsActive)

		// Шаг 8: Создание нового PR чтобы проверить, что неактивный пользователь не назначается
		prReq2 := struct {
			PullRequestID   string `json:"pull_request_id"`
			PullRequestName string `json:"pull_request_name"`
			AuthorID        string `json:"author_id"`
		}{
			PullRequestID:   prID2,
			PullRequestName: "E2E Test Pull Request 2",
			AuthorID:        "e2e_user_2",
		}

		resp, err = http.Post(
			fmt.Sprintf("%s/pullRequest/create", testBaseURL),
			"application/json",
			toJSON(prReq2),
		)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var createPRResp2 struct {
			PR *models.PullRequest `json:"pr"`
		}
		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err)
		err = json.Unmarshal(body, &createPRResp2)
		require.NoError(t, err)

		assert.Equal(t, prID2, createPRResp2.PR.PullRequestID)
		assert.Equal(t, "e2e_user_2", createPRResp2.PR.AuthorID)

		// Проверяем, что неактивный пользователь (e2e_user_5) не назначен ревьювером
		for _, reviewerID := range createPRResp2.PR.AssignedReviewers {
			assert.NotEqual(t, "e2e_user_5", reviewerID, "Неактивный пользователь не должен быть назначен ревьювером")
		}
	})
}

// toJSON преобразует объект в io.Reader с JSON данными
func toJSON(v interface{}) io.Reader {
	data, _ := json.Marshal(v)
	return bytes.NewReader(data)
}
