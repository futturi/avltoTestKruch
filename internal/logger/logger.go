package logger

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ctxLogger struct{}

func ContextWithLogger(ctx context.Context, l *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, ctxLogger{}, l)
}

func LoggerFromContext(ctx context.Context) *zap.SugaredLogger {
	if l, ok := ctx.Value(ctxLogger{}).(*zap.SugaredLogger); ok {
		return l
	}
	return zap.S()
}

func LoggerMiddleware(log *zap.SugaredLogger) gin.HandlerFunc {
	defer func() {
		err := log.Sync()
		if err != nil {
			log.Debugw("Failed to sync logger", zap.Error(err))
		}
	}()
	return func(c *gin.Context) {
		c.Request = c.Request.Clone(context.WithValue(c.Request.Context(), ctxLogger{}, log))
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		log.Infow("[GIN]",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.RequestURI),
			zap.Int("statusCode", c.Writer.Status()),
			zap.String("latency", latency.String()),
		)
	}
}

func InitLogger() *zap.SugaredLogger {
	cfg := zap.NewProductionConfig()

	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.OutputPaths = []string{"stdout"}
	logger, _ := cfg.Build()
	return logger.Sugar()
}
