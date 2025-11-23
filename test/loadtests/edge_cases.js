import http from 'k6/http';
import { check } from 'k6';

export const options = {
    vus: 3,
    duration: '1m',
};

export default function () {
    // Тест 1: Создание PR без автора
    const missingAuthor = http.post('http://localhost:8080/pullRequest/create', JSON.stringify({
        pull_request_id: 'test-pr-1',
        pull_request_name: 'Test PR',
        // author_id отсутствует
    }), { headers: { 'Content-Type': 'application/json' } });

    check(missingAuthor, {
        'missing author returns 400': (r) => r.status === 400,
    });

    // Тест 2: Создание PR с пустым идентификатором
    const emptyIdPR = http.post('http://localhost:8080/pullRequest/create', JSON.stringify({
        pull_request_id: '',
        pull_request_name: 'Test PR',
        author_id: 'user-1',
    }), { headers: { 'Content-Type': 'application/json' } });

    check(emptyIdPR, {
        'empty PR ID returns 400': (r) => r.status === 400,
    });

    // Тест 3: Мерж несуществующего PR
    const mergeFakePR = http.post('http://localhost:8080/pullRequest/merge', JSON.stringify({
        pull_request_id: 'non-existent-pr-id',
    }), { headers: { 'Content-Type': 'application/json' } });

    check(mergeFakePR, {
        'merge non-existent PR returns 404': (r) => r.status === 404,
    });

    // Тест 4: Переназначение ревьювера у несуществующего PR
    const reassignPayload = JSON.stringify({
        pull_request_id: 'non-existent-pr-id',
        old_user_id: 'reviewer-1',
    });

    const reassignResponse = http.post(
        'http://localhost:8080/pullRequest/reassign',
        reassignPayload,
        { headers: { 'Content-Type': 'application/json' } }
    );

    check(reassignResponse, {
        'reassign from non-existent PR returns error': (r) => r.status === 404,
    });

    // Тест 5: Попытка переназначения ревьювера у слитого PR
    // Сначала создаем и сливаем PR
    const createPRPayload = JSON.stringify({
        pull_request_id: 'test-merged-pr',
        pull_request_name: 'Test Merged PR',
        author_id: 'user-1',
    });

    const createPRRes = http.post('http://localhost:8080/pullRequest/create', createPRPayload,
        { headers: { 'Content-Type': 'application/json' } }
    );

    if (createPRRes.status === 201) {
        // Сливаем PR
        const mergePayload = JSON.stringify({
            pull_request_id: 'test-merged-pr',
        });
        http.post('http://localhost:8080/pullRequest/merge', mergePayload,
            { headers: { 'Content-Type': 'application/json' } }
        );

        // Пытаемся переназначить ревьювера у слитого PR
        const reassignMerged = http.post(
            'http://localhost:8080/pullRequest/reassign',
            JSON.stringify({
                pull_request_id: 'test-merged-pr',
                old_user_id: 'user-2',
            }),
            { headers: { 'Content-Type': 'application/json' } }
        );

        check(reassignMerged, {
            'reassign from merged PR returns conflict': (r) => r.status === 409,
        });
    }
}