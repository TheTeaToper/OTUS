package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

type EventResponse struct {
	Events []storage.Event `json:"events"`
}

func TestCalendar_IntegrationFlow(t *testing.T) {
	// Настройка хостов для работы как внутри докера, так и локально
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	apiHost := os.Getenv("API_HOST")
	if apiHost == "" {
		apiHost = "localhost"
	}

	// Подключение к БД для очистки и проверок
	db, err := sqlx.Open("pgx", fmt.Sprintf("postgres://calendar_user:calendar_password@%s:5432/calendar?sslmode=disable", dbHost))
	require.NoError(t, err)
	defer db.Close()

	// Очищаем таблицы перед тестами
	_, _ = db.Exec("TRUNCATE TABLE events CASCADE;")
	_, _ = db.Exec("TRUNCATE TABLE notifications_log CASCADE;")

	// СЦЕНАРИЙ 1 & 2: Создание события и получение списка
	t.Run("Create an event and verify list for day", func(t *testing.T) {
		tomorrowDate := time.Now().Add(24 * time.Hour)

		reqBody := storage.Event{
			Title:       "Integration Test Meeting",
			StartTime:   tomorrowDate,
			EndTime:     tomorrowDate.Add(1 * time.Hour),
			Description: "Testing complete flow",
			User:        "integration_user_1",
		}

		body, _ := json.Marshal(reqBody)
		resp, err := http.Post(fmt.Sprintf("http://%s:8888/create_event", apiHost), "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		// Проверяем списое событий на завтра
		dateStr := tomorrowDate.Format("2006-01-02")
		respList, err := http.Get(fmt.Sprintf("http://%s:8888/events_for_day?day=%s", apiHost, dateStr))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, respList.StatusCode)

		var listResp EventResponse
		err = json.NewDecoder(respList.Body).Decode(&listResp)
		require.NoError(t, err)

		// Проверяем, что событие вернулось в листинге
		require.NotEmpty(t, listResp.Events)
		require.Equal(t, "Integration Test Meeting", listResp.Events[0].Title)
		require.Equal(t, "integration_user_1", listResp.Events[0].User)
	})

	// СЦЕНАРИЙ 3: Обработка бизнес-ошибок
	t.Run("'Create Event' business error validation", func(t *testing.T) {
		// время окончания меньше времени начала
		invalidEvent := storage.Event{
			Title:     "Bad Event",
			StartTime: time.Now().Add(5 * time.Hour),
			EndTime:   time.Now().Add(2 * time.Hour),
			User:      "integration_user_1",
		}

		body, _ := json.Marshal(invalidEvent)
		resp, err := http.Post(fmt.Sprintf("http://%s:8888/create_event", apiHost), "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)

		// сервер отдаст 500 статус StatusInternalServerError (согласно handleCreateEvent)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	// Проверка отправки уведомления sender`ом
	t.Run("Notification logging test", func(t *testing.T) {
		// событие, которое наступит через 2 секунды, чтобы scheduler сразу его подхватил
		triggerTime := time.Now().Add(2 * time.Second)

		alertEvent := storage.Event{
			Title:       "Urgent Alert Test",
			StartTime:   triggerTime,
			EndTime:     triggerTime.Add(30 * time.Minute),
			Description: "Testing RMQ and Sender logging",
			User:        "alert_user_999",
		}

		body, _ := json.Marshal(alertEvent)
		resp, err := http.Post(fmt.Sprintf("http://%s:8888/create_event", apiHost), "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		// scheduler (интервал 20с) считает из БД -> бросит в RabbitMQ -> sender залогирует в БД
		t.Log("Waiting for scheduler and sender processing...")
		time.Sleep(20 * time.Second)

		// Проверяем таблицу логов уведомлений, которую заполнил sender
		var status string
		err = db.Get(&status,
			"SELECT status FROM notifications_log WHERE user_id = $1 AND event_title = $2",
			"alert_user_999", "Urgent Alert Test")

		require.NoError(t, err, "Sender did not log the notification status in time")
		require.Equal(t, "sent", status)
	})
}
