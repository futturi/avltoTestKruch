package logger

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestContextWithLogger(t *testing.T) {
	ctx := context.Background()
	testLogger := zap.NewExample().Sugar()

	ctxWithLogger := ContextWithLogger(ctx, testLogger)
	retrievedLogger := LoggerFromContext(ctxWithLogger)

	if !reflect.DeepEqual(testLogger, retrievedLogger) {
		t.Errorf("expected logger %v, got %v", testLogger, retrievedLogger)
	}
}

func TestLoggerFromContextWithoutLogger(t *testing.T) {
	ctx := context.Background()
	loggerFromCtx := LoggerFromContext(ctx)
	if loggerFromCtx == nil {
		t.Error("expected non-nil logger from context")
	}
}

func TestLoggerMiddleware(t *testing.T) {
	obsCore, logs := observer.New(zapcore.InfoLevel)
	testLogger := zap.New(obsCore).Sugar()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test-path", nil)
	mw := LoggerMiddleware(testLogger)
	c.Set("dummy", "value")
	mw(c)
	entries := logs.All()
	assert.Greater(t, len(entries), 0, "expected at least one log entry")
	entry := entries[0]
	fields := entry.Context

	var hasMethod, hasPath, hasStatus, hasLatency bool
	for _, f := range fields {
		switch f.Key {
		case "method":
			if f.String == "GET" {
				hasMethod = true
			}
		case "path":
			if f.String == "/test-path" {
				hasPath = true
			}
		case "statusCode":
			if f.Integer == 200 {
				hasStatus = true
			}
		case "latency":
			if f.String != "" {
				hasLatency = true
			}
		}
	}
	assert.True(t, hasMethod, "expected log field 'method' with value GET")
	assert.True(t, hasPath, "expected log field 'path' with value /test-path")
	assert.True(t, hasStatus, "expected log field 'statusCode' with value 200")
	assert.True(t, hasLatency, "expected log field 'latency' with non-empty value")
}

func TestInitLogger(t *testing.T) {
	l := InitLogger()
	assert.NotNil(t, l, "expected non-nil logger from InitLogger")
	assert.NotPanics(t, func() {
		l.Infow("test message", "key", "value")
	})
}
