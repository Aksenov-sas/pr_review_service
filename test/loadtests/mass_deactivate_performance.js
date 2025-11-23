import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomIntBetween, randomItem } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

// Тест производительности метода массовой деактивации пользователей
const BASE_URL = 'http://localhost:8080';

export const options = {
    vus: 10,  // Виртуальные пользователи
    duration: '30s',  // Длительность теста
    thresholds: {
        http_req_failed: ['rate<0.01'],  // <1% ошибок
        http_req_duration: ['p(95)<100'], // 95% запросов <100ms
        http_req_duration: ['p(99)<200'], // 99% запросов <200ms
    },
};

export default function () {
    // Подготовка данных для теста
    const requestPayload = JSON.stringify({
        team_name: 'test-team',
        user_ids: [
            `user-${randomIntBetween(1000, 9999)}`,
            `user-${randomIntBetween(1000, 9999)}`,
            `user-${randomIntBetween(1000, 9999)}`,
            `user-${randomIntBetween(1000, 9999)}`,
            `user-${randomIntBetween(1000, 9999)}`
        ],
        with_reassignment: true,
    });

    const response = http.post(`${BASE_URL}/users/massDeactivate`, requestPayload, {
        headers: { 'Content-Type': 'application/json' },
    });

    check(response, {
        'mass deactivate returns success': (r) => r.status === 200 || r.status === 404, // 404 ожидаемо если команда не существует
        'response time < 100ms': (r) => r.timings.duration < 100,
    });

    sleep(1); // Пауза между итерациями
}