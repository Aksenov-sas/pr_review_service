package cache

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

// Тестирование создания нового Redis кэша
func TestNewRedisCache(t *testing.T) {
	// Создаем mock Redis сервер
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Не удалось запустить mock Redis: %v", err)
	}
	defer mr.Close()

	logger := slog.New(slog.Default().Handler())

	cache, err := NewRedisCache("redis://"+mr.Addr(), logger)
	if err != nil {
		t.Errorf("Ожидалось успешное создание Redis кэша, получили ошибку: %v", err)
	}

	if cache == nil {
		t.Error("Ожидался объект кэша, получили nil")
	}

	// Проверяем, что клиент подключен
	err = cache.Close()
	if err != nil {
		t.Errorf("Ожидалось успешное закрытие кэша, получили ошибку: %v", err)
	}
}

// Тестирование установки значения в кэш
func TestRedisCacheSet(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Не удалось запустить mock Redis: %v", err)
	}
	defer mr.Close()

	logger := slog.New(slog.Default().Handler())
	cache, err := NewRedisCache("redis://"+mr.Addr(), logger)
	if err != nil {
		t.Fatalf("Не удалось создать Redis кэш: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()
	testData := map[string]string{"key": "value", "test": "data"}

	err = cache.Set(ctx, "test-key", testData, 10*time.Minute)
	if err != nil {
		t.Errorf("Ожидалось успешное сохранение в кэш, получили ошибку: %v", err)
	}

	// Проверяем, что данные были сохранены в Redis
	storedData, err := mr.Get("test-key")
	if err != nil {
		t.Errorf("Ошибка получения данных из mock Redis: %v", err)
		return
	}
	if storedData == "" {
		t.Error("Данные не были сохранены в Redis")
	}

	// Проверяем, что данные корректно сериализованы
	var deserializedData map[string]string
	err = json.Unmarshal([]byte(storedData), &deserializedData)
	if err != nil {
		t.Errorf("Ошибка десериализации данных: %v", err)
	}

	if deserializedData["key"] != "value" || deserializedData["test"] != "data" {
		t.Errorf("Ожидались данные map[string]string{'key': 'value', 'test': 'data'}, получили %v", deserializedData)
	}
}

// Тестирование получения значения из кэша
func TestRedisCacheGet(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Не удалось запустить mock Redis: %v", err)
	}
	defer mr.Close()

	logger := slog.New(slog.Default().Handler())
	cache, err := NewRedisCache("redis://"+mr.Addr(), logger)
	if err != nil {
		t.Fatalf("Не удалось создать Redis кэш: %v", err)
	}
	defer cache.Close()

	// Сначала сохраняем данные
	ctx := context.Background()
	testData := struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}{
		ID:   1,
		Name: "test",
	}

	err = cache.Set(ctx, "get-test-key", testData, 10*time.Minute)
	if err != nil {
		t.Fatalf("Ошибка сохранения в кэш: %v", err)
	}

	// Теперь извлекаем данные
	var result struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	err = cache.Get(ctx, "get-test-key", &result)
	if err != nil {
		t.Errorf("Ожидалось успешное получение из кэша, получили ошибку: %v", err)
	}

	if result.ID != 1 || result.Name != "test" {
		t.Errorf("Ожидались данные {ID: 1, Name: 'test'}, получили %+v", result)
	}
}

// Тестирование получения несуществующего ключа
func TestRedisCacheGetNonExistentKey(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Не удалось запустить mock Redis: %v", err)
	}
	defer mr.Close()

	logger := slog.New(slog.Default().Handler())
	cache, err := NewRedisCache("redis://"+mr.Addr(), logger)
	if err != nil {
		t.Fatalf("Не удалось создать Redis кэш: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()
	var result string

	err = cache.Get(ctx, "non-existent-key", &result)
	if err == nil {
		t.Error("Ожидалась ошибка при попытке получить несуществующий ключ")
	}
}

// Тестирование удаления значения из кэша
func TestRedisCacheDelete(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Не удалось запустить mock Redis: %v", err)
	}
	defer mr.Close()

	logger := slog.New(slog.Default().Handler())
	cache, err := NewRedisCache("redis://"+mr.Addr(), logger)
	if err != nil {
		t.Fatalf("Не удалось создать Redis кэш: %v", err)
	}
	defer cache.Close()

	// Сохраняем данные
	ctx := context.Background()
	err = cache.Set(ctx, "delete-test-key", "test-value", 10*time.Minute)
	if err != nil {
		t.Fatalf("Ошибка сохранения в кэш: %v", err)
	}

	// Проверяем, что ключ существует
	exists, err := cache.Exists(ctx, "delete-test-key")
	if err != nil {
		t.Fatalf("Ошибка проверки существования ключа: %v", err)
	}
	if !exists {
		t.Error("Ключ должен существовать перед удалением")
	}

	// Удаляем ключ
	err = cache.Delete(ctx, "delete-test-key")
	if err != nil {
		t.Errorf("Ожидалось успешное удаление ключа, получили ошибку: %v", err)
	}

	// Проверяем, что ключ больше не существует
	exists, err = cache.Exists(ctx, "delete-test-key")
	if err != nil {
		t.Fatalf("Ошибка проверки существования ключа после удаления: %v", err)
	}
	if exists {
		t.Error("Ключ не должен существовать после удаления")
	}
}

