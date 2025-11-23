import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomIntBetween, randomItem } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

const BASE_URL = 'http://localhost:8080';

// Подготавливаем тестовые данные
function setupData() {
    console.log('Setting up test data...');

    // Создаем тестовую команду
    const teamRes = http.post(`${BASE_URL}/team/add`, JSON.stringify({
        team_name: `load-test-team-${randomIntBetween(10000, 99999)}`,
        members: [
            {user_id: `user-1-${Date.now()}`, username: 'User 1', is_active: true},
            {user_id: `user-2-${Date.now()}`, username: 'User 2', is_active: true},
            {user_id: `user-3-${Date.now()}`, username: 'User 3', is_active: true},
            {user_id: `user-4-${Date.now()}`, username: 'User 4', is_active: true},
            {user_id: `user-5-${Date.now()}`, username: 'User 5', is_active: true}
        ]
    }), { headers: { 'Content-Type': 'application/json' } });

    const teamData = JSON.parse(teamRes.body);

    return {
        teamName: teamData.team.team_name,
        userIds: ['user-1', 'user-2', 'user-3', 'user-4', 'user-5']
    };
}

export function setup() {
    return setupData();
}

export const options = {
    scenarios: {
        normal_load: {
            executor: 'constant-vus',
            vus: 5,  // 5 RPS как в требованиях
            duration: '3m',
            exec: 'normalLoad',
        },
        spike_load: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '1m', target: 20 },  // Пиковая нагрузка
                { duration: '1m', target: 20 },
                { duration: '1m', target: 5 },   // Возврат к нормальной
            ],
            exec: 'spikeLoad',
            startTime: '3m',  // Начинаем после normal_load
        },
    },
    thresholds: {
        http_req_failed: ['rate<0.005'],      // <0.5% ошибок
        http_req_duration: ['p(95)<250'],     // 95% < 250ms
        http_req_duration: ['p(99)<300'],     // 99% < 300ms (требование)
    },
};

// Нормальная нагрузка
export function normalLoad(data) {
    // 80% - создание PR
    if (Math.random() < 0.8) {
        createPR(data.userIds);
    }
    // 20% - другие операции
    else {
        if (Math.random() < 0.5) {
            getPRsForUser(data.userIds);
        } else {
            getTeamData(data.teamName);
        }
    }

    sleep(randomIntBetween(1, 3));
}

// Пиковая нагрузка
export function spikeLoad(data) {
    createPR(data.userIds);
    sleep(0.5); // Меньше паузы для имитации пика
}

function createPR(userIds) {
    const payload = JSON.stringify({
        pull_request_id: `pr-${randomIntBetween(1000, 9999)}`,
        pull_request_name: `Load Test PR ${randomIntBetween(1000, 9999)}`,
        author_id: randomItem(userIds),
    });

    const res = http.post(`${BASE_URL}/pullRequest/create`, payload, {
        headers: { 'Content-Type': 'application/json' },
        tags: { endpoint: 'create_pr' },
    });

    check(res, {
        'create PR: status 201': (r) => r.status === 201,
        'create PR: duration < 300ms': (r) => r.timings.duration < 300,
    });
}

function getPRsForUser(userIds) {
    const userId = randomItem(userIds);
    const res = http.get(`${BASE_URL}/users/getReview?user_id=${userId}`, {
        tags: { endpoint: 'get_user_prs' },
    });

    check(res, {
        'get user PRs: status 200': (r) => r.status === 200,
    });
}

function getTeamData(teamName) {
    const res = http.get(`${BASE_URL}/team/get?team_name=${teamName}`, {
        tags: { endpoint: 'get_team' },
    });

    check(res, {
        'get team: status 200': (r) => r.status === 200,
    });
}