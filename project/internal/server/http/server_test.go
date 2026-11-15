package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/TheTeaToper/OTUS/project/internal/app"
)

func TestHTTP_ValidateAuth(t *testing.T) {
	// Создаем сервис с жесткими лимитами: 1 запрос для логина/пароля/IP
	cfg := app.Config{LoginLimit: 1, PasswordLimit: 1, IpLimit: 1}
	service := app.NewAntibruteforceService(cfg, time.Minute*10)
	server := NewServer(service)

	// Подготавливаем тестовый запрос
	reqBody, _ := json.Marshal(AuthRequest{
		Login:    "admin",
		Password: "password123",
		IP:       "192.1.1.5",
	})

	// --- ПЕРВЫЙ ЗАПРОС (Должен быть OK=true) ---
	req1 := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/auth/validate", bytes.NewBuffer(reqBody))
	w1 := httptest.NewRecorder()
	server.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("Ожидался статус 200, получен %d", w1.Code)
	}

	var resp1 AuthResponse
	_ = json.NewDecoder(w1.Body).Decode(&resp1)
	if !resp1.OK {
		t.Error("Первая попытка авторизации должна возвращать ok=true")
	}

	// --- ВТОРОЙ ЗАПРОС ПОДРЯД (Должен быть OK=false, превышен лимит) ---
	req2 := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/auth/validate", bytes.NewBuffer(reqBody))
	w2 := httptest.NewRecorder()
	server.ServeHTTP(w2, req2)

	var resp2 AuthResponse
	_ = json.NewDecoder(w2.Body).Decode(&resp2)
	if resp2.OK {
		t.Error("Вторая попытка должна возвращать ok=false (bruteforce)")
	}

	// --- ТРЕТИЙ ЗАПРОС ПОДРЯД (Должен быть OK=false, превышен лимит 2) ---
	req3 := httptest.NewRequestWithContext(context.Background(), "POST", "/api/v1/auth/validate", bytes.NewBuffer(reqBody))
	w3 := httptest.NewRecorder()
	server.ServeHTTP(w3, req3)

	var resp3 AuthResponse
	_ = json.NewDecoder(w3.Body).Decode(&resp3)
	if resp3.OK {
		t.Error("Третья попытка должна возвращать ok=false (bruteforce)")
	}
}