// Тестирование проверки существования ключа
func TestRedisCacheExists(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Не удалось запустить mock Redis: %v", err)
	}
	defer mr.Close()

	logger := slog.New(slog.Default().Handler())
	cache, err := NewRedisCache("redis://"+mr.Addr(), logger)
	if err != nil {
		t.Fatalf("Не удалось создать Redis кэш: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	// Проверяем несуществующий ключ
	exists, err := cache.Exists(ctx, "non-existent-key")
	if err != nil {
		t.Errorf("Ожидалась успешная проверка несуществующего ключа, получили ошибку: %v", err)
	}
	if exists {
		t.Error("Несуществующий ключ должен возвращать false")
	}

	// Сохраняем данные
	err = cache.Set(ctx, "exists-test-key", "test-value", 10*time.Minute)
	if err != nil {
		t.Fatalf("Ошибка сохранения в кэш: %v", err)
	}

	// Проверяем существующий ключ
	exists, err = cache.Exists(ctx, "exists-test-key")
	if err != nil {
		t.Errorf("Ожидалась успешная проверка существующего ключа, получили ошибку: %v", err)
	}
	if !exists {
		t.Error("Существующий ключ должен возвращать true")
	}
}

// Тестирование истечения срока действия ключа
func TestRedisCacheExpiration(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Не удалось запустить mock Redis: %v", err)
	}
	defer mr.Close()

	logger := slog.New(slog.Default().Handler())
	cache, err := NewRedisCache("redis://"+mr.Addr(), logger)
	if err != nil {
		t.Fatalf("Не удалось создать Redis кэш: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	// Устанавливаем ключ с коротким временем жизни (100 мс)
	err = cache.Set(ctx, "expiring-key", "expiring-value", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Ошибка сохранения в кэш: %v", err)
	}

	// Для miniredis нужно вручную обработать истечение срока
	mr.FastForward(150 * time.Millisecond)

	var result string
	err = cache.Get(ctx, "expiring-key", &result)
	if err == nil {
		t.Error("Ожидалась ошибка при попытке получить истекший ключ")
	}
}

// Тестирование сериализации различных типов данных
func TestRedisCacheVariousDataTypes(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Не удалось запустить mock Redis: %v", err)
	}
	defer mr.Close()

	logger := slog.New(slog.Default().Handler())
	cache, err := NewRedisCache("redis://"+mr.Addr(), logger)
	if err != nil {
		t.Fatalf("Не удалось создать Redis кэш: %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	// Тестируем строку
	testString := "test string"
	err = cache.Set(ctx, "string-key", testString, 10*time.Minute)
	if err != nil {
		t.Errorf("Ошибка сохранения строки: %v", err)
	}

	var resultString string
	err = cache.Get(ctx, "string-key", &resultString)
	if err != nil {
		t.Errorf("Ошибка получения строки: %v", err)
	}
	if resultString != testString {
		t.Errorf("Ожидалась строка '%s', получили '%s'", testString, resultString)
	}

	// Тестируем число
	testInt := 42
	err = cache.Set(ctx, "int-key", testInt, 10*time.Minute)
	if err != nil {
		t.Errorf("Ошибка сохранения числа: %v", err)
	}

	var resultInt int
	err = cache.Get(ctx, "int-key", &resultInt)
	if err != nil {
		t.Errorf("Ошибка получения числа: %v", err)
	}
	if resultInt != testInt {
		t.Errorf("Ожидалось число %d, получили %d", testInt, resultInt)
	}

	// Тестируем сложную структуру
	type TestStruct struct {
		ID       int                    `json:"id"`
		Name     string                 `json:"name"`
		Metadata map[string]interface{} `json:"metadata"`
	}

	testStruct := TestStruct{
		ID:   123,
		Name: "test-struct",
		Metadata: map[string]interface{}{
			"version": 1,
			"active":  true,
		},
	}

	err = cache.Set(ctx, "struct-key", testStruct, 10*time.Minute)
	if err != nil {
		t.Errorf("Ошибка сохранения структуры: %v", err)
	}

	var resultStruct TestStruct
	err = cache.Get(ctx, "struct-key", &resultStruct)
	if err != nil {
		t.Errorf("Ошибка получения структуры: %v", err)
	}

	if resultStruct.ID != testStruct.ID || resultStruct.Name != testStruct.Name {
		t.Errorf("Структура не совпадает: ожидается %+v, получено %+v", testStruct, resultStruct)
	}
}

// Тестирование закрытия соединения
func TestRedisCacheClose(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Не удалось запустить mock Redis: %v", err)
	}
	defer mr.Close()

	logger := slog.New(slog.Default().Handler())
	cache, err := NewRedisCache("redis://"+mr.Addr(), logger)
	if err != nil {
		t.Fatalf("Не удалось создать Redis кэш: %v", err)
	}

	// Клиент должен быть в рабочем состоянии
	if cache.client.Ping(context.Background()).Err() != nil {
		t.Error("Клиент должен быть подключен до вызова Close")
	}

	err = cache.Close()
	if err != nil {
		t.Errorf("Ожидалось успешное закрытие, получили ошибку: %v", err)
	}

	// После закрытия клиент не должен быть доступен
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	
	err = cache.client.Ping(ctx).Err()
	if err == nil {
		t.Error("Клиент должен быть отключен после вызова Close")
	}
}

// Тестирование повторного закрытия (идемпотентность)
func TestRedisCacheCloseIdempotent(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Не удалось запустить mock Redis: %v", err)
	}
	defer mr.Close()

	logger := slog.New(slog.Default().Handler())
	cache, err := NewRedisCache("redis://"+mr.Addr(), logger)
	if err != nil {
		t.Fatalf("Не удалось создать Redis кэш: %v", err)
	}

	// Закрываем первый раз
	err = cache.Close()
	if err != nil {
		t.Errorf("Ожидалось успешное первое закрытие, получили ошибку: %v", err)
	}

	// Пытаемся закрыть второй раз
	err = cache.Close()
	// Не все клиенты Redis возвращают ошибку при повторном закрытии, 
	// поэтому мы просто проверяем, что программа не падает
}