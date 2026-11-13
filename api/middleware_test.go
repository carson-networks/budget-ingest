package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/carson-networks/budget-ingest/internal/logging"
)

func newTestLogger() (*bytes.Buffer, *Server) {
	buf := &bytes.Buffer{}
	logger := logging.SetupLogging()
	logger.Out = buf
	return buf, &Server{Logger: logger, Port: "0"}
}

func logLines(t *testing.T, buf *bytes.Buffer) []map[string]interface{} {
	var lines []map[string]interface{}
	for _, raw := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		line := map[string]interface{}{}
		require.NoError(t, json.Unmarshal([]byte(raw), &line))
		lines = append(lines, line)
	}
	return lines
}

func TestLoggingMiddleware(t *testing.T) {
	buf, srv := newTestLogger()
	handler := LoggingMiddleware(srv.Logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotNil(t, logging.GetLogData(r.Context()))
		w.WriteHeader(http.StatusTeapot)
	}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/brew", nil))

	lines := logLines(t, buf)
	require.Len(t, lines, 2)

	start := lines[0]
	assert.Equal(t, "Handler.Start", start["msg"])
	assert.Equal(t, "info", start["loglevel"])
	assert.Equal(t, "POST", start["method"])
	assert.Equal(t, "/brew", start["path"])

	complete := lines[1]
	assert.Equal(t, "Handler.Complete", complete["msg"])
	assert.Equal(t, "info", complete["loglevel"])
	assert.Equal(t, "POST", complete["method"])
	assert.Equal(t, "/brew", complete["path"])
	assert.Equal(t, float64(http.StatusTeapot), complete["status"])
	assert.Contains(t, complete, "durationMs")
}

func TestLoggingMiddlewareDefaultStatus(t *testing.T) {
	buf, srv := newTestLogger()
	handler := LoggingMiddleware(srv.Logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("body"))
	}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	lines := logLines(t, buf)
	require.Len(t, lines, 2)
	assert.Equal(t, float64(http.StatusOK), lines[1]["status"])
}
