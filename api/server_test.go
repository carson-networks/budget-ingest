package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthz(t *testing.T) {
	buf, srv := newTestLogger()
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ok", rec.Body.String())

	lines := logLines(t, buf)
	require.Len(t, lines, 2)
	assert.Equal(t, "Handler.Start", lines[0]["msg"])
	assert.Equal(t, "Handler.Complete", lines[1]["msg"])
	assert.Equal(t, "/healthz", lines[1]["path"])
	assert.Equal(t, float64(http.StatusOK), lines[1]["status"])
}

func TestHealthzRejectsOtherMethods(t *testing.T) {
	_, srv := newTestLogger()
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/healthz", nil))

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestUnknownRouteIsNotFound(t *testing.T) {
	_, srv := newTestLogger()
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/nope", nil))

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
