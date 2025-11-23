package services

import (
	"context"
	"log/slog"
	"math/rand"
	"time"

	"pr-reviewer-assignment-service/internal/interfaces"
	"pr-reviewer-assignment-service/internal/models"
)

// reviewerService - реализация сервиса назначения ревьюверов
type reviewerService struct {
	storage interfaces.Repository
	cache   interfaces.Cache
	logger  *slog.Logger
}

// NewReviewerService - создает новый экземпляр сервиса
func NewReviewerService(storage interfaces.Repository, cache interfaces.Cache, messenger interface{}, logger *slog.Logger) interfaces.ReviewerService {
	return &reviewerService{
		storage: storage,
		cache:   cache,
		logger:  logger,
	}
}

// CreateTeam - создает новую команду
func (s *reviewerService) CreateTeam(team *models.Team) (*models.Team, error) {
	s.logger.Info("Создание команды", slog.String("team_name", team.Name))

	err := s.storage.CreateTeam(team)
	if err != nil {
		s.logger.Error("Ошибка при создании команды", slog.String("team_name", team.Name), slog.String("error", err.Error()))
		return nil, err
	}

	// Очистим кэш, если используется
	if s.cache != nil {
		ctx := context.Background()
		cacheKey := "team:" + team.Name
		_ = s.cache.Delete(ctx, cacheKey)
	}

	s.logger.Info("Команда успешно создана", slog.String("team_name", team.Name))
	return team, nil
}

// GetTeam - возвращает команду по имени
func (s *reviewerService) GetTeam(teamName string) (*models.Team, error) {
	s.logger.Info("Получение команды", slog.String("team_name", teamName))

	var team *models.Team
	var err error

	// Попробуем получить из кэша, если он используется
	if s.cache != nil {
		ctx := context.Background()
		cacheKey := "team:" + teamName

		if err := s.cache.Get(ctx, cacheKey, &team); err == nil {
			s.logger.Info("Команда получена из кэша", slog.String("team_name", teamName))
			return team, nil
		}
	}

	// Получаем из хранилища
	team, err = s.storage.GetTeam(teamName)
	if err != nil {
		s.logger.Error("Ошибка при получении команды", slog.String("team_name", teamName), slog.String("error", err.Error()))
		return nil, err
	}

	// Сохраним в кэше, если он используется
	if s.cache != nil {
		ctx := context.Background()
		cacheKey := "team:" + teamName
		_ = s.cache.Set(ctx, cacheKey, team, 10*time.Minute) // Кэшируем на 10 минут
	}

	s.logger.Info("Команда успешно получена", slog.String("team_name", teamName))
	return team, nil
}

// SetUserActive - устанавливает флаг активности пользователя
func (s *reviewerService) SetUserActive(userID string, isActive bool) (*models.User, error) {
	s.logger.Info("Изменение статуса активности пользователя", slog.String("user_id", userID), slog.Bool("is_active", isActive))

	user, err := s.storage.GetUser(userID)
	if err != nil {
		s.logger.Error("Ошибка при получении пользователя", slog.String("user_id", userID), slog.String("error", err.Error()))
		return nil, err
	}

	// Обновляем статус активности
	user.IsActive = isActive

	err = s.storage.UpdateUser(user)
	if err != nil {
		s.logger.Error("Ошибка при обновлении пользователя", slog.String("user_id", userID), slog.String("error", err.Error()))
		return nil, err
	}

	// Очистим кэш, если используется
	if s.cache != nil {
		ctx := context.Background()
		cacheKey := "user:" + userID
		_ = s.cache.Delete(ctx, cacheKey)
	}

	s.logger.Info("Статус активности пользователя успешно обновлен", slog.String("user_id", userID), slog.Bool("is_active", isActive))
	return user, nil
}

// GetPRsForReviewer - возвращает список PR для ревьювера
func (s *reviewerService) GetPRsForReviewer(userID string) ([]*models.PullRequestShort, error) {
	s.logger.Info("Получение списка PR для ревьювера", slog.String("user_id", userID))

	prs, err := s.storage.GetPRsForReviewer(userID)
	if err != nil {
		s.logger.Error("Ошибка при получении PR для ревьювера", slog.String("user_id", userID), slog.String("error", err.Error()))
		return nil, err
	}

	prShortList := make([]*models.PullRequestShort, 0, len(prs))
	for _, pr := range prs {
		prShortList = append(prShortList, pr.ToPullRequestShort())
	}

	s.logger.Info("Список PR для ревьювера успешно получен", slog.String("user_id", userID), slog.Int("count", len(prShortList)))
	return prShortList, nil
}

