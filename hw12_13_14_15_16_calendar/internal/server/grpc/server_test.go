package internalgrpc

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/api"
	"github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/app"
	"github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type mockStorage struct {
	events []storage.Event
}

func (m *mockStorage) CreateEvent(e *storage.Event) error {
	e.ID = int64(len(m.events)) + 1
	m.events = append(m.events, *e)
	return nil
}
func (m *mockStorage) UpdateEvent(e *storage.Event) error { return nil }
func (m *mockStorage) DeleteEvent(id int64) error         { return nil }
func (m *mockStorage) ListEvents() ([]storage.Event, error) {
	return m.events, nil
}

func TestGRPC_CreateAndListEvents(t *testing.T) {
	mockLogger := logger.New("DEBUG")
	// Размер буфера памяти
	buffer := 1024 * 1024
	listener := bufconn.Listen(buffer)

	mockStorage := &mockStorage{}
	calendar := app.New(mockLogger, mockStorage)

	// gRPC-сервер на буфере памяти
	grpcServer := grpc.NewServer()
	server := NewServer(calendar, mockLogger)
	api.RegisterEventsServiceServer(grpcServer, server)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()
	defer grpcServer.GracefulStop()

	// gRPC-клиент для подключения к буферу
	ctx := context.Background()
	dialer := func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}

	conn, err := grpc.DialContext(
		ctx, "bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := api.NewEventsServiceClient(conn)

	// Тест 1: CreateEvent
	now := time.Now().UTC()
	createResponse, err := client.CreateEvent(ctx, &api.CreateEventRequest{
		Title:       "gRPC Тест",
		StartTime:   timestamppb.New(now),
		EndTime:     timestamppb.New(now.Add(time.Hour)),
		Description: "Проверка работы gRPC",
		UserId:      "grpc_user",
	})
	if err != nil {
		t.Fatalf("CreateEvent failed: %v", err)
	}

	if createResponse.Event.Title != "gRPC Тест" {
		t.Errorf("Expected title 'gRPC Тест', got '%s'", createResponse.Event.Title)
	}

	if createResponse.Event.Id != "1" {
		t.Errorf("Expected event ID '1', got '%s'", createResponse.Event.Id)
	}

	// Тест 2: EventsForDay
	listResp, err := client.EventsForDay(ctx, &api.EventsForDayRequest{
		Day: timestamppb.New(now),
	})
	if err != nil {
		t.Fatalf("EventsForDay failed: %v", err)
	}

	if len(listResp.Events) != 1 {
		t.Errorf("Expected 1 event in gRPC list, got %d", len(listResp.Events))
	}
}
