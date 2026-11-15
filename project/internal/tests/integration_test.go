package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

const baseURL = "http://localhost:8080/api/v1"

// Структуры DTO для отправки запросов
type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	IP       string `json:"ip"`
}

type AuthResponse struct {
	OK bool `json:"ok"`
}

type CidrRequest struct {
	Subnet string `json:"subnet"`
}

type ResetRequest struct {
	Login string `json:"login"`
	IP    string `json:"ip"`
}

func TestIntegration_AntiBruteForce(t *testing.T) {
	// Перед запуском тестов проверяем, доступен ли сервер
	client := &http.Client{Timeout: 2 * time.Second}
	_, err := client.Get(baseURL + "/auth/validate")
	if err != nil {
		t.Skip("Пропуск интеграционных тестов: сервер Анти-брутфорс не запущен на порту 8080. Запустите 'make run' перед тестированием.")
	}

	// Данные для тестов
	testLogin := "integration_user"
	testPassword := "secure_pass"
	testIP := "192.168.1.10"
	testSubnet := "192.168.1.0/24"

	// Убедимся, что бакеты чистые перед началом
	resetBuckets(t, client, testLogin, testIP)

	// ==========================================
	// 1. ТЕСТ: Базовая проверка лимитов API
	// ==========================================
	t.Run("Test Rate Limits via API", func(t *testing.T) {
		// По дефолту лимит логина N=10 запросов в минуту.
		// Сделаем 10 успешных запросов подряд.
		for i := 1; i <= 10; i++ {
			resp := validateAuth(t, client, testLogin, testPassword, testIP)
			if !resp.OK {
				t.Fatalf("Запрос %d должен был вернуть ok=true, лимит еще не исчерпан", i)
			}
		}

		// 11-й запрос должен заблокироваться
		respFails := validateAuth(t, client, testLogin, testPassword, testIP)
		if respFails.OK {
			t.Error("11-й запрос подряд должен был вернуть ok=false (превышен лимит)")
		}
	})

	// ==========================================
	// 2. ТЕСТ: Работа со списками (Blacklist)
	// ==========================================
	t.Run("Test Blacklist API", func(t *testing.T) {
		// Добавляем подсеть в черный список
		modifyList(t, client, "POST", "/blacklist", testSubnet)

		// Другой пользователь с IP из этой подсети должен сразу получить блок (ok=false)
		resp := validateAuth(t, client, "other_user", "some_pass", "192.168.1.55")
		if resp.OK {
			t.Error("Пользователь из Blacklist подсети должен быть заблокирован")
		}

		// Удаляем подсеть из черного списка
		modifyList(t, client, "DELETE", "/blacklist", testSubnet)

		// Теперь запрос оттуда должен пройти (бакет для нового юзера чистый)
		respAfterDelete := validateAuth(t, client, "other_user", "some_pass", "192.168.1.55")
		if !respAfterDelete.OK {
			t.Error("После удаления подсети из Blacklist запрос должен пройти успешно")
		}
	})

	// ==========================================
	// 3. ТЕСТ: Работа со списками (Whitelist)
	// ==========================================
	t.Run("Test Whitelist API", func(t *testing.T) {
		// Наш основной testLogin сейчас заблокирован лимитерами из первого теста.
		// Добавим его подсеть в Whitelist.
		modifyList(t, client, "POST", "/whitelist", testSubnet)

		// Запрос должен безусловно пройти (ok=true), несмотря на переполненные бакеты
		resp := validateAuth(t, client, testLogin, testPassword, testIP)
		if !resp.OK {
			t.Error("Запрос с IP из Whitelist должен успешно проходить, даже если лимиты бакетов исчерпаны")
		}

		// Чистим за собой вайтлист
		modifyList(t, client, "DELETE", "/whitelist", testSubnet)
	})

	// ==========================================
	// 4. ТЕСТ: Сброс бакетов (Reset API)
	// ==========================================
	t.Run("Test Reset Buckets API", func(t *testing.T) {
		// Наш IP убран из вайтлиста, бакеты все еще переполнены -> запрос заблокирован
		respBlocked := validateAuth(t, client, testLogin, testPassword, testIP)
		if respBlocked.OK {
			t.Error("Запрос должен быть заблокирован перед сбросом")
		}

		// Вызываем метод сброса бакетов через API
		resetBuckets(t, client, testLogin, testIP)

		// Теперь запрос снова должен стать успешным
		respUnblocked := validateAuth(t, client, testLogin, testPassword, testIP)
		if !respUnblocked.OK {
			t.Error("После вызова Reset API бакеты должны очиститься, а запрос — пройти")
		}
	})
}

// --- Хелперы для выполнения HTTP запросов к API ---

func validateAuth(t *testing.T, client *http.Client, login, password, ip string) AuthResponse {
	body, _ := json.Marshal(AuthRequest{Login: login, Password: password, IP: ip})
	req, _ := http.NewRequest("POST", baseURL+"/auth/validate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Ошибка отправки запроса на валидацию: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Сервер вернул статус %d при валидации", resp.StatusCode)
	}

	var result AuthResponse
	_ = json.NewDecoder(resp.Body).Decode(&result)
	return result
}

func modifyList(t *testing.T, client *http.Client, method, endpoint, subnet string) {
	body, _ := json.Marshal(CidrRequest{Subnet: subnet})
	req, _ := http.NewRequest(method, baseURL+endpoint, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Ошибка изменения списка %s: %v", endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Сервер вернул статус %d при изменении списка %s", resp.StatusCode, endpoint)
	}
}

func resetBuckets(t *testing.T, client *http.Client, login, ip string) {
	body, _ := json.Marshal(ResetRequest{Login: login, IP: ip})
	req, _ := http.NewRequest("POST", baseURL+"/buckets/reset", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Ошибка сброса бакетов: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Сервер вернул статус %d при сбросе бакетов", resp.StatusCode)
	}
}
