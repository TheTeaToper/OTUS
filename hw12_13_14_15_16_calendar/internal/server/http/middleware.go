package internalhttp

import (
	"fmt"
	"net"
	"net/http"
	"time"
)

// Реализация кастомного ResponseWriter для перехвата статус-кода и размера ответа
type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (rw *responseWriterInterceptor) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriterInterceptor) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Оборачиваем оригинальный ResponseWriter, чтобы узнать код ответа
		interceptor := &responseWriterInterceptor{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // дефолтный статус, если WriteHeader не вызвана явно
		}

		next.ServeHTTP(interceptor, r)

		latency := time.Since(start)

		// Получаем чистый IP клиента
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		// Форматируем время в точности как в примере: [25/Feb/2020:19:11:24 +0600]
		timeStr := start.Format("02/Jan/2006:15:04:05 -0700")

		// Формируем строчку лога в соответствии с требованиями ТЗ
		logLine := fmt.Sprintf(
			"%s [%s] %s %s %s %d %d \"%s\" (Latency: %v)",
			ip,
			timeStr,
			r.Method,
			r.URL.RequestURI(),
			r.Proto,
			interceptor.statusCode,
			interceptor.bytesWritten,
			r.UserAgent(),
			latency,
		)

		s.logger.Info(logLine)
	})
}
