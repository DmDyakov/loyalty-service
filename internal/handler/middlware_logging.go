package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// ZapLogFormatter реализует интерфейс middleware.LogFormatter
type ZapLogFormatter struct {
	logger *zap.Logger
}

func (m *Middleware) withLogging(next http.Handler) http.Handler {
	return middleware.RequestLogger(&ZapLogFormatter{
		logger: m.logger,
	})(next)
}

func (f *ZapLogFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	return &ZapLogEntry{
		logger: f.logger,
		req:    r,
	}
}

type ZapLogEntry struct {
	logger *zap.Logger
	req    *http.Request
}

func (e *ZapLogEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	e.logger.Info("request completed",
		zap.String("method", e.req.Method),
		zap.String("path", e.req.URL.Path),
		zap.Int("status", status),
		zap.Int("bytes", bytes),
		zap.Duration("duration", elapsed),
		zap.String("request_id", middleware.GetReqID(e.req.Context())),
		zap.String("remote_addr", e.req.RemoteAddr),
		zap.String("user_agent", e.req.UserAgent()),
	)
}

func (e *ZapLogEntry) Panic(v interface{}, stack []byte) {
	e.logger.Error("request panicked",
		zap.Any("panic", v),
		zap.String("stack", string(stack)),
	)
}
