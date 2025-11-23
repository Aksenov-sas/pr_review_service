import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomIntBetween, randomItem } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

// Тестовые данные
const BASE_URL = 'http://localhost:8080';
const TEAMS = ['team-1', 'team-2', 'team-3'];
const USERS = ['user-1', 'user-2', 'user-3', 'user-4', 'user-5'];

export const options = {
    vus: 5,  // Виртуальные пользователи
    duration: '1m',  // Длительность теста
    thresholds: {
        http_req_failed: ['rate<0.01'],  // <1% ошибок
        http_req_duration: ['p(95)<300'], // 95% запросов <300ms
    },
};

export default function () {
    // Подготовка данных для теста
    const teamName = `load_test_team_${randomIntBetween(1000, 9999)}`;

    // Создание команды
    const createTeamPayload = JSON.stringify({
        team_name: teamName,
        members: [
            {user_id: 'user-1', username: 'User 1', is_active: true},
            {user_id: 'user-2', username: 'User 2', is_active: true},
            {user_id: 'user-3', username: 'User 3', is_active: true},
            {user_id: 'user-4', username: 'User 4', is_active: true},
            {user_id: 'user-5', username: 'User 5', is_active: true}
        ]
    });

    const createTeamResponse = http.post(`${BASE_URL}/team/add`, createTeamPayload, {
        headers: { 'Content-Type': 'application/json' },
    });

    check(createTeamResponse, {
        'Team created successfully': (r) => r.status === 201,
    });

    // 1. Создание PR
    const createPRPayload = JSON.stringify({
        pull_request_id: `pr-${randomIntBetween(1000, 9999)}`,
        pull_request_name: `Load Test PR ${randomIntBetween(1, 1000)}`,
        author_id: randomItem(USERS),
    });

    const createPRResponse = http.post(`${BASE_URL}/pullRequest/create`, createPRPayload, {
        headers: { 'Content-Type': 'application/json' },
    });

    check(createPRResponse, {
        'PR created successfully': (r) => r.status === 201,
        'response time < 300ms': (r) => r.timings.duration < 300,
    });

    if (createPRResponse.status === 201) {
        const pr = JSON.parse(createPRResponse.body);
        const prId = pr.pr.pull_request_id;

        // 2. Получение PR для ревьюера (50% случаев)
        if (Math.random() > 0.5) {
            const reviewerId = randomItem(USERS);
            const getReviewerPRs = http.get(`${BASE_URL}/users/getReview?user_id=${reviewerId}`);
            check(getReviewerPRs, {
                'Reviewer PRs retrieved': (r) => r.status === 200,
            });
        }

        // 3. Переназначение ревьювера (25% случаев)
        if (Math.random() > 0.75 && pr.pr.assigned_reviewers && pr.pr.assigned_reviewers.length > 0) {
            const oldReviewer = pr.pr.assigned_reviewers[0];
            const reassignPayload = JSON.stringify({
                pull_request_id: prId,
                old_user_id: oldReviewer,
            });

            const reassignResponse = http.post(`${BASE_URL}/pullRequest/reassign`, reassignPayload, {
                headers: { 'Content-Type': 'application/json' },
            });

            check(reassignResponse, {
                'Reviewer reassigned successfully': (r) => r.status === 200,
            });
        }

        // 4. Слияние PR (25% случаев)
        if (Math.random() > 0.75) {
            const mergePayload = JSON.stringify({
                pull_request_id: prId,
            });

            const mergeResponse = http.post(`${BASE_URL}/pullRequest/merge`, mergePayload, {
                headers: { 'Content-Type': 'application/json' },
            });

            check(mergeResponse, {
                'PR merged successfully': (r) => r.status === 200,
            });
        }
    }

    sleep(1); // Пауза между итерациями
}