// CreatePullRequest - создает новый Pull Request и автоматически назначает ревьюверов
func (s *reviewerService) CreatePullRequest(ctx context.Context, prID, prName, authorID string) (*models.PullRequest, error) {
	s.logger.Info("Создание нового PR", slog.String("pr_id", prID), slog.String("author_id", authorID))

	// Получаем автора
	author, err := s.storage.GetUser(authorID)
	if err != nil {
		s.logger.Error("Ошибка при получении автора PR", slog.String("pr_id", prID), slog.String("author_id", authorID), slog.String("error", err.Error()))
		return nil, err
	}

	// Получаем команду автора
	users, err := s.storage.GetUsersByTeam(author.TeamName)
	if err != nil {
		s.logger.Error("Ошибка при получении пользователей команды автора", slog.String("pr_id", prID), slog.String("team_name", author.TeamName), slog.String("error", err.Error()))
		return nil, err
	}

	// Создаем PR
	now := time.Now()
	pr := &models.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            models.PullRequestStatusOpen,
		AssignedReviewers: []string{},
		CreatedAt:         &now,
	}

	// Назначаем ревьюверов из команды автора (исключая самого автора)
	reviewers := s.assignReviewers(users, authorID, 2)
	pr.AssignedReviewers = reviewers

	err = s.storage.CreatePullRequest(pr)
	if err != nil {
		s.logger.Error("Ошибка при создении PR", slog.String("pr_id", prID), slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.Info("PR успешно создан с назначенными ревьюверами", slog.String("pr_id", prID), slog.Any("reviewers", reviewers))
	return pr, nil
}

// MergePullRequest - помечает PR как MERGED (идемпотентная операция)
func (s *reviewerService) MergePullRequest(ctx context.Context, prID string) (*models.PullRequest, error) {
	s.logger.Info("Слияние PR", slog.String("pr_id", prID))

	pr, err := s.storage.GetPullRequest(prID)
	if err != nil {
		s.logger.Error("Ошибка при получении PR", slog.String("pr_id", prID), slog.String("error", err.Error()))
		return nil, err
	}

	// Если PR уже слит, возвращаем его без изменений (идемпотентность)
	if pr.Status == models.PullRequestStatusMerged {
		s.logger.Info("PR уже слит", slog.String("pr_id", prID))
		return pr, nil
	}

	// Обновляем статус
	now := time.Now()
	pr.Status = models.PullRequestStatusMerged
	pr.MergedAt = &now

	err = s.storage.UpdatePullRequest(pr)
	if err != nil {
		s.logger.Error("Ошибка при обновлении статуса PR", slog.String("pr_id", prID), slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.Info("PR успешно слит", slog.String("pr_id", prID))
	return pr, nil
}

// ReassignReviewer - переназначает конкретного ревьювера на другого из его команды
func (s *reviewerService) ReassignReviewer(ctx context.Context, prID, oldUserID string) (*models.PullRequest, string, error) {
	s.logger.Info("Переназначение ревьювера", slog.String("pr_id", prID), slog.String("old_user_id", oldUserID))

	pr, err := s.storage.GetPullRequest(prID)
	if err != nil {
		s.logger.Error("Ошибка при получении PR", slog.String("pr_id", prID), slog.String("error", err.Error()))
		return nil, "", err
	}

	// Проверяем, что PR не слит (после слияния ревьюверов менять нельзя)
	if pr.Status == models.PullRequestStatusMerged {
		err := models.NewAppError(models.ErrorCodePRMerged, "Нельзя менять ревьюверов после слияния PR")
		s.logger.Error("Попытка изменить ревьюверов у слитого PR", slog.String("pr_id", prID), slog.String("error", err.Error()))
		return nil, "", err
	}

	// Проверяем, что старый пользователь действительно назначен ревьювером
	reviewerFound := false
	for _, reviewerID := range pr.AssignedReviewers {
		if reviewerID == oldUserID {
			reviewerFound = true
			break
		}
	}

	if !reviewerFound {
		err := models.NewAppError(models.ErrorCodeNotAssigned, "Пользователь не назначен ревьювером для этого PR")
		s.logger.Error("Пользователь не назначен ревьювером", slog.String("pr_id", prID), slog.String("user_id", oldUserID), slog.String("error", err.Error()))
		return nil, "", err
	}

	// Получаем пользователя, которого нужно заменить
	oldUser, err := s.storage.GetUser(oldUserID)
	if err != nil {
		s.logger.Error("Ошибка при получении старого ревьювера", slog.String("user_id", oldUserID), slog.String("error", err.Error()))
		return nil, "", err
	}

	// Получаем пользователей из команды старого ревьювера
	users, err := s.storage.GetUsersByTeam(oldUser.TeamName)
	if err != nil {
		s.logger.Error("Ошибка при получении пользователей из команды ревьювера", slog.String("team_name", oldUser.TeamName), slog.String("error", err.Error()))
		return nil, "", err
	}

	// Ищем кандидата для замены (активного пользователя, не являющегося автором или другим ревьювером)
	var candidate *models.User
	for _, user := range users {
		// Пропускаем неактивных пользователей
		if !user.IsActive {
			continue
		}

		// Пропускаем старого ревьювера
		if user.UserID == oldUserID {
			continue
		}

		// Пропускаем автора PR
		if user.UserID == pr.AuthorID {
			continue
		}

		// Пропускаем уже назначенных ревьюверов
		isAssigned := false
		for _, reviewerID := range pr.AssignedReviewers {
			if user.UserID == reviewerID {
				isAssigned = true
				break
			}
		}
		if isAssigned {
			continue
		}

		// Нашли подходящего кандидата
		candidate = user
		break
	}

	if candidate == nil {
		err := models.NewAppError(models.ErrorCodeNoCandidate, "Нет доступных активных кандидатов для переназначения в команде")
		s.logger.Error("Нет доступных кандидатов для переназначения", slog.String("pr_id", prID), slog.String("team_name", oldUser.TeamName), slog.String("error", err.Error()))
		return nil, "", err
	}

	// Заменяем старого ревьювера на нового
	newReviewerID := candidate.UserID
	newAssignedReviewers := make([]string, len(pr.AssignedReviewers))
	copy(newAssignedReviewers, pr.AssignedReviewers)

	for i, reviewerID := range newAssignedReviewers {
		if reviewerID == oldUserID {
			newAssignedReviewers[i] = newReviewerID
			break
		}
	}

	pr.AssignedReviewers = newAssignedReviewers

	err = s.storage.UpdatePullRequest(pr)
	if err != nil {
		s.logger.Error("Ошибка при обновлении PR после переназначения", slog.String("pr_id", prID), slog.String("error", err.Error()))
		return nil, "", err
	}

	// Отправим сообщение в Kafka о переназначении ревьювера

	s.logger.Info("Ревьювер успешно переназначен", slog.String("pr_id", prID), slog.String("old_user_id", oldUserID), slog.String("new_user_id", newReviewerID))
	return pr, newReviewerID, nil
}

// MassDeactivateUsers - массово деактивирует пользователей команды и при необходимости переназначает их ревью на других активных пользователей
func (s *reviewerService) MassDeactivateUsers(ctx context.Context, teamName string, userIDs []string, withReassignment bool) (*models.MassDeactivateResult, error) {
	s.logger.Info("Массовая деактивация пользователей",
		slog.String("team_name", teamName),
		slog.Any("user_ids", userIDs),
		slog.Bool("with_reassignment", withReassignment))

	result := &models.MassDeactivateResult{
		DeactivatedCount:    0,
		ReassignedCount:     0,
		FailedDeactivations: make(map[string]string),
		FailedReassignments: make(map[string]string),
	}

	// Получаем всех пользователей команды для определения, у кого из них есть открытые PR
	allUsers, err := s.storage.GetUsersByTeam(teamName)
	if err != nil {
		s.logger.Error("Ошибка при получении пользователей команды",
			slog.String("team_name", teamName),
			slog.String("error", err.Error()))
		return nil, err
	}

	// Создаем мапу для быстрого поиска пользователей
	userMap := make(map[string]*models.User)
	activeUsersMap := make(map[string]*models.User)
	activeUsersList := make([]*models.User, 0)

	for _, user := range allUsers {
		userMap[user.UserID] = user
		// Собираем активных пользователей, не входящих в список деактивации
		if user.IsActive && !s.isInList(user.UserID, userIDs) {
			activeUsersMap[user.UserID] = user
			activeUsersList = append(activeUsersList, user)
		}
	}

	// Деактивируем указанных пользователей
	for _, userID := range userIDs {
		user, exists := userMap[userID]
		if !exists {
			result.FailedDeactivations[userID] = "Пользователь не найден в команде"
			continue
		}

		// Обновляем статус активности
		user.IsActive = false

		err := s.storage.UpdateUser(user)
		if err != nil {
			s.logger.Error("Ошибка при деактивации пользователя",
				slog.String("user_id", userID),
				slog.String("error", err.Error()))
			result.FailedDeactivations[userID] = err.Error()
			continue
		}

		// Очистим кэш, если используется
		if s.cache != nil {
			cacheKey := "user:" + userID
			_ = s.cache.Delete(ctx, cacheKey)
		}

		result.DeactivatedCount++
	}

	// Если запрошена переназначаемость, обрабатываем открытые PR
	if withReassignment {
		// Получаем все открытые PR, назначеные на деактивируемых пользователей
		openPRs, err := s.getOpenPRsAssignedToUsers(userIDs)
		if err != nil {
			s.logger.Error("Ошибка при получении открытых PR для переназначения",
				slog.String("team_name", teamName),
				slog.String("error", err.Error()))
			return nil, err
		} else {
			// Для каждого открытого PR обрабатываем переназначение
			for _, pr := range openPRs {
				// Создаем новый список назначенных ревьюверов
				newAssignedReviewers := make([]string, len(pr.AssignedReviewers))
				copy(newAssignedReviewers, pr.AssignedReviewers)

				// Проходим по ревьюверам и переназначаем тех, кто в списке деактивации
				for i, reviewerID := range newAssignedReviewers {
					// Проверяем, входит ли ревьювер в список деактивируемых
					if s.isInList(reviewerID, userIDs) {
						// Находим нового активного кандидата
						newReviewer, err := s.findActiveReviewer(pr, activeUsersList)
						if err != nil {
							s.logger.Error("Не удалось найти нового ревьювера",
								slog.String("pr_id", pr.PullRequestID),
								slog.String("old_reviewer_id", reviewerID),
								slog.String("error", err.Error()))
							result.FailedReassignments[pr.PullRequestID] = err.Error()
							continue
						}

						// Заменяем старого ревьювера на нового
						newAssignedReviewers[i] = newReviewer.UserID
					}
				}

				// Если были изменения в списке ревьюверов, обновляем PR
				if !s.assignedReviewersEqual(pr.AssignedReviewers, newAssignedReviewers) {
					// Создаем копию PR для обновления
					updatedPR := *pr
					updatedPR.AssignedReviewers = newAssignedReviewers

					err = s.storage.UpdatePullRequest(&updatedPR)
					if err != nil {
						s.logger.Error("Ошибка при обновлении PR после переназначения",
							slog.String("pr_id", pr.PullRequestID),
							slog.String("error", err.Error()))
						result.FailedReassignments[pr.PullRequestID] = err.Error()
						continue
					}

					result.ReassignedCount++
				}
			}
		}
	}

	s.logger.Info("Массовая деактивация завершена",
		slog.String("team_name", teamName),
		slog.Int("deactivated_count", result.DeactivatedCount),
		slog.Int("reassigned_count", result.ReassignedCount))

	return result, nil
}

// assignedReviewersEqual проверяет, равны ли два списка назначенных ревьюверов
func (s *reviewerService) assignedReviewersEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// isInList - проверяет, содержится ли элемент в списке
func (s *reviewerService) isInList(userID string, userIDs []string) bool {
	for _, id := range userIDs {
		if id == userID {
			return true
		}
	}
	return false
}

// getOpenPRsAssignedToUsers - возвращает открытые PR, назначенные на указанных пользователей
func (s *reviewerService) getOpenPRsAssignedToUsers(userIDs []string) ([]*models.PullRequest, error) {
	return s.storage.GetOpenPRsAssignedToUsers(userIDs)
}

// findActiveReviewer - находит активного ревьювера для PR из доступных пользователей
func (s *reviewerService) findActiveReviewer(pr *models.PullRequest, availableUsers []*models.User) (*models.User, error) {
	for _, user := range availableUsers {
		// Пропускаем автора PR
		if user.UserID == pr.AuthorID {
			continue
		}

		// Пропускаем уже назначенных ревьюверов
		isAssigned := false
		for _, reviewerID := range pr.AssignedReviewers {
			if user.UserID == reviewerID {
				isAssigned = true
				break
			}
		}
		if isAssigned {
			continue
		}

		// Нашли подходящего кандидата
		return user, nil
	}

	return nil, models.NewAppError(models.ErrorCodeNoCandidate, "Нет доступных активных кандидатов для переназначения")
}

// assignReviewers - назначает ревьюверов из команды (исключая автора)
func (s *reviewerService) assignReviewers(users []*models.User, authorID string, maxReviewers int) []string {
	var candidates []*models.User

	// Отбираем активных пользователей, исключая автора
	for _, user := range users {
		// Проверяем, что пользователь активен и не является автором
		if user.IsActive && user.UserID != authorID {
			candidates = append(candidates, user)
		}
	}

	// Если у нас больше кандидатов, чем нужно, выбираем случайных
	if len(candidates) > maxReviewers {
		// Перемешиваем кандидатов
		rand.Shuffle(len(candidates), func(i, j int) {
			candidates[i], candidates[j] = candidates[j], candidates[i]
		})

		// Берем первые maxReviewers кандидатов
		candidates = candidates[:maxReviewers]
	}

	// Извлекаем ID ревьюверов
	reviewerIDs := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		reviewerIDs = append(reviewerIDs, candidate.UserID)
	}

	return reviewerIDs
}
