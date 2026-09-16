package internalgrpc

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/api"
	"github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/app"
	"github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/TheTeaToper/OTUS/hw12_13_14_15_16_calendar/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	api.UnimplementedEventsServiceServer
	calendarApp *app.App
	logger      *logger.Logger
	grpcServer  *grpc.Server
}

func NewServer(calendarApp *app.App, logger *logger.Logger) *Server {
	return &Server{
		calendarApp: calendarApp,
		logger:      logger,
	}
}

func (s *Server) loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	s.logger.Debug("gRPC %s | ВРЕМЯ: %v | ОШИБКА: %v", info.FullMethod, time.Since(start), err)
	return resp, err
}

func (s *Server) Start(ctx context.Context, addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.grpcServer = grpc.NewServer(grpc.UnaryInterceptor(s.loggingInterceptor))
	api.RegisterEventsServiceServer(s.grpcServer, s)
	s.logger.Debug("gRPC сервер запущен на %s", addr)

	errorChan := make(chan error, 1)
	go func() {
		if err := s.grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			errorChan <- err
		}
	}()

	// Ожидаем либо ошибку старта, либо сигнал отмены контекста для остановки
	select {
	case err := <-errorChan:
		return err
	case <-ctx.Done():
		s.logger.Info("Остановка gRPC сервера...")
		s.grpcServer.GracefulStop()
		return nil
	}
}

func (s *Server) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
}

func (s *Server) CreateEvent(ctx context.Context, req *api.CreateEventRequest) (*api.CreateEventResponse, error) {
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title не может быть пустым")
	}

	dbEvent := &storage.Event{
		Title:       req.Title,
		StartTime:   req.StartTime.AsTime(),
		EndTime:     req.EndTime.AsTime(),
		Description: req.Description,
		User:        req.UserId,
	}

	if err := s.calendarApp.CreateEvent(ctx, dbEvent); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.CreateEventResponse{Event: s.convertToProto(*dbEvent)}, nil
}

func (s *Server) UpdateEvent(ctx context.Context, req *api.UpdateEventRequest) (*api.UpdateEventResponse, error) {
	if req.Id == "" || req.Event == nil {
		return nil, status.Error(codes.InvalidArgument, "id и объект event обязательны")
	}

	id, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "невалидный формат id")
	}

	dbEvent := &storage.Event{
		ID:          id,
		Title:       req.Event.Title,
		StartTime:   req.Event.StartTime.AsTime(),
		EndTime:     req.Event.EndTime.AsTime(),
		Description: req.Event.Description,
		User:        req.Event.UserId,
	}

	if err := s.calendarApp.UpdateEvent(ctx, dbEvent); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.UpdateEventResponse{Event: s.convertToProto(*dbEvent)}, nil
}

func (s *Server) DeleteEvent(ctx context.Context, req *api.DeleteEventRequest) (*api.DeleteEventResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id обязателен")
	}

	if err := s.calendarApp.DeleteEvent(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.DeleteEventResponse{Success: true}, nil
}

func (s *Server) EventsForDay(ctx context.Context, req *api.EventsForDayRequest) (*api.EventsResponse, error) {
	if req.Day == nil {
		return nil, status.Error(codes.InvalidArgument, "day обязателен")
	}

	events, err := s.calendarApp.ListEventsForDay(ctx, req.Day.AsTime())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.EventsResponse{Events: s.convertToProtoSlice(events)}, nil
}

func (s *Server) EventsForWeek(ctx context.Context, req *api.EventsForWeekRequest) (*api.EventsResponse, error) {
	if req.WeekFirstDay == nil {
		return nil, status.Error(codes.InvalidArgument, "WeekFirstDay обязателен")
	}

	events, err := s.calendarApp.ListEventsForWeek(ctx, req.WeekFirstDay.AsTime())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.EventsResponse{Events: s.convertToProtoSlice(events)}, nil
}

func (s *Server) EventsForMonth(ctx context.Context, req *api.EventsForMonthRequest) (*api.EventsResponse, error) {
	if req.MonthFirstDay == nil {
		return nil, status.Error(codes.InvalidArgument, "MonthFirstDay обязателен")
	}

	events, err := s.calendarApp.ListEventsForMonth(ctx, req.MonthFirstDay.AsTime())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.EventsResponse{Events: s.convertToProtoSlice(events)}, nil
}

func (s *Server) convertToProto(e storage.Event) *api.Event {
	return &api.Event{
		Id:          strconv.FormatInt(e.ID, 10),
		Title:       e.Title,
		StartTime:   timestamppb.New(e.StartTime),
		EndTime:     timestamppb.New(e.EndTime),
		Description: e.Description,
		UserId:      e.User,
	}
}

func (s *Server) convertToProtoSlice(events []storage.Event) []*api.Event {
	protoEvents := make([]*api.Event, len(events))
	for i, e := range events {
		protoEvents[i] = s.convertToProto(e)
	}
	return protoEvents
}
